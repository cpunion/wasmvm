// Package module defines the owned, decoded Core 2.0 representation. A decoded
// Module is unvalidated and must never be executed or lowered directly.
package module

import "github.com/cpunion/wasmvm/pkg/wasm"

type FuncType struct{ Params, Results []wasm.ValueType }

// SignatureKey serializes full reference metadata, not truncated binary codes.
// labels optionally normalizes acyclic module-local function type identities.
func SignatureKey(t FuncType, labels []uint32) string {
	b := make([]byte, 0, 8*(len(t.Params)+len(t.Results)+1))
	appendType := func(v wasm.ValueType) {
		if index, ok := v.TypeIndex(); ok && !v.Canonical() && labels != nil {
			v = v.WithTypeIndex(labels[index], false)
		}
		for n := uint(0); n < 64; n += 8 {
			b = append(b, byte(uint64(v)>>n))
		}
	}
	for _, v := range t.Params {
		appendType(v)
	}
	appendType(0)
	for _, v := range t.Results {
		appendType(v)
	}
	return string(b)
}

type Limits struct {
	Min, Max uint64
	HasMax   bool
	Index64  bool
	Shared   bool
}

func (l Limits) AddressType() wasm.ValueType {
	if l.Index64 {
		return wasm.I64
	}
	return wasm.I32
}

func (l Limits) Address(bits uint64) uint64 {
	if l.Index64 {
		return bits
	}
	return uint64(uint32(bits))
}

func (l Limits) Failure() uint64 {
	if l.Index64 {
		return ^uint64(0)
	}
	return uint64(^uint32(0))
}

type TableType struct {
	Element wasm.ValueType
	Limits  Limits
}
type GlobalType struct {
	Value   wasm.ValueType
	Mutable bool
}
type Import struct {
	Module, Name string
	Kind         byte
	Type         uint32
	Table        TableType
	Memory       Limits
	Global       GlobalType
}
type Export struct {
	Name  string
	Kind  byte
	Index uint32
}
type Local struct {
	Count uint32
	Type  wasm.ValueType
}
type Function struct {
	Type       uint32
	Locals     []Local
	LocalCount uint32
	Body       Expr
}
type Global struct {
	Type GlobalType
	Init Expr
}
type Custom struct {
	Name string
	Data []byte
}

type SegmentMode byte

const (
	Active SegmentMode = iota
	Passive
	Declarative
)

type Element struct {
	Mode        SegmentMode
	Table       uint32
	Offset      Expr
	Type        wasm.ValueType
	Functions   []uint32
	Expressions []Expr
}
type Data struct {
	Mode   SegmentMode
	Memory uint32
	Offset Expr
	Bytes  []byte
}

// Instr retains raw numeric bits, byte offsets and unvalidated immediates.
// Prefixed operations use prefix<<16 | subopcode. A and B hold index/memarg
// operands in their binary order; Memory identifies scalar/SIMD memargs.
// Block holds the signed block type. Vector
// holds v128/shuffle bytes; Lane holds a lane immediate. Labels excludes the
// default br_table target, which is in A. Types is the typed-select vector.
// Catch describes one ordered try_table arm. Kinds 0/1 are tagged, and
// kinds 1/3 append a non-null exception reference to the branch tuple.
type Catch struct {
	Kind       byte
	Tag, Label uint32
}

type Instr struct {
	Op     uint32
	Offset uint64
	A, B   uint32
	Memory uint32 // memory index for scalar and SIMD memargs
	// The upper displacement word fills padding on the native targets.
	// B retains the low word so the Core 2.0 representation stays compact.
	MemoryOffsetHigh uint32
	Block            int64
	Bits             uint64
	Vector           [16]byte
	Lane             byte
	Labels           []uint32
	Types            []wasm.ValueType
	Catches          []Catch
}

func (in *Instr) MemoryOffset() uint64 { return uint64(in.B) | uint64(in.MemoryOffsetHigh)<<32 }

type Expr []Instr

// Module owns all retained bytes; callers must not mutate it after validation
// and publication. It has no runtime state and cannot execute by itself.
type Module struct {
	Types        []FuncType
	TypeDefs     []TypeDef
	TypeGroups   []TypeGroup
	Imports      []Import
	Functions    []Function
	Tables       []TableType
	TableInit    []Expr // nil entries use the element type's default value
	Memories     []Limits
	Tags         []uint32
	Globals      []Global
	Exports      []Export
	HasStart     bool
	Start        uint32
	Elements     []Element
	Data         []Data
	Custom       []Custom
	HasDataCount bool
	DataCount    uint32
}

// ReferenceType preserves old manually constructed instruction fixtures while
// using the full type word for explicitly decoded reference instructions.
func (in *Instr) ReferenceType() wasm.ValueType {
	if in.Bits != 0 {
		return wasm.ValueType(in.Bits)
	}
	return wasm.ValueType(in.A)
}

// BlockResult decodes the internal block representation, including a typed
// reference result packed below the sign bit. Type-index blocks stay positive.
func BlockResult(t int64) wasm.ValueType {
	if t < -128 {
		return wasm.ValueType(uint64(t) & ((1 << 63) - 1))
	}
	return wasm.ValueType(byte(t) & 0x7f)
}

func IsCall(op uint32) bool     { return op >= 0x10 && op <= 0x15 }
func IsTailCall(op uint32) bool { return op == 0x12 || op == 0x13 || op == 0x15 }
