package codegen

import "github.com/cpunion/wasmvm/pkg/wasm"

// The constrained x86 lowering needs RDX/R8 only for select/popcount in this
// integer subset. Proven default-memory accesses also need no length register.
// Borrowing is confined to a nontrapping native loop; ordinary continuations
// retain the complete constrained inventory. Unknown operations lend nothing.
func (e *functionEmitter) loopAuxiliaryRegisters(start int, r memoryRegion) []byte {
	if e.a.arch != "amd64" {
		return nil
	}
	available := uint16(1<<2 | 1<<8)
	if e.usesMemory && (!e.memoryViews || e.sharedDefaultMemory) {
		available |= 1 << 11
	}
	for pc := start; pc <= r.end; pc++ {
		if r.elided[pc] {
			continue
		}
		in := e.f.Code[pc]
		for _, s := range []uint32{uint32(in.Dst), uint32(in.A), uint32(in.B), uint32(in.C)} {
			if s < uint32(len(e.f.Slots)) && e.f.Slots[s] != wasm.I32 && e.f.Slots[s] != wasm.I64 {
				return nil
			}
		}
		switch op := in.Op; {
		case op == 1 || op == 0x1a || op >= 0x20 && op <= 0x22 || op == 0x41 || op == 0x42:
		case op == 0x1b || op == 0x1c:
			available &^= 1 << 2
		case op >= 0x45 && op <= 0x5a || op >= 0x67 && op <= 0x8a:
			if scalarCanTrap(op) {
				return nil
			}
			if op == 0x69 || op == 0x7b {
				available &^= 1<<2 | 1<<8 // Baseline SWAR popcount.
			}
		case op == 0xa7 || op == 0xac || op == 0xad || op >= 0xc0 && op <= 0xc4:
		case isMemoryAccess(op) && r.proven[pc]:
		default:
			return nil
		}
	}
	var pool []byte
	for _, reg := range []byte{2, 8, 11} {
		if available&(1<<reg) != 0 {
			pool = append(pool, reg)
		}
	}
	return pool
}

func (e *functionEmitter) loopBorrowedMemoryLength() bool {
	if e.a.arch == "amd64" && e.usesMemory {
		for _, reg := range e.registers {
			if reg == 11 {
				return true
			}
		}
	}
	return false
}
