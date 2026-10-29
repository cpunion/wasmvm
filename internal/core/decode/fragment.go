package decode

import (
	"errors"

	"github.com/cpunion/wasmvm/internal/binary"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Fragment reads Core grammar embedded in a Component Model type definition.
// The enclosing decoder owns input and aggregate budgets. Types and Reader are
// advanced only by this fragment's bounded grammar; discard them on failure.
// Core semantic validation remains a separate required step.
type Fragment struct {
	Reader   *binary.Reader
	Types    *module.Module
	Limits   Limits
	Features wasm.Features
}

func (f Fragment) read(parse func(*decoder)) (nodes uint64, err error) {
	if f.Features & ^wasm.SupportedFeatures != 0 {
		return 0, wasm.ErrUnsupported
	}
	d := decoder{r: *f.Reader, limits: f.Limits, features: f.Features, definition: f.Types}
	defer func() {
		if failure := recover(); failure != nil {
			switch e := failure.(type) {
			case *wasm.Error:
				err = e
			case *binary.Error:
				kind := wasm.ErrMalformed
				if errors.Is(e, binary.ErrLimit) {
					kind = wasm.ErrLimit
				}
				err = &wasm.Error{Kind: kind, Offset: e.Offset, Detail: e.Reason.Error()}
			default:
				panic(failure)
			}
		}
		*f.Reader = d.r
		nodes = d.nodes
	}()
	parse(&d)
	return d.nodes, nil
}

func (f Fragment) TypeGroup() (uint64, error) {
	return f.read(func(d *decoder) {
		d.charge(1)
		if d.features&wasm.GC != 0 {
			start := len(f.Types.Types)
			d.typeGroup(f.Types)
			resolveTypeReferencesFrom(f.Types, start)
		} else {
			d.expect(0x60)
			f.Types.Types = append(f.Types.Types, module.FuncType{Params: d.types(), Results: d.types()})
		}
	})
}

// Module decodes a complete embedded Core module while borrowing the input
// owned by its enclosing component. Its returned slices retain that ownership.
func (f Fragment) Module() (result *module.Module, nodes uint64, err error) {
	nodes, err = f.read(func(d *decoder) { result = d.module() })
	if err != nil {
		result = nil
	}
	return
}

func (f Fragment) ValueType() (result wasm.ValueType, nodes uint64, err error) {
	nodes, err = f.read(func(d *decoder) { result = d.valueType() })
	return
}

func (f Fragment) ExternType() (result module.Import, nodes uint64, err error) {
	nodes, err = f.read(func(d *decoder) {
		result.Kind = d.byte()
		switch result.Kind {
		case 0:
			result.Type = d.u32()
		case 1:
			result.Table = d.table()
		case 2:
			result.Memory = d.bounds(true)
		case 3:
			result.Global = d.globalType()
		case 4:
			result.Type = d.tagType()
		default:
			d.fail("unknown Core external type")
		}
	})
	return
}
