package validate

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (c *context) continuation(index uint32) module.FuncType {
	d := c.composite(index, module.ContinuationKind)
	return c.typ(d.Continuation)
}

func (c *context) resultsSubtype(actual, expected []wasm.ValueType) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i, t := range actual {
		if !c.matches(t, expected[i]) {
			return false
		}
	}
	return true
}

func (c *context) functionShapeSubtype(actual, expected module.FuncType) bool {
	return c.resultsSubtype(expected.Params, actual.Params) && c.resultsSubtype(actual.Results, expected.Results)
}

func (c *context) continuationReference(t wasm.ValueType) module.FuncType {
	index, ok := t.TypeIndex()
	if !ok || !t.IsContinuationReference() {
		c.fail("concrete continuation reference required")
	}
	return c.continuation(index)
}

func (v *checker) stackSwitch(in module.Instr) {
	if v.features&wasm.StackSwitching == 0 {
		v.fail("stack switching is disabled")
	}
	if in.Op == 0xe2 {
		tag := v.tag(in.A)
		v.pops(tag.Params)
		v.push(tag.Results...)
		return
	}
	sig := v.continuation(in.A)
	ref := v.m.Reference(in.A, true)
	switch in.Op {
	case 0xe0:
		d := v.m.TypeDef(in.A)
		v.pop(v.m.Reference(d.Continuation, true))
		v.push(ref.NonNull().WithExact(v.features&wasm.CustomDescriptors != 0))
	case 0xe1:
		target := v.continuation(in.B)
		n := len(sig.Params) - len(target.Params)
		if n < 0 || !v.functionShapeSubtype(module.FuncType{Params: sig.Params[max(0, n):], Results: sig.Results}, target) {
			v.fail("continuation binding type mismatch")
		}
		v.pop(ref)
		v.pops(sig.Params[:n])
		v.push(v.m.Reference(in.B, false).WithExact(v.features&wasm.CustomDescriptors != 0))
	case 0xe3, 0xe4, 0xe5:
		for _, handler := range in.On {
			tag := v.tag(handler.Tag)
			if handler.Kind == 1 {
				if len(tag.Params) != 0 || !v.resultsSubtype(sig.Results, tag.Results) || !v.resultsSubtype(tag.Results, sig.Results) {
					v.fail("switch handler type mismatch")
				}
				continue
			}
			if handler.Kind != 0 {
				v.fail("invalid continuation handler")
			}
			label := v.label(handler.Label)
			if len(label) != len(tag.Params)+1 {
				v.fail("continuation handler arity")
			}
			dest := v.continuationReference(label[len(label)-1])
			if !v.resultsSubtype(tag.Params, label[:len(label)-1]) || !v.functionShapeSubtype(module.FuncType{Params: tag.Results, Results: sig.Results}, dest) {
				v.fail("continuation handler type mismatch")
			}
		}
		v.pop(ref)
		if in.Op == 0xe3 {
			v.pops(sig.Params)
		} else if in.Op == 0xe4 {
			v.pops(v.exceptionTag(in.B).Params)
		} else {
			v.pop(wasm.ExnRef)
		}
		v.push(sig.Results...)
	case 0xe6:
		tag := v.tag(in.B)
		if len(tag.Params) != 0 || len(sig.Params) == 0 {
			v.fail("invalid switch continuation or tag")
		}
		back := v.continuationReference(sig.Params[len(sig.Params)-1])
		if !v.resultsSubtype(sig.Results, tag.Results) || !v.resultsSubtype(tag.Results, back.Results) {
			v.fail("switch result type mismatch")
		}
		v.pop(ref)
		v.pops(sig.Params[:len(sig.Params)-1])
		v.push(back.Params...)
	default:
		v.fail("invalid continuation opcode")
	}
}
