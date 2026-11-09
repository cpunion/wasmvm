package vm

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// A background invocation needs only a sticky cancellation bit until a channel,
// value or public context is observed. This state belongs to one operation and
// is never rebound. Cancelable/custom parents retain ordinary WithCancel.
type operationContext struct {
	parent   context.Context
	canceled atomic.Bool
	once     sync.Once
	child    atomic.Pointer[operationCancellation]
	storage  operationCancellation
}

type operationCancellation struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *operationContext) standard() context.Context {
	child := c.child.Load()
	if child == nil {
		c.once.Do(func() {
			c.storage.ctx, c.storage.cancel = context.WithCancel(c.parent)
			c.child.Store(&c.storage)
		})
		child = c.child.Load()
	}
	// Publication precedes this check. Racing Cancel either finds the published
	// child or leaves the sticky bit for this path; neither can lose cancellation.
	if c.canceled.Load() {
		child.cancel()
	}
	return child.ctx
}

func (c *operationContext) cancel() {
	c.canceled.Store(true)
	if child := c.child.Load(); child != nil {
		child.cancel()
	}
}
func (c *operationContext) Deadline() (time.Time, bool) { return c.parent.Deadline() }
func (c *operationContext) Done() <-chan struct{}       { return c.standard().Done() }
func (c *operationContext) Value(key any) any           { return c.standard().Value(key) }
func (c *operationContext) Err() error {
	if !c.canceled.Load() {
		return nil
	}
	// Complete channel publication before exposing an error to a waiter.
	if child := c.child.Load(); child != nil {
		child.cancel()
	}
	return context.Canceled
}

func (op *operation) cancelContext() {
	if op.cancelParent != nil {
		op.cancelParent()
	} else {
		op.background.cancel()
	}
}
func (op *operation) publicContext() context.Context {
	if op.cancelParent != nil {
		return op.ctx
	}
	return op.background.standard()
}
