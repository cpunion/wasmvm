package execute

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPortableComparisonEdgesPreserveCancellationAndFuel(t *testing.T) {
	body := module.Expr{
		{Op: 0x41}, {Op: 0x24},
		{Op: 3, Block: -64},
		{Op: 0x23}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0x24},
		{Op: 0x23}, {Op: 0x20}, {Op: 0x49}, {Op: 0xd},
		{Op: 0xb}, {Op: 0x23}, {Op: 0xb},
	}
	m := simple(body, wasm.I32)
	m.Types[0].Params = []wasm.ValueType{wasm.I32}
	m.Globals = []module.Global{{
		Type: module.GlobalType{Value: wasm.I32, Mutable: true},
		Init: module.Expr{{Op: 0x41}, {Op: 0xb}},
	}}
	store := state.New(state.DefaultLimits())
	defer store.Close()
	in := instance(t, store, m, nil)
	fn, _ := store.Function(in.Functions[0])
	var compare *slots.Instruction
	for i := range fn.Code.Code {
		if fn.Code.Code[i].Op == 0x49 {
			compare = &fn.Code.Code[i]
			break
		}
	}
	if compare == nil || compare.Count != 1 {
		t.Fatal("expected an adjacent comparison and branch")
	}
	type outcome struct {
		status                      Status
		steps, fuel, global, result uint64
		err                         string
	}
	run := func(hint uint32, input, steps uint64, metered bool, fuel uint64, ctx context.Context) outcome {
		t.Helper()
		compare.Count = hint
		e := New(store, ctx, Config{Limits: DefaultLimits(), Metered: metered, Fuel: fuel})
		defer e.Close()
		if err := e.Start(in.Functions[0], []value.Value{{Lo: input}}); err != nil {
			t.Fatal(err)
		}
		e.steps = steps
		status := e.Run()
		got := outcome{status: status, steps: e.steps, fuel: e.Fuel(),
			global: in.Globals[0].Value.Lo, err: fmt.Sprint(e.Err())}
		if status == Done {
			got.result = e.Results()[0].Lo
		}
		return got
	}
	for _, input := range []uint64{0, 5} {
		for _, steps := range []uint64{0, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, math.MaxUint64 - 4} {
			for _, cancelsAt := range []int{2, 4, 100} {
				want := run(0, input, steps, false, 0, &cancelAfter{at: cancelsAt})
				if got := run(1, input, steps, false, 0, &cancelAfter{at: cancelsAt}); got != want {
					t.Fatalf("input %d, step %d, cancel %d: got %+v, want %+v", input, steps, cancelsAt, got, want)
				}
			}
		}
		for fuel := uint64(0); fuel <= 48; fuel++ {
			want := run(0, input, 0, true, fuel, context.Background())
			if got := run(1, input, 0, true, fuel, context.Background()); got != want {
				t.Fatalf("input %d, fuel %d: got %+v, want %+v", input, fuel, got, want)
			}
		}
	}
}
