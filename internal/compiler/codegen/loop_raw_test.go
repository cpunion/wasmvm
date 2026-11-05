package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLoopRawRegisterPlanBanksAndEligibility(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, mode := range []string{"select", "proof", "no-values", "no-plan", "variable", "full-bank", "deferred", "untracked", "cold"} {
			e := functionEmitter{a: &assembler{arch: arch}, f: &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I64, wasm.F64, wasm.V128, wasm.ExternRef, wasm.I32, wasm.I64}}}
			for range 5 {
				e.f.Code = append(e.f.Code, slots.Instruction{Op: 0x7c, Dst: 1, A: 1, B: 5, C: slots.NoSlot})
			}
			fp := e.floatRegisterPool()
			old := &registerPlan{homes: map[slots.Slot]byte{0: e.integerRegisterPool()[0]}, floats: map[slots.Slot]byte{2: fp[0]}, vectors: map[slots.Slot]byte{3: fp[1]}, raw: map[slots.Slot]byte{6: fp[2]}, slots: []slots.Slot{0, 2, 3, 6}}
			e.useRegisterPlan(old)
			e.vectorConstants = map[[2]uint64]byte{{1, 2}: fp[3]}
			v := &loopValues{bySlot: map[slots.Slot][]int{}}
			for s := range e.f.Slots {
				v.bySlot[slots.Slot(s)] = []int{0}
			}
			v.bySlot[slots.NoSlot] = []int{0}
			r := memoryRegion{end: len(e.f.Code) - 2}
			var proof *memoryLoop
			switch mode {
			case "proof":
				proof = &memoryLoop{limit: 64, stride: 4}
			case "no-values":
				v = nil
			case "no-plan":
				e.useRegisterPlan(nil)
			case "variable":
				r.values = &loopValuePlans{}
			case "full-bank":
				for i, reg := range fp {
					e.vectorConstants[[2]uint64{uint64(i), 3}] = reg
				}
			case "deferred":
				r.copies = &loopFrameCopies{deferred: map[slots.Slot]bool{5: true}}
			case "untracked":
				delete(v.bySlot, 1)
				delete(v.bySlot, 5)
			case "cold":
				e.f.Code = e.f.Code[:1]
				r.end = -1
			}
			plan := e.loopRawRegisterPlan(v, r, proof)
			if mode != "select" && mode != "proof" && mode != "deferred" {
				if plan != nil {
					t.Fatal(arch, mode, "ineligible raw plan")
				}
				continue
			}
			if plan == nil || plan.raw[1] != fp[4] || len(old.raw) != 1 || len(old.slots) != 4 {
				t.Fatal(arch, mode, "lost independent or deterministic homes", plan, old)
			}
			want := 3
			if mode == "deferred" {
				want = 2
			}
			if len(plan.raw) != want || want == 3 && plan.raw[5] != fp[5] {
				t.Fatal(arch, mode, "wrong raw selection", plan.raw)
			}
			for s, reg := range plan.raw {
				if s != 6 && s != 1 && (s != 5 || mode == "deferred") || reg < fp[4] && s != 6 {
					t.Fatal(arch, mode, "overlapping bank or invalid raw slot", s, reg)
				}
			}
		}
	}
}
