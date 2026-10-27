package codegen

import (
	"math"
	"sort"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

func (e *functionEmitter) literalRegister(in slots.Instruction) (byte, bool) {
	v := e.f.Sources[in.Source].Wasm.Bits
	if in.Op == 0x41 {
		v = uint64(uint32(v))
	}
	r, ok := e.constantRegisters[v]
	return r, ok
}

// AArch64's otherwise unused caller-saved x1-x8 retain expensive integer
// constants for one native entry. They are immutable and need no spill/root
// state. Every continuation reloads the same pool before dispatch.
func (e *functionEmitter) prepareConstants() {
	if e.a.arch != "arm64" {
		return
	}
	weights := map[uint64]int{}
	for _, in := range e.f.Code {
		if in.Op != 0x41 && in.Op != 0x42 {
			continue
		}
		v := e.f.Sources[in.Source].Wasm.Bits
		if in.Op == 0x41 {
			v = uint64(uint32(v))
		}
		probe := assembler{arch: "arm64"}
		probe.aconst(1, v, v > math.MaxUint32)
		if len(probe.code) > 4 {
			weights[v] += len(probe.code) - 4
		}
	}
	values := make([]uint64, 0, len(weights))
	for v := range weights {
		values = append(values, v)
	}
	sort.Slice(values, func(i, j int) bool {
		if weights[values[i]] != weights[values[j]] {
			return weights[values[i]] > weights[values[j]]
		}
		return values[i] < values[j]
	})
	e.constantRegisters = map[uint64]byte{}
	for i, v := range values[:min(8, len(values))] {
		r := byte(i + 1)
		e.a.aconst(uint32(r), v, v > math.MaxUint32)
		e.constantRegisters[v] = r
	}
}
