package codegen

import "sort"

type registerLoop struct{ first, last int }

func enclosingRegisterLoop(loops []registerLoop, pc int) registerLoop {
	first, last := 0, len(loops)
	for first < last {
		mid := first + (last-first)/2
		if loops[mid].first <= pc {
			first = mid + 1
		} else {
			last = mid
		}
	}
	if first != 0 && pc <= loops[first-1].last {
		return loops[first-1]
	}
	return registerLoop{pc, pc}
}

// Keep residency stable through a backedge, including its branch block.
// Merge crossing and nested spans first, so extending an endpoint requires
// only one lookup rather than repeatedly traversing loops to a fixed point.
func extendRegisterLoops(intervals []registerInterval, loops []registerLoop) {
	if len(loops) == 0 {
		return
	}
	sort.Slice(loops, func(i, j int) bool {
		if loops[i].first != loops[j].first {
			return loops[i].first < loops[j].first
		}
		return loops[i].last < loops[j].last
	})
	merged := loops[:0]
	for _, loop := range loops {
		if len(merged) != 0 && loop.first <= merged[len(merged)-1].last {
			merged[len(merged)-1].last = max(merged[len(merged)-1].last, loop.last)
		} else {
			merged = append(merged, loop)
		}
	}
	for i := range intervals {
		v := &intervals[i]
		if v.first < 0 {
			continue
		}
		v.first = enclosingRegisterLoop(merged, v.first).first
		v.last = enclosingRegisterLoop(merged, v.last).last
	}
}
