package codegen

import (
	"encoding/binary"
	"math"

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
	if config.SharedDispatch {
		e.sharedDispatchAddress()
	} else {
		e.callAddress(e.linkedDispatch, 1)
	}
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
	if !e.hasLinkedDispatch || config.SharedDispatch {
		return
	}
	e.a.mark(e.linkedDispatch)
	e.emitDispatchRows(config)
}

func (e *functionEmitter) emitDispatchRows(config Config) {
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

// sharedDispatchAddress uses a forward image-relative address, without assuming
// page alignment of a publication inside the macOS JIT arena. Work registers
// are already spilled; register 3 is scratch until the selected row is formed.
func (e *functionEmitter) sharedDispatchAddress() {
	label := e.a.label()
	e.a.mark(label)
	e.dispatchRelocations = append(e.dispatchRelocations, label)
	if e.a.arch == "amd64" {
		e.a.bytes(0x48, 0x8d, 0x0d)
		e.a.word(0)
	} else {
		e.a.word(0x1000000a) // adr x10, .
		e.a.word(0x5280000e) // movz w14, #0
		e.a.word(0x72a0000e) // movk w14, #0, lsl #16
		e.a.word(0x8b0e014a) // add x10, x10, x14
	}
}

// LinkedDispatchTable emits the same checked rows used by standalone bodies.
// Its branches must be resolved with LinkCalls against the matching policy.
func LinkedDispatchTable(arch string, config Config) ([]byte, []CallRelocation, error) {
	if arch != "amd64" && arch != "arm64" {
		return nil, nil, wasm.ErrUnsupported
	}
	count := uint32(len(config.LinkedSignatures))
	if count == 0 || count > config.FunctionCount || uint64(count)*32 > math.MaxUint32 {
		return nil, nil, wasm.ErrInvalid
	}
	first := config.FunctionCount - count
	for i := range config.LinkedSignatures {
		if config.LinkedCallees[first+uint32(i)] == nil {
			return nil, nil, wasm.ErrInvalid
		}
	}
	a := &assembler{arch: arch}
	e := &functionEmitter{a: a, start: a.label()}
	a.mark(e.start)
	e.emitDispatchRows(config)
	code, err := a.finish()
	if err != nil {
		return nil, nil, err
	}
	relocations := make([]CallRelocation, 0, count)
	for _, call := range e.linkedRelocations {
		position, _ := a.labelAt(call.label)
		relocations = append(relocations, CallRelocation{uint32(position), call.target})
	}
	return code, relocations, nil
}

// LinkDispatch resolves only compiler-owned forward table addresses. Validate
// all sites before touching the unpublished image, including the full arm64
// template; neither an entry address nor a mutable runtime pointer is exposed.
func LinkDispatch(arch string, code []byte, sites []uint32, target uint32) error {
	for _, site := range sites {
		pos := uint64(site)
		if target <= site || uint64(target) >= uint64(len(code)) {
			return wasm.ErrInvalid
		}
		switch arch {
		case "amd64":
			if pos+7 > uint64(len(code)) || code[pos] != 0x48 || code[pos+1] != 0x8d || code[pos+2] != 0x0d {
				return wasm.ErrInvalid
			}
			if int64(target)-int64(site)-7 > math.MaxInt32 {
				return wasm.ErrLimit
			}
		case "arm64":
			if site%4 != 0 || target%4 != 0 || pos+16 > uint64(len(code)) {
				return wasm.ErrInvalid
			}
			for i, word := range []uint32{0x1000000a, 0x5280000e, 0x72a0000e, 0x8b0e014a} {
				if binary.LittleEndian.Uint32(code[pos+uint64(i*4):]) != word {
					return wasm.ErrInvalid
				}
			}
		default:
			return wasm.ErrUnsupported
		}
	}
	for _, site := range sites {
		delta := target - site
		if arch == "amd64" {
			binary.LittleEndian.PutUint32(code[site+3:], delta-7)
		} else {
			binary.LittleEndian.PutUint32(code[site+4:], 0x5280000e|(delta&0xffff)<<5)
			binary.LittleEndian.PutUint32(code[site+8:], 0x72a0000e|(delta>>16)<<5)
		}
	}
	return nil
}
