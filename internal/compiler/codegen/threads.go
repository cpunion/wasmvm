package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Mark only shared-memory accesses before optimization. Every other instruction
// still emits normally; memory guards, fusion and raw descriptors cannot bypass
// the synchronization domain. The runtime retains the unchanged original code.
func sharedMemoryHelpers(f *slots.Function, shared []bool) *slots.Function {
	enabled := false
	for _, v := range shared {
		enabled = enabled || v
	}
	if !enabled {
		return f
	}
	var copyFunction *slots.Function
	for pc, in := range f.Code {
		if !isMemoryAccess(in.Op) && in.Op != 0x3f && in.Op != 0x40 {
			continue
		}
		source := f.Sources[in.Source].Wasm
		index := source.Memory
		if in.Op == 0x3f || in.Op == 0x40 {
			index = source.A
		}
		if uint64(index) >= uint64(len(shared)) || !shared[index] {
			continue
		}
		if copyFunction == nil {
			g := *f
			g.Code = append([]slots.Instruction(nil), f.Code...)
			copyFunction = &g
		}
		copyFunction.Code[pc].Op = slots.SharedMemoryOp
	}
	if copyFunction != nil {
		return copyFunction
	}
	return f
}
