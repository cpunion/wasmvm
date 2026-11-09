package vm

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLoopCounterPrivateSharedCancellationAndReuse(t *testing.T) {
	for _, kind := range []string{"private", "shared"} {
		data, err := os.ReadFile("../../benchmarks/modern/threads-" + kind + "-memory.wasm")
		if err != nil {
			t.Fatal(err)
		}
		withBackends(t, func(t *testing.T, newEngine func() *Engine) {
			t.Run(kind, func(t *testing.T) {
				cfg := newEngine().config
				cfg.Profile, cfg.Features = "core3", wasm.Threads
				engine, err := NewEngineWithConfig(cfg)
				if err != nil {
					t.Fatal(err)
				}
				defer engine.Close()
				compiled, err := engine.Compile(nil, data)
				if err != nil {
					t.Fatal(err)
				}
				defer compiled.Close()
				store, err := engine.NewStore(DefaultStoreConfig())
				if err != nil {
					t.Fatal(err)
				}
				defer store.Close()
				instance := instantiate(t, NewLinker(store), compiled)
				fn := export(t, instance, "run")
				for _, seed := range []uint32{0, 17, 0xffffffff} {
					for _, n := range []uint32{0, 1, 2, 31, 1000} {
						r, err := fn.Call(nil, []Value{I32(seed), I32(n)}, CallOptions{})
						if err != nil {
							t.Fatal(err)
						}
						count := n
						if count == 0 {
							count = 1
						}
						if len(r.Values) != 1 {
							t.Fatal(r)
						}
						got, hi, err := r.Values[0].Bits()
						if err != nil || hi != 0 || got != uint64(seed+count) {
							t.Fatal(seed, n, r, err)
						}
					}
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				_, err = fn.Call(ctx, []Value{I32(17), I32(0xffffffff)}, CallOptions{})
				cancel()
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal("loop cancellation", err)
				}
				r, err := fn.Call(nil, []Value{I32(2), I32(3)}, CallOptions{})
				if err != nil || len(r.Values) != 1 {
					t.Fatal(r, err)
				}
				got, _, err := r.Values[0].Bits()
				if err != nil || got != 5 {
					t.Fatal("post-cancellation state", r, err)
				}
			})
		})
	}
}
