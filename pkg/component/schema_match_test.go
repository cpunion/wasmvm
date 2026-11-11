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

func TestSchemaMatchRechecksNominalBindingsAndBudgets(t *testing.T) {
	in := &Instance{config: DefaultConfig()}
	identity := &canonical.ResourceType{}
	left := &namespace{resources: map[uint64]*resourceType{1: {identity: identity}}}
	right := &namespace{resources: map[uint64]*resourceType{2: {identity: identity}}}
	a := &model.Type{Kind: model.Own, ResourceID: 1}
	b := &model.Type{Kind: model.Own, ResourceID: 2}
	if ok, err := in.matchSchema(a, b, left, right); !ok || err != nil || len(in.schemaMatches) != 1 {
		t.Fatal(ok, err, in.schemaMatches)
	}
	for _, binding := range []*resourceType{nil, {identity: &canonical.ResourceType{}}} {
		right.resources[2] = binding
		if ok, err := in.matchSchema(a, b, left, right); ok || err != nil {
			t.Fatal("cached a stale nominal identity", ok, err)
		}
	}
	right.resources[2] = &resourceType{identity: identity}
	if ok, err := in.matchSchema(a, b, left, right); !ok || err != nil {
		t.Fatal(ok, err)
	}
	budget := in.config.Validate.MaxNodes
	in.config.Validate.MaxNodes = 0
	if ok, err := in.matchSchema(a, b, left, right); ok || !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("cached a different budget", ok, err)
	}
	in.config.Validate.MaxNodes = budget
	bad := &model.Type{Kind: model.Borrow, ResourceID: 2}
	if ok, err := in.matchSchema(a, bad, left, right); ok || err != nil || len(in.schemaMatches) != 1 {
		t.Fatal("cached a structural mismatch", ok, err)
	}
	right.resources[2] = nil
	fresh := &model.Type{Kind: model.Own, ResourceID: 2}
	if ok, err := in.matchSchema(a, fresh, left, right); ok || err != nil || len(in.schemaMatches) != 1 {
		t.Fatal("cached a nominal mismatch", ok, err)
	}
	right.resources[2] = &resourceType{identity: identity}
	if ok, err := in.matchSchema(a, fresh, left, right); !ok || err != nil || len(in.schemaMatches) != 2 {
		t.Fatal("failed comparison was retained", ok, err)
	}
	right.resources[1] = &resourceType{identity: identity}
	if ok, err := in.matchSchema(a, a, left, right); !ok || err != nil {
		t.Fatal("same schema pointer", ok, err)
	}
	right.resources[1] = &resourceType{identity: &canonical.ResourceType{}}
	if ok, err := in.matchSchema(a, a, left, right); ok || err != nil {
		t.Fatal("same pointer retained a stale nominal identity", ok, err)
	}
}

func TestSchemaMatchCacheDrainsAfterActiveCallsAndSourceCleanup(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			in, typ := sourceType(t, backend, "host-source-stream-u8")
			var cleaned, retainedDuringSourceCleanup bool
			_, err := in.NewSource(nil, typ, AsyncSource{
				Read: func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrType },
				Close: func() {
					cleaned = true
					retainedDuringSourceCleanup = len(in.schemaMatches) != 0
				},
			})
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
					identity := &canonical.ResourceType{}
					left := &namespace{resources: map[uint64]*resourceType{1: {identity: identity}}}
					right := &namespace{resources: map[uint64]*resourceType{2: {identity: identity}}}
					a := &model.Type{Kind: model.Own, ResourceID: 1}
					b := &model.Type{Kind: model.Own, ResourceID: 2}
					if equal, err := in.matchSchema(a, b, left, right); err != nil || !equal {
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
				t.Fatal("call ended before populating cache", err)
			}
			in.Close()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := in.WaitClosed(ctx); !errors.Is(err, context.Canceled) {
				t.Fatal("active call shutdown", err)
			}
			close(release)
			if err := <-done; err != nil && !errors.Is(err, wasm.ErrClosed) && !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := in.WaitClosed(ctx); err != nil {
				t.Fatal(err)
			}
			if !cleaned || !retainedDuringSourceCleanup || in.schemaMatches != nil || in.schemaNominalPairs != 0 {
				t.Fatal("shutdown retained schema state or source", cleaned, len(in.schemaMatches), in.schemaNominalPairs)
			}
		})
	}
}

func TestSchemaMatchCacheAndNominalStorageStayBounded(t *testing.T) {
	in := &Instance{config: DefaultConfig()}
	for range maxSchemaMatches + 8 {
		a, b := &model.Type{Kind: model.U8}, &model.Type{Kind: model.U8}
		if ok, err := in.matchSchema(a, b, nil, nil); !ok || err != nil {
			t.Fatal(ok, err)
		}
	}
	if len(in.schemaMatches) != maxSchemaMatches {
		t.Fatal(len(in.schemaMatches))
	}
	in = &Instance{config: DefaultConfig()}
	identity := &canonical.ResourceType{}
	left := &namespace{resources: map[uint64]*resourceType{1: {identity: identity}}}
	right := &namespace{resources: map[uint64]*resourceType{2: {identity: identity}}}
	a, b := &model.Type{Kind: model.Own, ResourceID: 1}, &model.Type{Kind: model.Own, ResourceID: 2}
	largeA, largeB := &model.Type{Kind: model.Tuple}, &model.Type{Kind: model.Tuple}
	for range maxSchemaNominalPairs + 1 {
		largeA.Fields = append(largeA.Fields, model.Field{Type: a})
		largeB.Fields = append(largeB.Fields, model.Field{Type: b})
	}
	if ok, err := in.matchSchema(largeA, largeB, left, right); !ok || err != nil || len(in.schemaMatches) != 0 || in.schemaNominalPairs != 0 {
		t.Fatal("cached an oversized nominal replay", ok, err)
	}
	largeA.Fields = largeA.Fields[:maxSchemaNominalPairs]
	largeB.Fields = largeB.Fields[:maxSchemaNominalPairs]
	if ok, err := in.matchSchema(largeA, largeB, left, right); !ok || err != nil || in.schemaNominalPairs != maxSchemaNominalPairs {
		t.Fatal("exact nominal bound", ok, err, in.schemaNominalPairs)
	}
	if ok, err := in.matchSchema(a, b, left, right); !ok || err != nil || len(in.schemaMatches) != 1 {
		t.Fatal("exceeded total nominal storage", ok, err)
	}
}
