// Package decode reads the complete Core 2.0 binary grammar. It checks encoding
// and resource bounds, not type correctness or runtime import compatibility.
package decode

import (
	"errors"
	"math"

	"github.com/cpunion/wasmvm/internal/binary"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Limits bounds input ownership and decoded structure. MaxNodes counts vector
// entries, instructions, and control frames in aggregate, not allocated bytes.
// Expanded locals are bounded separately without expanding local groups.
// Zero fields are literal zero limits; use DefaultLimits for normal decoding.
type Limits struct {
	MaxBytes     uint64
	MaxNodes     uint64
	MaxNameBytes uint32
	MaxNesting   uint32
	MaxLocals    uint32
}

func DefaultLimits() Limits { return Limits{32 << 20, 2 << 20, 1 << 20, 10000, 1 << 20} }

type decoder struct {
	r              binary.Reader
	limits         Limits
	features       wasm.Features
	nodes          uint64
	needsDataCount bool
	inCode         bool
}

// Decode takes a snapshot of input. It returns no partial module on failure.
// Returned data/custom slices never alias the caller's bytes. A successful
// decode still requires complete semantic validation before use.
func Decode(input []byte, limits Limits) (*module.Module, error) {
	return DecodeWithFeatures(input, limits, 0)
}

// DecodeWithFeatures preserves the base grammar unless extensions are explicit.
func DecodeWithFeatures(input []byte, limits Limits, features wasm.Features) (result *module.Module, err error) {
	if features & ^wasm.SupportedFeatures != 0 {
		return nil, wasm.ErrUnsupported
	}
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
		return nil, &wasm.Error{Kind: wasm.ErrLimit, Detail: "module bytes"}
	}
	owned := append([]byte(nil), input...)
	d := decoder{r: binary.New(owned), limits: limits, features: features}
	header := d.bytes(8)
	if string(header[:4]) != "\x00asm" {
		d.failAt(0, "bad magic")
	}
	if string(header[4:]) != "\x01\x00\x00\x00" {
		d.failAt(4, "unknown binary version")
	}
	m := new(module.Module)
	last, codeCount := byte(0), uint32(0)
	for d.r.Remaining() != 0 {
		at := d.r.Offset()
		id := d.byte()
		if id > 12 {
			d.failAt(at, "unknown section")
		}
		rank := id
		if id == 12 {
			rank = 10
		} else if id >= 10 {
			rank++
		}
		if id != 0 {
			if rank <= last {
				d.failAt(at, "duplicate or out-of-order section")
			}
			last = rank
		}
		size := d.u32()
		sub := must(d.r.Sub(uint64(size)))
		parent := d.r
		d.r = sub
		switch id {
		case 0:
			d.charge(1)
			name := d.name()
			m.Custom = append(m.Custom, module.Custom{Name: name, Data: d.bytes(uint64(d.r.Remaining()))})
		case 1:
			m.Types = make([]module.FuncType, d.count())
			for i := range m.Types {
				d.expect(0x60)
				m.Types[i] = module.FuncType{Params: d.types(), Results: d.types()}
			}
		case 2:
			m.Imports = make([]module.Import, d.count())
			for i := range m.Imports {
				im := &m.Imports[i]
				im.Module, im.Name, im.Kind = d.name(), d.name(), d.byte()
				switch im.Kind {
				case 0:
					im.Type = d.u32()
				case 1:
					im.Table = d.table()
				case 2:
					im.Memory = d.bounds()
				case 3:
					im.Global = d.globalType()
				default:
					d.fail("unknown import kind")
				}
			}
		case 3:
			m.Functions = make([]module.Function, d.count())
			for i := range m.Functions {
				m.Functions[i].Type = d.u32()
			}
		case 4:
			m.Tables = make([]module.TableType, d.count())
			for i := range m.Tables {
				m.Tables[i] = d.table()
			}
		case 5:
			m.Memories = make([]module.Limits, d.count())
			for i := range m.Memories {
				m.Memories[i] = d.bounds()
			}
		case 6:
			m.Globals = make([]module.Global, d.count())
			for i := range m.Globals {
				m.Globals[i] = module.Global{Type: d.globalType(), Init: d.expr()}
			}
		case 7:
			m.Exports = make([]module.Export, d.count())
			for i := range m.Exports {
				e := &m.Exports[i]
				e.Name, e.Kind = d.name(), d.byte()
				if e.Kind > 3 {
					d.fail("unknown export kind")
				}
				e.Index = d.u32()
			}
		case 8:
			m.HasStart, m.Start = true, d.u32()
		case 9:
			m.Elements = make([]module.Element, d.count())
			for i := range m.Elements {
				m.Elements[i] = d.element()
			}
		case 10:
			codeCount = d.count()
			if uint64(codeCount) != uint64(len(m.Functions)) {
				d.fail("function/code count mismatch")
			}
			for i := range m.Functions {
				size := d.u32()
				body := must(d.r.Sub(uint64(size)))
				section := d.r
				d.r = body
				f := &m.Functions[i]
				f.Locals = make([]module.Local, d.count())
				var total uint64
				for j := range f.Locals {
					f.Locals[j] = module.Local{Count: d.u32(), Type: d.valueType()}
					total += uint64(f.Locals[j].Count)
					if total > math.MaxUint32 {
						d.fail("too many locals")
					}
				}
				if total > uint64(d.limits.MaxLocals) {
					d.limit("locals")
				}
				d.inCode = true
				f.LocalCount, f.Body = uint32(total), d.expr()
				d.inCode = false
				check(d.r.Finish())
				d.r = section
			}
		case 11:
			m.Data = make([]module.Data, d.count())
			for i := range m.Data {
				s := &m.Data[i]
				flag := d.u32()
				switch flag {
				case 0:
					s.Offset = d.expr()
				case 1:
					s.Mode = module.Passive
				case 2:
					s.Memory, s.Offset = d.u32(), d.expr()
				default:
					d.fail("unknown data segment flag")
				}
				s.Bytes = d.bytes(uint64(d.u32()))
			}
		case 12:
			m.HasDataCount, m.DataCount = true, d.u32()
		}
		check(d.r.Finish())
		d.r = parent
	}
	if uint64(codeCount) != uint64(len(m.Functions)) {
		d.fail("function/code count mismatch")
	}
	if m.HasDataCount && uint64(m.DataCount) != uint64(len(m.Data)) {
		d.fail("data count mismatch")
	}
	if d.needsDataCount && !m.HasDataCount {
		d.fail("data count section required")
	}
	return m, nil
}

func must[T any](v T, err error) T { check(err); return v }
func check(err error) {
	if err != nil {
		panic(err)
	}
}
func (d *decoder) failAt(at uint64, detail string) {
	panic(&wasm.Error{Kind: wasm.ErrMalformed, Offset: at, Detail: detail})
}
func (d *decoder) fail(detail string) { d.failAt(d.r.Offset(), detail) }
func (d *decoder) limit(detail string) {
	panic(&wasm.Error{Kind: wasm.ErrLimit, Offset: d.r.Offset(), Detail: detail})
}
func (d *decoder) byte() byte            { return must(d.r.Byte()) }
func (d *decoder) u32() uint32           { return must(d.r.U32()) }
func (d *decoder) bytes(n uint64) []byte { return must(d.r.Bytes(n)) }
func (d *decoder) name() string          { return must(d.r.Name(d.limits.MaxNameBytes)) }
func (d *decoder) expect(b byte) {
	at := d.r.Offset()
	if d.byte() != b {
		d.failAt(at, "unexpected reserved byte")
	}
}
func (d *decoder) charge(n uint64) {
	if n > d.limits.MaxNodes-d.nodes {
		d.limit("decoded nodes")
	}
	d.nodes += n
}
func (d *decoder) count() uint32 {
	n := d.u32()
	// Every vector element consumes at least one input byte. Reject impossible
	// claims before allocating, even if a caller selected a very large quota.
	if uint64(n) > uint64(d.r.Remaining()) {
		d.fail("vector exceeds remaining input")
	}
	d.charge(uint64(n))
	return n
}
func (d *decoder) valueType() wasm.ValueType {
	at := d.r.Offset()
	t := wasm.ValueType(d.byte())
	if !t.Valid() {
		d.failAt(at, "unknown value type")
	}
	return t
}
func (d *decoder) refType() wasm.ValueType {
	at := d.r.Offset()
	t := wasm.ValueType(d.byte())
	if !t.IsReference() {
		d.failAt(at, "reference type required")
	}
	return t
}
func (d *decoder) types() []wasm.ValueType {
	ts := make([]wasm.ValueType, d.count())
	for i := range ts {
		ts[i] = d.valueType()
	}
	return ts
}
func (d *decoder) indexes() []uint32 {
	xs := make([]uint32, d.count())
	for i := range xs {
		xs[i] = d.u32()
	}
	return xs
}
func (d *decoder) bounds() module.Limits {
	flag := d.byte()
	if flag > 1 {
		d.fail("unknown limits flag")
	}
	l := module.Limits{Min: d.u32(), HasMax: flag == 1}
	if l.HasMax {
		l.Max = d.u32()
	}
	return l
}
func (d *decoder) table() module.TableType {
	return module.TableType{Element: d.refType(), Limits: d.bounds()}
}
func (d *decoder) globalType() module.GlobalType {
	t := d.valueType()
	mut := d.byte()
	if mut > 1 {
		d.fail("unknown mutability")
	}
	return module.GlobalType{Value: t, Mutable: mut == 1}
}
func (d *decoder) element() module.Element {
	flag := d.u32()
	if flag > 7 {
		d.fail("unknown element segment flag")
	}
	s := module.Element{Type: wasm.FuncRef}
	if flag&1 == 0 {
		if flag&2 != 0 {
			s.Table = d.u32()
		}
		s.Offset = d.expr()
	} else if flag&2 == 0 {
		s.Mode = module.Passive
	} else {
		s.Mode = module.Declarative
	}
	if flag&4 == 0 {
		if flag != 0 {
			d.expect(0)
		}
		s.Functions = d.indexes()
	} else {
		if flag != 4 {
			s.Type = d.refType()
		}
		s.Expressions = make([]module.Expr, d.count())
		for i := range s.Expressions {
			s.Expressions[i] = d.expr()
		}
	}
	return s
}
