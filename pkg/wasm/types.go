package wasm

import "fmt"

// ValueType identifies a Core 2.0 value type by its binary encoding.
type ValueType byte

const (
	I32       ValueType = 0x7f
	I64       ValueType = 0x7e
	F32       ValueType = 0x7d
	F64       ValueType = 0x7c
	V128      ValueType = 0x7b
	FuncRef   ValueType = 0x70
	ExternRef ValueType = 0x6f
)

func (t ValueType) IsReference() bool { return t == FuncRef || t == ExternRef }
func (t ValueType) Valid() bool       { return t >= V128 && t <= I32 || t.IsReference() }
func (t ValueType) String() string {
	switch t {
	case I32:
		return "i32"
	case I64:
		return "i64"
	case F32:
		return "f32"
	case F64:
		return "f64"
	case V128:
		return "v128"
	case FuncRef:
		return "funcref"
	case ExternRef:
		return "externref"
	default:
		return fmt.Sprintf("type(%#x)", byte(t))
	}
}
