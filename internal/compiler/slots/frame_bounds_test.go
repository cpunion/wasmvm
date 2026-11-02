package slots

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestFrameProofCoversEachInputAndOutput(t *testing.T) {
	for _, size := range []int{0, 1, 17} {
		for field := 0; field < 4; field++ {
			for _, slot := range []Slot{NoSlot, 0, Slot(size), NoSlot - 1} {
				in := Instruction{Dst: NoSlot, A: NoSlot, B: NoSlot, C: NoSlot}
				fields := [...]*Slot{&in.Dst, &in.A, &in.B, &in.C}
				*fields[field] = slot
				f := &Function{Slots: make([]wasm.ValueType, size), Code: []Instruction{in}}
				budget := ir.VerifyBudget{Remaining: 4}
				err := verifyFrameAccesses(f, &budget)
				valid := slot == NoSlot || uint64(slot) < uint64(size)
				if valid && err != nil || !valid && !errors.Is(err, wasm.ErrInvalid) {
					t.Fatal(size, field, slot, err)
				}
				if budget.Remaining != 0 {
					t.Fatal("unchecked verification work", budget.Remaining)
				}
			}
		}
	}
}

func TestFrameProofHasAggregateWorkLimit(t *testing.T) {
	f := &Function{Code: []Instruction{{Dst: NoSlot, A: NoSlot, B: NoSlot, C: NoSlot}}}
	budget := ir.VerifyBudget{Remaining: 7}
	if err := verifyFrameAccesses(f, &budget); err != nil {
		t.Fatal(err)
	}
	if err := verifyFrameAccesses(f, &budget); !errors.Is(err, wasm.ErrLimit) || budget.Remaining != 3 {
		t.Fatal(budget, err)
	}
}
