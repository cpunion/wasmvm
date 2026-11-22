package execute

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// All operands, intermediate writes, and underlying memory effects belong to
// the current Store lease. A mismatch resumes ordinary per-opcode dispatch.
func updateLocalI32(instance *state.Instance, code *slots.Function, values []value.Value, pc uint32) (bool, wasm.TrapCode) {
	seq := code.Code[pc : pc+slots.LocalI32Update+1]
	load, literal, compute, store := &seq[0], &seq[1], &seq[2], &seq[3]
	address := values[load.A].Lo
	if address != values[store.A].Lo {
		return false, wasm.NoTrap
	}
	im := &code.Sources[load.Source].Wasm
	immediate := uint64(literal.Extra)
	old, next, trap, applied := instance.Memories[im.Memory].UpdateLocalI32(address, im.MemoryOffset(), compute.Op, immediate)
	if !applied || trap != wasm.NoTrap {
		return applied, trap
	}
	values[load.Dst] = value.Value{Lo: old}
	values[literal.Dst] = value.Value{Lo: immediate}
	values[compute.Dst] = value.Value{Lo: next}
	return true, wasm.NoTrap
}
