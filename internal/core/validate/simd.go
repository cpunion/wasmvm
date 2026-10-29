package validate

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (v *checker) lane(in module.Instr, count byte) {
	if in.Lane >= count {
		v.fail("SIMD lane out of range")
	}
}
func (v *checker) simd(in module.Instr) {
	op := in.Op & 0xffff
	if op >= 0x100 && op <= 0x113 {
		if v.features&wasm.RelaxedSIMD == 0 {
			v.fail("relaxed SIMD is disabled")
		}
		if op >= 0x105 && op <= 0x10c || op == 0x113 {
			v.pop(wasm.V128)
		}
		if op >= 0x101 && op <= 0x104 {
			v.unary(wasm.V128, wasm.V128)
		} else {
			v.binary(wasm.V128, wasm.V128)
		}
		return
	}
	switch {
	case op <= 11 || op == 92 || op == 93:
		var align uint32
		switch {
		case op == 0 || op == 11:
			align = 4
		case op <= 6:
			align = 3
		case op <= 10:
			align = op - 7
		case op == 92:
			align = 2
		case op == 93:
			align = 3
		}
		v.alignment(in, align)
		if op == 11 {
			v.pop(wasm.V128)
		}
		v.pop(v.memories[in.Memory].AddressType())
		if op != 11 {
			v.push(wasm.V128)
		}
	case op == 12:
		v.push(wasm.V128)
	case op == 13:
		for _, lane := range in.Vector {
			if lane >= 32 {
				v.fail("shuffle lane out of range")
			}
		}
		v.binary(wasm.V128, wasm.V128)
	case op >= 15 && op <= 20:
		types := [...]wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I64, wasm.F32, wasm.F64}
		v.unary(types[op-15], wasm.V128)
	case op >= 21 && op <= 34:
		var t wasm.ValueType
		var count byte
		switch {
		case op <= 23:
			t, count = wasm.I32, 16
		case op <= 26:
			t, count = wasm.I32, 8
		case op <= 28:
			t, count = wasm.I32, 4
		case op <= 30:
			t, count = wasm.I64, 2
		case op <= 32:
			t, count = wasm.F32, 4
		default:
			t, count = wasm.F64, 2
		}
		v.lane(in, count)
		if op == 23 || op == 26 || op == 28 || op == 30 || op == 32 || op == 34 {
			v.pop(t)
			v.unary(wasm.V128, wasm.V128)
		} else {
			v.unary(wasm.V128, t)
		}
	case op >= 84 && op <= 91:
		align := (op - 84) % 4
		v.alignment(in, align)
		v.lane(in, byte(16>>align))
		v.pop(wasm.V128)
		v.pop(v.memories[in.Memory].AddressType())
		if op <= 87 {
			v.push(wasm.V128)
		}
	case op == 82:
		v.pop(wasm.V128)
		v.binary(wasm.V128, wasm.V128)
	case op == 83 || op == 99 || op == 100 || op == 131 || op == 132 || op == 163 || op == 164 || op == 195 || op == 196:
		v.unary(wasm.V128, wasm.I32)
	case op >= 107 && op <= 109 || op >= 139 && op <= 141 || op >= 171 && op <= 173 || op >= 203 && op <= 205:
		v.pop(wasm.I32)
		v.unary(wasm.V128, wasm.V128)
	case op == 77 || op >= 94 && op <= 98 || op >= 103 && op <= 106 || op == 116 || op == 117 || op == 122 || op >= 124 && op <= 129 || op >= 135 && op <= 138 || op == 148 || op == 160 || op == 161 || op >= 167 && op <= 170 || op == 192 || op == 193 || op >= 199 && op <= 202 || op == 224 || op == 225 || op == 227 || op == 236 || op == 237 || op == 239 || op >= 248 && op <= 255:
		v.unary(wasm.V128, wasm.V128)
	case op == 14 || op >= 35 && op <= 76 || op >= 78 && op <= 81 || op == 101 || op == 102 || op >= 110 && op <= 115 || op >= 118 && op <= 121 || op == 123 || op == 130 || op == 133 || op == 134 || op >= 142 && op <= 147 || op >= 149 && op <= 153 || op >= 155 && op <= 159 || op == 174 || op == 177 || op >= 181 && op <= 186 || op >= 188 && op <= 191 || op == 206 || op == 209 || op >= 213 && op <= 223 || op >= 228 && op <= 235 || op >= 240 && op <= 247:
		v.binary(wasm.V128, wasm.V128)
	default:
		v.fail("unknown SIMD instruction")
	}
}
