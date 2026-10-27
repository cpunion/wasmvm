package codegen

import (
	"math"
	"sort"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
)

// Table metadata and function identities cannot change inside a native entry:
// every mutating operation/host call exits to the serialized driver. Reuse spare
// arm64 constant registers for these words, refreshing them on every resume.
// Elements themselves remain live loads. No additional register spills or roots
// are needed, and a missing runtime view installs zero lengths for checked exits.
func (e *functionEmitter) prepareRuntimeRegisters(config Config) {
	if e.a.arch != "arm64" || config.FunctionCount == 0 {
		return
	}
	weights := map[uint32]int{}
	for _, in := range e.f.Code {
		if in.Op != 0x11 {
			continue
		}
		source := e.f.Sources[in.Source].Wasm
		targets := config.IndirectCallees[source.A]
		offset := uint64(config.FunctionCount)*8 + uint64(source.B)*16
		if len(targets) == 0 || offset > math.MaxInt32-8 {
			continue
		}
		weights[uint32(offset)] += 2
		weights[uint32(offset+8)] += 2
		for _, index := range targets[:min(4, len(targets))] {
			if index < config.FunctionCount {
				weights[index*8]++
			}
		}
	}
	var offsets []uint32
	for offset := range weights {
		offsets = append(offsets, offset)
	}
	sort.Slice(offsets, func(i, j int) bool {
		if weights[offsets[i]] != weights[offsets[j]] {
			return weights[offsets[i]] > weights[offsets[j]]
		}
		return offsets[i] < offsets[j]
	})
	used := map[byte]bool{}
	for _, r := range e.constantRegisters {
		used[r] = true
	}
	var available []byte
	for r := byte(1); r <= 8; r++ {
		if !used[r] {
			available = append(available, r)
		}
	}
	offsets = offsets[:min(len(offsets), len(available))]
	if len(offsets) == 0 {
		return
	}
	e.runtimeRegisters = map[uint32]byte{}
	missing, done := e.a.label(), e.a.label()
	e.contextLoad(0, nativeabi.RuntimeDataOffset)
	e.test(true)
	e.branch(missing, "eq")
	e.registerMove(e.operandRegister(2), e.operandRegister(0), true)
	for i, offset := range offsets {
		e.pointerLoad(0, 2, offset)
		e.registerMove(available[i], e.operandRegister(0), true)
		e.runtimeRegisters[offset] = available[i]
	}
	e.a.branch(done, false, 0)
	e.a.mark(missing)
	for _, r := range available[:len(offsets)] {
		e.a.aconst(uint32(r), 0, false)
	}
	e.a.mark(done)
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
