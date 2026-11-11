package execute

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
)

// A compiled call veneer only binds a constant table operand and performs the
// ordinary call/return transitions. Keep its real guest frame, including while
// a host call is parked. No scalar instruction interpreter runs in this path.
func (e *Execution) runVeneer(backend Backend) (Status, bool) {
	f := &e.frames[len(e.frames)-1]
	plan := f.function.Code.Veneer
	if plan == nil || len(e.frames) < 2 || f.function.Instance == e.frames[len(e.frames)-2].function.Instance ||
		f.pc != 0 && f.pc != plan.ReturnPC {
		return Running, false
	}
	owner, ok := backend.(interface{ CanRunVeneer(*state.Function) bool })
	if !ok || !owner.CanRunVeneer(f.function) {
		return Running, false
	}
	fuel := plan.ReturnFuel
	if f.pc == 0 {
		fuel = plan.CallFuel
	}
	if e.config.Metered && e.fuel < fuel {
		// The original artifact preserves the exact partial-progress fuel and
		// source location when a limit falls inside the folded prelude.
		return Running, false
	}
	if err := e.poll(); err != nil {
		return e.fail(err), true
	}
	if e.config.Metered {
		e.fuel -= fuel
	}
	e.steps++
	if f.pc == 0 {
		if plan.Index != slots.NoSlot {
			e.values[f.base+int(plan.Index)] = value.Value{Lo: plan.Bits}
		}
		f.pc = plan.CallPC
		return e.runCall(), true
	}
	return e.returnFrame(), true
}
