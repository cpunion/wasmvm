package codegen

import "testing"

func TestRegisterLoopClosureIncludesCrossingAndNestedBackedges(t *testing.T) {
	loops := []registerLoop{{6, 8}, {2, 5}, {3, 4}, {4, 6}, {10, 11}}
	for first := 0; first <= 12; first++ {
		for last := first; last <= 12; last++ {
			wantFirst, wantLast := first, last
			for changed := true; changed; {
				changed = false
				for _, loop := range loops {
					if loop.first <= wantLast && loop.last >= wantFirst {
						a, b := min(wantFirst, loop.first), max(wantLast, loop.last)
						changed = changed || a != wantFirst || b != wantLast
						wantFirst, wantLast = a, b
					}
				}
			}
			intervals := []registerInterval{{0, first, last, 0}, {1, -1, -1, 0}}
			extendRegisterLoops(intervals, append([]registerLoop(nil), loops...))
			if intervals[0].first != wantFirst || intervals[0].last != wantLast || intervals[1].first != -1 {
				t.Fatal(first, last, intervals, wantFirst, wantLast)
			}
		}
	}
	intervals := []registerInterval{{0, 1, 2, 0}}
	extendRegisterLoops(intervals, nil)
	if intervals[0].first != 1 || intervals[0].last != 2 {
		t.Fatal("acyclic interval changed")
	}
}

func TestLinearScanKeepsLoopHeaderBodyAndBackedgeInOneResidencyPlan(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, always := range []bool{false, true} {
			f := coalescedArithmeticLoopFixture(false, true, always)
			branch := len(f.Code) - 2
			e := functionEmitter{f: f, a: &assembler{arch: arch}, unmetered: true}
			e.prepareLinearScan(map[int]memoryRegion{1: {end: branch - 1}}, 1<<18)
			if len(e.allocationPlans) != len(f.Code) {
				t.Fatal("missing residency plans")
			}
			for pc := 1; pc <= branch; pc++ {
				if !equalRegisterPlans(e.allocationPlans[1], e.allocationPlans[pc]) {
					t.Fatal(arch, always, pc, "loop contains a spill/reload transition")
				}
			}
		}
	}
}
