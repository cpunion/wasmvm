package codegen

import (
	"encoding/binary"
	"math"
	"slices"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type CallRelocation struct{ Offset, Function uint32 }
type nativeCallResume struct {
	label, pc    int
	plan         *registerPlan
	result       slots.Slot
	destinations []slots.Slot
}
type linkedRelocation struct {
	label  int
	target uint32
}

// LinkCalls resolves compiler-owned body branches after module layout. Validate
// every relocation before changing the unpublished code buffer.
func LinkCalls(arch string, code []byte, relocations []CallRelocation, targets map[uint32]uint32) error {
	for _, r := range relocations {
		target, ok := targets[r.Function]
		if !ok || uint64(target) >= uint64(len(code)) {
			return wasm.ErrInvalid
		}
		pos := uint64(r.Offset)
		switch arch {
		case "amd64":
			if pos+5 > uint64(len(code)) || code[pos] != 0xe9 {
				return wasm.ErrInvalid
			}
			delta := int64(target) - int64(pos) - 5
			if delta < math.MinInt32 || delta > math.MaxInt32 {
				return wasm.ErrLimit
			}
		case "arm64":
			if pos%4 != 0 || target%4 != 0 || pos+4 > uint64(len(code)) || binary.LittleEndian.Uint32(code[pos:])>>26 != 5 {
				return wasm.ErrInvalid
			}
			delta := int64(target) - int64(pos)
			if delta < -(1<<27) || delta >= 1<<27 {
				return wasm.ErrLimit
			}
		default:
			return wasm.ErrUnsupported
		}
	}
	for _, r := range relocations {
		delta := int64(targets[r.Function]) - int64(r.Offset)
		if arch == "amd64" {
			binary.LittleEndian.PutUint32(code[r.Offset+1:], uint32(delta-5))
		} else {
			binary.LittleEndian.PutUint32(code[r.Offset:], 0x14000000|uint32(delta/4)&0x3ffffff)
		}
	}
	return nil
}

func (e *functionEmitter) callAddFrame(n byte) {
	r, base := e.callReg(n), e.callFrame()
	if e.a.arch == "amd64" {
		e.a.bytes(0x48|(base>>3)<<2|r>>3, 0x01, 0xc0|(base&7)<<3|r&7)
	} else {
		e.a.word(0x8b000000 | uint32(base)<<16 | uint32(r)<<5 | uint32(r))
	}
}

// Copy indexed complete values. Work registers 3/4 retain parent metadata and
// destination frame, 6/7 walk immutable source/destination index tables.
func (e *functionEmitter) callCopy(count uint32, cancelled int) {
	if count == 0 {
		return
	}
	e.callConst(5, uint64(count))
	loop := e.a.label()
	e.a.mark(loop)
	e.callMemory(0, e.callReg(6), 0, false, true)
	e.callAdvance(6, 4)
	e.callShift(0, 4)
	e.callAddFrame(0)
	e.callPair(1, 2, e.callReg(0), 0, false)
	e.callMemory(0, e.callReg(7), 0, false, true)
	e.callAdvance(7, 4)
	e.callShift(0, 4)
	e.callAdd(0, 0, 4)
	e.callPair(1, 2, e.callReg(0), 0, true)
	e.callCancelled(cancelled)
	e.callAdvance(5, -1)
	e.callTest(5)
	e.branch(loop, "ne")
}
func (e *functionEmitter) callPlan(label int, indices []slots.Slot) {
	e.a.mark(label)
	e.callIndices(indices)
}

func (e *functionEmitter) callIndices(indices []slots.Slot) {
	for _, slot := range indices {
		e.a.word(uint32(slot))
	}
}

// Linked callers have checked depth, slot quotas and physical arena capacity.
// The shared runtime view was validated at the ordinary entry and cannot move
// inside this native run. Restore frame/root state and the compiler-known PC
// directly, without reloading a just-written continuation from memory.
func (e *functionEmitter) callEntryFrame(pc uint64) {
	e.registerMove(e.callFrame(), e.callReg(4), true)
	// Ordinary entry established root map zero for this native run. Private
	// calls never change it; every body uses the same conservative root map.
	e.constant(0, pc)
}

func (e *functionEmitter) linkedCall(pc int, callee *slots.Function) bool {
	in := e.f.Code[pc]
	indirect := in.Op == 0x11 || in.Op == 0x14
	count := in.Count
	if indirect {
		count--
	} else if callee == nil || int(count) != len(callee.Signature.Params) || int(count) > len(callee.Slots) {
		return false
	}
	resultsCount := len(e.f.Sources[in.Source].Outputs)
	if callee != nil {
		resultsCount = len(callee.Signature.Results)
	}
	end := uint64(in.Extra) + uint64(in.Count) + uint64(resultsCount)
	if end > uint64(len(e.f.Args)) {
		return false
	}
	a := e.a
	fallback, cancelled, resume := a.label(), a.label(), a.label()
	inputs, parameters := a.label(), a.label()
	if e.nativeFunction == 0 {
		// Wasm function indices occupy one word in the private call record.
		// Reject a malformed standalone context instead of truncating its ID.
		e.callMemory(0, e.callContext(), uint32(nativeabi.FunctionOffset)+4, false, true)
		e.callTest(0)
		e.branch(fallback, "ne")
	}
	e.callLoad(2, e.callContext(), uint32(nativeabi.CallViewOffset))
	e.callTest(2)
	e.branch(fallback, "eq")
	e.callLoad(1, e.callReg(2), nativeabi.CallCapacityOffset)
	if e.frameRegisters {
		e.callCompare(callTop, 1)
	} else {
		e.callTopLoad(0)
		e.callCompare(0, 1)
	}
	e.branch(fallback, "hs")
	e.callLoad(1, e.callReg(2), nativeabi.CallDepthOffset)
	if e.frameRegisters {
		e.callCompare(callDepth, 1)
	} else {
		e.callStateLoad(0, nativeabi.DepthOffset)
		e.callCompare(0, 1)
	}
	e.branch(fallback, "hs")
	e.callStateLoad(0, nativeabi.UsedOffset)
	if indirect {
		e.callLoad(1, e.callContext(), uint32(nativeabi.PayloadOffset))
		e.callMemory(1, e.callReg(1), 0, false, true)
		e.callConst(7, uint64(len(e.f.Slots)))
		e.callAdd(1, 1, 7)
		e.callShift(1, 4)
	} else {
		e.callConst(1, uint64(len(e.f.Slots)+len(callee.Slots))*16)
	}
	e.callAdd(0, 0, 1)
	e.callLoad(1, e.callReg(2), nativeabi.CallSlotsOffset)
	e.callCompare(1, 0)
	e.branch(fallback, "lo")
	e.callLoad(1, e.callContext(), uint32(nativeabi.CapacityOffset))
	e.callCompare(1, 0)
	e.branch(fallback, "lo")
	if !indirect {
		e.spillRegisters()
	}
	if e.frameRegisters {
		// The current frame is already Arena+Used. Quotas above checked
		// the complete child before deriving its address from that frame.
		e.callConst(4, uint64(len(e.f.Slots))*16)
		e.callAddFrame(4)
	} else {
		e.callStateLoad(0, nativeabi.UsedOffset)
		e.callConst(1, uint64(len(e.f.Slots))*16)
		e.callAdd(0, 0, 1)
		e.callLoad(4, e.callContext(), uint32(nativeabi.ArenaOffset))
		e.callAdd(4, 4, 0)
	}
	if indirect {
		e.callLoad(5, e.callContext(), uint32(nativeabi.PayloadOffset))
		e.callMemory(5, e.callReg(5), 0, false, true)
		e.callZeroDynamic(count, cancelled)
	} else {
		e.callZero(uint32(len(callee.Slots)), count, cancelled)
	}
	args := e.f.Args[in.Extra : in.Extra+count]
	if len(args) <= 8 {
		for i, slot := range args {
			e.callPair(0, 1, e.callFrame(), uint32(slot)*16, false)
			e.callPair(0, 1, e.callReg(4), uint32(i)*16, true)
		}
	} else {
		e.callAddress(inputs, 6)
		e.callAddress(parameters, 7)
		e.callCopy(count, cancelled)
	}
	e.callMeta(false)
	if e.nativeFunction != 0 {
		// Both halves are compiler constants. Publish their complete packed
		// word with one store; the remaining record words stay full width.
		e.callConst(0, uint64(uint32(pc))<<32|(e.nativeFunction-1))
		e.callPair(0, callUsed, e.callReg(3), nativeabi.CallFunctionOffset, true)
	} else {
		e.callMemory(0, e.callContext(), uint32(nativeabi.FunctionOffset), false, true)
		e.callMemory(0, e.callReg(3), nativeabi.CallFunctionOffset, true, true)
		e.callConst(0, uint64(pc))
		e.callMemory(0, e.callReg(3), nativeabi.CallPCOffset, true, true)
		if e.frameRegisters {
			e.callStore(callUsed, e.callReg(3), nativeabi.CallUsedOffset)
		} else {
			e.callStateLoad(0, nativeabi.UsedOffset)
			e.callStore(0, e.callReg(3), nativeabi.CallUsedOffset)
		}
	}
	e.callAddress(resume, 0)
	e.callStore(0, e.callReg(3), nativeabi.CallResumeOffset)
	if e.frameRegisters {
		e.callAdvance(callTop, 1)
	} else {
		e.callTopLoad(0)
		e.callAdvance(0, 1)
		e.callStore(0, e.callReg(2), nativeabi.CallTopOffset)
	}
	e.callPushState()
	if !e.frameRegisters {
		e.contextConst(nativeabi.ContinuationOffset, 0)
	}
	if indirect {
		e.callLoad(1, e.callContext(), uint32(nativeabi.PayloadOffset))
		if e.nativeFunction == 0 {
			e.callMemory(0, e.callReg(1), 4, false, true)
			e.callStore(0, e.callContext(), uint32(nativeabi.FunctionOffset))
		}
		e.callAdvance(1, 16)
		e.callEntryFrame(0)
		e.callJump(1)
	} else {
		target := e.f.Sources[in.Source].Wasm.A
		if e.nativeFunction == 0 {
			e.contextConst(nativeabi.FunctionOffset, uint64(target))
		}
		e.callEntryFrame(0)
		jump := a.label()
		a.mark(jump)
		a.branch(e.start, false, 0)
		e.linkedRelocations = append(e.linkedRelocations, linkedRelocation{jump, target})
	}
	if e.unmetered {
		plan := e.currentRegisterPlan()
		if e.allocationPlans != nil {
			plan = e.allocationPlans[pc+1]
		}
		result := slots.NoSlot
		if e.frameRegisters && resultsCount == 1 {
			slot := e.f.Args[in.Extra+in.Count]
			if typ := e.f.Slots[slot]; typ != wasm.F32 && typ != wasm.F64 && typ != wasm.V128 {
				result = slot
			}
		}
		e.callResumes = append(e.callResumes, nativeCallResume{label: resume, pc: pc + 1, plan: plan, result: result, destinations: e.f.Args[in.Extra+in.Count : uint32(end)]})
	} else {
		e.callIndices(e.f.Args[in.Extra+in.Count : uint32(end)])
		a.mark(resume)
		e.callEntryFrame(uint64(pc + 1))
		a.branch(e.linkedStart, false, 0)
	}
	if len(args) > 8 {
		e.callPlan(inputs, args)
		a.mark(parameters)
		for i := range args {
			a.word(uint32(i))
		}
	}
	a.mark(cancelled)
	if a.arch == "arm64" {
		a.word(0xd50339bf)
	}
	e.contextConst(nativeabi.SourceOffset, e.f.Sources[in.Source].Wasm.Offset)
	e.callRawLeave(nativeabi.Call, uint64(pc), uint64(pc+1))
	a.mark(fallback)
	e.contextConst(nativeabi.SourceOffset, e.f.Sources[in.Source].Wasm.Offset)
	if indirect {
		e.callRawLeave(nativeabi.Call, uint64(pc), uint64(pc+1))
	} else {
		e.leave(nativeabi.Call, uint64(pc), uint64(pc+1))
	}
	return true
}

func (e *functionEmitter) linkedReturn(pc int) {
	a := e.a
	fallback, cancelled, sources := a.label(), a.label(), a.label()
	e.callLoad(2, e.callContext(), uint32(nativeabi.CallViewOffset))
	e.callTest(2)
	e.branch(fallback, "eq")
	if e.frameRegisters {
		e.callTest(callTop)
	} else {
		e.callTopLoad(0)
		e.callTest(0)
	}
	e.branch(fallback, "eq")
	in := e.f.Code[pc]
	scalar := false
	var source slots.Slot
	var cached byte
	var hasCached bool
	if e.frameRegisters && in.Count == 1 {
		source = e.f.Args[in.Extra]
		typ := e.f.Slots[source]
		scalar = typ != wasm.F32 && typ != wasm.F64 && typ != wasm.V128
		cached, hasCached = e.registers[source]
	}
	deferred := e.callReturnSpills(in, scalar && hasCached)
	if scalar && hasCached {
		// Work word 5 survives parent-metadata preparation. Capture the
		// already materialized low word without loading its spill again.
		e.registerMove(e.callReg(5), cached, true)
	}
	e.callMeta(true)
	e.callLoad(4, e.callReg(3), nativeabi.CallUsedOffset)
	e.callLoad(0, e.callContext(), uint32(nativeabi.ArenaOffset))
	e.callAdd(4, 4, 0)
	if scalar {
		e.callScalarResult(source, hasCached, cancelled)
	} else if in.Count != 0 {
		e.callAddress(sources, 6)
		e.callDestinations(in.Count)
		e.callCopy(in.Count, cancelled)
	}
	if e.nativeFunction == 0 {
		e.callMemory(0, e.callReg(3), nativeabi.CallFunctionOffset, false, true)
		e.callStore(0, e.callContext(), uint32(nativeabi.FunctionOffset))
	}
	e.callPopState()
	if !e.unmetered {
		e.callMemory(0, e.callReg(3), nativeabi.CallPCOffset, false, true)
		e.callAdvance(0, 1)
		e.callStore(0, e.callContext(), uint32(nativeabi.ContinuationOffset))
	}
	if e.frameRegisters {
		e.callAdvance(callTop, -1)
	} else {
		e.callLoad(2, e.callContext(), uint32(nativeabi.CallViewOffset))
		e.callTopLoad(0)
		e.callAdvance(0, -1)
		e.callStore(0, e.callReg(2), nativeabi.CallTopOffset)
	}
	e.callLoad(0, e.callReg(3), nativeabi.CallResumeOffset)
	e.callJump(0)
	e.callPlan(sources, e.f.Args[in.Extra:in.Extra+in.Count])
	a.mark(cancelled)
	if a.arch == "arm64" {
		a.word(0xd50339bf)
	}
	e.spillRegisterSlots(deferred)
	e.callRawLeave(nativeabi.Return, uint64(pc), uint64(pc))
	a.mark(fallback)
	e.leave(nativeabi.Return, uint64(pc), uint64(pc))
}

// Keep the complete canonical result in the parent frame and also carry its
// low word in work register 1 until the caller's private continuation. The high
// word remains authoritative in the frame, including opaque reference values.
func (e *functionEmitter) callScalarResult(source slots.Slot, cached bool, cancelled int) {
	if cached {
		e.callMove(1, 5)
	} else {
		e.callLoad(1, e.callFrame(), uint32(source)*16)
	}
	e.callLoad(2, e.callFrame(), uint32(source)*16+8)
	e.callDestinations(1)
	e.callMemory(0, e.callReg(7), 0, false, true)
	e.callShift(0, 4)
	e.callAdd(0, 0, 4)
	e.callPair(1, 2, e.callReg(0), 0, true)
	e.callCancelled(cancelled)
}

// The caller's immutable result indices immediately precede its continuation.
// Structural signature checks establish the count before a private call.
func (e *functionEmitter) callDestinations(count uint32) {
	e.callLoad(7, e.callReg(3), nativeabi.CallResumeOffset)
	bytes := uint64(count) * 4
	if bytes <= 4095 || e.a.arch == "amd64" {
		e.callAdvance(7, -int(bytes))
	} else {
		e.callConst(0, bytes)
		e.callSubtract(7, 0)
	}
}

// Private returns have a compiler-known PC and a fully materialized caller
// frame. Restore the ordinary memory/constants/register environment, then
// branch to that exact continuation without the host-entry dispatcher.
func (e *functionEmitter) emitCallResumes(pcs []int) {
	previous := e.currentRegisterPlan()
	defer e.useRegisterPlan(previous)
	for _, resume := range e.callResumes {
		e.callIndices(resume.destinations)
		e.a.mark(resume.label)
		e.callEntryFrame(0)
		var incoming *registerPlan
		if resume.result != slots.NoSlot && resume.plan != nil {
			if reg, ok := resume.plan.homes[resume.result]; ok && slices.Contains(e.integerRegisterPool(), reg) {
				// Install the value before scratch/binding initialization. The
				// selected guest home is disjoint from those register classes.
				e.registerMove(reg, e.callReg(1), true)
				incoming = &registerPlan{homes: map[slots.Slot]byte{resume.result: reg}, slots: []slots.Slot{resume.result}}
			}
		}
		e.callEnvironment()
		e.useRegisterPlan(incoming)
		e.transitionRegisters(resume.plan)
		e.a.branch(pcs[resume.pc], false, 0)
	}
}

// A fresh private call has a checked, initialized frame and policy established
// by the ordinary entry. Reuse the same binding setup as private returns, then
// enter the first original PC without a context-PC load or dispatch. Cold host
// resumptions retain the ordinary checked entry and all original continuations.
func (e *functionEmitter) emitCallStart(pcs []int, initial *registerPlan) int {
	if !e.frameRegisters {
		return e.linkedStart
	}
	previous := e.currentRegisterPlan()
	defer e.useRegisterPlan(previous)
	entry := e.a.label()
	e.a.mark(entry)
	e.useRegisterPlan(nil)
	e.callEnvironment()
	if e.allocationPlans != nil {
		initial = e.allocationPlans[0]
	}
	// Allocation entry stubs reload and dispatch an arbitrary context PC.
	// Fresh calls establish only the first plan, then enter the original PC.
	e.transitionRegisters(initial)
	e.a.branch(pcs[0], false, 0)
	return entry
}

func (e *functionEmitter) callEnvironment() {
	if e.usesMemory && (!e.memoryViews || e.sharedDefaultMemory) {
		// Match the ordinary pre-allocation entry: cached runtime words
		// still belong to the callee until the reload below.
		runtimeRegisters := e.runtimeRegisters
		e.runtimeRegisters = nil
		e.prepareMemory()
		e.runtimeRegisters = runtimeRegisters
	}
	e.reloadVectorConstants()
	e.prepareConstants()
	e.reloadRuntimeRegisters()
	e.prepareHeapCaches()
}

// A successful private return retires its child frame. Materialize results and
// any homes overwritten by the return's work registers now; preserve the rest
// in registers until the last cancellation poll. The cold exit materializes
// those surviving homes before exposing the still-live child to the driver.
func (e *functionEmitter) callReturnSpills(in slots.Instruction, scalarCarrier bool) []slots.Slot {
	if !e.frameRegisters {
		e.spillRegisters()
		return nil
	}
	var work [32]bool
	for n := byte(0); n < 8; n++ {
		work[e.callReg(n)] = true
	}
	results := map[slots.Slot]bool{}
	if !scalarCarrier {
		for _, slot := range e.f.Args[in.Extra : in.Extra+in.Count] {
			results[slot] = true
		}
	}
	var early, deferred []slots.Slot
	for _, slot := range e.registerSlots {
		reg, integer := e.registers[slot]
		if results[slot] || integer && work[reg] {
			early = append(early, slot)
		} else {
			deferred = append(deferred, slot)
		}
	}
	e.spillRegisterSlots(early)
	return deferred
}
