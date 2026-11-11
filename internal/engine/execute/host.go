package execute

import (
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
)

// One request belongs to each suspended execution. Public callback arguments
// are encoded separately, and reentry owns a different execution. Large valid
// tuples remain transient; idle request storage retains at most 1 KiB.
const maxCachedHostArgs = 64

func (e *Execution) prepareHost(fn *state.Function, caller *state.Instance, count int) []value.Value {
	if count > cap(e.hostArgs) {
		e.hostArgs = make([]value.Value, count)
	} else {
		e.hostArgs = e.hostArgs[:count]
	}
	e.hostRequest = HostCall{Function: fn, Caller: caller, Args: e.hostArgs}
	e.pending = &e.hostRequest
	return e.hostArgs
}

func (e *Execution) clearHost() {
	e.pending = nil
	e.hostRequest = HostCall{}
	clear(e.hostArgs)
	if cap(e.hostArgs) > maxCachedHostArgs {
		e.hostArgs = nil
	} else {
		e.hostArgs = e.hostArgs[:0]
	}
}
