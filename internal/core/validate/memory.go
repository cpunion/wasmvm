package validate

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (v *checker) alignment(in module.Instr, max uint32) {
	v.memory(0)
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
	v.pop(wasm.I32)
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
		v.memory(0)
		v.index(in.A, len(v.m.Data), "data segment")
		v.threeI32()
	case sub == 9:
		v.index(in.A, len(v.m.Data), "data segment")
	case sub == 10 || sub == 11:
		v.memory(0)
		v.threeI32()
	case sub == 12:
		v.index(in.A, len(v.m.Elements), "element segment")
		if v.table(in.B).Element != v.m.Elements[in.A].Type {
			v.fail("table.init element type mismatch")
		}
		v.threeI32()
	case sub == 13:
		v.index(in.A, len(v.m.Elements), "element segment")
	case sub == 14:
		if v.table(in.A).Element != v.table(in.B).Element {
			v.fail("table.copy element type mismatch")
		}
		v.threeI32()
	case sub == 15:
		t := v.table(in.A).Element
		v.pop(wasm.I32)
		v.pop(t)
		v.push(wasm.I32)
	case sub == 16:
		v.table(in.A)
		v.push(wasm.I32)
	case sub == 17:
		t := v.table(in.A).Element
		v.pop(wasm.I32)
		v.pop(t)
		v.pop(wasm.I32)
	default:
		v.fail("unknown miscellaneous instruction")
	}
}
