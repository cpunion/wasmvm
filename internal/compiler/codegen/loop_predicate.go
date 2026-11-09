package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// A private body may retain recoverable words until an actual observation.
// Both ordinary exits and cold polls use this contract before publishing the
// complete frame. Original PC entries still execute the original operations.
type frameSnapshot struct {
	copies    []loopSlotCopy
	predicate *loopPredicate
	countdown *loopCountdown
}

type loopPredicate struct {
	in      slots.Instruction
	literal *slots.Instruction
}

func (e *functionEmitter) planLoopPredicate(v *loopValues, r memoryRegion) *loopPredicate {
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
	return &loopPredicate{in: in}
}

// Only the adjacent literal is removed. Its entry word must also be unread,
// and no recoverable copy may rely on the omitted write as its final anchor.
func (e *functionEmitter) deferPredicateLiteral(p *loopPredicate, v *loopValues, r memoryRegion) {
	pc := r.end - 1
	if pc < v.first || r.elided[pc] || p.in.Op == 0x45 || p.in.Op == 0x50 {
		return
	}
	in := e.f.Code[pc]
	op := uint32(0x41)
	if integerWide(p.in.Op) {
		op = 0x42
	}
	if in.Op != op || in.Dst != p.in.B || in.Dst == p.in.A || e.coalescedRegisterSlots()[in.Dst] {
		return
	}
	versions := v.bySlot[in.Dst]
	if len(versions) < 2 || v.versions[versions[0]].last >= v.versions[versions[0]].first {
		return
	}
	if _, ok := comparisonImmediate(e.a.arch, integerWide(p.in.Op), e.f.Sources[in.Source].Wasm.Bits); !ok {
		return
	}
	if r.copies != nil {
		for _, copy := range r.copies.final {
			if copy.src == in.Dst {
				return
			}
		}
	}
	p.literal = &in
}

func (e *functionEmitter) predicateFlags(p *loopPredicate) {
	in := p.in
	left, right := e.frameRead(in.A), e.frameRead(in.B)
	l, lok := e.registers[left]
	r, rok := e.registers[right]
	if !lok {
		e.slotLoad(left, 0, false, integerWide(in.Op))
		l = e.operandRegister(0)
	}
	if p.literal != nil {
		e.registerCompareImmediateFlags(l, in.Op, e.f.Sources[p.literal.Source].Wasm.Bits)
		return
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
	if s.countdown != nil {
		e.restoreLoopCountdown(s.countdown, !known)
	}
	// A deferred literal can feed a copy recipe, so restore it before copies.
	// When it aliases the comparison result, only cold recomputation needs it.
	if s.predicate != nil && s.predicate.literal != nil && (!known || s.predicate.literal.Dst != s.predicate.in.Dst) {
		e.literal(*s.predicate.literal)
	}
	e.materializeFrameCopies(s.copies)
	if s.predicate == nil {
		return
	}
	if !known {
		e.integer(s.predicate.in)
		return
	}
	bits := uint64(0)
	if condition {
		bits = 1
	}
	e.constant(0, bits)
	e.slotStore(s.predicate.in.Dst, 0, false)
}
