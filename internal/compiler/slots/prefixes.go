package slots

import "github.com/cpunion/wasmvm/internal/compiler/ir"

// Keep original PCs and semantics. A coalesced local read changes no value;
// constants can supply an adjacent pure or atomic consumer. Unmetered
// interpretation commits each prefix before dispatching the consumer.
func cachePrefixes(f *Function, budget *budget) {
	budget.charge(uint64(len(f.Code)))
	for pc := 0; pc+1 < len(f.Code); pc++ {
		in, next := &f.Code[pc], f.Code[pc+1]
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
