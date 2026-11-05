package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// A private body may retain recoverable words until an actual observation.
// Both ordinary exits and cold polls use this contract before publishing the
// complete frame. Original PC entries still execute the original operations.
type frameSnapshot struct {
	copies    []loopSlotCopy
	predicate *slots.Instruction
}

func (e *functionEmitter) loopPredicate(v *loopValues, r memoryRegion) *slots.Instruction {
	if v == nil || r.loop == nil || !r.loop.flags {
		return nil
	}
	in := e.f.Code[r.end]
	versions := v.bySlot[in.Dst]
	if len(versions) < 2 || v.versions[versions[0]].last >= v.versions[versions[0]].first || e.coalescedRegisterSlots()[in.Dst] {
		return nil
	}
	// The final comparison is nontrapping and its inputs remain intact when
	// its write is omitted. No entry read may observe the stale result on the
	// next iteration. The local graph also excludes interior/header polls.
	return &in
}

func (e *functionEmitter) predicateFlags(in slots.Instruction) {
	left, right := e.frameRead(in.A), e.frameRead(in.B)
	l, lok := e.registers[left]
	r, rok := e.registers[right]
	if !lok {
		e.slotLoad(left, 0, false, integerWide(in.Op))
		l = e.operandRegister(0)
	}
	if !rok && in.Op != 0x45 && in.Op != 0x50 {
		e.slotLoad(right, 1, false, integerWide(in.Op))
		r = e.operandRegister(1)
	}
	e.registerCompareFlags(l, r, in.Op)
}

func (e *functionEmitter) materializeFrameSnapshot(s *frameSnapshot, known, condition bool) {
	if s == nil {
		return
	}
	e.materializeFrameCopies(s.copies)
	if s.predicate == nil {
		return
	}
	if !known {
		e.integer(*s.predicate)
		return
	}
	bits := uint64(0)
	if condition {
		bits = 1
	}
	e.constant(0, bits)
	e.slotStore(s.predicate.Dst, 0, false)
}
