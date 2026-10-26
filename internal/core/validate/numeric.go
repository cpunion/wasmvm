package validate

import "github.com/cpunion/wasmvm/pkg/wasm"

func (v *checker) numeric(op uint32) {
	switch {
	case op == 0x45:
		v.unary(wasm.I32, wasm.I32)
	case op <= 0x4f:
		v.binary(wasm.I32, wasm.I32)
	case op == 0x50:
		v.unary(wasm.I64, wasm.I32)
	case op <= 0x5a:
		v.binary(wasm.I64, wasm.I32)
	case op <= 0x60:
		v.binary(wasm.F32, wasm.I32)
	case op <= 0x66:
		v.binary(wasm.F64, wasm.I32)
	case op <= 0x69:
		v.unary(wasm.I32, wasm.I32)
	case op <= 0x78:
		v.binary(wasm.I32, wasm.I32)
	case op <= 0x7b:
		v.unary(wasm.I64, wasm.I64)
	case op <= 0x8a:
		v.binary(wasm.I64, wasm.I64)
	case op <= 0x91:
		v.unary(wasm.F32, wasm.F32)
	case op <= 0x98:
		v.binary(wasm.F32, wasm.F32)
	case op <= 0x9f:
		v.unary(wasm.F64, wasm.F64)
	case op <= 0xa6:
		v.binary(wasm.F64, wasm.F64)
	case op <= 0xbf:
		conversion := conversions[op-0xa7]
		v.unary(conversion[0], conversion[1])
	case op <= 0xc1:
		v.unary(wasm.I32, wasm.I32)
	default:
		v.unary(wasm.I64, wasm.I64)
	}
}

// Input/output types for the consecutive conversion/reinterpret opcodes a7–bf.
var conversions = [...][2]wasm.ValueType{
	{wasm.I64, wasm.I32},
	{wasm.F32, wasm.I32}, {wasm.F32, wasm.I32}, {wasm.F64, wasm.I32}, {wasm.F64, wasm.I32},
	{wasm.I32, wasm.I64}, {wasm.I32, wasm.I64},
	{wasm.F32, wasm.I64}, {wasm.F32, wasm.I64}, {wasm.F64, wasm.I64}, {wasm.F64, wasm.I64},
	{wasm.I32, wasm.F32}, {wasm.I32, wasm.F32}, {wasm.I64, wasm.F32}, {wasm.I64, wasm.F32}, {wasm.F64, wasm.F32},
	{wasm.I32, wasm.F64}, {wasm.I32, wasm.F64}, {wasm.I64, wasm.F64}, {wasm.I64, wasm.F64}, {wasm.F32, wasm.F64},
	{wasm.F32, wasm.I32}, {wasm.F64, wasm.I64}, {wasm.I32, wasm.F32}, {wasm.I64, wasm.F64},
}
