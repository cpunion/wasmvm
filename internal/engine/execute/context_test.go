package execute

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/engine/state"
)

func TestCancellationFlagDoesNotSurviveBinding(t *testing.T) {
	for _, quiesce := range []bool{false, true} {
		s := state.New(state.DefaultLimits())
		e := New(s, context.Background(), Config{Limits: DefaultLimits()})
		var previous atomic.Bool
		e.SetCancellationFlag(&previous)
		if quiesce {
			e.Quiesce()
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := e.Rebind(ctx, DefaultLimits(), false, 0); err != nil {
			t.Fatal(err)
		}
		if !errors.Is(e.poll(), context.Canceled) {
			t.Fatal("previous cancellation flag hid the new context's error", quiesce)
		}
		e.Close()
		s.Close()
	}
}

func TestCancellationFlagRetainsContextError(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	e := New(s, ctx, Config{Limits: DefaultLimits()})
	defer e.Close()
	var flag atomic.Bool
	e.SetCancellationFlag(&flag)
	if err := e.poll(); err != nil {
		t.Fatal(err)
	}
	cause := errors.New("owner canceled")
	flag.Store(true)
	cancel(cause)
	if !errors.Is(e.poll(), context.Canceled) || context.Cause(e.context) != cause {
		t.Fatal("cancellation signal lost the context error or cause")
	}
}

func TestContextMetadataAndRetainedBinding(t *testing.T) {
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
	if err := e.Rebind(ctx, DefaultLimits(), false, 0); err != nil {
		t.Fatal(err)
	}
	retained := e.context
	if got, ok := retained.Deadline(); !ok || !got.Equal(deadline) || retained.Value(key{}) != "first" || retained.Done() != ctx.Done() || retained.Err() != nil {
		t.Fatal("lost context metadata", got, ok)
	}
	if retained != ctx {
		t.Fatal("replaced the invocation context")
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
