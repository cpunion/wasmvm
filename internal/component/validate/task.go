package validate

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (c *componentContext) taskBuiltin(def *model.Canonical) error {
	info := &CanonicalFunction{Definition: def}
	s := c.result.scope
	switch def.Op {
	case 9:
		for _, option := range def.Options {
			if option.Kind > 3 {
				return invalid("task.return canonical option")
			}
		}
		if _, err := c.canonicalOptions(info); err != nil {
			return err
		}
		t, err := s.resolve(&model.TypeDef{Kind: model.Function, Result: def.Results})
		if err != nil {
			return err
		}
		info.Type = t
		var params []*model.Type
		if t.Result != nil {
			params = []*model.Type{t.Result}
		}
		b := c.canonicalLayouts(info.Memory64)
		info.ABI, err = b.Function(params, nil, false)
		if err != nil {
			return err
		}
		pointers, err := c.hasPointers(t.Result)
		if err != nil {
			return err
		}
		if (pointers || info.ABI.ParameterMemory != nil) && info.Memory == nil {
			return invalid("task.return requires memory")
		}
		info.Signature.Params = info.ABI.Params
	case 0x0a, 0x0b:
		if (def.CoreValue != wasm.I32 && def.CoreValue != wasm.I64) || def.Index > 1 {
			return invalid("canonical context type or slot")
		}
		if c.contextType != 0 && c.contextType != def.CoreValue {
			return invalid("canonical context widths differ")
		}
		c.contextType = def.CoreValue
		if def.Op == 0x0a {
			info.Signature.Results = []wasm.ValueType{def.CoreValue}
		} else {
			info.Signature.Params = []wasm.ValueType{def.CoreValue}
		}
	case 0x0c:
		info.Signature.Results = []wasm.ValueType{wasm.I32}
	}
	if err := c.addSignature(module.FuncType{Params: info.Signature.Params, Results: info.Signature.Results}); err != nil {
		return err
	}
	c.result.Canonical = append(c.result.Canonical, info)
	return nil
}
