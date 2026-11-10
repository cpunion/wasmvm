package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Definitions have consecutive Store identities. Check the actual instance's
// identity after the range lookup so a different allocation layout falls back
// safely. Imports and structural subtypes use the common checked dispatcher.
func (e *functionEmitter) linkedIndirect(pc int, config Config) bool {
	count := uint32(len(config.LinkedSignatures))
	if count == 0 || count > config.FunctionCount || len(config.LinkedCallees) != int(count) {
		return false
	}
	in := e.f.Code[pc]
	source := e.f.Sources[in.Source].Wasm
	if uint64(source.A) >= uint64(len(config.LinkedTypeLabels)) || in.Count == 0 {
		return false
	}
	first := config.FunctionCount - count
	offset, valid := nativeabi.TablePairOffset(config.FunctionCount, source.B)
	if in.Op == 0x11 && !valid {
		return false
	}
	if !e.hasLinkedDispatch {
		e.linkedDispatch, e.hasLinkedDispatch = e.a.label(), true
	}
	fallback := e.a.label()
	e.spillRegisters()
	e.callLoad(2, e.callContext(), uint32(nativeabi.RuntimeDataOffset))
	e.callTest(2)
	e.branch(fallback, "eq")
	input := e.f.Args[in.Extra+in.Count-1]
	if in.Op == 0x11 {
		e.callMemory(0, e.callFrame(), uint32(input)*16, false, e.f.Slots[input] != wasm.I64)
		e.callLoad(1, e.callReg(2), uint32(offset))
		e.callCompare(0, 1)
		e.branch(fallback, "hs")
		e.callLoad(1, e.callReg(2), uint32(offset+8))
		e.callShift(0, 3)
		e.callAdd(0, 0, 1)
		e.callLoad(0, e.callReg(0), 0)
	} else {
		e.callLoad(0, e.callFrame(), uint32(input)*16)
	}
	e.callMove(5, 0)
	e.callLoad(1, e.callReg(2), first*8)
	e.callSubtract(0, 1)
	e.callConst(1, uint64(count))
	e.callCompare(0, 1)
	e.branch(fallback, "hs")
	// Verify the identity at the selected module index before trusting its row.
	e.callConst(1, uint64(first))
	e.callAdd(0, 0, 1)
	e.callShift(0, 3)
	e.callAdd(0, 0, 2)
	e.callLoad(0, e.callReg(0), 0)
	e.callCompare(0, 5)
	e.branch(fallback, "ne")
	e.callMove(0, 5)
	e.callLoad(1, e.callReg(2), first*8)
	e.callSubtract(0, 1)
	e.callShift(0, 5)
	e.callAddress(e.linkedDispatch, 1)
	e.callAdd(3, 0, 1)
	e.callMemory(0, e.callReg(3), 8, false, true)
	e.callConst(1, uint64(config.LinkedTypeLabels[source.A]))
	e.callCompare(0, 1)
	e.branch(fallback, "ne")
	// Payload[0] is scratch until a native exit publishes its instruction PC.
	e.callStore(3, e.callContext(), uint32(nativeabi.PayloadOffset))
	e.linkedCall(pc, nil)
	e.a.mark(fallback)
	e.contextConst(nativeabi.SourceOffset, source.Offset)
	e.callRawLeave(nativeabi.Call, uint64(pc), uint64(pc+1))
	return true
}

// One immutable table per emitted function keeps lookup and frame setup
// independent of the number of possible callees. Rows contain only validated
// slot counts, structural identities and relocated compiler-owned branches.
func (e *functionEmitter) emitLinkedDispatch(config Config) {
	if !e.hasLinkedDispatch {
		return
	}
	e.a.mark(e.linkedDispatch)
	first := config.FunctionCount - uint32(len(config.LinkedSignatures))
	for i, signature := range config.LinkedSignatures {
		index := first + uint32(i)
		callee := config.LinkedCallees[index]
		if callee == nil {
			panic(&wasm.Error{Kind: wasm.ErrInvalid, Detail: "missing native indirect definition"})
		}
		e.a.word(uint32(len(callee.Slots)))
		e.a.word(index)
		e.a.word(signature)
		e.a.word(0)
		jump := e.a.label()
		e.a.mark(jump)
		start := len(e.a.code)
		e.a.branch(e.start, false, 0)
		e.linkedRelocations = append(e.linkedRelocations, linkedRelocation{jump, index})
		for len(e.a.code)-start < 16 {
			e.a.bytes(0)
		}
	}
}
