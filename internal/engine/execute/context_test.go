package execute

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/engine/state"
)

func TestCancellationViewMetadataAndRetainedBinding(t *testing.T) {
	type key struct{}
	deadline := time.Now().Add(time.Hour)
	parent, stop := context.WithDeadline(context.WithValue(context.Background(), key{}, "first"), deadline)
	defer stop()
	ctx, cancel := context.WithCancelCause(parent)
	defer cancel(nil)
	s := state.New(state.DefaultLimits())
	defer s.Close()
	e := New(s, ctx, Config{Limits: DefaultLimits()})
	defer e.Close()
	retained := e.context
	if got, ok := retained.Deadline(); !ok || !got.Equal(deadline) || retained.Value(key{}) != "first" || retained.Done() != ctx.Done() || retained.Err() != nil {
		t.Fatal("lost context metadata", got, ok)
	}
	if view, done := cancellationView(retained); view != retained || done != retained.Done() {
		t.Fatal("duplicated an immutable cancellation view")
	}
	e.Quiesce()
	second, cancelSecond := context.WithCancel(context.WithValue(context.Background(), key{}, "second"))
	defer cancelSecond()
	if err := e.Rebind(second, DefaultLimits(), false, 0); err != nil {
		t.Fatal(err)
	}
	cause := errors.New("first invocation stopped")
	cancel(cause)
	if !errors.Is(retained.Err(), context.Canceled) || context.Cause(retained) != cause || retained.Value(key{}) != "first" || e.context.Err() != nil || e.context.Value(key{}) != "second" {
		t.Fatal("rebinding mutated a retained context")
	}
	cancelSecond()
	if !errors.Is(e.context.Err(), context.Canceled) || !errors.Is(context.Cause(e.context), context.Canceled) {
		t.Fatal("lost rebound cancellation")
	}
}
