package validate

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (c *context) composite(index uint32, kind byte) module.TypeDef {
	c.index(index, len(c.m.Types), "composite type")
	d := c.m.TypeDef(index)
	if d.Kind != kind {
		c.fail("composite type kind mismatch")
	}
	return d
}
func (c *context) array(index uint32) module.FieldType {
	return c.composite(index, module.ArrayKind).Fields[0]
}
func (c *context) field(index, field uint32) module.FieldType {
	d := c.composite(index, module.StructKind)
	c.index(field, len(d.Fields), "struct field")
	return d.Fields[field]
}
func (c *context) refTop(t wasm.ValueType) wasm.ValueType {
	if t.IsExceptionReference() {
		return wasm.ExnRef
	}
	if t.IsFunctionReference() {
		return wasm.FuncRef
	}
	if t.IsExternalReference() {
		return wasm.ExternRef
	}
	return wasm.AnyRef
}
func (v *checker) gc(in module.Instr) {
	if v.features&wasm.GC == 0 {
		v.fail("GC is disabled")
	}
	op := in.Op & 0xffff
	if op >= 32 && v.features&wasm.CustomDescriptors == 0 {
		v.fail("custom descriptors are disabled")
	}
	ref := func(i uint32) wasm.ValueType { return v.m.Reference(i, true) }
	allocated := func(i uint32) wasm.ValueType {
		return ref(i).NonNull().WithExact(v.features&wasm.CustomDescriptors != 0)
	}
	descriptor := func(t wasm.ValueType) wasm.ValueType {
		i, ok := t.TypeIndex()
		if !ok {
			v.fail("defined type required for descriptor cast")
		}
		d := v.composite(i, module.StructKind)
		if d.Descriptor == 0 {
			v.fail("type has no descriptor")
		}
		return ref(uint32(d.Descriptor - 1)).WithExact(t.Exact())
	}
	defaultable := func(f module.FieldType) {
		if !f.Unpacked().Defaultable() {
			v.fail("nondefaultable field")
		}
	}
	numeric := func(f module.FieldType) {
		if f.Unpacked().IsReference() {
			v.fail("numeric array required")
		}
	}
	mutable := func(f module.FieldType) {
		if !f.Mutable {
			v.fail("immutable field")
		}
	}
	packed := func(f module.FieldType, extended bool) {
		if (f.Packed != 0) != extended {
			v.fail("packed access mismatch")
		}
	}
	switch op {
	case 0, 1, 32, 33:
		d := v.composite(in.A, module.StructKind)
		if (d.Descriptor != 0) != (op >= 32) {
			v.fail("struct allocation descriptor mismatch")
		}
		if op >= 32 {
			v.pop(descriptor(ref(in.A).WithExact(true)))
		}
		fs := d.Fields
		if op == 0 || op == 32 {
			for i := len(fs) - 1; i >= 0; i-- {
				v.pop(fs[i].Unpacked())
			}
		} else {
			for _, f := range fs {
				defaultable(f)
			}
		}
		v.push(allocated(in.A))
	case 2, 3, 4:
		f := v.field(in.A, in.B)
		packed(f, op != 2)
		v.pop(ref(in.A))
		v.push(f.Unpacked())
	case 5:
		f := v.field(in.A, in.B)
		mutable(f)
		v.pop(f.Unpacked())
		v.pop(ref(in.A))
	case 6, 7:
		f := v.array(in.A)
		v.pop(wasm.I32)
		if op == 6 {
			v.pop(f.Unpacked())
		} else {
			defaultable(f)
		}
		v.push(allocated(in.A))
	case 8:
		f := v.array(in.A)
		if uint64(in.B) > uint64(v.limits.MaxStack) {
			v.limit("array fixed operands")
		}
		for range in.B {
			v.pop(f.Unpacked())
		}
		v.push(allocated(in.A))
	case 9, 10:
		f := v.array(in.A)
		if op == 9 {
			numeric(f)
			v.index(in.B, len(v.m.Data), "data segment")
			if !v.m.HasDataCount {
				v.fail("data count required")
			}
		} else {
			v.index(in.B, len(v.m.Elements), "element segment")
			if !v.matches(v.m.Elements[in.B].Type, f.Unpacked()) {
				v.fail("array element type mismatch")
			}
		}
		v.pop(wasm.I32)
		v.pop(wasm.I32)
		v.push(allocated(in.A))
	case 11, 12, 13:
		f := v.array(in.A)
		packed(f, op != 11)
		v.pop(wasm.I32)
		v.pop(ref(in.A))
		v.push(f.Unpacked())
	case 14:
		f := v.array(in.A)
		mutable(f)
		v.pop(f.Unpacked())
		v.pop(wasm.I32)
		v.pop(ref(in.A))
	case 15:
		v.pop(wasm.ArrayRef)
		v.push(wasm.I32)
	case 16:
		f := v.array(in.A)
		mutable(f)
		v.pop(wasm.I32)
		v.pop(f.Unpacked())
		v.pop(wasm.I32)
		v.pop(ref(in.A))
	case 17:
		dst, src := v.array(in.A), v.array(in.B)
		mutable(dst)
		if dst.Packed != src.Packed || dst.Packed == 0 && !v.matches(src.Value, dst.Value) {
			v.fail("array copy type mismatch")
		}
		v.pop(wasm.I32)
		v.pop(wasm.I32)
		v.pop(ref(in.B))
		v.pop(wasm.I32)
		v.pop(ref(in.A))
	case 18, 19:
		f := v.array(in.A)
		mutable(f)
		if op == 18 {
			numeric(f)
			v.index(in.B, len(v.m.Data), "data segment")
			if !v.m.HasDataCount {
				v.fail("data count required")
			}
		} else {
			v.index(in.B, len(v.m.Elements), "element segment")
			if !v.matches(v.m.Elements[in.B].Type, f.Unpacked()) {
				v.fail("array element type mismatch")
			}
		}
		v.pop(wasm.I32)
		v.pop(wasm.I32)
		v.pop(wasm.I32)
		v.pop(ref(in.A))
	case 20, 21, 22, 23, 35, 36:
		t := in.ReferenceType()
		v.valueType(t, len(v.m.Types))
		if !t.IsReference() {
			v.fail("cast reference type")
		}
		if op >= 35 {
			v.pop(descriptor(t))
		}
		v.pop(v.refTop(t))
		if op <= 21 {
			v.push(wasm.I32)
		} else {
			v.push(t)
		}
	case 24, 25, 37, 38:
		if len(in.Types) != 2 {
			v.fail("cast branch reference types")
		}
		a, b := in.Types[0], in.Types[1]
		v.valueType(a, len(v.m.Types))
		v.valueType(b, len(v.m.Types))
		compatible := v.matches(b, a)
		if v.features&wasm.CustomDescriptors != 0 {
			compatible = v.refTop(a) == v.refTop(b)
		}
		if !a.IsReference() || !b.IsReference() || !compatible {
			v.fail("cast branch type mismatch")
		}
		diff := a
		if b.Nullable() {
			diff = diff.NonNull()
		}
		branch, fall := b, diff
		if op == 25 || op == 38 {
			branch, fall = diff, b
		}
		ts := v.label(in.A)
		if len(ts) == 0 || !v.matches(branch, ts[len(ts)-1]) {
			v.fail("cast branch label type")
		}
		if op >= 37 {
			v.pop(descriptor(b))
		}
		v.pop(a)
		v.pops(ts[:len(ts)-1])
		v.push(ts[:len(ts)-1]...)
		v.push(fall)
	case 34:
		t := v.pop(ref(in.A))
		exact := v.matches(t, ref(in.A).WithExact(true))
		v.push(descriptor(ref(in.A).WithExact(exact)).NonNull())
	case 26, 27:
		a, b := wasm.ExternRef, wasm.AnyRef
		if op == 27 {
			a, b = b, a
		}
		t := v.pop(a)
		if t == 0 || !t.Nullable() {
			b = b.NonNull()
		}
		v.push(b)
	case 28:
		v.pop(wasm.I32)
		v.push(wasm.I31Ref.NonNull())
	case 29, 30:
		v.pop(wasm.I31Ref)
		v.push(wasm.I32)
	default:
		v.fail("unknown GC instruction")
	}
}
