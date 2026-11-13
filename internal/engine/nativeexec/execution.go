package nativeexec

import (
	"context"
	"sync/atomic"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Execution struct {
	ctx             context.Context
	native          *native.Context
	modules         map[*slots.Module]*Lease
	maximum         uint64
	maxFrames       uint64
	metered         bool
	stop            func() bool
	interrupted     atomic.Bool
	reusable        bool
	values          []value.Value
	runtimeData     *native.Memory
	runtimeWords    []uint64
	allocationView  uint64
	runtimeInstance *state.Instance
	runtimeFunction *slots.Function
	runtimeGlobals  []globalBinding
	scalarData      *native.Memory
	scalarWords     []uint64
	scalarInstance  *state.Instance
	scalarBindings  []scalarBinding
	scalarAccesses  []scalarAccess
	entryState      native.State
	callData        *native.Memory
	callWords       []uint64
	callFrames      []state.NativeFrame
	callEnabled     bool
	callGrow        bool
	resumeToken     uint64
	nextResume      uint64
	resumeDepth     int
}

func New(ctx context.Context, modules map[*slots.Module]*Lease, maxSlots, maxFrames uint32, metered bool) (*Execution, error) {
	return newExecution(ctx, modules, maxSlots, maxFrames, metered, false)
}

// NewOwned delegates cancellation delivery to the invocation owner. The owner
// must interrupt on cancellation and drain that delivery before Close or reuse.
func NewOwned(ctx context.Context, modules map[*slots.Module]*Lease, maxSlots, maxFrames uint32, metered bool) (*Execution, error) {
	return newExecution(ctx, modules, maxSlots, maxFrames, metered, true)
}

func newExecution(ctx context.Context, modules map[*slots.Module]*Lease, maxSlots, maxFrames uint32, metered, owned bool) (*Execution, error) {
	c, err := native.NewContext(0, uint64(maxSlots)*16, uint64(max(1, maxFrames)))
	if err != nil {
		return nil, err
	}
	e := &Execution{native: c, modules: modules, maximum: uint64(maxSlots) * 16, maxFrames: uint64(maxFrames), reusable: true}
	if err := e.begin(ctx, metered, owned); err != nil {
		e.Close()
		return nil, err
	}
	return e, nil
}

// Begin binds an idle context to a new invocation. A context whose cancellation
// callback started is never reused: Suspend returns false and the owner closes it.
func (e *Execution) Begin(ctx context.Context, metered bool) error {
	return e.begin(ctx, metered, false)
}

// BeginOwned binds a new invocation with the same owner-delivered cancellation
// contract as NewOwned. No per-workspace watcher is registered.
func (e *Execution) BeginOwned(ctx context.Context, metered bool) error {
	return e.begin(ctx, metered, true)
}

func (e *Execution) begin(ctx context.Context, metered, owned bool) error {
	if e.native == nil || e.ctx != nil || !e.reusable || e.interrupted.Load() {
		return wasm.ErrInvalid
	}
	if ctx == nil {
		ctx = context.Background()
	}
	e.entryState = native.State{}
	e.ctx, e.metered = ctx, metered
	e.reusable = false
	if !owned {
		e.stop = context.AfterFunc(ctx, e.native.Cancel)
	}
	return nil
}

// Interrupt is concurrent with Run. The owner serializes it with Close and
// rebinding. Native cancellation is sticky, so interrupted storage is retired.
func (e *Execution) Interrupt() {
	e.interrupted.Store(true)
	e.native.Cancel()
}

// Suspend removes the former cancellation writer and releases native-visible
// instance memory. Only a callback stopped before it started permits reuse;
// the native context's sticky control word is never cleared.
func (e *Execution) Suspend() bool {
	if e.native == nil {
		return false
	}
	if e.ctx == nil {
		return e.reusable
	}
	stopped := !e.interrupted.Load()
	if e.stop != nil {
		stopped = e.stop() && stopped
	}
	e.stop, e.ctx, e.runtimeInstance = nil, nil, nil
	e.entryState = native.State{}
	e.resetCalls()
	e.resetRuntimeGlobals()
	e.resetScalarStorage()
	e.reusable = e.native.BindMemory(nil) == nil && stopped
	return e.reusable
}

func (e *Execution) Resize(length, capacity int) ([]value.Value, error) {
	if e.native == nil {
		return nil, native.ErrClosed
	}
	if length < 0 || capacity < length || uint64(capacity) > e.maximum/16 {
		return nil, wasm.ErrLimit
	}
	buffer, err := e.native.ResizeArena(uint64(capacity)*16, e.maximum)
	if err != nil {
		return nil, err
	}
	// Value contains exactly two uint64s and no Go references. Only this
	// execution retains the view, and all accesses stop during native entry.
	e.values = unsafe.Slice((*value.Value)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/16)
	return e.values[:length], nil
}

func (e *Execution) Close() {
	e.reusable = false
	e.ctx = nil
	e.entryState = native.State{}
	e.resetCalls()
	if e.callData != nil {
		e.callData.Close()
		e.callData = nil
	}
	e.callWords, e.callFrames = nil, nil
	if e.stop != nil {
		e.stop()
		e.stop = nil
	}
	e.values = nil
	if e.native != nil {
		e.native.Close()
		e.native = nil
	}
	if e.runtimeData != nil {
		e.runtimeData.Close()
		e.runtimeData = nil
	}
	e.runtimeWords, e.runtimeInstance = nil, nil
	e.resetRuntimeGlobals()
	e.resetScalarStorage()
	if e.scalarData != nil {
		e.scalarData.Close()
		e.scalarData = nil
	}
	e.scalarWords = nil
}

func (e *Execution) Run(fn *state.Function, base, depth int, pc uint32, fuel uint64) (uint32, uint64, error) {
	e.resetCalls()
	return e.runFrames(fn, base, depth, pc, fuel, e.maximum/16, e.maxFrames)
}

// RunFrames uses the invocation's current budgets, which may be smaller than
// the workspace's original capacity after a suspended parent shares its quota.
// Frame transfer is handled by the common machine before runtime effects.
func (e *Execution) RunFrames(frames *state.NativeFrames) error {
	resume := e.resumeCalls(frames)
	if !resume {
		e.resetCalls()
	}
	e.resumeToken = 0
	frames.Resume, frames.Parents = 0, 0
	e.callEnabled = true
	fn, base, depth := frames.Function, frames.Base, frames.Depth
	if resume {
		frames.Parents = uint32(len(e.callFrames) - 1)
		frames.Function, frames.Base, frames.Depth = e.callFrames[0].Function, e.callFrames[0].Base, depth-int(frames.Parents)
	}
	var err error
	frames.PC, frames.Fuel, err = e.runFrames(fn, base, depth, frames.PC, frames.Fuel,
		min(e.maximum/16, uint64(frames.MaxSlots)), min(e.maxFrames, uint64(frames.MaxFrames)))
	if err == nil && len(e.callWords) != 0 && (e.callWords[0] != 0 || resume) {
		err = e.publishCallFrames(frames)
	}
	if len(e.callWords) != 0 {
		e.callGrow = e.callWords[0] == e.callWords[1]
		if err == nil && e.callWords[0] != 0 && e.nextResume != ^uint64(0) {
			e.nextResume++
			e.resumeToken, frames.Resume = e.nextResume, e.nextResume
			e.resumeDepth = frames.Depth + len(e.callFrames) - 1
		}
	}
	if err != nil {
		e.resetCalls()
	}
	e.callEnabled = false
	return err
}

func (e *Execution) runFrames(fn *state.Function, base, depth int, pc uint32, fuel, maxSlots, maxFrames uint64) (uint32, uint64, error) {
	if e.native == nil {
		return 0, fuel, native.ErrClosed
	}
	if e.ctx == nil {
		return 0, fuel, wasm.ErrInvalid
	}
	if fn == nil || fn.Code == nil || fn.Instance == nil || base < 0 || depth < 1 {
		return 0, fuel, wasm.ErrInvalid
	}
	if uint64(depth) > maxFrames || uint64(base)+uint64(len(fn.Code.Slots)) > maxSlots {
		return 0, fuel, wasm.ErrLimit
	}
	module := e.modules[fn.Instance.Compiled]
	if module == nil {
		return 0, fuel, wasm.ErrUnsupported
	}
	entries, ok := module.function(fn)
	if !ok {
		return 0, fuel, wasm.ErrInvalid
	}
	var memory *native.Memory
	if len(fn.Instance.Memories) != 0 {
		memory = fn.Instance.Memories[0].Native()
		if memory == nil && fn.Instance.Memories[0].Shared() {
			memory = fn.Instance.Memories[0].NativeShared()
		}
		if memory == nil {
			return 0, fuel, wasm.ErrInvalid
		}
		if !entries.nativeMemory {
			// Preserve ownership/retirement validation without binding data that
			// the selected artifact cannot access.
			if _, err := memory.Size(); err != nil {
				return 0, fuel, err
			}
			memory = nil
		}
	}
	if err := e.native.BindMemory(memory); err != nil {
		return 0, fuel, err
	}
	e.allocationView = 0
	if !e.metered && entries.allocationBatch != 0 {
		return e.runAllocations(module, entries, fn, base, depth, pc, fuel, maxSlots, maxFrames)
	}
	return e.run(module, entries, fn, base, depth, pc, fuel, maxSlots, maxFrames)
}

// Only allocating entries need an unwind frame. It still covers every error
// and panic after reservation, including binding and pre-entry cancellation.
func (e *Execution) runAllocations(module *Lease, entries *functionEntries, fn *state.Function, base, depth int, pc uint32, fuel, maxSlots, maxFrames uint64) (uint32, uint64, error) {
	var err error
	e.allocationView, err = fn.Instance.Owner.PrepareNativeAllocations(entries.allocationBatch)
	if err != nil {
		return 0, fuel, err
	}
	defer fn.Instance.Owner.FinishNativeAllocations(fn.Instance.Types)
	return e.run(module, entries, fn, base, depth, pc, fuel, maxSlots, maxFrames)
}

func (e *Execution) run(module *Lease, entries *functionEntries, fn *state.Function, base, depth int, pc uint32, fuel, maxSlots, maxFrames uint64) (uint32, uint64, error) {
	entry := entries.ordinary
	needsData := entries.directRuntimeData
	e.entryState = native.State{Used: uint64(base) * 16, Depth: uint64(depth), Function: uint64(fn.Index), Continuation: uint64(pc), Fuel: fuel, Budget: 64}
	s := &e.entryState
	if e.metered {
		entry = entries.metered
		s.Payload[3] = 1
	} else if entries.hasExpansion && uint64(depth) <= maxFrames &&
		uint64(entries.inlineFrames) <= maxFrames-uint64(depth) &&
		uint64(base)+uint64(len(fn.Code.Slots))+uint64(entries.inlineSlots) <= maxSlots {
		entry = entries.expanded
		needsData = needsData || entries.usesRuntimeData
	}
	if needsData {
		var err error
		binding := fn.Code
		if module.linked {
			binding = nil
		}
		s.Payload[2], err = e.bindRuntimeGlobals(fn.Instance, binding, entries.globals)
		if err != nil {
			return 0, fuel, err
		}
	}
	if module.linked && e.callEnabled {
		var err error
		s.Payload[1], err = e.prepareCalls(maxSlots, maxFrames)
		if err != nil {
			return 0, fuel, err
		}
	}
	instance := fn.Instance
	for {
		if err := e.ctx.Err(); err != nil {
			return 0, s.Fuel, err
		}
		previousFuel := s.Fuel
		exit, err := e.native.RunState(module.code, entry, s)
		if e.allocationView != 0 {
			if finishErr := fn.Instance.Owner.FinishNativeAllocations(fn.Instance.Types); finishErr != nil {
				return 0, s.Fuel, finishErr
			}
		}
		if len(entries.globals) != 0 || len(e.scalarBindings) != 0 {
			e.flushRuntimeGlobals()
		}
		if err != nil {
			return 0, previousFuel, err
		}
		if s.Function != uint64(fn.Index) {
			if !module.linked || !e.callEnabled {
				return 0, s.Fuel, wasm.ErrInvalid
			}
			var ok bool
			fn, ok = linkedFunction(instance, s.Function)
			if !ok {
				return 0, s.Fuel, wasm.ErrInvalid
			}
		}
		if module.linked && e.callEnabled {
			if s.Used%16 != 0 || s.Used/16+uint64(len(fn.Code.Slots)) > uint64(len(e.values)) {
				return 0, s.Fuel, wasm.ErrInvalid
			}
			base = int(s.Used / 16)
			active, exists := module.function(fn)
			if !exists {
				return 0, s.Fuel, wasm.ErrInvalid
			}
			entry = active.ordinary
			if e.metered {
				entry = active.metered
			}
		}
		switch exit {
		case native.Poll:
			s.Budget = 64
			continue
		case native.Move:
			if s.Payload[0] >= uint64(len(fn.Code.Edges)) {
				return 0, s.Fuel, wasm.ErrInvalid
			}
			edge := fn.Code.Edges[s.Payload[0]]
			for i, move := range fn.Code.Moves[edge.Start : edge.Start+edge.Count] {
				if i%256 == 0 {
					if err := e.ctx.Err(); err != nil {
						return 0, s.Fuel, err
					}
				}
				e.values[base+int(move.Dst)] = e.values[base+int(move.Src)]
			}
			continue
		case native.Trap:
			return 0, s.Fuel, &wasm.Trap{Code: wasm.TrapCode(s.Payload[0]), Function: fn.Index, Offset: s.Source}
		case native.FuelExhausted:
			return 0, s.Fuel, &wasm.Error{Kind: wasm.ErrLimit, Offset: s.Source, Detail: "execution fuel"}
		case native.Return, native.Call, native.GrowMemory, native.RuntimeHelper:
			pc := s.Payload[0]
			if pc >= uint64(len(fn.Code.Code)) {
				return 0, s.Fuel, wasm.ErrInvalid
			}
			op := fn.Code.Code[pc].Op
			// Reject mismatched exits instead of interpreting arbitrary code.
			valid := exit == native.Return && op == slots.ControlBase+uint32(ir.Return) ||
				exit == native.Call && (op >= 0x10 && op <= 0x15) || exit == native.GrowMemory && op == 0x40 ||
				exit == native.RuntimeHelper && (op >= 0xe0 && op <= 0xe6 || op>>16 == 0xfe || sharedMemoryBoundary(fn, fn.Code.Code[pc]) || op == 8 || op == 10 || op >= 0x23 && op <= 0x26 || op == 0xd2 || op == 0xd3 || op>>16 == 0xfb || op == slots.ControlBase+uint32(ir.IfCast) || op >= 0xfc0008 && op <= 0xfc0011 || op>>16 == 0xfd)
			if !valid {
				return 0, s.Fuel, wasm.ErrInvalid
			}
			return uint32(pc), s.Fuel, nil
		default:
			return 0, s.Fuel, wasm.ErrInvalid
		}
	}
}

func sharedMemoryBoundary(fn *state.Function, in slots.Instruction) bool {
	op := in.Op
	memoryAccess := op >= 0x28 && op <= 0x40 || op>>16 == 0xfd && (op&0xffff <= 11 || op&0xffff >= 84 && op&0xffff <= 93)
	if !memoryAccess {
		return false
	}
	var scratch ir.Source
	source := fn.Code.ReadSource(in.Source, &scratch).Wasm
	index := source.Memory
	if op == 0x3f || op == 0x40 {
		index = source.A
	}
	return uint64(index) < uint64(len(fn.Instance.Memories)) && fn.Instance.Memories[index].Shared()
}
