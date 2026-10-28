package nativeexec

import (
	"context"
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
	reusable        bool
	values          []value.Value
	runtimeData     *native.Memory
	runtimeWords    []uint64
	runtimeInstance *state.Instance
	runtimeFunction *slots.Function
	runtimeGlobals  []globalBinding
}

func New(ctx context.Context, modules map[*slots.Module]*Lease, maxSlots, maxFrames uint32, metered bool) (*Execution, error) {
	c, err := native.NewContext(0, uint64(maxSlots)*16, uint64(max(1, maxFrames)))
	if err != nil {
		return nil, err
	}
	e := &Execution{native: c, modules: modules, maximum: uint64(maxSlots) * 16, maxFrames: uint64(maxFrames), reusable: true}
	if err := e.Begin(ctx, metered); err != nil {
		e.Close()
		return nil, err
	}
	return e, nil
}

// Begin binds an idle context to a new invocation. A context whose cancellation
// callback started is never reused: Suspend returns false and the owner closes it.
func (e *Execution) Begin(ctx context.Context, metered bool) error {
	if e.native == nil || e.stop != nil || !e.reusable {
		return wasm.ErrInvalid
	}
	if ctx == nil {
		ctx = context.Background()
	}
	e.ctx, e.metered = ctx, metered
	e.reusable = false
	e.stop = context.AfterFunc(ctx, e.native.Cancel)
	return nil
}

// Suspend removes the former cancellation writer and releases native-visible
// instance memory. Only a callback stopped before it started permits reuse;
// the native context's sticky control word is never cleared.
func (e *Execution) Suspend() bool {
	if e.native == nil {
		return false
	}
	if e.stop == nil {
		return e.reusable
	}
	stopped := e.stop()
	e.stop, e.ctx, e.runtimeInstance = nil, nil, nil
	e.resetRuntimeGlobals()
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
}

func (e *Execution) Run(fn *state.Function, base, depth int, pc uint32, fuel uint64) (uint32, uint64, error) {
	if e.native == nil {
		return 0, fuel, native.ErrClosed
	}
	if e.stop == nil {
		return 0, fuel, wasm.ErrInvalid
	}
	module := e.modules[fn.Instance.Compiled]
	if module == nil {
		return 0, fuel, wasm.ErrUnsupported
	}
	entries, ok := module.entries[fn.Code]
	if !ok {
		return 0, fuel, wasm.ErrInvalid
	}
	var memory *native.Memory
	if len(fn.Instance.Memories) != 0 {
		memory = fn.Instance.Memories[0].Native()
		if memory == nil {
			return 0, fuel, wasm.ErrInvalid
		}
	}
	if err := e.native.BindMemory(memory); err != nil {
		return 0, fuel, err
	}
	entry := entries.ordinary
	needsData := len(entries.globals) != 0
	s := native.State{Used: uint64(base) * 16, Depth: uint64(depth), Function: uint64(fn.Index), Continuation: uint64(pc), Fuel: fuel, Budget: 64}
	if e.metered {
		entry = entries.metered
		s.Payload[3] = 1
	} else if entries.hasExpansion && uint64(depth) < e.maxFrames &&
		uint64(base)+uint64(len(fn.Code.Slots))+uint64(entries.inlineSlots) <= e.maximum/16 {
		entry = entries.expanded
		needsData = needsData || entries.usesRuntimeData
	}
	if needsData {
		var err error
		s.Payload[2], err = e.bindRuntimeGlobals(fn.Instance, fn.Code, entries.globals)
		if err != nil {
			return 0, fuel, err
		}
	}
	for {
		if err := e.ctx.Err(); err != nil {
			return 0, s.Fuel, err
		}
		if err := e.native.SetState(s); err != nil {
			return 0, s.Fuel, err
		}
		exit, err := e.native.Enter(module.code, entry)
		if len(entries.globals) != 0 {
			e.flushRuntimeGlobals()
		}
		if err != nil {
			return 0, s.Fuel, err
		}
		s, err = e.native.State()
		if err != nil {
			return 0, fuel, err
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
				exit == native.Call && (op == 0x10 || op == 0x11) || exit == native.GrowMemory && op == 0x40 ||
				exit == native.RuntimeHelper && (op >= 0x23 && op <= 0x26 || op == 0xd2 || op >= 0xfc0008 && op <= 0xfc0011 || op>>16 == 0xfd)
			if !valid {
				return 0, s.Fuel, wasm.ErrInvalid
			}
			return uint32(pc), s.Fuel, nil
		default:
			return 0, s.Fuel, wasm.ErrInvalid
		}
	}
}
