package wasm

import "fmt"

// ValueType retains legacy binary codes and compact typed-reference metadata.
// Indexed types are module-local until resolved against a Store's type domain.
type ValueType uint64

const (
	I32         ValueType = 0x7f
	I64         ValueType = 0x7e
	F32         ValueType = 0x7d
	F64         ValueType = 0x7c
	V128        ValueType = 0x7b
	FuncRef     ValueType = 0x70
	ExternRef   ValueType = 0x6f
	AnyRef      ValueType = 0x6e
	EqRef       ValueType = 0x6d
	I31Ref      ValueType = 0x6c
	StructRef   ValueType = 0x6b
	ArrayRef    ValueType = 0x6a
	NoneRef     ValueType = 0x71
	NoExternRef ValueType = 0x72
	NoFuncRef   ValueType = 0x73
)

func (t ValueType) IsReference() bool {
	return t.referenceEncoding() && (byte(t) >= byte(ArrayRef) && byte(t) <= byte(NoFuncRef))
}
func (t ValueType) Valid() bool { return t >= V128 && t <= I32 || t.IsReference() }
func (t ValueType) String() string {
	if t.IsReference() && t > 255 {
		heap := t.HeapName()
		if index, ok := t.TypeIndex(); ok {
			heap = fmt.Sprintf("type[%d]", index)
			if t.Canonical() {
				heap = fmt.Sprintf("canonical[%d]", index)
			}
		}
		if t.Nullable() {
			return "(ref null " + heap + ")"
		}
		return "(ref " + heap + ")"
	}
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
	case NoneRef:
		return "nullref"
	case NoFuncRef:
		return "nullfuncref"
	case NoExternRef:
		return "nullexternref"
	case AnyRef, EqRef, I31Ref, StructRef, ArrayRef:
		return t.HeapName() + "ref"
	case FuncRef:
		return "funcref"
	case ExternRef:
		return "externref"
	default:
		return fmt.Sprintf("type(%#x)", uint64(t))
	}
}
