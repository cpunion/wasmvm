package codegen

import (
	"slices"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (e *functionEmitter) floatRegisterPool() []byte {
	if e.a.arch == "arm64" {
		return []byte{8, 9, 10, 11, 12, 13, 14, 15, 16, 17}
	}
	return []byte{3, 4, 5, 6, 7, 8, 9, 10}
}

func registerBankPlan(f *slots.Function, integers, floats *scanBank, pinned *registerPlan) *registerPlan {
	plan := &registerPlan{homes: map[slots.Slot]byte{}, floats: map[slots.Slot]byte{}, vectors: map[slots.Slot]byte{}}
	if pinned != nil {
		plan = cloneRegisterPlan(pinned)
	}
	for i, resident := range integers.resident {
		if resident.valid {
			s := resident.interval.slot
			plan.homes[s] = integers.pool[i]
			plan.slots = append(plan.slots, s)
		}
	}
	for i, resident := range floats.resident {
		if resident.valid {
			s := resident.interval.slot
			if f.Slots[s] == wasm.V128 {
				plan.vectors[s] = floats.pool[i]
			} else {
				plan.floats[s] = floats.pool[i]
			}
			plan.slots = append(plan.slots, s)
		}
	}
	slices.Sort(plan.slots)
	return plan
}

// Initial residency prefers existing physical homes. Subsequent expiration
// and cost eviction use the same typed scan as ordinary global allocation.
func (b *scanBank) seed(plan *registerPlan, initial []registerInterval, floating bool) []registerInterval {
	b.resident = make([]registerResident, len(b.pool))
	pending := initial[:0]
	for _, interval := range initial {
		kind, reg := registerLocation(plan, interval.slot)
		if kind != 0 && (kind != 1) == floating {
			if index := slices.Index(b.pool, reg); index >= 0 {
				b.resident[index] = registerResident{interval: interval, valid: true}
				continue
			}
		}
		pending = append(pending, interval)
	}
	return pending
}
