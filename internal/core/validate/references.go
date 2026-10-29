package validate

import "github.com/cpunion/wasmvm/pkg/wasm"

func (c *context) valueType(t wasm.ValueType, bound int) {
	if !t.Valid() {
		c.fail("invalid value type")
	}
	if t <= 255 {
		return
	}
	if c.features&wasm.FunctionReferences == 0 {
		c.fail("typed references are disabled")
	}
	if t.Canonical() {
		c.fail("canonical type in decoded module")
	}
	if index, indexed := t.TypeIndex(); indexed {
		c.index(index, bound, "reference type")
	}
}

func (c *context) matches(actual, expected wasm.ValueType) bool {
	return actual == expected || actual == 0 || wasm.ReferenceSubtype(actual, expected, c.typeLabels)
}

func (c *context) functionReference(index uint32) wasm.ValueType {
	if c.features&wasm.FunctionReferences == 0 {
		return wasm.FuncRef
	}
	return wasm.FunctionReference(c.funcTypeIndices[index], false)
}

func (v *checker) resetInitializers(mark int) {
	for _, index := range v.setLocals[mark:] {
		v.initialized[index] = false
	}
	v.setLocals = v.setLocals[:mark]
}

func (v *checker) referenceBranch(op, index uint32) {
	t := v.pop(0)
	if t != 0 && !t.IsReference() {
		v.fail("reference required")
	}
	ts := v.label(index)
	if op == 0xd5 {
		v.pops(ts)
		v.push(ts...)
		v.push(t.NonNull())
	} else {
		if len(ts) == 0 || !ts[len(ts)-1].IsReference() || !v.matches(t.NonNull(), ts[len(ts)-1]) {
			v.fail("nonnull branch target type")
		}
		v.pops(ts[:len(ts)-1])
		v.push(ts[:len(ts)-1]...)
	}
}
