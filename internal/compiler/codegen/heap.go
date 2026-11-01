package codegen

import (
	"math"
	"math/bits"
	"sort"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type heapAccess struct {
	offset                       uint32
	width                        uint32
	array, length, store, signed bool
	cacheClass                   uint64
}

type heapCache struct{ key, address byte }

// Cache only within one native entry. No host call, reclamation or relocation
// can occur during that interval; bounded bump allocations preserve addresses.
// Full handle equality is
// checked on every hit; each miss uses the ordinary null/index/generation
// checks. Entry resets prevent a moved address surviving a safepoint.
func (e *functionEmitter) prepareHeapCaches() {
	if !e.unmetered || e.a.arch != "arm64" || len(e.heapAccesses) == 0 {
		return
	}
	free := e.spareCallerRegisters()
	if len(free) < 2 {
		return
	}
	weights := map[uint64]uint64{}
	frequencies := instructionFrequencies(e.f)
	for pc, in := range e.f.Code {
		if access, ok := e.heapAccesses[in.Source]; ok && in.Op>>16 == 0xfb {
			weights[access.cacheClass] += frequency(frequencies, pc)
		}
	}
	var classes []uint64
	for class, weight := range weights {
		if weight >= 2 {
			classes = append(classes, class)
		}
	}
	sort.Slice(classes, func(i, j int) bool {
		if weights[classes[i]] != weights[classes[j]] {
			return weights[classes[i]] > weights[classes[j]]
		}
		return classes[i] < classes[j]
	})
	for _, class := range classes[:min(2, len(classes))] {
		if len(free) < 2 {
			break
		}
		if e.heapCaches == nil {
			e.heapCaches = map[uint64]heapCache{}
		}
		e.heapCaches[class] = heapCache{free[0], free[1]}
		// A non-null reference cannot match this initial key. The address is
		// written only after all miss checks succeed, before installing a key.
		e.a.aconst(uint32(free[0]), 0, false)
		free = free[2:]
	}
}

func (e *functionEmitter) prepareHeapAccesses(config Config) {
	// Cache each used struct prefix once, avoiding quadratic compilation for
	// repeated field accesses in large types. Validation owns type/mutability
	// proofs; subtypes retain exactly the same packed supertype prefix.
	layouts := map[uint32][]uint64{}
	for _, in := range e.f.Code {
		op := in.Op
		if op < 0xfb0002 || op > 0xfb0005 && (op < 0xfb000b || op > 0xfb000f) {
			continue
		}
		a := heapAccess{array: op >= 0xfb000b, length: op == 0xfb000f, store: op == 0xfb0005 || op == 0xfb000e, signed: op == 0xfb0003 || op == 0xfb000c}
		a.cacheClass = math.MaxUint64 // array.len has no concrete type immediate.
		if !a.length {
			src := e.f.Sources[in.Source].Wasm
			a.cacheClass = uint64(src.A)
			if uint64(src.A) >= uint64(len(config.HeapTypes)) {
				continue
			}
			t := config.HeapTypes[src.A]
			if a.array {
				if t.Kind != module.ArrayKind || len(t.Fields) != 1 {
					continue
				}
				a.width = uint32(t.Fields[0].HeapBytes())
			} else {
				if t.Kind != module.StructKind || uint64(src.B) >= uint64(len(t.Fields)) {
					continue
				}
				offsets := layouts[src.A]
				if offsets == nil {
					offsets = make([]uint64, len(t.Fields)+1)
					for i, f := range t.Fields {
						offsets[i+1] = offsets[i] + uint64(f.HeapBytes())
					}
					layouts[src.A] = offsets
				}
				if offsets[src.B] > math.MaxInt32 {
					continue
				}
				a.offset, a.width = uint32(offsets[src.B]), uint32(t.Fields[src.B].HeapBytes())
			}
			if a.width != 1 && a.width != 2 && a.width != 4 && a.width != 8 && a.width != 16 {
				continue
			}
		}
		if e.heapAccesses == nil {
			e.heapAccesses = map[uint32]heapAccess{}
		}
		e.heapAccesses[in.Source] = a
	}
}

func (e *functionEmitter) heapInstruction(in slots.Instruction) bool {
	if in.Op>>16 != 0xfb || in.Op != e.f.Sources[in.Source].Wasm.Op {
		return false
	}
	a, ok := e.heapAccesses[in.Source]
	if !ok {
		return false
	}
	e.slotLoad(in.A, 0, false, true)
	cache, cached := e.heapCaches[a.cacheClass]
	ready := 0
	if cached {
		ready = e.a.label()
		// Null must miss even when an entry's cache key is initially zero.
		// CCMP sets Z only for a non-null, complete matching handle.
		e.test(true)
		e.a.word(0xfa401120 | uint32(cache.key)<<16) // ccmp x9,key,#0,ne
		e.branch(ready, "eq")
	}
	e.test(true)
	e.branch(e.trapExit(wasm.TrapNullReference, e.f.Sources[in.Source].Wasm.Offset), "eq")
	// Validate the complete generation/tag word before loading a payload.
	// A stale or forged internal frame is a bad state, never a host address.
	if _, cached := e.runtimeRegisters[e.heapPair]; !cached {
		e.contextLoad(2, nativeabi.RuntimeDataOffset)
	}
	e.runtimeLoad(1, e.heapPair)
	e.registerMove(e.operandRegister(2), e.operandRegister(0), false)
	if e.a.arch == "amd64" {
		e.a.bytes(0x48, 0x39, 0xca) // cmp rdx, rcx
	} else {
		e.a.word(0xeb0a017f) // cmp x11, x10
	}
	e.branch(e.bad, "hs")
	if base, cached := e.runtimeRegisters[e.heapPair+8]; cached {
		e.registerMove(e.operandRegister(1), base, true)
	} else {
		e.contextLoad(1, nativeabi.RuntimeDataOffset)
		e.pointerLoad(1, 1, e.heapPair+8)
	}
	if e.a.arch == "amd64" {
		e.a.bytes(0x48, 0xc1, 0xe2, 5, 0x48, 0x01, 0xca) // shl rdx,5; add rdx,rcx
	} else {
		e.a.word(0x8b0b154b) // add x11, x10, x11, lsl #5
	}
	e.pointerLoad(1, 2, 0)
	e.compare(true)
	e.branch(e.bad, "ne")
	if !a.array {
		e.pointerLoad(2, 2, 8)
	}
	if cached {
		e.registerMove(cache.address, 11, true)
		e.registerMove(cache.key, 9, true)
		e.a.mark(ready)
	}
	base := e.operandRegister(2)
	if cached {
		if !a.array && a.offset == 0 {
			base = cache.address
		} else {
			e.registerMove(11, cache.address, true)
		}
	}
	if a.length {
		e.pointerLoad(0, 2, 16)
		e.heapResult(in.Dst)
		return true
	}
	if a.array {
		e.slotLoad(in.B, 0, false, false)
		e.pointerLoad(1, 2, 16)
		e.compare(true)
		e.branch(e.trapExit(wasm.TrapArrayBounds, e.f.Sources[in.Source].Wasm.Offset), "hs")
		e.pointerLoad(2, 2, 8)
		shift := uint32(bits.TrailingZeros32(a.width))
		if e.a.arch == "amd64" {
			if shift != 0 {
				e.a.bytes(0x48, 0xc1, 0xe0, byte(shift))
			}
			e.a.bytes(0x48, 0x01, 0xc2) // add rdx,rax
		} else {
			e.a.word(0x8b09016b | shift<<10) // add x11,x11,x9,lsl #stride
		}
	} else {
		if a.offset != 0 {
			if e.a.arch == "amd64" {
				e.a.bytes(0x48, 0x81, 0xc2)
				e.a.word(a.offset)
			} else {
				e.constant(0, uint64(a.offset))
				e.a.word(0x8b09016b)
			}
		}
	}
	if a.width == 16 {
		slot := in.Dst
		if a.store {
			slot = in.B
			if a.array {
				slot = e.f.Args[in.Extra+2]
			}
		}
		var reg byte
		if a.store {
			reg = e.vectorOperand(slot, 0)
		} else {
			reg = e.vectorDestination(slot)
		}
		if e.a.arch == "amd64" {
			e.a.bytes(0xf3)
			if reg >= 8 {
				e.a.bytes(0x44)
			}
			op := byte(0x6f)
			if a.store {
				op = 0x7f
			}
			e.a.bytes(0x0f, op, (reg&7)<<3|2) // movdqu xmm,[rdx] / [rdx],xmm
		} else {
			op := uint32(0x3dc00000)
			if a.store {
				op = 0x3d800000
			}
			e.a.word(op | uint32(base)<<5 | uint32(reg))
		}
		if !a.store {
			e.vectorStore(slot, reg)
		}
	} else if a.store {
		operand := in.B
		if a.array {
			operand = e.f.Args[in.Extra+2]
		}
		e.slotLoad(operand, 0, false, true)
		e.heapStore(a.width, base)
	} else {
		e.heapLoad(a.width, a.signed, base)
		e.heapResult(in.Dst)
	}
	return true
}

func (e *functionEmitter) heapResult(dst slots.Slot) {
	e.slotStore(dst, 0, false)
	if !e.unmetered {
		e.constant(0, 0)
		e.slotStore(dst, 0, true)
	}
}

func (e *functionEmitter) heapLoad(width uint32, signed bool, base byte) {
	if e.a.arch == "amd64" {
		if width >= 8 {
			e.a.bytes(0x48, 0x8b, 0x02)
		} else if width == 4 {
			e.a.bytes(0x8b, 0x02)
		} else {
			op := byte(0xb6)
			if width == 2 {
				op++
			}
			if signed {
				op += 8
			}
			e.a.bytes(0x0f, op, 0x02)
		}
	} else {
		op := uint32(0x39400009) | uint32(base)<<5 | uint32(bits.TrailingZeros32(width))<<30
		if signed && width < 4 {
			op |= 1 << 23
		}
		e.a.word(op)
	}
}

func (e *functionEmitter) heapStore(width uint32, base byte) {
	if e.a.arch == "amd64" {
		if width == 2 {
			e.a.bytes(0x66)
		} else if width == 8 {
			e.a.bytes(0x48)
		}
		op := byte(0x89)
		if width == 1 {
			op = 0x88
		}
		e.a.bytes(op, 0x02)
	} else {
		e.a.word(0x39000009 | uint32(base)<<5 | uint32(bits.TrailingZeros32(width))<<30)
	}
}
