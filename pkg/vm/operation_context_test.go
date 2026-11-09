package vm

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestOperationContextLazyCancellationAndStandardPublication(t *testing.T) {
	for _, parent := range []context.Context{context.Background(), context.TODO()} {
		for _, early := range []bool{false, true} {
			c := &operationContext{parent: parent}
			if c.Err() != nil || c.child.Load() != nil {
				t.Fatal("eager cancellation state")
			}
			if deadline, ok := c.Deadline(); ok || !deadline.IsZero() {
				t.Fatal("background deadline")
			}
			if early {
				c.cancel()
				if c.Err() != context.Canceled || c.child.Load() != nil {
					t.Fatal("cancel allocated an unobserved child")
				}
			}
			standard := c.standard()
			if c.standard() != standard || c.Done() != standard.Done() || c.Value(struct{}{}) != nil {
				t.Fatal("unstable public context")
			}
			if !early {
				if standard.Err() != nil || context.Cause(c) != nil {
					t.Fatal("premature cancellation")
				}
				c.cancel()
			}
			c.cancel()
			select {
			case <-standard.Done():
			default:
				t.Fatal("unpublished cancellation")
			}
			if c.Err() != context.Canceled || standard.Err() != context.Canceled || context.Cause(c) != context.Canceled || context.Cause(standard) != context.Canceled {
				t.Fatal("wrong cancellation cause")
			}
			child, stop := context.WithCancel(c)
			if child.Err() != context.Canceled {
				t.Fatal("late child revived")
			}
			stop()
		}
	}
}

func TestOperationContextConcurrentPublicationAndCancellation(t *testing.T) {
	for range 64 {
		c := &operationContext{parent: context.Background()}
		start := make(chan struct{})
		contexts := make(chan context.Context, 8)
		var workers sync.WaitGroup
		for range 8 {
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				contexts <- c.standard()
				<-c.Done()
				if c.Err() != context.Canceled {
					t.Error("lost concurrent cancellation")
				}
			}()
		}
		close(start)
		c.cancel()
		workers.Wait()
		close(contexts)
		var first context.Context
		for ctx := range contexts {
			if first == nil {
				first = ctx
			}
			if ctx != first || ctx.Err() != context.Canceled || context.Cause(ctx) != context.Canceled {
				t.Fatal("different or live published context")
			}
		}
	}
	c := &operationContext{parent: context.Background()}
	called := make(chan struct{})
	stop := context.AfterFunc(c, func() { close(called) })
	c.cancel()
	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("AfterFunc did not observe cancellation")
	}
	if stop() {
		t.Fatal("completed AfterFunc stopped")
	}
}

type nonComparableOperationParent struct {
	context.Context
	cells []int
}

func TestOperationContextParentSelectionAndUnobservedRetirement(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, store, _ := setupWithEngine(t, newEngine)
		for _, parent := range []context.Context{nil, context.Background(), context.TODO()} {
			op, err := store.begin(parent, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if op.cancelParent != nil || op.ctx != &op.background || op.background.child.Load() != nil {
				t.Fatal("standard background was not lazy")
			}
			op.finish()
			if op.ctx.Err() != context.Canceled || op.background.child.Load() != nil {
				t.Fatal("unobserved operation allocated or survived")
			}
			late := op.publicContext()
			if late.Err() != context.Canceled || context.Cause(late) != context.Canceled {
				t.Fatal("late public context revived")
			}
		}
		key := struct{}{}
		parent := context.WithValue(context.Background(), key, "retained")
		for _, ctx := range []context.Context{parent, nonComparableOperationParent{Context: parent, cells: []int{1}}} {
			op, err := store.begin(ctx, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			retained := op.publicContext()
			if op.cancelParent == nil || retained.Value(key) != "retained" {
				t.Fatal("custom parent changed")
			}
			op.finish()
			if !errors.Is(retained.Err(), context.Canceled) || ctx.Err() != nil {
				t.Fatal("custom parent lifetime")
			}
		}
	})
}
