package execute

import (
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
)

// Compiled frame transitions can bind a call veneer's constant table operand,
// or return an already suspended caller without another native entry. Keep
// real guest frames and checked quotas. No scalar interpreter runs in this path.
func (e *Execution) runVeneer(backend Backend) (Status, bool) {
	f := &e.frames[len(e.frames)-1]
	code := f.function.Code
	plan := code.Veneer
	returning := f.suspended && code.Code[f.pc].Op == slots.ControlBase+uint32(ir.Return)
	if !returning && (plan == nil || f.pc != 0 || !e.veneerNeedsBoundary(f, plan)) {
		return Running, false
	}
	owner, ok := backend.(interface{ CanRunFramePlan(*state.Function) bool })
	if !ok || !owner.CanRunFramePlan(f.function) {
		return Running, false
	}
	var fuel uint64
	if returning {
		fuel = code.Sources[code.Code[f.pc].Source].Fuel()
		if plan != nil && f.pc == plan.ReturnPC {
			fuel = plan.ReturnFuel
		}
	} else if f.pc == 0 {
		fuel = plan.CallFuel
	} else {
		fuel = plan.ReturnFuel
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
	if !returning && f.pc == 0 {
		if plan.Index != slots.NoSlot {
			e.values[f.base+int(plan.Index)] = value.Value{Lo: plan.Bits}
		}
		f.pc = plan.CallPC
		return e.runCall(), true
	}
	return e.returnFrame(), true
}

// Prefer the native batch for same-instance guest callees. Forwarders to host
// or cross-instance functions would immediately exit that batch anyway. This
// selection reads the live target; runCall retains the full type/trap checks.
func (e *Execution) veneerNeedsBoundary(f *frame, plan *slots.CallVeneer) bool {
	in := f.function.Instance
	call := f.function.Code.Code[plan.CallPC]
	im := f.function.Code.Sources[call.Source].Wasm
	var target uint64
	if call.Op == 0x10 {
		target = in.Functions[im.A]
	} else {
		table := in.Tables[im.B]
		index := table.Type.Limits.Address(plan.Bits)
		if index >= uint64(len(table.Values)) {
			return false
		}
		target = table.Values[index]
	}
	callee, ok := e.store.Function(target)
	return ok && (callee.Code == nil || callee.Instance != in)
}
