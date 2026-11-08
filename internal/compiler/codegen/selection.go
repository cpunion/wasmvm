package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// One ordered selector pipeline serves ordinary code and bounded native regions.
// Callers establish mapping, guard, remapping and shared-write boundaries before
// selection, then retain an independent continuation for each consumed PC.
func (e *functionEmitter) combinePlain(pc, start int, out *CompiledFunction) (int, bool) {
	in := e.f.Code[pc]
	if end, ok := e.combineVectorAffine(pc); ok {
		out.FusedVectors++
		return end, true
	}
	if end, ok := e.combineMaskShift(pc); ok {
		out.FusedIntegers++
		return end, true
	}
	if end, ok := e.combineShiftedInteger(pc); ok {
		out.FusedIntegers++
		return end, true
	}
	if e.logicalShift(pc, in) {
		out.FusedIntegers++
		return pc + 2, true
	}
	if e.regionEnd != 0 {
		if end, ok := e.combineDeadImmediate(pc); ok {
			out.ForwardedScalars++
			return end, true
		}
		if pc > start {
			prev := e.f.Code[pc-1]
			if (prev.Op == 0x41 || prev.Op == 0x42) && prev.Dst == in.B && in.B != slots.NoSlot && e.registerImmediate(in.Op, in.Dst, in.A, e.f.Sources[prev.Source].Wasm.Bits) {
				return pc, true
			}
		}
	}
	if end, ok := e.combineScalarCopies(pc); ok {
		out.ForwardedScalars++
		return end, true
	}
	if end, ok := e.forwardHeapStoreLoad(pc); ok {
		out.NativeHeap += 2
		out.ForwardedHeap++
		return end, true
	}
	if end, ok := e.forwardMemoryStoreLoad(pc, in); ok {
		return end, true
	}
	if end, ok := e.combineI31(pc); ok {
		out.FusedIntegers++
		return end, true
	}
	if end, ok := e.combineMultiplyAdd(pc); ok {
		out.FusedIntegers++
		return end, true
	}
	if end, ok := e.combineIntegerConstant(pc); ok {
		out.FusedIntegers++
		return end, true
	}
	if e.forwardScalarResult(pc) {
		out.ForwardedScalars++
		return pc + 1, true
	}
	return 0, false
}
