package execute

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
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
func (b *movingFrameBackend) Run(_ *state.Function, _, _ int, pc uint32, fuel uint64) (uint32, uint64, error) {
	return pc, fuel, nil
}
func (b *movingFrameBackend) Close() { b.arena = nil }

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
