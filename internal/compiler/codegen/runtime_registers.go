package codegen

import (
	"slices"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Metadata is immutable within one native entry; mutable elements remain loads.
func (e *functionEmitter) runtimePriorities(config Config, frequencies []uint64) map[uint32]uint64 {
	weights := map[uint32]uint64{}
	for pc, in := range e.f.Code {
		w := frequency(frequencies, pc)
		if e.memoryViews && (isMemoryAccess(in.Op) || in.Op == 0x3f || in.Op>>16 == 0xfe && (in.Op&0xffff <= 2 || in.Op&0xffff >= 0x10)) {
			index := e.f.Sources[in.Source].Wasm.Memory
			if in.Op == 0x3f {
				index = e.f.Sources[in.Source].Wasm.A
			}
			if offset, ok := nativeabi.MemoryPairOffset(config.FunctionCount, config.TableCount, uint32(len(config.Globals)), e.memoryBindings, index); ok {
				weights[offset] += 2 * w
				weights[offset+8] += w
			}
		}
		if _, ok := e.heapAccesses[in.Source]; ok && in.Op>>16 == 0xfb {
			weights[e.heapPair] += 2 * w
			weights[e.heapPair+8] += 2 * w
		}
		if offset, _, ok := globalCell(e.f, in, config); ok {
			weights[offset] += 2 * w
		}
		if offset, ok := referenceCell(e.f, in, config); ok {
			weights[offset] += 2 * w
			if in.Op != 0xd2 {
				weights[offset+8] += 2 * w
			}
		}
		if in.Op != 0x11 || config.FunctionCount == 0 {
			continue
		}
		source := e.f.Sources[in.Source].Wasm
		targets := config.IndirectCallees[source.A]
		offset, valid := nativeabi.TablePairOffset(config.FunctionCount, source.B)
		if len(targets) == 0 || !valid {
			continue
		}
		weights[offset] += 2 * w
		weights[offset+8] += 2 * w
		for _, index := range targets[:min(4, len(targets))] {
			if index < config.FunctionCount {
				weights[index*8] += w
			}
		}
	}
	return weights
}

// Reload the planned words on every entry. Indirect-only missing views install
// zeros for checked fallback; direct operations have already validated the view.
func (e *functionEmitter) prepareRuntimeRegisters(config Config) {
	if len(e.runtimeRegisters) == 0 {
		return
	}
	if e.unmetered && e.a.arch == "arm64" {
		for _, in := range e.f.Code {
			offset, typ, ok := globalCell(e.f, in, config)
			if !ok || in.Op != 0x23 || typ == wasm.V128 || config.Globals[e.f.Sources[in.Source].Wasm.A].Mutable {
				continue
			}
			if _, cached := e.runtimeRegisters[offset]; cached {
				if e.runtimeGlobalValues == nil {
					e.runtimeGlobalValues = map[uint32]bool{}
				}
				e.runtimeGlobalValues[offset] = true
			}
		}
	}
	e.reloadRuntimeRegisters()
}

// Same-instance private returns reuse the selected immutable binding plan.
func (e *functionEmitter) reloadRuntimeRegisters() {
	offsets := make([]uint32, 0, len(e.runtimeRegisters))
	for offset := range e.runtimeRegisters {
		offsets = append(offsets, offset)
	}
	slices.Sort(offsets)
	if len(offsets) == 0 {
		return
	}
	missing, done := e.a.label(), e.a.label()
	e.contextLoad(0, nativeabi.RuntimeDataOffset)
	e.test(true)
	e.branch(missing, "eq")
	e.registerMove(e.operandRegister(2), e.operandRegister(0), true)
	for _, offset := range offsets {
		e.pointerLoad(0, 2, offset)
		if e.runtimeGlobalValues[offset] {
			// Binding guarantees a live cell; immutable imports match actual
			// mutability. Retain its complete value in the same planned word.
			e.pointerLoad(0, 0, 0)
		}
		e.registerMove(e.runtimeRegisters[offset], e.operandRegister(0), true)
	}
	e.a.branch(done, false, 0)
	e.a.mark(missing)
	for _, offset := range offsets {
		e.a.aconst(uint32(e.runtimeRegisters[offset]), 0, false)
	}
	e.a.mark(done)
}

// Return the physical register containing a trusted native cell/base pointer.
func (e *functionEmitter) runtimeBase(offset uint32) byte {
	if r, cached := e.runtimeRegisters[offset]; cached {
		return r
	}
	e.contextLoad(2, nativeabi.RuntimeDataOffset)
	e.runtimeLoad(2, offset)
	return e.operandRegister(2)
}

func (e *functionEmitter) runtimeLoad(dst byte, offset uint32) {
	if r, ok := e.runtimeRegisters[offset]; ok {
		e.registerMove(e.operandRegister(dst), r, true)
	} else {
		e.pointerLoad(dst, 2, offset)
	}
}

func (e *functionEmitter) runtimeCompare(offset uint32) {
	if r, ok := e.runtimeRegisters[offset]; ok {
		e.a.word(0xeb00013f | uint32(r)<<16) // cmp x9, cached word
	} else {
		e.runtimeLoad(1, offset)
		e.compare(true)
	}
}
