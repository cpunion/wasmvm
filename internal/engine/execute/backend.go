package execute

import (
	"context"
	"errors"

	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Quiesce releases guest frames and the invocation context, retaining bounded
// frame/result capacity for a Store-owned idle workspace. Results stay valid
// until the next Start; drivers must encode them before starting another call.
func (e *Execution) Quiesce() {
	e.Abort(nil)
	e.context = context.Background()
	e.status = Done
}

// Interrupt requests a concurrent native exit after the invocation context has
// been canceled. The owner keeps this frame registered until delivery drains;
// its backend cannot be closed or rebound concurrently. Interpreters poll ctx.
func (e *Execution) Interrupt() {
	if e.config.Backend != nil {
		e.config.Backend.Interrupt()
	}
}

// SetLimits redistributes a session's logical frame/slot budget between parked
// executions. The driver calls it only while this execution is not running.
func (e *Execution) SetLimits(limits Limits) error {
	used := e.Footprint()
	if used.MaxFrames > limits.MaxFrames || used.MaxSlots > limits.MaxSlots {
		return wasm.ErrLimit
	}
	e.config.Limits = limits
	return nil
}

// Rebind changes per-invocation policy only after all frames have left. Native
// storage remains owned by the same backend; swapping it requires a new driver.
func (e *Execution) Rebind(ctx context.Context, limits Limits, metered bool, fuel uint64) error {
	if e.status == Running || e.status == HostRequest {
		return errors.New("execution is not an idle workspace")
	}
	e.context = contextOrBackground(ctx)
	e.config = Config{Limits: limits, Metered: metered, Fuel: fuel, Backend: e.config.Backend}
	e.fuel, e.steps = fuel, 0
	e.status, e.err = Done, nil
	return nil
}

// Backend executes compiled code through the shared explicit-frame packet.
// Its returned runtime boundary has already consumed fuel; the common machine owns
// runtime effects and validates the complete frame/arena transfer. Interrupt
// requests bounded cancellation; frame registration owns delivery lifetime.
// Ordinary scalar work must execute in the backend, not return for portable
// fallback. An absent backend selects slot dispatch. Resize owns pointer-free storage.
type Backend interface {
	Resize(length, capacity int) ([]value.Value, error)
	RunFrames(*state.NativeFrames) error
	Interrupt()
	Close()
}

// Close releases active frames and backend storage. Results remain ordinary
// owned Go values. The execution must not be used after Close.
func (e *Execution) Close() {
	e.Abort(nil)
	e.values = nil
	e.tailArgs = nil
	e.gcArgs = nil
	e.hostArgs = nil
	if e.config.Backend != nil {
		e.config.Backend.Close()
	}
}
