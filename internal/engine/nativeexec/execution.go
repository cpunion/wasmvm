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
	ctx     context.Context
	native  *native.Context
	modules map[*slots.Module]*Lease
	maximum uint64
	metered bool
	stop    func() bool
	values  []value.Value
}

func New(ctx context.Context, modules map[*slots.Module]*Lease, maxSlots, maxFrames uint32, metered bool) (*Execution, error) {
	c, err := native.NewContext(0, uint64(maxSlots)*16, uint64(max(1, maxFrames)))
	if err != nil {
		return nil, err
	}
	e := &Execution{ctx: ctx, native: c, modules: modules, maximum: uint64(maxSlots) * 16, metered: metered}
	e.stop = context.AfterFunc(ctx, c.Cancel)
	return e, nil
}

func (e *Execution) Resize(length, capacity int) ([]value.Value, error) {
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

func (e *Execution) Close() { e.stop(); e.values = nil; e.native.Close() }

func (e *Execution) Run(fn *state.Function, base, depth int, pc uint32, fuel uint64) (uint32, uint64, error) {
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
	tier := 0
	s := native.State{Used: uint64(base) * 16, Depth: uint64(depth), Function: uint64(fn.Index), Continuation: uint64(pc), Fuel: fuel, Budget: 64}
	if e.metered {
		tier = 1
		s.Payload[3] = 1
	}
	for {
		if err := e.ctx.Err(); err != nil {
			return 0, s.Fuel, err
		}
		if err := e.native.SetState(s); err != nil {
			return 0, s.Fuel, err
		}
		exit, err := e.native.Enter(module.code, entries[tier])
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
