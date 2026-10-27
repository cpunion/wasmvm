package execute

import (
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
)

// Backend executes compiled code up to a typed runtime boundary. The returned
// instruction has already consumed fuel; only its explicit runtime effect is
// serviced by the common frame machine. An absent backend selects slot dispatch.
// Resize owns pointer-free frame storage and preserves it across growth.
type Backend interface {
	Resize(length, capacity int) ([]value.Value, error)
	Run(fn *state.Function, base, depth int, pc uint32, fuel uint64) (uint32, uint64, error)
	Close()
}

// Close releases active frames and backend storage. Results remain ordinary
// owned Go values. The execution must not be used after Close.
func (e *Execution) Close() {
	e.Abort(nil)
	e.values = nil
	if e.config.Backend != nil {
		e.config.Backend.Close()
	}
}
