package slots

// Keep both original PCs and semantics. Only uninterrupted unmetered execution
// can use this hint; native and precise-fuel execution retain their own paths.
func cacheIntegerPairs(f *Function, budget *budget) {
	budget.charge(uint64(len(f.Code)))
	for pc := 0; pc+1 < len(f.Code); pc++ {
		in, next := &f.Code[pc], f.Code[pc+1]
		if in.Op != 0x41 || in.Dst == NoSlot || next.B != in.Dst || next.A == NoSlot || next.A == in.Dst {
			continue
		}
		op := next.Op
		if op >= 0x46 && op <= 0x4f || op == 0x6a || op == 0x6b || op == 0x6c || op >= 0x71 && op <= 0x78 {
			in.Count = 1
		}
	}
}
