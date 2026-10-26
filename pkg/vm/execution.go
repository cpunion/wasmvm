package vm

import (
	"context"
	"sync"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Execution binds a function and per-run policy. Run may be repeated after
// completion but cannot overlap another Run on this object or its Store.
// Close cancels this execution without closing the Store. Returned reference
// Values are owned by the caller and must be released independently.
type Execution struct {
	mu       sync.Mutex
	function Extern
	options  CallOptions
	closed   bool
	cancel   context.CancelFunc
}

func NewExecution(function Extern, options CallOptions) (*Execution, error) {
	if _, err := function.FunctionType(); err != nil {
		return nil, err
	}
	return &Execution{function: function, options: options}, nil
}
func (e *Execution) Run(ctx context.Context, args []Value) (CallResult, error) {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return CallResult{}, wasm.ErrClosed
	}
	if e.cancel != nil {
		e.mu.Unlock()
		return CallResult{}, ErrBusy
	}
	ctx, cancel := context.WithCancel(contextOrBackground(ctx))
	e.cancel = cancel
	e.mu.Unlock()
	defer func() { cancel(); e.mu.Lock(); e.cancel = nil; e.mu.Unlock() }()
	return e.function.Call(ctx, args, e.options)
}
func (e *Execution) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	if e.cancel != nil {
		e.cancel()
	}
}
