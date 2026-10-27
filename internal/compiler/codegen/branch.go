package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

func (e *functionEmitter) branchZero(slot slots.Slot, label int) {
	e.branchCondition(slot, label, false)
}

// Use only the low i32 condition bits, directly in the allocated register.
func (e *functionEmitter) branchCondition(slot slots.Slot, label int, nonzero bool) {
	r, ok := e.registers[slot]
	if !ok {
		e.slotLoad(slot, 0, false, false)
		r = e.operandRegister(0)
	}
	if e.a.arch == "arm64" {
		e.a.compareBranch(label, r, false, nonzero)
	} else {
		if rex := byte(0x40) | (r>>3)<<2 | r>>3; rex != 0x40 {
			e.a.bytes(rex)
		}
		e.a.bytes(0x85, 0xc0|(r&7)<<3|(r&7))
		cc := byte(4)
		if nonzero {
			cc = 5
		}
		e.a.branch(label, true, cc)
	}
}

// Prefer a direct conditional edge and let the other arm fall through. Edges
// with copies keep their complete parallel-move plan on the selected path.
func (e *functionEmitter) conditionalEdges(in slots.Instruction, pcs []int, next int) {
	e.conditionalEdgesUsing(in, pcs, next, e.branchCondition)
}

func (e *functionEmitter) conditionalEdgesUsing(in slots.Instruction, pcs []int, next int, branch func(slots.Slot, int, bool)) {
	yes, no := e.f.Edges[in.Extra], e.f.Edges[in.Extra+1]
	if yes.Count == 0 && (int(no.PC) == next || no.Count != 0 || int(yes.PC) != next) {
		branch(in.A, pcs[yes.PC], true)
		e.edge(in.Extra+1, pcs, next)
	} else if no.Count == 0 {
		branch(in.A, pcs[no.PC], false)
		e.edge(in.Extra, pcs, next)
	} else {
		other := e.a.label()
		branch(in.A, other, false)
		e.edge(in.Extra, pcs, -1)
		e.a.mark(other)
		e.edge(in.Extra+1, pcs, next)
	}
}

// Only normal fallthrough consumes the predecessor's integer comparison flags.
// AArch64's interrupt load/CBNZ sequence preserves NZCV. The canonical boolean
// remains materialized for polls, edge copies and arbitrary cold continuations.
func (e *functionEmitter) compareBranchFallthrough(pc int, pcs []int) bool {
	if !e.unmetered || e.a.arch != "arm64" || pc == 0 {
		return false
	}
	in, previous := e.f.Code[pc], e.f.Code[pc-1]
	if in.Op != slots.ControlBase+uint32(ir.If) || in.A != previous.Dst || previous.Op < 0x45 || previous.Op > 0x5a {
		return false
	}
	cc := armIntegerCondition(previous.Op)
	e.interrupt(uint32(pc))
	e.conditionalEdgesUsing(in, pcs, pc+1, func(_ slots.Slot, label int, nonzero bool) {
		condition := cc
		if !nonzero {
			condition ^= 1
		}
		e.a.branch(label, true, condition)
	})
	return true
}
