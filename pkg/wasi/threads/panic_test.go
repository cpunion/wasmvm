//go:build linux || darwin

package threads

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasi/preview1"
)

type panicInput struct{ preview1.Stream }

func (panicInput) Read(context.Context, []byte) (int, error) { panic("host-read") }

func TestGroupHostPanicsJoinInitializationAndWorkers(t *testing.T) {
	engines(t, func(t *testing.T, e *vm.Engine, options vm.CallOptions) {
		for _, scenario := range []string{"root-start", "child-start", "root-read", "worker-read"} {
			path := "testdata/host-panic-start.wasm"
			pc := preview1.DefaultConfig()
			if scenario == "root-read" || scenario == "worker-read" {
				pc.Stdin = panicInput{preview1.BytesInput(nil)}
				name := "wasi_threads_exit_nonmain_wasi_read.wasm"
				if scenario == "worker-read" {
					name = "wasi_threads_exit_main_wasi_read.wasm"
				}
				path = "../../../artifacts/wasi-threads/official/" + name
			} else {
				calls := 0
				pc.Clock = preview1.Clock{
					Now: func(uint32) (uint64, error) {
						calls++
						if scenario == "root-start" || calls == 2 {
							panic("host-init")
						}
						return 0, nil
					},
					Resolution: func(uint32) (uint64, error) { return 1, nil },
					Wait:       func(context.Context, uint32, uint64) error { return nil },
				}
			}
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			m, err := e.Compile(nil, b)
			if err != nil {
				t.Fatal(err)
			}
			p, err := preview1.NewProcess(pc)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			cfg := DefaultConfig()
			cfg.Execution = options
			g, err := New(ctx, e, m, p, cfg)
			if err != nil {
				t.Fatal(err)
			}
			err = g.Run()
			var panicErr *HostPanicError
			if !errors.As(err, &panicErr) || panicErr.Value == nil || panicErr.Error() == "" {
				t.Fatal(scenario, err)
			}
			if err := g.WaitClosed(ctx); err != nil {
				t.Fatal("panic reclamation", scenario, err)
			}
			if g.active != 0 || len(g.stores) != 0 {
				t.Fatal("panic leaked a reserved/published Store")
			}
			cancel()
			m.Close()
		}
	})
}
