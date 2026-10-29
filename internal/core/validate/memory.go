package validate

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (v *checker) alignment(in module.Instr, max uint32) {
	v.memory(in.Memory)
	if !v.memories[in.Memory].Index64 && in.MemoryOffsetHigh != 0 {
		v.fail("memory offset exceeds address width")
	}
	if in.A > max {
		v.fail("alignment exceeds natural alignment")
	}
}
func (v *checker) scalarMemory(in module.Instr) {
	types := [...]wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I64, wasm.I64, wasm.I64, wasm.I64, wasm.I64, wasm.I64, wasm.I32, wasm.I64, wasm.F32, wasm.F64, wasm.I32, wasm.I32, wasm.I64, wasm.I64, wasm.I64}
	align := [...]uint32{2, 3, 2, 3, 0, 0, 1, 1, 0, 0, 1, 1, 2, 2, 2, 3, 2, 3, 0, 1, 0, 1, 2}
	i := in.Op - 0x28
	v.alignment(in, align[i])
	if in.Op >= 0x36 {
		v.pop(types[i])
	}
	v.pop(v.memories[in.Memory].AddressType())
	if in.Op < 0x36 {
		v.push(types[i])
	}
}
func (v *checker) threeI32() { v.pop(wasm.I32); v.pop(wasm.I32); v.pop(wasm.I32) }
func (v *checker) misc(in module.Instr) {
	sub := in.Op & 0xffff
	switch {
	case sub < 8:
		input, output := wasm.F32, wasm.I32
		if sub&2 != 0 {
			input = wasm.F64
		}
		if sub >= 4 {
			output = wasm.I64
		}
		v.unary(input, output)
	case sub == 8:
		v.memory(in.B)
		v.index(in.A, len(v.m.Data), "data segment")
		v.pop(wasm.I32)
		v.pop(wasm.I32)
		v.pop(v.memories[in.B].AddressType())
	case sub == 9:
		v.index(in.A, len(v.m.Data), "data segment")
	case sub == 10:
		v.memory(in.A)
		v.memory(in.B)
		dst, src := v.memories[in.A].AddressType(), v.memories[in.B].AddressType()
		v.pop(copyAddressType(dst, src))
		v.pop(src)
		v.pop(dst)
	case sub == 11:
		v.memory(in.A)
		at := v.memories[in.A].AddressType()
		v.pop(at)
		v.pop(wasm.I32)
		v.pop(at)
	case sub == 12:
		v.index(in.A, len(v.m.Elements), "element segment")
		if !v.matches(v.m.Elements[in.A].Type, v.table(in.B).Element) {
			v.fail("table.init element type mismatch")
		}
		v.pop(wasm.I32)
		v.pop(wasm.I32)
		v.pop(v.table(in.B).Limits.AddressType())
	case sub == 13:
		v.index(in.A, len(v.m.Elements), "element segment")
	case sub == 14:
		if !v.matches(v.table(in.B).Element, v.table(in.A).Element) {
			v.fail("table.copy element type mismatch")
		}
		dst, src := v.table(in.A).Limits.AddressType(), v.table(in.B).Limits.AddressType()
		v.pop(copyAddressType(dst, src))
		v.pop(src)
		v.pop(dst)
	case sub == 15:
		t := v.table(in.A).Element
		at := v.table(in.A).Limits.AddressType()
		v.pop(at)
		v.pop(t)
		v.push(at)
	case sub == 16:
		v.table(in.A)
		v.push(v.table(in.A).Limits.AddressType())
	case sub == 17:
		t := v.table(in.A).Element
		at := v.table(in.A).Limits.AddressType()
		v.pop(at)
		v.pop(t)
		v.pop(at)
	default:
		v.fail("unknown miscellaneous instruction")
	}
}

func copyAddressType(dst, src wasm.ValueType) wasm.ValueType {
	if dst == wasm.I32 || src == wasm.I32 {
		return wasm.I32
	}
	return wasm.I64
}
