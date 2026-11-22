package slots

import "github.com/cpunion/wasmvm/internal/compiler/ir"

// LocalI32Update marks a load followed by a literal, arithmetic, and store.
// It does not grant access to a memory or affect the original PCs.
const LocalI32Update = 3

// Keep original PCs and semantics. A coalesced local read changes no value;
// constants can supply an adjacent pure or atomic consumer. Unmetered
// interpretation commits each prefix before dispatching the consumer.
func cachePrefixes(f *Function, budget *budget) {
	budget.charge(uint64(len(f.Code)))
	for pc := 0; pc+1 < len(f.Code); pc++ {
		in, next := &f.Code[pc], f.Code[pc+1]
		if in.Op == 0x28 && pc+LocalI32Update < len(f.Code) && cacheLocalI32Update(f, pc) {
			in.Count = LocalI32Update
			// Preserve both address writes while dispatching their common
			// consumer. No other prefix is needed for this group.
			if pc > 0 {
				addr := &f.Code[pc-1]
				if addr.Op == 0x41 && addr.Dst == in.A {
					addr.Count = 1
					if pc > 1 {
						dest := &f.Code[pc-2]
						if dest.Op == 0x41 && dest.Dst == f.Code[pc+LocalI32Update].A {
							dest.Count = 1
						}
					}
				}
			}
			continue
		}
		if in.Op == 0x20 && in.Dst != NoSlot && uint32(in.Dst) == in.Extra {
			in.Count = 1
			continue
		}
		if in.Op >= 0x45 && in.Op <= 0x5a && in.Dst != NoSlot &&
			next.Op == ControlBase+uint32(ir.If) && next.A == in.Dst {
			in.Count = 1
			continue
		}
		if in.Op == 0x41 && in.Dst != NoSlot && next.Op>>16 == 0xfe &&
			(next.A == in.Dst || next.B == in.Dst) {
			in.Count = 1
			if pc > 0 && next.B == in.Dst && next.A != in.Dst {
				prev := &f.Code[pc-1]
				if prev.Op == 0x41 && prev.Dst != NoSlot && prev.Dst == next.A {
					prev.Count = 1
				}
			}
			continue
		}
		if in.Op != 0x41 || in.Dst == NoSlot || next.B != in.Dst || next.A == NoSlot || next.A == in.Dst {
			continue
		}
		op := next.Op
		if op >= 0x46 && op <= 0x4f || op == 0x6a || op == 0x6b || op == 0x6c || op >= 0x71 && op <= 0x78 {
			in.Count = 1
		}
	}
}

func cacheLocalI32Update(f *Function, pc int) bool {
	seq := f.Code[pc : pc+LocalI32Update+1]
	load, literal, compute, store := &seq[0], &seq[1], &seq[2], &seq[3]
	if literal.Op != 0x41 || store.Op != 0x36 || load.Dst == NoSlot || literal.Dst == NoSlot ||
		compute.Dst == NoSlot || compute.A != load.Dst || compute.B != literal.Dst ||
		compute.A == compute.B || store.B != compute.Dst ||
		store.A == load.Dst || store.A == literal.Dst || store.A == compute.Dst {
		return false
	}
	switch compute.Op {
	case 0x6a, 0x6b, 0x6c, 0x71, 0x72, 0x73:
	default:
		return false
	}
	read, write := &f.Sources[load.Source].Wasm, &f.Sources[store.Source].Wasm
	return read.Memory == write.Memory && read.MemoryOffset() == write.MemoryOffset()
}
