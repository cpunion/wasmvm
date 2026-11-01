package vm

import (
	"os"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestGCNativeAllocationBatchesAndForcedCollection(t *testing.T) {
	input, err := os.ReadFile("testdata/gc.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		base := newEngine()
		cfg := base.config
		cfg.Features = wasm.GC
		e, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		m, err := e.Compile(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		for _, forced := range []bool{false, true} {
			cfg := DefaultStoreConfig()
			cfg.CollectEveryAllocation = forced
			cfg.Limits.MaxHeapObjects, cfg.Limits.MaxHeapBytes = 70, 70*128
			s, err := e.NewStore(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			in := instantiate(t, NewLinker(s), m)
			for _, count := range []uint32{1, 63, 64, 65, 1000} {
				out, err := export(t, in, "stress").Call(nil, []Value{I32(count)}, CallOptions{})
				if err != nil || len(out.Values) != 1 || bits(t, out.Values[0]) != 1 {
					t.Fatal("batch result", count, forced, out, err)
				}
				if err := s.Collect(); err != nil {
					t.Fatal(err)
				}
			}
			batches, objects := s.state.NativeAllocationStats()
			if cfg := e.config; cfg.Backend == "native" && !forced {
				if objects < 1100 || batches == 0 || batches >= objects/2 {
					t.Fatal("allocations still exit individually", batches, objects)
				}
			} else if batches != 0 || objects != 0 {
				t.Fatal("batching bypassed policy", batches, objects, forced)
			}
		}
	})
}
