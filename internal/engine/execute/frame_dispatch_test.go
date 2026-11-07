package execute

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Model an owned arena that invalidates its old view after relocating it.
// Call arguments must be read from the preserved new view after push.
type movingFrameBackend struct {
	arena []value.Value
	moves int
}

func (b *movingFrameBackend) Resize(length, capacity int) ([]value.Value, error) {
	next := make([]value.Value, capacity)
	copy(next, b.arena)
	clear(b.arena)
	b.arena = next
	b.moves++
	return next[:length], nil
}

// This bounded model implements the ordinary operations used by the frame
// tests. It returns an already-charged runtime boundary, like a native backend;
// production code never interprets arbitrary scalar instructions at that exit.
func (b *movingFrameBackend) advance(out *state.NativeFrames, charge func(*state.NativeFrames, *state.Function, uint32) error) error {
	fn := out.Function
	code := fn.Code
	for {
		pc := out.PC
		if charge != nil {
			if err := charge(out, fn, pc); err != nil {
				return err
			}
		}
		in := code.Code[pc]
		values := out.Values[out.Base : out.Base+len(code.Slots)]
		var result value.Value
		switch op := in.Op; op {
		case 0x20:
			result = values[in.Extra]
		case 0x21, 0x22:
			result = values[in.A]
			values[in.Extra] = result
		case 0x41:
			result.Lo = uint64(in.Extra)
		case slots.ControlBase + uint32(ir.Jump), slots.ControlBase + uint32(ir.If):
			index := uint32(0)
			if op == slots.ControlBase+uint32(ir.If) && uint32(values[in.A].Lo) == 0 {
				index = 1
			}
			edge := code.Edges[in.Extra+index]
			for _, move := range code.Moves[edge.Start : edge.Start+edge.Count] {
				values[move.Dst] = values[move.Src]
			}
			out.PC = edge.PC
			continue
		default:
			if op < 0x45 || op > 0xc4 {
				return nil
			}
			var a, c uint64
			if in.A != slots.NoSlot {
				a = values[in.A].Lo
			}
			if in.B != slots.NoSlot {
				c = values[in.B].Lo
			}
			var trap wasm.TrapCode
			result.Lo, trap = numeric.Eval(op, a, c)
			if trap != wasm.NoTrap {
				return &wasm.Trap{Code: trap, Function: fn.Index, Offset: code.Sources[in.Source].Wasm.Offset}
			}
		}
		if in.Dst != slots.NoSlot {
			values[in.Dst] = result
		}
		out.PC++
	}
}

func (b *movingFrameBackend) RunFrames(out *state.NativeFrames) error { return b.advance(out, nil) }
func (b *movingFrameBackend) Interrupt()                              {}
func (b *movingFrameBackend) Close()                                  { b.arena = nil }

func TestFrameDispatchAcrossArenaGrowth(t *testing.T) {
	for _, owned := range []bool{false, true} {
		s := state.New(state.DefaultLimits())
		m := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}}}
		for i := 0; i < 4; i++ {
			fn := module.Function{Body: module.Expr{{Op: 0x20}}}
			if i > 0 {
				n := uint32(64) << (i - 1)
				fn.Locals = []module.Local{{Count: n, Type: wasm.I32}}
				fn.LocalCount = n
			}
			if i < 3 {
				fn.Body = append(fn.Body, module.Instr{Op: 0x10, A: uint32(i + 1)})
			}
			fn.Body = append(fn.Body, module.Instr{Op: 0x41, Bits: uint64(1) << i}, module.Instr{Op: 0x6a}, module.Instr{Op: 0xb})
			m.Functions = append(m.Functions, fn)
		}
		in := instance(t, s, m, nil)
		var backend Backend
		moving := &movingFrameBackend{}
		if owned {
			backend = moving
		}
		e := New(s, nil, Config{Limits: DefaultLimits(), Backend: backend})
		for _, arg := range []uint32{7, 19, 0xffffffff} {
			if err := e.Start(in.Functions[0], []value.Value{{Lo: uint64(arg)}}); err != nil {
				t.Fatal(err)
			}
			if e.Run() != Done || e.Results()[0].Lo != uint64(arg+15) {
				t.Fatal(owned, arg, e.Err(), e.Results())
			}
		}
		if owned && moving.moves < 3 {
			t.Fatal("arena did not relocate across nested calls", moving.moves)
		}
		e.Close()
		s.Close()
	}
}
