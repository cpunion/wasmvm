// Package module defines the owned, decoded Core 2.0 representation. A decoded
// Module is unvalidated and must never be executed or lowered directly.
package module

import "github.com/cpunion/wasmvm/pkg/wasm"

type FuncType struct{ Params, Results []wasm.ValueType }
type Limits struct {
	Min, Max uint64
	HasMax   bool
	Index64  bool
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
}

func (in *Instr) MemoryOffset() uint64 { return uint64(in.B) | uint64(in.MemoryOffsetHigh)<<32 }

type Expr []Instr

// Module owns all retained bytes; callers must not mutate it after validation
// and publication. It has no runtime state and cannot execute by itself.
type Module struct {
	Types        []FuncType
	Imports      []Import
	Functions    []Function
	Tables       []TableType
	Memories     []Limits
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
