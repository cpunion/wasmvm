//go:build linux || darwin

package threads

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasi/preview1"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func engines(t *testing.T, run func(*testing.T, *vm.Engine, vm.CallOptions)) {
	t.Helper()
	for _, tier := range []string{"interpreter", "native", "fuel"} {
		if tier != "interpreter" && !native.Available {
			continue
		}
		t.Run(tier, func(t *testing.T) {
			cfg := vm.DefaultEngineConfig()
			cfg.Profile, cfg.Features = "core3", wasm.Threads
			cfg.Native.MaxCodeBytes = 256 << 20
			if tier != "interpreter" {
				cfg.Backend = "native"
			}
			e, err := vm.NewEngineWithConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			run(t, e, vm.CallOptions{Metered: tier == "fuel", Fuel: 1 << 32})
		})
	}
}

func newGroup(t *testing.T, e *vm.Engine, ctx context.Context, config Config, mode uint32) (*Group, vm.Extern) {
	t.Helper()
	bytes, err := os.ReadFile("testdata/group.wasm")
	if err != nil {
		t.Fatal(err)
	}
	m, err := e.Compile(ctx, bytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Close)
	p, err := preview1.NewProcess(preview1.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	g, err := New(ctx, e, m, p, config)
	if err != nil {
		p.Close()
		m.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { g.Close(); g.WaitClosed(nil) })
	observer, err := e.NewStore(vm.DefaultStoreConfig())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { observer.Close(); observer.WaitClosed(nil) })
	memory, err := observer.ImportMemory(g.memories[[2]string{"env", "memory"}])
	if err != nil {
		t.Fatal(err)
	}
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], mode)
	if err := memory.WriteMemory(0, b[:]); err != nil {
		t.Fatal(err)
	}
	return g, memory
}

func word(t *testing.T, memory vm.Extern, at uint64) uint32 {
	t.Helper()
	var b [4]byte
	if err := memory.ReadMemory(at, b[:]); err != nil {
		t.Fatal(err)
	}
	return binary.LittleEndian.Uint32(b[:])
}

func TestGroupInstancesInitializationAndReturn(t *testing.T) {
	engines(t, func(t *testing.T, e *vm.Engine, options vm.CallOptions) {
		for _, mode := range []uint32{0, 5, 6} {
			cfg := DefaultConfig()
			cfg.Execution = options
			g, memory := newGroup(t, e, nil, cfg, mode)
			if err := g.Run(); err != nil {
				t.Fatal(mode, err)
			}
			if value := word(t, memory, 16); value != 42 {
				t.Fatal("worker result", value)
			}
			if value := word(t, memory, 8); value != 123456 {
				t.Fatal("child reset initialized shared data", value)
			}
			if mode == 6 && word(t, memory, 20) != 20 {
				t.Fatal("monotonic unique TIDs")
			}
			if g.active != 0 || len(g.stores) != 0 {
				t.Fatal("live workers after join")
			}
			if err := g.Run(); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal("closed group", err)
			}
		}
	})
}

func TestGroupTrapExitAndCancel(t *testing.T) {
	engines(t, func(t *testing.T, e *vm.Engine, options vm.CallOptions) {
		for _, mode := range []uint32{1, 2, 3, 9} {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			cfg := DefaultConfig()
			cfg.Execution = options
			g, _ := newGroup(t, e, ctx, cfg, mode)
			err := g.Run()
			cancel()
			switch mode {
			case 1, 9:
				if !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("worker trap", err)
				}
			case 2:
				var exit *preview1.ExitError
				if !errors.As(err, &exit) || exit.Code != 7 {
					t.Fatal("worker process exit", err)
				}
			case 3:
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal("pure worker cancellation", err)
				}
			}
			if err := g.WaitClosed(ctx); err != nil {
				t.Fatal("completed close wins canceled context", err)
			}
		}
	})
}

func TestGroupLimitsCloseAndRunArbitration(t *testing.T) {
	engines(t, func(t *testing.T, e *vm.Engine, options vm.CallOptions) {
		cfg := DefaultConfig()
		cfg.MaxThreads, cfg.Execution = 1, options
		g, memory := newGroup(t, e, nil, cfg, 4)
		if err := g.Run(); err != nil || int32(word(t, memory, 24)) != -int32(preview1.ErrAgain) {
			t.Fatal("thread quota", err)
		}
		cfg.MaxThreads = 4
		g, _ = newGroup(t, e, nil, cfg, 3)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := g.WaitClosed(ctx); !errors.Is(err, context.Canceled) {
			t.Fatal("pending close", err)
		}
		result := make(chan error, 1)
		go func() { result <- g.Run() }()
		for {
			g.mu.Lock()
			running := g.running
			g.mu.Unlock()
			if running {
				break
			}
			time.Sleep(time.Millisecond)
		}
		if err := g.Run(); !errors.Is(err, vm.ErrBusy) {
			t.Fatal("concurrent Run", err)
		}
		var closes sync.WaitGroup
		for i := 0; i < 4; i++ {
			closes.Add(1)
			go func() { defer closes.Done(); g.Close() }()
		}
		closes.Wait()
		if err := <-result; !errors.Is(err, context.Canceled) {
			t.Fatal("manual close reported success", err)
		}
		if err := g.WaitClosed(nil); err != nil {
			t.Fatal(err)
		}
		g, _ = newGroup(t, e, nil, cfg, 0)
		g.Close()
		if err := g.WaitClosed(nil); err != nil {
			t.Fatal("close before Run", err)
		}
	})
}
