package nativeabi

import (
	"math"
	"unsafe"
)

// GlobalPointerOffset locates a cell pointer after the immutable function IDs
// and table length/address pairs. Both emitters use signed 32-bit displacements.
func GlobalPointerOffset(functions, tables, index uint32) (uint32, bool) {
	offset := uint64(functions)*8 + uint64(tables)*16 + uint64(index)*8
	return uint32(offset), offset <= math.MaxInt32
}

// TablePairOffset locates the length/base pair following the function IDs.
func TablePairOffset(functions, index uint32) (uint32, bool) {
	offset := uint64(functions)*8 + uint64(index)*16
	return uint32(offset), offset <= math.MaxInt32-8
}

// MemoryPairOffset appends length/base pairs after any numeric-global view.
// Existing function/table/global offsets remain unchanged.
func MemoryPairOffset(functions, tables, globals, bindings, index uint32) (uint32, bool) {
	words := uint64(functions) + uint64(tables)*2
	if bindings != 0 {
		words += uint64(globals) + uint64(bindings)*2
	}
	offset := (words + uint64(index)*2) * 8
	return uint32(offset), offset <= math.MaxInt32-8
}

type Exit uint32

const (
	Return Exit = iota + 1
	Trap
	HostCall
	GrowMemory
	GrowStack
	Poll
	FuelExhausted
	GuestGC
	RuntimeHelper
	Call
	BadState
	Move
)

const ABIVersion = 2

// State contains scalar bits, guest handles and explicitly designated native
// runtime addresses. Continuation and RootMap are compiler-owned IDs, never
// addresses supplied by guest memory. The driver accesses State only while idle.
type State struct {
	// Used is the byte offset of the current frame in Arena.
	Used, Depth, Function, Continuation uint64
	Source, RootMap, Fuel, Budget       uint64
	ResultLo, ResultHi                  uint64
	// Payload[2] is a trusted native runtime view, never a guest address.
	Payload [4]uint64
}

// Record is the fixed C ABI, verified against the C layout on each target. The
// control word is accessed only by adapter atomics or emitted atomic loads.
// Arena is a pointer to owned native storage; no Go pointer crosses this ABI.
type Record struct {
	Version, Arena, Capacity, MaxDepth uint64
	State
	Control                uint64
	MemoryBase, MemorySize uint64
}

const (
	RecordSize         = unsafe.Sizeof(Record{})
	ArenaOffset        = unsafe.Offsetof(Record{}.Arena)
	CapacityOffset     = unsafe.Offsetof(Record{}.Capacity)
	MaxDepthOffset     = unsafe.Offsetof(Record{}.MaxDepth)
	UsedOffset         = unsafe.Offsetof(Record{}.State) + unsafe.Offsetof(State{}.Used)
	DepthOffset        = unsafe.Offsetof(Record{}.State) + unsafe.Offsetof(State{}.Depth)
	FunctionOffset     = unsafe.Offsetof(Record{}.State) + unsafe.Offsetof(State{}.Function)
	ContinuationOffset = unsafe.Offsetof(Record{}.State) + unsafe.Offsetof(State{}.Continuation)
	SourceOffset       = unsafe.Offsetof(Record{}.State) + unsafe.Offsetof(State{}.Source)
	RootMapOffset      = unsafe.Offsetof(Record{}.State) + unsafe.Offsetof(State{}.RootMap)
	FuelOffset         = unsafe.Offsetof(Record{}.State) + unsafe.Offsetof(State{}.Fuel)
	BudgetOffset       = unsafe.Offsetof(Record{}.State) + unsafe.Offsetof(State{}.Budget)
	ResultLoOffset     = unsafe.Offsetof(Record{}.State) + unsafe.Offsetof(State{}.ResultLo)
	ResultHiOffset     = unsafe.Offsetof(Record{}.State) + unsafe.Offsetof(State{}.ResultHi)
	PayloadOffset      = unsafe.Offsetof(Record{}.State) + unsafe.Offsetof(State{}.Payload)
	RuntimeDataOffset  = PayloadOffset + 16
	ControlOffset      = unsafe.Offsetof(Record{}.Control)
	MemoryBaseOffset   = unsafe.Offsetof(Record{}.MemoryBase)
	MemorySizeOffset   = unsafe.Offsetof(Record{}.MemorySize)
)
