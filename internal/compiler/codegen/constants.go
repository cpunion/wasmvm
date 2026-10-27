package codegen

import (
	"math"
	"sort"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

// AArch64's otherwise unused caller-saved x1-x8 retain expensive integer
// constants for one native entry. They are immutable and need no spill/root
// state. Every continuation reloads the same pool before dispatch.
func (e *functionEmitter) prepareConstants(config Config) {
	if e.a.arch != "arm64" {
		return
	}
	weights := map[uint64]int{}
	functions := []*slots.Function{e.f}
	for _, in := range e.f.Code {
		if in.Op == 0x10 {
			callee := config.DirectCallees[e.f.Sources[in.Source].Wasm.A]
			if _, ok := leafExpressions(callee, e.f.Args[in.Extra:in.Extra+in.Count]); ok {
				functions = append(functions, callee)
			}
		} else if in.Op == 0x11 {
			for _, index := range config.IndirectCallees[e.f.Sources[in.Source].Wasm.A] {
				if callee := config.DirectCallees[index]; CanExpandLeaf(callee) {
					functions = append(functions, callee)
				}
			}
		}
	}
	for _, f := range functions {
		for _, in := range f.Code {
			if in.Op != 0x41 && in.Op != 0x42 {
				continue
			}
			v := f.Sources[in.Source].Wasm.Bits
			if in.Op == 0x41 {
				v = uint64(uint32(v))
			}
			probe := assembler{arch: "arm64"}
			probe.aconst(1, v, v > math.MaxUint32)
			if len(probe.code) > 4 {
				weights[v] += len(probe.code) - 4
			}
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

// Select an immutable constant operand without changing the permanent slot
// allocation. The caller proves the constant's old slot is dead on this path.
func (e *functionEmitter) pooledInteger(in slots.Instruction, v uint64) bool {
	if !integerWide(in.Op) {
		v = uint64(uint32(v))
	}
	r, ok := e.constantRegisters[v]
	if !ok || in.Dst == in.B || in.A == in.B {
		return false
	}
	old, had := e.registers[in.B]
	e.registers[in.B] = r
	defer func() {
		if had {
			e.registers[in.B] = old
		} else {
			delete(e.registers, in.B)
		}
	}()
	return e.registerInteger(in.Op, in.Dst, in.A, in.B)
}

func (e *functionEmitter) literalRegister(in slots.Instruction) (byte, bool) {
	v := e.f.Sources[in.Source].Wasm.Bits
	if in.Op == 0x41 {
		v = uint64(uint32(v))
	}
	r, ok := e.constantRegisters[v]
	return r, ok
}
