package execute

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLocalWordUpdateRespectsPollFuelAndIndependentAddresses(t *testing.T) {
	m := simple(module.Expr{
		{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x28, Offset: 701},
		{Op: 0x41, Bits: 3}, {Op: 0x6a}, {Op: 0x36, Offset: 702},
		{Op: 0x20}, {Op: 0x28}, {Op: 0xb},
	}, wasm.I32)
	m.Types[0].Params = []wasm.ValueType{wasm.I32, wasm.I32}
	m.Memories = []module.Limits{{Min: 1, Max: 1, HasMax: true, Shared: true}}
	s := state.New(state.DefaultLimits())
	defer s.Close()
	in := instanceWithFeatures(t, s, m, nil, wasm.Core3Features|wasm.Threads)
	fn, _ := s.Function(in.Functions[0])
	if fn.Code.Code[2].Op != 0x28 || fn.Code.Code[2].Count != slots.LocalI32Update {
		t.Fatal("local memory update hint is absent", fn.Code.Code)
	}
	for _, q := range []struct {
		name          string
		dest, src     uint64
		initial, fuel uint64
		cancel        bool
		steps         uint64
		want          uint64
		failed        error
		trapOffset    uint64
	}{
		{name: "same unaligned word", dest: 1, src: 1, want: 16},
		{name: "distinct words", dest: 1, src: 8, want: 24},
		{name: "poll before store", dest: 1, src: 1, initial: 58, cancel: true, steps: 64, want: 13, failed: context.Canceled},
		{name: "poll after group", dest: 1, src: 1, initial: 57, cancel: true, steps: 64, want: 16, failed: context.Canceled},
		{name: "fuel before store", dest: 1, src: 1, fuel: 5, steps: 5, want: 13, failed: wasm.ErrLimit},
		{name: "source trap", dest: 65536, src: 65536, steps: 3, want: 13, trapOffset: 701},
		{name: "destination trap", dest: 65536, src: 1, steps: 6, want: 13, trapOffset: 702},
	} {
		t.Run(q.name, func(t *testing.T) {
			mem := in.Memories[0]
			if mem.Shareable() {
				t.Fatal("fixture memory is not Store-local")
			}
			if mem.Store(0x36, 1, 0, 0, value.Value{Lo: 13}) != wasm.NoTrap ||
				mem.Store(0x36, 8, 0, 0, value.Value{Lo: 21}) != wasm.NoTrap {
				t.Fatal("initialize guest words")
			}
			var ctx context.Context
			if q.cancel {
				ctx = &cancelAfter{at: 2}
			}
			e := New(s, ctx, Config{Limits: DefaultLimits(), Metered: q.fuel != 0, Fuel: q.fuel})
			defer e.Close()
			if err := e.Start(in.Functions[0], []value.Value{{Lo: q.dest}, {Lo: q.src}}); err != nil {
				t.Fatal(err)
			}
			e.steps = q.initial
			status := e.Run()
			switch {
			case q.trapOffset != 0:
				var trap *wasm.Trap
				if status != Failed || !errors.As(e.Err(), &trap) || trap.Code != wasm.TrapMemoryBounds ||
					trap.Offset != q.trapOffset || e.steps != q.steps {
					t.Fatal("trap or charged steps changed", status, e.Err(), e.steps)
				}
			case q.failed != nil:
				if status != Failed || !errors.Is(e.Err(), q.failed) || e.steps != q.steps {
					t.Fatal("boundary moved", status, e.Err(), e.steps)
				}
			default:
				if status != Done || e.Results()[0].Lo != q.want {
					t.Fatal("return changed", status, e.Results(), e.Err())
				}
			}
			if old, failure := mem.Load(0x28, 1, 0, 0, value.Value{}); failure != wasm.NoTrap || old.Lo != q.want {
				t.Fatal("destination word", old, failure, q.want)
			}
		})
	}
}
