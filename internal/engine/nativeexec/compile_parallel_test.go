package nativeexec

import (
	"bytes"
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type panickingCompileContext struct{ context.Context }

func (panickingCompileContext) Err() error { panic("compile-context-panic") }

func TestCompilationOrderAndOptionalSelection(t *testing.T) {
	d := &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}}}
	for i := 0; i < 11; i++ {
		d.Functions = append(d.Functions, module.Function{Body: module.Expr{{Op: 0x41, Bits: uint64(i)}, {Op: 0xb}}})
	}
	m := callSlots(t, d)
	limits := codegen.DefaultLimits()
	config := codegen.Config{Unmetered: true}
	for _, processors := range []int{1, 4} {
		previous := runtime.GOMAXPROCS(processors)
		func() {
			defer runtime.GOMAXPROCS(previous)
			for _, optional := range []bool{false, true} {
				wanted := map[*slots.Function]bool{m.Functions[0]: true, m.Functions[3]: true, m.Functions[8]: true}
				p := newFunctionCompiler(context.Background(), m.Functions, limits, config, optional, wanted)
				defer p.close()
				for _, f := range m.Functions {
					if optional && !wanted[f] {
						continue
					}
					got := p.next(limits.MaxCodeBytes)
					want, err := codegen.CompileFunctionWithConfig(runtime.GOARCH, f, limits, config)
					if err != nil || got.err != nil || got.panicked || !bytes.Equal(got.body.Entry.Code, want.Entry.Code) {
						t.Fatal("compilation reordered or changed a body", processors, optional, err, got.err, got.panicked)
					}
					if p.queued-p.taken > len(p.results) {
						t.Fatal("unbounded outstanding bodies")
					}
				}
				p.close()
				p.close()
			}
		}()
	}
}

func TestCompilationCancellationPanicAndAbandonment(t *testing.T) {
	previous := runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(previous)
	for _, processors := range []int{1, 4} {
		runtime.GOMAXPROCS(processors)
		for _, canceled := range []bool{false, true} {
			ctx, cancel := context.WithCancel(context.Background())
			if canceled {
				cancel()
			}
			// A context panic during compilation must reach the publishing
			// caller, without killing the process or leaving other workers
			// blocked on abandoned results.
			p := newFunctionCompiler(ctx, make([]*slots.Function, 11), codegen.DefaultLimits(), codegen.Config{}, false, nil)
			workerContext := p.ctx
			if !canceled {
				// Jobs have not been submitted, so no worker can read ctx yet.
				p.ctx = panickingCompileContext{p.ctx}
			}
			result := p.next(1 << 20)
			if canceled {
				if !errors.Is(result.err, context.Canceled) || result.panicked {
					t.Fatal("lost cancellation", result)
				}
			} else if !result.panicked || result.panicValue != "compile-context-panic" {
				t.Fatal("lost compiler panic", result)
			}
			p.close()
			p.close()
			cancel()
			if p.cancel != nil && !errors.Is(workerContext.Err(), context.Canceled) {
				t.Fatal("workers outlived their compilation")
			}
		}
	}
	// Closing before the first request also joins all idle workers.
	p := newFunctionCompiler(context.Background(), make([]*slots.Function, 11), codegen.DefaultLimits(), codegen.Config{}, false, nil)
	p.close()
	p.close()
}
