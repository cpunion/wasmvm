package slots

import (
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// The interpreter accesses these four fields without repeated slice checks.
// Compile verifies every emitted instruction after allocation and rewriting;
// published instruction and slot metadata must remain immutable.
func verifyFrameAccesses(f *Function, budget *ir.VerifyBudget) error {
	work := uint64(len(f.Code)) * 4
	if work > budget.Remaining {
		return &wasm.Error{Kind: wasm.ErrLimit, Detail: "frame slot verification work"}
	}
	budget.Remaining -= work
	limit := uint64(len(f.Slots))
	for _, in := range f.Code {
		for _, slot := range [...]Slot{in.Dst, in.A, in.B, in.C} {
			if slot != NoSlot && uint64(slot) >= limit {
				return &wasm.Error{Kind: wasm.ErrInvalid, Detail: "compiled frame slot bounds"}
			}
		}
	}
	return nil
}
