package codegen

import (
	"math"
	"sort"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

// Score literal materialization savings in native instructions. Leaf analysis
// is bounded separately; exhausting it only loses optional cache candidates.
func (e *functionEmitter) constantPriorities(config Config, frequencies []uint64) map[uint64]uint64 {
	weights := map[uint64]uint64{}
	add := func(op uint32, v, w uint64) {
		if op != 0x41 && op != 0x42 {
			return
		}
		if op == 0x41 {
			v = uint64(uint32(v))
		}
		probe := assembler{arch: "arm64"}
		probe.aconst(1, v, v > math.MaxUint32)
		if len(probe.code) > 4 {
			weights[v] += uint64(len(probe.code)/4-1) * w
		}
	}
	for pc, in := range e.f.Code {
		if in.Op == 0x41 || in.Op == 0x42 {
			add(in.Op, e.f.Sources[in.Source].Wasm.Bits, frequency(frequencies, pc))
		}
	}
	work := uint64(len(e.f.Code)) * 8
	for pc, in := range e.f.Code {
		var callees []*slots.Function
		if in.Op == 0x10 {
			callee := config.DirectCallees[e.f.Sources[in.Source].Wasm.A]
			if _, ok := leafExpressions(callee, e.f.Args[in.Extra:in.Extra+in.Count]); ok {
				callees = append(callees, callee)
			}
		} else if in.Op == 0x11 {
			targets := config.IndirectCallees[e.f.Sources[in.Source].Wasm.A]
			for _, index := range targets[:min(4, len(targets))] {
				if callee := config.DirectCallees[index]; CanExpandLeaf(callee) {
					callees = append(callees, callee)
				}
			}
		}
		for _, callee := range callees {
			if uint64(len(callee.Code)) > work {
				return weights
			}
			work -= uint64(len(callee.Code))
			for _, literal := range callee.Code {
				add(literal.Op, callee.Sources[literal.Source].Wasm.Bits, frequency(frequencies, pc))
			}
		}
	}
	return weights
}

// Planned literals are immutable and reload on every independent entry.
func (e *functionEmitter) prepareConstants() {
	values := make([]uint64, 0, len(e.constantRegisters))
	for value := range e.constantRegisters {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	for _, value := range values {
		e.a.aconst(uint32(e.constantRegisters[value]), value, value > math.MaxUint32)
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
