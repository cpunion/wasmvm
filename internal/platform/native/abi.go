// Package native provides the platform C entry boundary for owned machine code.
// It does not decode Wasm or execute an interpreter. Only compiler-produced,
// trusted code and C-ABI entry offsets may be published through this package.
package native

import (
	"errors"
	"unsafe"
)

var (
	ErrUnavailable = errors.New("native execution unavailable")
	ErrClosed      = errors.New("native resource closed")
	ErrBusy        = errors.New("native context already executing")
	ErrInvalid     = errors.New("invalid native artifact or execution record")
	ErrLimit       = errors.New("native allocation limit")
)

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

// record is the fixed C ABI, verified against the C layout on each target. The
// control word is accessed only by adapter atomics, including during execution.
// Arena is a pointer to owned native storage; no Go pointer crosses this ABI.
type record struct {
	Version, Arena, Capacity, MaxDepth uint64
	State
	control uint64
}

const (
	RecordSize         = unsafe.Sizeof(record{})
	ArenaOffset        = unsafe.Offsetof(record{}.Arena)
	CapacityOffset     = unsafe.Offsetof(record{}.Capacity)
	MaxDepthOffset     = unsafe.Offsetof(record{}.MaxDepth)
	UsedOffset         = unsafe.Offsetof(record{}.State) + unsafe.Offsetof(State{}.Used)
	DepthOffset        = unsafe.Offsetof(record{}.State) + unsafe.Offsetof(State{}.Depth)
	FunctionOffset     = unsafe.Offsetof(record{}.State) + unsafe.Offsetof(State{}.Function)
	ContinuationOffset = unsafe.Offsetof(record{}.State) + unsafe.Offsetof(State{}.Continuation)
	SourceOffset       = unsafe.Offsetof(record{}.State) + unsafe.Offsetof(State{}.Source)
	RootMapOffset      = unsafe.Offsetof(record{}.State) + unsafe.Offsetof(State{}.RootMap)
	FuelOffset         = unsafe.Offsetof(record{}.State) + unsafe.Offsetof(State{}.Fuel)
	BudgetOffset       = unsafe.Offsetof(record{}.State) + unsafe.Offsetof(State{}.Budget)
	ResultLoOffset     = unsafe.Offsetof(record{}.State) + unsafe.Offsetof(State{}.ResultLo)
	ResultHiOffset     = unsafe.Offsetof(record{}.State) + unsafe.Offsetof(State{}.ResultHi)
	PayloadOffset      = unsafe.Offsetof(record{}.State) + unsafe.Offsetof(State{}.Payload)
	ControlOffset      = unsafe.Offsetof(record{}.control)
)
