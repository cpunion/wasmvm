// Package decode reads the pinned Component Model binary grammar. It owns a
// snapshot of input and bounds aggregate nodes, names and lexical nesting.
// Successful decoding is not semantic validation or component execution.
package decode

import (
	"errors"

	"github.com/cpunion/wasmvm/internal/binary"
	"github.com/cpunion/wasmvm/internal/component/model"
	coredecode "github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Limits struct {
	MaxBytes, MaxNodes       uint64
	MaxNameBytes, MaxNesting uint32
}

func DefaultLimits() Limits { return Limits{64 << 20, 1 << 20, 1 << 20, 128} }

type decoder struct {
	r      binary.Reader
	limits Limits
	nodes  uint64
	depth  uint32
	core   *module.Module
}

func Decode(input []byte, limits Limits) (result *model.Binary, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			switch e := failure.(type) {
			case *wasm.Error:
				result, err = nil, e
			case *binary.Error:
				kind := wasm.ErrMalformed
				if errors.Is(e, binary.ErrLimit) {
					kind = wasm.ErrLimit
				}
				result, err = nil, &wasm.Error{Kind: kind, Offset: e.Offset, Detail: e.Reason.Error()}
			default:
				panic(failure)
			}
		}
	}()
	if uint64(len(input)) > limits.MaxBytes {
		return nil, &wasm.Error{Kind: wasm.ErrLimit, Detail: "component bytes"}
	}
	owned := append([]byte(nil), input...)
	d := decoder{r: binary.New(owned), limits: limits}
	return d.component(), nil
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
func (d *decoder) fail(detail string) {
	panic(&wasm.Error{Kind: wasm.ErrMalformed, Offset: d.r.Offset(), Detail: detail})
}
func (d *decoder) charge(n uint64) {
	if n > d.limits.MaxNodes-d.nodes {
		panic(&wasm.Error{Kind: wasm.ErrLimit, Offset: d.r.Offset(), Detail: "component nodes"})
	}
	d.nodes += n
}
func (d *decoder) enter() {
	if d.depth >= d.limits.MaxNesting {
		panic(&wasm.Error{Kind: wasm.ErrLimit, Offset: d.r.Offset(), Detail: "component nesting"})
	}
	d.depth++
}
func (d *decoder) byte() byte            { return must(d.r.Byte()) }
func (d *decoder) u32() uint32           { return must(d.r.U32()) }
func (d *decoder) bytes(n uint64) []byte { return must(d.r.Bytes(n)) }
func (d *decoder) name() string          { return must(d.r.Name(d.limits.MaxNameBytes)) }
func (d *decoder) expect(want byte) {
	if d.byte() != want {
		d.fail("invalid reserved byte")
	}
}
func (d *decoder) count() uint32 {
	n := d.u32()
	if uint64(n) > uint64(d.r.Remaining()) {
		d.fail("vector exceeds remaining component input")
	}
	d.charge(uint64(n))
	return n
}
func (d *decoder) optional() bool {
	b := d.byte()
	if b > 1 {
		d.fail("invalid option discriminant")
	}
	return b == 1
}

func (d *decoder) component() *model.Binary {
	d.enter()
	defer func() { d.depth-- }()
	if string(d.bytes(8)) != "\x00asm\x0d\x00\x01\x00" {
		d.fail("invalid component preamble")
	}
	result := &model.Binary{CoreTypes: new(module.Module)}
	oldCore := d.core
	d.core = result.CoreTypes
	defer func() { d.core = oldCore }()
	for d.r.Remaining() != 0 {
		at := d.r.Offset()
		id := d.byte()
		size := d.u32()
		sub := must(d.r.Sub(uint64(size)))
		parent := d.r
		d.r = sub
		switch id {
		case 0:
			d.charge(1)
			d.name()
			d.bytes(uint64(d.r.Remaining()))
		case 1, 4:
			d.charge(1)
			def := model.Definition{Kind: id, Offset: at}
			if id == 1 {
				reader := d.r
				def.Bytes = d.bytes(uint64(d.r.Remaining()))
				d.r = reader
				m, n, err := d.fragment().Module()
				if err != nil {
					panic(err)
				}
				d.charge(n)
				def.CoreModule = m
			} else {
				def.Component = d.component()
			}
			result.Definitions = append(result.Definitions, def)
		case 9:
			d.charge(1)
			s := &model.Start{Function: d.u32()}
			s.Arguments = make([]uint32, d.count())
			for i := range s.Arguments {
				s.Arguments[i] = d.u32()
			}
			s.Results = d.u32()
			result.Definitions = append(result.Definitions, model.Definition{Kind: id, Offset: at, Start: s})
		case 2, 3, 5, 6, 7, 8, 10, 11, 12:
			n := d.count()
			for range n {
				def := model.Definition{Kind: id, Offset: d.r.Offset()}
				switch id {
				case 2, 5:
					def.Instance = d.instance(id == 2)
				case 3:
					def.CoreType = d.coreType()
				case 6:
					def.Alias = d.alias()
				case 7:
					def.Type = d.typeDef()
				case 8:
					def.Canonical = d.canonical()
				case 10:
					def.Import = d.importType()
				case 11:
					e := &model.Export{Name: d.externName(), Item: d.sortIndex()}
					if d.optional() {
						t := d.externType()
						e.Type = &t
					}
					def.Export = e
				case 12:
					v := &model.ValueDefinition{Type: d.valueType()}
					v.Bytes = d.bytes(uint64(d.u32()))
					def.Value = v
				}
				result.Definitions = append(result.Definitions, def)
			}
		default:
			d.fail("unknown component section")
		}
		if err := d.r.Finish(); err != nil {
			panic(err)
		}
		d.r = parent
	}
	return result
}

func (d *decoder) fragment() coredecode.Fragment {
	return coredecode.Fragment{Reader: &d.r, Types: d.core, Features: wasm.SupportedFeatures, Limits: coredecode.Limits{MaxBytes: d.limits.MaxBytes, MaxNodes: d.limits.MaxNodes - d.nodes, MaxNameBytes: d.limits.MaxNameBytes, MaxNesting: d.limits.MaxNesting, MaxLocals: 1 << 20}}
}

func (d *decoder) coreValue() wasm.ValueType {
	v, n, err := d.fragment().ValueType()
	if err != nil {
		panic(err)
	}
	d.charge(n)
	return v
}
func (d *decoder) coreExtern() module.Import {
	v, n, err := d.fragment().ExternType()
	if err != nil {
		panic(err)
	}
	d.charge(n)
	return v
}
func (d *decoder) coreGroup() (uint32, uint32) {
	start := uint32(len(d.core.Types))
	n, err := d.fragment().TypeGroup()
	if err != nil {
		panic(err)
	}
	d.charge(n)
	return start, uint32(len(d.core.Types)) - start
}
