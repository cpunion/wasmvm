package codegen

import (
	"encoding/binary"
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type CallRelocation struct{ Offset, Function uint32 }
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
	e.callLoad(1, e.callReg(0), 0)
	e.callLoad(2, e.callReg(0), 8)
	e.callMemory(0, e.callReg(7), 0, false, true)
	e.callAdvance(7, 4)
	e.callShift(0, 4)
	e.callAdd(0, 0, 4)
	e.callStore(1, e.callReg(0), 0)
	e.callStore(2, e.callReg(0), 8)
	e.callCancelled(cancelled)
	e.callAdvance(5, -1)
	e.callTest(5)
	e.branch(loop, "ne")
}
func (e *functionEmitter) callPlan(label int, indices []slots.Slot) {
	e.a.mark(label)
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
	e.contextConst(nativeabi.RootMapOffset, 0)
	if pc != 0 {
		e.constant(0, pc)
	}
}

func (e *functionEmitter) linkedCall(pc int, callee *slots.Function) bool {
	in := e.f.Code[pc]
	indirect := in.Op == 0x11 || in.Op == 0x14
	count := in.Count
	if indirect {
		count--
	} else if callee == nil || int(count) != len(callee.Signature.Params) {
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
	results, inputs, parameters := a.label(), a.label(), a.label()
	e.callLoad(2, e.callContext(), uint32(nativeabi.CallViewOffset))
	e.callTest(2)
	e.branch(fallback, "eq")
	e.callLoad(0, e.callReg(2), nativeabi.CallTopOffset)
	e.callLoad(1, e.callReg(2), nativeabi.CallCapacityOffset)
	e.callCompare(0, 1)
	e.branch(fallback, "hs")
	e.callLoad(0, e.callContext(), uint32(nativeabi.DepthOffset))
	e.callLoad(1, e.callReg(2), nativeabi.CallDepthOffset)
	e.callCompare(0, 1)
	e.branch(fallback, "hs")
	e.callLoad(0, e.callContext(), uint32(nativeabi.UsedOffset))
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
	e.callLoad(0, e.callContext(), uint32(nativeabi.UsedOffset))
	e.callConst(1, uint64(len(e.f.Slots))*16)
	e.callAdd(0, 0, 1)
	e.callLoad(4, e.callContext(), uint32(nativeabi.ArenaOffset))
	e.callAdd(4, 4, 0)
	if indirect {
		e.callLoad(5, e.callContext(), uint32(nativeabi.PayloadOffset))
		e.callMemory(5, e.callReg(5), 0, false, true)
		e.callZeroDynamic(cancelled)
	} else {
		e.callZero(uint32(len(callee.Slots)), cancelled)
	}
	args := e.f.Args[in.Extra : in.Extra+count]
	if len(args) <= 8 {
		for i, slot := range args {
			e.callLoad(0, e.callFrame(), uint32(slot)*16)
			e.callLoad(1, e.callFrame(), uint32(slot)*16+8)
			e.callStore(0, e.callReg(4), uint32(i)*16)
			e.callStore(1, e.callReg(4), uint32(i)*16+8)
		}
	} else {
		e.callAddress(inputs, 6)
		e.callAddress(parameters, 7)
		e.callCopy(count, cancelled)
	}
	e.callMeta(false)
	e.callLoad(0, e.callContext(), uint32(nativeabi.FunctionOffset))
	e.callStore(0, e.callReg(3), nativeabi.CallFunctionOffset)
	e.callLoad(0, e.callContext(), uint32(nativeabi.UsedOffset))
	e.callStore(0, e.callReg(3), nativeabi.CallUsedOffset)
	e.callConst(0, uint64(pc))
	e.callStore(0, e.callReg(3), nativeabi.CallPCOffset)
	e.callAddress(resume, 0)
	e.callStore(0, e.callReg(3), nativeabi.CallResumeOffset)
	e.callAddress(results, 0)
	e.callStore(0, e.callReg(3), nativeabi.CallDestinationsOffset)
	e.callLoad(0, e.callReg(2), nativeabi.CallTopOffset)
	e.callAdvance(0, 1)
	e.callStore(0, e.callReg(2), nativeabi.CallTopOffset)
	e.callLoad(0, e.callContext(), uint32(nativeabi.UsedOffset))
	e.callConst(1, uint64(len(e.f.Slots))*16)
	e.callAdd(0, 0, 1)
	e.callStore(0, e.callContext(), uint32(nativeabi.UsedOffset))
	e.callLoad(0, e.callContext(), uint32(nativeabi.DepthOffset))
	e.callAdvance(0, 1)
	e.callStore(0, e.callContext(), uint32(nativeabi.DepthOffset))
	e.contextConst(nativeabi.ContinuationOffset, 0)
	if indirect {
		e.callLoad(1, e.callContext(), uint32(nativeabi.PayloadOffset))
		e.callMemory(0, e.callReg(1), 4, false, true)
		e.callStore(0, e.callContext(), uint32(nativeabi.FunctionOffset))
		e.callAdvance(1, 16)
		e.callEntryFrame(0)
		e.callJump(1)
	} else {
		target := e.f.Sources[in.Source].Wasm.A
		e.contextConst(nativeabi.FunctionOffset, uint64(target))
		e.callEntryFrame(0)
		jump := a.label()
		a.mark(jump)
		a.branch(e.start, false, 0)
		e.linkedRelocations = append(e.linkedRelocations, linkedRelocation{jump, target})
	}
	a.mark(resume)
	e.callEntryFrame(uint64(pc + 1))
	a.branch(e.linkedStart, false, 0)
	e.callPlan(results, e.f.Args[in.Extra+in.Count:uint32(end)])
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
	e.callLoad(0, e.callReg(2), nativeabi.CallTopOffset)
	e.callTest(0)
	e.branch(fallback, "eq")
	e.spillRegisters()
	e.callMeta(true)
	e.callLoad(4, e.callReg(3), nativeabi.CallUsedOffset)
	e.callLoad(0, e.callContext(), uint32(nativeabi.ArenaOffset))
	e.callAdd(4, 4, 0)
	in := e.f.Code[pc]
	if in.Count != 0 {
		e.callAddress(sources, 6)
		e.callLoad(7, e.callReg(3), nativeabi.CallDestinationsOffset)
		e.callCopy(in.Count, cancelled)
	}
	e.callLoad(0, e.callReg(3), nativeabi.CallFunctionOffset)
	e.callStore(0, e.callContext(), uint32(nativeabi.FunctionOffset))
	e.callLoad(0, e.callReg(3), nativeabi.CallUsedOffset)
	e.callStore(0, e.callContext(), uint32(nativeabi.UsedOffset))
	e.callLoad(0, e.callReg(3), nativeabi.CallPCOffset)
	e.callAdvance(0, 1)
	e.callStore(0, e.callContext(), uint32(nativeabi.ContinuationOffset))
	e.callLoad(0, e.callContext(), uint32(nativeabi.DepthOffset))
	e.callAdvance(0, -1)
	e.callStore(0, e.callContext(), uint32(nativeabi.DepthOffset))
	e.callLoad(2, e.callContext(), uint32(nativeabi.CallViewOffset))
	e.callLoad(0, e.callReg(2), nativeabi.CallTopOffset)
	e.callAdvance(0, -1)
	e.callStore(0, e.callReg(2), nativeabi.CallTopOffset)
	e.callLoad(0, e.callReg(3), nativeabi.CallResumeOffset)
	e.callJump(0)
	e.callPlan(sources, e.f.Args[in.Extra:in.Extra+in.Count])
	a.mark(cancelled)
	if a.arch == "arm64" {
		a.word(0xd50339bf)
	}
	e.callRawLeave(nativeabi.Return, uint64(pc), uint64(pc))
	a.mark(fallback)
	e.leave(nativeabi.Return, uint64(pc), uint64(pc))
}
