package component

import (
	"errors"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHostCallsRetainIndependentTransactionsThroughCollection(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		config := DefaultConfig()
		config.Core.Backend = backend
		engine, err := NewEngine(config)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { engine.Close() })
		module, err := engine.Compile(nil, fixture(t, "ownership"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { module.Close() })
		in, err := module.Instantiate(nil, Imports{}, CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { in.Close() })
		exported, err := in.Function("get")
		if err != nil {
			t.Fatal(err)
		}
		for _, scope := range []*namespace{nil, exported.scope} {
			_, err := in.store.WithCaller(nil, CallOptions{}, func(core *vm.Caller) error {
				f := &Function{instance: in, scope: scope, typ: &validate.Type{Kind: model.Function}}
				first, err := f.hostCall(core, nil)
				if err != nil {
					return err
				}
				second, err := f.hostCall(core, nil)
				if err != nil {
					return err
				}
				if first == second || &first.codec == &second.codec || (first.codec.Resources != nil) != (scope != nil) {
					t.Fatal("host transaction identity or resource policy changed", backend)
				}
				if first.codec.Resources != nil && first.codec.Resources == second.codec.Resources {
					t.Fatal("host calls share resource ownership state")
				}
				first.codec.MaxValues, second.codec.MaxValues = 1, 1
				typ := &model.Type{Kind: model.U32}
				if _, err := first.codec.Clone(typ, Value{Bits: 17}); err != nil {
					return err
				}
				first.account()
				runtime.GC()
				if _, err := first.codec.Clone(typ, Value{Bits: 19}); !errors.Is(err, wasm.ErrLimit) {
					t.Fatal("first call lost its budget", err)
				}
				if err := first.finish(core, true, nil); err != nil {
					return err
				}
				if _, err := second.codec.Clone(typ, Value{Bits: 23}); err != nil {
					return err
				}
				second.account()
				if second.closed {
					t.Fatal("finishing one call closed another")
				}
				return second.finish(core, true, nil)
			})
			if err != nil {
				t.Fatal(backend, err)
			}
		}
	}
}
