// Package model defines component types independently of binary decoding and
// Core execution. Compilation owns these graphs; public metadata uses copies.
package model

type Kind byte

const (
	Resource     Kind = 0x3f
	Function     Kind = 0x40
	Component    Kind = 0x41
	Instance     Kind = 0x42
	ErrorContext Kind = 0x64
	Future       Kind = 0x65
	Stream       Kind = 0x66
	Borrow       Kind = 0x68
	Own          Kind = 0x69
	Result       Kind = 0x6a
	Option       Kind = 0x6b
	Enum         Kind = 0x6d
	Flags        Kind = 0x6e
	Tuple        Kind = 0x6f
	List         Kind = 0x70
	Variant      Kind = 0x71
	Record       Kind = 0x72
	String       Kind = 0x73
	Char         Kind = 0x74
	F64          Kind = 0x75
	F32          Kind = 0x76
	U64          Kind = 0x77
	S64          Kind = 0x78
	U32          Kind = 0x79
	S32          Kind = 0x7a
	U16          Kind = 0x7b
	S16          Kind = 0x7c
	U8           Kind = 0x7d
	S8           Kind = 0x7e
	Bool         Kind = 0x7f
)

type Field struct {
	Name string
	Type *Type
}

// Type is a resolved component value type. Resource identity is nominal;
// records/tuples and variants retain declaration order. Nil case types denote
// a case without payload, not an implicit unit type.
type Type struct {
	Kind       Kind
	Fields     []Field
	Cases      []Field
	Labels     []string
	Element    *Type
	ResourceID uint64
}
