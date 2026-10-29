package validate

import (
	"math"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type frame struct {
	op                   uint32
	height               int
	in, out              []wasm.ValueType
	unreachable, hadElse bool
	initMark             int
}
type checker struct {
	*context
	locals          []wasm.ValueType
	stack           []wasm.ValueType
	control         []frame
	info            FunctionInfo
	record          bool
	inputs, outputs []wasm.ValueType
	initialized     []bool
	setLocals       []uint32
}

func (c *context) body(f *module.Function) FunctionInfo {
	sig := c.typ(f.Type)
	n := uint64(len(sig.Params))
	for _, group := range f.Locals {
		n += uint64(group.Count)
		c.valueType(group.Type, len(c.m.Types))
	}
	if n > math.MaxUint32 {
		c.fail("too many locals including parameters")
	}
	if n > uint64(c.limits.MaxLocals) {
		c.limit("locals including parameters")
	}
	v := checker{context: c, locals: make([]wasm.ValueType, int(n))}
	copy(v.locals, sig.Params)
	at := len(sig.Params)
	for _, group := range f.Locals {
		for j := uint32(0); j < group.Count; j++ {
			v.locals[at] = group.Type
			at++
		}
	}
	for _, typ := range v.locals[len(sig.Params):] {
		if !typ.Defaultable() {
			v.initialized = make([]bool, len(v.locals))
			for i, typ := range v.locals {
				v.initialized[i] = i < len(sig.Params) || typ.Defaultable()
			}
			break
		}
	}
	v.enter(0, nil, sig.Results)
	for pc, in := range f.Body {
		c.at = in.Offset
		if len(v.control) == 0 {
			c.fail("instruction after function end")
		}
		v.inputs, v.outputs = v.inputs[:0], v.outputs[:0]
		v.record = c.visitor != nil && (in.Op == 1 || in.Op == 8 || in.Op == 10 || module.IsCall(in.Op) || in.Op >= 0x1a)
		v.instruction(in)
		if c.visitor != nil {
			for i, j := 0, len(v.inputs)-1; i < j; i, j = i+1, j-1 {
				v.inputs[i], v.inputs[j] = v.inputs[j], v.inputs[i]
			}
			if err := c.visitor(pc, in, v.inputs, v.outputs); err != nil {
				panic(visitorFailure{err})
			}
		}
	}
	if len(v.control) != 0 {
		c.fail("missing function end")
	}
	return v.info
}
func (v *checker) top() *frame { return &v.control[len(v.control)-1] }
func (v *checker) pop(want wasm.ValueType) wasm.ValueType {
	f := v.top()
	if len(v.stack) == f.height {
		if f.unreachable {
			if v.record {
				v.inputs = append(v.inputs, want)
			}
			return 0
		}
		v.fail("operand stack underflow")
	}
	got := v.stack[len(v.stack)-1]
	v.stack = v.stack[:len(v.stack)-1]
	if want != 0 && got != 0 && !v.matches(got, want) {
		v.fail("operand type mismatch")
	}
	if v.record {
		v.inputs = append(v.inputs, got)
	}
	return got
}
func (v *checker) push(ts ...wasm.ValueType) {
	if uint64(len(v.stack))+uint64(len(ts)) > uint64(v.limits.MaxStack) {
		v.limit("operand stack")
	}
	v.stack = append(v.stack, ts...)
	if v.record {
		v.outputs = append(v.outputs, ts...)
	}
	if uint32(len(v.stack)) > v.info.MaxStack {
		v.info.MaxStack = uint32(len(v.stack))
	}
}
func (v *checker) pops(ts []wasm.ValueType) {
	for i := len(ts) - 1; i >= 0; i-- {
		v.pop(ts[i])
	}
}
func (v *checker) enter(op uint32, in, out []wasm.ValueType) {
	if uint64(len(v.control)) >= uint64(v.limits.MaxControl) {
		v.limit("control stack")
	}
	v.control = append(v.control, frame{op: op, height: len(v.stack), in: in, out: out, initMark: len(v.setLocals)})
	v.push(in...)
	if uint32(len(v.control)) > v.info.MaxControl {
		v.info.MaxControl = uint32(len(v.control))
	}
}
func (v *checker) unreachable() { f := v.top(); v.stack = v.stack[:f.height]; f.unreachable = true }
func (v *checker) label(index uint32) []wasm.ValueType {
	v.index(index, len(v.control), "label")
	f := v.control[len(v.control)-1-int(index)]
	if f.op == 3 {
		return f.in
	}
	return f.out
}
func (v *checker) block(t int64) module.FuncType {
	if t == -64 {
		return module.FuncType{}
	}
	if t >= 0 {
		if uint64(t) > math.MaxUint32 {
			v.fail("invalid block type index")
		}
		return v.typ(uint32(t))
	}
	typ := module.BlockResult(t)
	v.valueType(typ, len(v.m.Types))
	return module.FuncType{Results: []wasm.ValueType{typ}}
}
func (v *checker) unary(in, out wasm.ValueType)  { v.pop(in); v.push(out) }
func (v *checker) binary(in, out wasm.ValueType) { v.pop(in); v.pop(in); v.push(out) }

func (v *checker) instruction(in module.Instr) {
	switch op := in.Op; {
	case op == 0:
		v.unreachable()
	case op == 1:
	case op == 2 || op == 3 || op == 4 || op == 0x1f:
		if op == 4 {
			v.pop(wasm.I32)
		}
		if op == 0x1f {
			v.catches(in.Catches)
		}
		sig := v.block(in.Block)
		v.pops(sig.Params)
		v.enter(op, sig.Params, sig.Results)
	case op == 8 || op == 10:
		if v.features&wasm.ExceptionHandling == 0 {
			v.fail("exceptions are disabled")
		}
		if op == 8 {
			v.pops(v.tag(in.A).Params)
		} else {
			v.pop(wasm.ExnRef)
		}
		v.unreachable()
	case op == 5:
		f := v.top()
		if f.op != 4 || f.hadElse {
			v.fail("unexpected else")
		}
		v.pops(f.out)
		if len(v.stack) != f.height {
			v.fail("extra operands at else")
		}
		f.unreachable = false
		f.hadElse = true
		v.resetInitializers(f.initMark)
		v.push(f.in...)
	case op == 0x0b:
		f := *v.top()
		v.pops(f.out)
		if len(v.stack) != f.height {
			v.fail("extra operands at end")
		}
		if f.op == 4 && !f.hadElse {
			if len(f.in) != len(f.out) {
				v.fail("implicit else result mismatch")
			}
			for i, t := range f.in {
				if !v.matches(t, f.out[i]) {
					v.fail("implicit else type mismatch")
				}
			}
		}
		v.resetInitializers(f.initMark)
		v.control = v.control[:len(v.control)-1]
		v.push(f.out...)
	case op == 0x0c || op == 0x0d:
		if op == 0x0d {
			v.pop(wasm.I32)
		}
		ts := v.label(in.A)
		v.pops(ts)
		if op == 0x0c {
			v.unreachable()
		} else {
			v.push(ts...)
		}
	case op == 0x0e:
		v.pop(wasm.I32)
		ts := v.label(in.A)
		// Preserve actual operand types, including unknowns in unreachable
		// code, while comparing every target. Equal arity alone is insufficient.
		values := make([]wasm.ValueType, len(ts))
		for i := len(ts) - 1; i >= 0; i-- {
			values[i] = v.pop(ts[i])
		}
		for _, label := range in.Labels {
			other := v.label(label)
			if len(other) != len(ts) {
				v.fail("branch table arity mismatch")
			}
			for i, t := range values {
				if t != 0 && !v.matches(t, other[i]) {
					v.fail("branch table type mismatch")
				}
			}
		}
		v.unreachable()
	case op == 0x0f:
		v.pops(v.control[0].out)
		v.unreachable()
	case module.IsCall(op):
		tail := module.IsTailCall(op)
		if tail && v.features&wasm.TailCall == 0 {
			v.fail("tail calls are disabled")
		}
		var sig module.FuncType
		if op == 0x10 || op == 0x12 {
			sig = v.function(in.A)
		} else if op == 0x14 || op == 0x15 {
			if v.features&(wasm.FunctionReferences|wasm.GC) == 0 {
				v.fail("function references are disabled")
			}
			sig = v.typ(in.A)
			v.pop(wasm.FunctionReference(in.A, true))
		} else {
			sig = v.typ(in.A)
			if !v.table(in.B).Element.IsFunctionReference() {
				v.fail("indirect call requires funcref table")
			}
			v.pop(v.table(in.B).Limits.AddressType())
		}
		v.pops(sig.Params)
		if tail {
			want := v.control[0].out
			if len(want) != len(sig.Results) {
				v.fail("tail call result arity")
			}
			for i, t := range want {
				if !v.matches(sig.Results[i], t) {
					v.fail("tail call result type")
				}
			}
			v.unreachable()
		} else {
			v.push(sig.Results...)
		}
	case op == 0x1a:
		v.pop(0)
	case op == 0x1b:
		v.pop(wasm.I32)
		b, a := v.pop(0), v.pop(0)
		if a.IsReference() || b.IsReference() {
			v.fail("untyped select requires numeric/vector values")
		}
		if a != 0 && b != 0 && a != b {
			v.fail("select operand mismatch")
		}
		if a == 0 {
			a = b
		}
		v.push(a)
	case op == 0x1c:
		if len(in.Types) != 1 || !in.Types[0].Valid() {
			v.fail("typed select requires one value type")
		}
		v.valueType(in.Types[0], len(v.m.Types))
		v.pop(wasm.I32)
		v.pop(in.Types[0])
		v.pop(in.Types[0])
		v.push(in.Types[0])
	case op >= 0x20 && op <= 0x22:
		v.index(in.A, len(v.locals), "local")
		t := v.locals[in.A]
		if op == 0x20 && v.initialized != nil && !v.initialized[in.A] {
			v.fail("uninitialized nonnullable local")
		}
		if op != 0x20 {
			v.pop(t)
			if v.initialized != nil && !v.initialized[in.A] {
				v.initialized[in.A] = true
				v.setLocals = append(v.setLocals, in.A)
			}
		}
		if op != 0x21 {
			v.push(t)
		}
	case op == 0x23 || op == 0x24:
		g := v.global(in.A)
		if op == 0x23 {
			v.push(g.Value)
		} else {
			if !g.Mutable {
				v.fail("immutable global")
			}
			v.pop(g.Value)
		}
	case op == 0x25 || op == 0x26:
		t := v.table(in.A).Element
		if op == 0x26 {
			v.pop(t)
		}
		v.pop(v.table(in.A).Limits.AddressType())
		if op == 0x25 {
			v.push(t)
		}
	case op >= 0x28 && op <= 0x3e:
		v.scalarMemory(in)
	case op == 0x3f:
		v.memory(in.A)
		v.push(v.memories[in.A].AddressType())
	case op == 0x40:
		v.memory(in.A)
		v.unary(v.memories[in.A].AddressType(), v.memories[in.A].AddressType())
	case op >= 0x41 && op <= 0x44:
		v.push([]wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64}[op-0x41])
	case op >= 0x45 && op <= 0xc4:
		v.numeric(op)
	case op == 0xd0:
		t := in.ReferenceType()
		v.valueType(t, len(v.m.Types))
		if !t.IsReference() {
			v.fail("invalid reference type")
		}
		v.push(t)
	case op == 0xd1:
		t := v.pop(0)
		if t != 0 && !t.IsReference() {
			v.fail("reference required")
		}
		v.push(wasm.I32)
	case op == 0xd2:
		v.function(in.A)
		if !v.refs[in.A] {
			v.fail("undeclared function reference")
		}
		v.push(v.functionReference(in.A))
	case op == 0xd3:
		if v.features&wasm.GC == 0 {
			v.fail("GC is disabled")
		}
		v.pop(wasm.EqRef)
		v.pop(wasm.EqRef)
		v.push(wasm.I32)
	case op == 0xd4:
		if v.features&(wasm.FunctionReferences|wasm.GC) == 0 {
			v.fail("function references are disabled")
		}
		t := v.pop(0)
		if t != 0 && !t.IsReference() {
			v.fail("reference required")
		}
		v.push(t.NonNull())
	case op == 0xd5 || op == 0xd6:
		if v.features&(wasm.FunctionReferences|wasm.GC) == 0 {
			v.fail("function references are disabled")
		}
		v.referenceBranch(op, in.A)
	case op>>16 == 0xfb:
		v.gc(in)
	case op>>16 == 0xfc:
		v.misc(in)
	case op>>16 == 0xfd:
		v.simd(in)
	default:
		v.fail("unknown instruction")
	}
}
