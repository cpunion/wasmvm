package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Register priorities estimate emitted hot work. Encoded right operands do not
// need a register read, and a dead literal does not need its own write. Retain
// one cold-use credit for every discounted operation. This only changes stable
// register selection; ordinary lowering and independent PCs remain unchanged.
func (e *functionEmitter) immediateRegisterCosts(weights, frequencies []uint64, regions map[int]memoryRegion) {
	if e.a.arch != "amd64" || !e.unmetered {
		return
	}
	scan := functionEmitter{f: e.f}
	region := memoryRegion{end: -1}
	for pc, literal := range e.f.Code {
		if next, ok := regions[pc]; ok {
			region = next
		} else if pc > region.end {
			region = memoryRegion{end: -1}
		}
		_, boundary := regions[pc+1]
		if (literal.Op != 0x41 && literal.Op != 0x42) || literal.Dst == slots.NoSlot || pc+1 >= len(e.f.Code) || boundary ||
			pc/256 != (pc+1)/256 || region.elided[pc] || region.elided[pc+1] ||
			region.end >= pc && pc+1 > region.end {
			continue
		}
		in := e.f.Code[pc+1]
		if in.B != literal.Dst || in.A == literal.Dst {
			continue
		}
		bits := e.f.Sources[literal.Source].Wasm.Bits
		_, arithmetic := x86ImmediateGroup(in.Op, bits)
		if !arithmetic && !(in.Op >= 0x74 && in.Op <= 0x78 || in.Op >= 0x86 && in.Op <= 0x8a) {
			continue
		}
		weights[literal.Dst] -= frequency(frequencies, pc+1) - 1
		scan.regionEnd = 0
		if region.end >= pc {
			scan.regionEnd = region.end + 1
		}
		scan.provenPCs, scan.guardPCs = region.proven, region.guards
		scan.addresses, scan.elidedPCs = region.addresses, region.elided
		if literal.Dst == in.Dst || scan.overwrittenBeforeExit(pc+2, literal.Dst) {
			weights[literal.Dst] -= frequency(frequencies, pc) - 1
		}
	}
}
