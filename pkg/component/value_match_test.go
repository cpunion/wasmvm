package component

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestValueMatchPlansRecheckResourcesAndBudgets(t *testing.T) {
	in := &Instance{config: DefaultConfig()}
	identity := &canonical.ResourceType{}
	left := &namespace{resources: map[uint64]*resourceType{1: {identity: identity}}}
	right := &namespace{resources: map[uint64]*resourceType{2: {identity: identity}}}
	a, b := &model.Type{Kind: model.Own, ResourceID: 1}, &model.Type{Kind: model.Own, ResourceID: 2}
	for _, binding := range []*resourceType{{identity: identity}, nil, {identity: &canonical.ResourceType{}}, {identity: identity}} {
		right.resources[2] = binding
		ok, err := in.matchValue(a, b, left, right)
		want := binding != nil && binding.identity == identity
		if err != nil || ok != want {
			t.Fatal("stale nominal binding", ok, err)
		}
	}
	in.config.Validate.MaxNodes = 0
	if ok, err := in.matchValue(a, b, left, right); ok || !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("budget ignored", ok, err)
	}
	right.resources[2] = nil
	if ok, err := in.matchValue(a, b, left, right); ok || err != nil {
		t.Fatal("nominal mismatch must precede node limit", ok, err)
	}
	for i := 0; i < 100; i++ {
		v := &model.Type{Kind: model.U8}
		if ok, err := in.matchValue(v, v, left, right); ok || !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("bounded fallback changed budget", ok, err)
		}
	}
	in.config.Validate.MaxNodes = 1
	right.resources[2] = &resourceType{identity: identity}
	if ok, err := in.matchValue(a, b, left, right); !ok || err != nil {
		t.Fatal("bounded fallback lost resource equality", ok, err)
	}
}

func TestValueMatchPlansDrainAfterSourceCleanup(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			in, typ := sourceType(t, backend, "host-source-stream-u8")
			var cleaned, retained bool
			_, err := in.NewSource(nil, typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrType }, Close: func() {
				cleaned, retained = true, in.comparisons.Len() != 0
			}})
			if err != nil {
				t.Fatal(err)
			}
			entered, release := make(chan struct{}), make(chan struct{})
			defer func() {
				select {
				case <-release:
				default:
					close(release)
				}
			}()
			done := make(chan error, 1)
			go func() {
				_, err := in.WithCaller(nil, CallOptions{}, func(*Caller) error {
					v := &model.Type{Kind: model.U8}
					if ok, err := in.matchValue(v, v, nil, nil); err != nil || !ok {
						return vm.ErrType
					}
					close(entered)
					<-release
					return nil
				})
				done <- err
			}()
			select {
			case <-entered:
			case err := <-done:
				t.Fatal("call ended before preparing comparisons", err)
			}
			in.Close()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := in.WaitClosed(ctx); !errors.Is(err, context.Canceled) {
				t.Fatal("active call was not drained", err)
			}
			close(release)
			if err := <-done; err != nil && !errors.Is(err, wasm.ErrClosed) && !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := in.WaitClosed(ctx); err != nil || !cleaned || !retained || in.comparisons.Len() != 0 {
				t.Fatal("comparison roots survived shutdown or retired early", cleaned, retained, err)
			}
		})
	}
}
