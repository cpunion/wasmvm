package validate

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (c *context) tagType(index uint32) module.FuncType {
	if c.features&wasm.ExceptionHandling == 0 {
		c.fail("tags are disabled")
	}
	t := c.typ(index)
	if len(t.Results) != 0 {
		c.fail("tag type has results")
	}
	return t
}
func (c *context) tag(index uint32) module.FuncType {
	if c.features&wasm.ExceptionHandling == 0 {
		c.fail("exceptions are disabled")
	}
	c.index(index, len(c.tags), "tag")
	return c.tags[index]
}
func (v *checker) catches(catches []module.Catch) {
	if v.features&wasm.ExceptionHandling == 0 {
		v.fail("exceptions are disabled")
	}
	for _, c := range catches {
		if c.Kind > 3 {
			v.fail("invalid catch kind")
		}
		var types []wasm.ValueType
		if c.Kind < 2 {
			types = v.tag(c.Tag).Params
		}
		target := v.label(c.Label)
		n := len(types)
		if c.Kind&1 != 0 {
			n++
		}
		if len(target) != n {
			v.fail("catch branch arity")
		}
		for i, t := range types {
			if !v.matches(t, target[i]) {
				v.fail("catch branch type")
			}
		}
		if c.Kind&1 != 0 && !v.matches(wasm.ExnRef.NonNull(), target[n-1]) {
			v.fail("catch reference type")
		}
	}
}
