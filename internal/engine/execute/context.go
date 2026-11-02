package execute

import "context"

// Runtime helpers receive the same cancellation view as the dispatcher. The
// pinned LLGo atomic.Value.Load allocates even on cancelCtx.Err's nil path.
// Keep this view immutable: a helper may retain its context across rebinding.
type cancellationContext struct {
	context.Context
	done <-chan struct{}
}

func (c *cancellationContext) Done() <-chan struct{} { return c.done }
func (c *cancellationContext) Err() error {
	select {
	case <-c.done:
		return c.Context.Err()
	default:
		return nil
	}
}

func cancellationView(ctx context.Context) (context.Context, <-chan struct{}) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c, ok := ctx.(*cancellationContext); ok {
		return c, c.done
	}
	done := ctx.Done()
	if done == nil {
		return ctx, nil // Retain background and Err-only custom contexts.
	}
	return &cancellationContext{Context: ctx, done: done}, done
}
