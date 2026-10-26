package nativeabi

import "unsafe"

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
)

const ABIVersion = 1

// State contains only scalar bits and guest handles. Continuation and RootMap
// are compiler-owned IDs, never addresses supplied by guest memory. The driver
// reads or replaces State only while execution is idle.
type State struct {
	Used, Depth, Function, Continuation uint64
	Source, RootMap, Fuel, Budget       uint64
	ResultLo, ResultHi                  uint64
	Payload                             [4]uint64
}

// Record is the fixed C ABI, verified against the C layout on each target. The
// control word is accessed only by adapter atomics, including during execution.
// Arena is a pointer to owned native storage; no Go pointer crosses this ABI.
type Record struct {
	Version, Arena, Capacity, MaxDepth uint64
	State
	Control uint64
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
	ControlOffset      = unsafe.Offsetof(Record{}.Control)
)
