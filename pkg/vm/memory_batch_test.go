package vm

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCallerMemoryBatchRangesOrderingAndLifetime(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		for _, shared := range []bool{false, true} {
			cfg := newEngine().config
			cfg.Features = wasm.Threads
			e, err := NewEngineWithConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(e.Close)
			s, err := e.NewStore(DefaultStoreConfig())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(s.Close)
			l := NewLinker(s)
			var saved *Caller
			var writes []MemoryWrite
			var memoryIndex uint32
			var stop context.CancelFunc
			host, err := s.NewScalarFunction(FunctionType{}, func(c *Caller, _ []uint64) error {
				saved = c
				if stop != nil {
					stop()
				}
				return c.WriteMemoryBatch(memoryIndex, writes)
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := l.Define("h", "write", host); err != nil {
				t.Fatal(err)
			}
			d := &module.Module{Types: []module.FuncType{{}}, Imports: []module.Import{{Module: "h", Name: "write", Kind: 0}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x10}, {Op: 0xb}}}}, Memories: []module.Limits{{Min: 1, Max: 1, HasMax: true, Shared: shared}}, Exports: []module.Export{{Name: "run", Index: 1}, {Name: "memory", Kind: 2}}}
			instance := instantiate(t, l, compileDefinitionWithFeatures(t, e, d, wasm.Threads))
			run, memory := export(t, instance, "run"), export(t, instance, "memory")
			writes = []MemoryWrite{{Offset: 1, Data: []byte{1, 2, 3}}, {Offset: 8, Data: []byte{4}}, {Offset: 2, Data: []byte{9}}, {Offset: 65536}}
			if _, err := run.Call(nil, nil, CallOptions{}); err != nil {
				t.Fatal(shared, err)
			}
			before := make([]byte, 16)
			if err := memory.ReadMemory(0, before); err != nil {
				t.Fatal(err)
			}
			want := []byte{0, 1, 9, 3, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0}
			if !slices.Equal(before, want) {
				t.Fatal("write order", before)
			}
			if err := saved.WriteMemoryBatch(0, nil); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal("retained caller", err)
			}
			for _, bad := range []MemoryWrite{{Offset: 65535, Data: []byte{1, 2}}, {Offset: ^uint64(0), Data: []byte{1}}, {Offset: 65537}} {
				writes = []MemoryWrite{{Offset: 1, Data: []byte{7}}, bad}
				if _, err := run.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("bad range", bad, err)
				}
				var after [16]byte
				if err := memory.ReadMemory(0, after[:]); err != nil || !slices.Equal(before, after[:]) {
					t.Fatal("partial write before bad range", after, err)
				}
			}
			memoryIndex = 1
			if _, err := run.Call(nil, nil, CallOptions{}); !errors.Is(err, ErrReference) {
				t.Fatal("bad memory index", err)
			}
			memoryIndex, writes = 0, nil
			if _, err := run.Call(nil, nil, CallOptions{}); err != nil {
				t.Fatal("empty batch", err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			stop = cancel
			if _, err := run.Call(ctx, nil, CallOptions{}); !errors.Is(err, context.Canceled) {
				t.Fatal("cancellation", err)
			}
			stop = nil
			if _, err := host.Call(nil, nil, CallOptions{}); !errors.Is(err, ErrReference) {
				t.Fatal("caller without instance", err)
			}
		}
	})
}
