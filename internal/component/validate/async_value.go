package validate

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (c *componentContext) asyncValueBuiltin(def *model.Canonical) error {
	t, err := c.result.scope.At(def.Index)
	if err != nil {
		return err
	}
	future := def.Op >= 0x15 && def.Op <= 0x1b || def.Op == 0x2f
	kind := model.Stream
	if future {
		kind = model.Future
	}
	if t.Kind != kind {
		return invalid("async endpoint type")
	}
	info := &CanonicalFunction{Definition: def, Type: t}
	op := def.Op
	if future && op <= 0x1b {
		op -= 7
	}
	f := module.FuncType{Params: []wasm.ValueType{wasm.I32}}
	switch op {
	case 0x0e:
		info.Async = true
		f.Params = nil
		f.Results = []wasm.ValueType{wasm.I64}
	case 0x0f, 0x10:
		for _, o := range def.Options {
			if o.Kind > 4 && o.Kind != 6 {
				return invalid("async copy canonical option")
			}
		}
		address, err := c.canonicalOptions(info)
		if err != nil {
			return err
		}
		needed, err := c.hasPointers(t.Value.Element)
		if err != nil {
			return err
		}
		if t.Value.Element != nil && info.Memory == nil || op == 0x0f && needed && info.Realloc == nil {
			return invalid("async copy requires memory/realloc")
		}
		f.Params = append(f.Params, address)
		if !future {
			f.Params = append(f.Params, address)
		}
		f.Results = []wasm.ValueType{address}
		if future {
			f.Results = []wasm.ValueType{wasm.I32}
		}
	case 0x11, 0x12:
		info.Async = def.Flag != 0
		f.Results = []wasm.ValueType{wasm.I32}
	case 0x13, 0x14:
	case 0x2e, 0x2f:
		f.Params = append(f.Params, wasm.I32)
	default:
		return invalid("async endpoint operation")
	}
	info.Signature = f
	if err := c.addSignature(f); err != nil {
		return err
	}
	c.result.Canonical = append(c.result.Canonical, info)
	return nil
}
