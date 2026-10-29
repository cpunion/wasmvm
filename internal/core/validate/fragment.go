package validate

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"math"
)

// Fragment validates a lexical Core type group embedded in a component. IDs
// binds preceding lexical types, including resolved outer aliases. Types may
// contain trusted canonical references only in those already-bound aliases;
// new decoded groups still use local indices. Registry is compilation-local.
type Fragment struct {
	Types    *module.Module
	IDs      []uint32
	Registry *module.TypeRegistry
	Features wasm.Features
	MaxTypes uint32
}

// Extern checks one Core external descriptor against already bound lexical
// types. It uses the same value, tag, memory and table rules as module checking.
func (f Fragment) Extern(t module.Import) (err error) {
	if f.Features & ^wasm.SupportedFeatures != 0 {
		return wasm.ErrUnsupported
	}
	defer func() {
		if failure := recover(); failure != nil {
			if e, ok := failure.(*wasm.Error); ok {
				err = e
			} else {
				panic(failure)
			}
		}
	}()
	c := context{m: f.Types, features: f.Features, typeLabels: f.IDs, lexicalTypes: true}
	if f.Types == nil || f.Registry == nil || len(f.IDs) != len(f.Types.Types) {
		c.fail("missing lexical Core external types")
	}
	c.registry, c.lexicalBound = *f.Registry, uint32(len(f.IDs))
	switch t.Kind {
	case 0, 4:
		c.typ(t.Type)
		if f.IDs[t.Type] == 0 || uint64(f.IDs[t.Type]) >= uint64(len(c.registry.Types)) {
			c.fail("unbound lexical function type")
		}
		if t.Kind == 4 {
			c.tagType(t.Type)
		}
	case 1:
		c.tableType(t.Table)
	case 2:
		c.memoryType(t.Memory)
	case 3:
		c.valueType(t.Global.Value, len(f.IDs))
	default:
		c.fail("invalid Core external descriptor")
	}
	return nil
}

func (c *context) memoryType(t module.Limits) {
	if t.Shared && (c.features&wasm.Threads == 0 || !t.HasMax) {
		c.fail("shared memory requires Threads and a maximum")
	}
	maximum := uint64(65536)
	if t.Index64 {
		maximum = 1 << 48
	}
	c.bounds(t, maximum)
}

func (c *context) tableType(t module.TableType) {
	if t.Limits.Shared {
		c.fail("shared tables are not Core Threads")
	}
	c.valueType(t.Element, len(c.m.Types))
	if !t.Element.IsReference() {
		c.fail("invalid table element type")
	}
	maximum := uint64(math.MaxUint32)
	if t.Limits.Index64 {
		maximum = math.MaxUint64
	}
	c.bounds(t.Limits, maximum)
}

func (f Fragment) TypeGroup(start, count uint32) (err error) {
	if f.Features & ^wasm.SupportedFeatures != 0 {
		return wasm.ErrUnsupported
	}
	c := context{m: f.Types, features: f.Features, typeLabels: f.IDs, lexicalTypes: true, lexicalBound: start}
	if f.Registry == nil {
		return &wasm.Error{Kind: wasm.ErrInvalid, Detail: "missing lexical Core type registry"}
	}
	f.Registry.Init()
	c.registry = *f.Registry
	mark := c.registry.Count()
	binding := false
	defer func() {
		if failure := recover(); failure != nil {
			if e, ok := failure.(*wasm.Error); ok {
				c.registry.Rollback(mark)
				if binding {
					clear(f.IDs[start : start+count])
				}
				err = e
			} else {
				panic(failure)
			}
		} else {
			*f.Registry = c.registry
		}
	}()
	end := uint64(start) + uint64(count)
	if f.Types == nil || end > uint64(len(f.Types.Types)) || len(f.IDs) != len(f.Types.Types) || len(f.Types.TypeDefs) != 0 && len(f.Types.TypeDefs) != len(f.Types.Types) {
		c.fail("invalid lexical Core type group")
	}
	if count == 0 {
		if f.Features&wasm.GC == 0 {
			c.fail("empty recursive Core group requires GC")
		}
		return nil
	}
	for i := start; uint64(i) < end; i++ {
		if f.IDs[i] != 0 {
			c.fail("lexical Core type is already bound")
		}
	}
	if len(f.Types.TypeDefs) != 0 && f.Features&wasm.GC == 0 {
		c.fail("composite lexical Core types require GC")
	}
	binding = true
	c.checkTypeGroup(module.TypeGroup{Start: start, Count: count}, f.MaxTypes)
	return nil
}
