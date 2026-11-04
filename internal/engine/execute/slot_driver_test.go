package execute

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type chargedBoundaryBackend struct {
	batchFrameBackend
	entries uint64
}

func (b *chargedBoundaryBackend) RunFrames(out *state.NativeFrames) error {
	b.entries++
	return b.batchFrameBackend.RunFrames(out)
}

// A native boundary is already charged. Even constants, coalesced local reads
// and backedges must return to that driver before another instruction executes.
func TestNativeBoundaryDriverPreservesExactFuelAndLoopContinuations(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	in := instance(t, s, &module.Module{
		Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}},
		Functions: []module.Function{{
			Locals: []module.Local{{Count: 1, Type: wasm.I32}}, LocalCount: 1,
			Body: module.Expr{
				{Op: 0x41, Bits: 7}, {Op: 0x21, A: 1},
				{Op: 2, Block: -64}, {Op: 3, Block: -64},
				{Op: 0x20}, {Op: 0x45}, {Op: 0xd, A: 1},
				{Op: 0x20, A: 1}, {Op: 0x41, Bits: 3}, {Op: 0x6a}, {Op: 0x21, A: 1},
				{Op: 0x20}, {Op: 0x41, Bits: 1}, {Op: 0x6b}, {Op: 0x21},
				{Op: 0xc}, {Op: 0xb}, {Op: 0xb}, {Op: 0x20, A: 1}, {Op: 0xb},
			},
		}},
	}, nil)
	for _, arg := range []uint64{0, 4} {
		for fuel := uint64(0); fuel <= 80; fuel++ {
			var wantStatus Status
			var wantFuel uint64
			for _, native := range []bool{false, true} {
				model := &chargedBoundaryBackend{batchFrameBackend: batchFrameBackend{store: s, metered: true}}
				config := Config{Limits: DefaultLimits(), Metered: true, Fuel: fuel}
				if native {
					config.Backend = model
				}
				e := New(s, nil, config)
				if err := e.Start(in.Functions[0], []value.Value{{Lo: arg}}); err != nil {
					t.Fatal(err)
				}
				status := e.Run()
				if native {
					if status != wantStatus || e.Fuel() != wantFuel {
						t.Fatal("changed charged boundary", arg, fuel, status, e.Fuel(), wantStatus, wantFuel)
					}
					entries := e.steps
					if status == Failed {
						entries++ // The backend rejects the next charge before slot dispatch.
					}
					if model.entries != entries {
						t.Fatal("slot dispatch crossed a native boundary", arg, fuel, model.entries, entries)
					}
				} else {
					wantStatus, wantFuel = status, e.Fuel()
				}
				if status == Done {
					if len(e.Results()) != 1 || e.Results()[0].Lo != 7+3*arg {
						t.Fatal("lost loop continuation", native, arg, fuel, e.Results())
					}
				} else if status != Failed || !errors.Is(e.Err(), wasm.ErrLimit) {
					t.Fatal(native, arg, fuel, status, e.Err())
				}
				e.Close()
			}
		}
	}
}
