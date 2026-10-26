// Package native provides the platform C entry boundary for owned machine code.
// It does not decode Wasm or execute an interpreter. Only compiler-produced,
// trusted code and C-ABI entry offsets may be published through this package.
package native

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
)

var (
	ErrUnavailable = errors.New("native execution unavailable")
	ErrClosed      = errors.New("native resource closed")
	ErrBusy        = errors.New("native context already executing")
	ErrInvalid     = errors.New("invalid native artifact or execution record")
	ErrLimit       = errors.New("native allocation limit")
)

type Exit = nativeabi.Exit
type State = nativeabi.State
type record = nativeabi.Record

const (
	Return             = nativeabi.Return
	Trap               = nativeabi.Trap
	HostCall           = nativeabi.HostCall
	GrowMemory         = nativeabi.GrowMemory
	GrowStack          = nativeabi.GrowStack
	Poll               = nativeabi.Poll
	FuelExhausted      = nativeabi.FuelExhausted
	GuestGC            = nativeabi.GuestGC
	RuntimeHelper      = nativeabi.RuntimeHelper
	Call               = nativeabi.Call
	BadState           = nativeabi.BadState
	Move               = nativeabi.Move
	ABIVersion         = nativeabi.ABIVersion
	RecordSize         = nativeabi.RecordSize
	ArenaOffset        = nativeabi.ArenaOffset
	CapacityOffset     = nativeabi.CapacityOffset
	MaxDepthOffset     = nativeabi.MaxDepthOffset
	UsedOffset         = nativeabi.UsedOffset
	DepthOffset        = nativeabi.DepthOffset
	FunctionOffset     = nativeabi.FunctionOffset
	ContinuationOffset = nativeabi.ContinuationOffset
	SourceOffset       = nativeabi.SourceOffset
	RootMapOffset      = nativeabi.RootMapOffset
	FuelOffset         = nativeabi.FuelOffset
	BudgetOffset       = nativeabi.BudgetOffset
	ResultLoOffset     = nativeabi.ResultLoOffset
	ResultHiOffset     = nativeabi.ResultHiOffset
	PayloadOffset      = nativeabi.PayloadOffset
	ControlOffset      = nativeabi.ControlOffset
	MemoryBaseOffset   = nativeabi.MemoryBaseOffset
	MemorySizeOffset   = nativeabi.MemorySizeOffset
)
