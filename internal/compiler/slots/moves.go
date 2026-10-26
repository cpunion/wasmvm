package slots

// parallelMoves orders simultaneous assignments in linear expected time.
// Destinations are unique and types agree (guaranteed by verified SSA edges).
// scratch returns a disjoint slot of the destination's type. Cycles use one
// temporary each; callers may reuse the same typed scratch between cycles.
func parallelMoves(input []Move, scratch func(Slot) Slot) []Move {
	pending := make(map[Slot]Slot, len(input))
	used := make(map[Slot]int, len(input))
	for _, m := range input {
		if m.Dst != m.Src {
			pending[m.Dst] = m.Src
			used[m.Src]++
		}
	}
	queue := make([]Slot, 0, len(pending))
	out := make([]Move, 0, len(pending))
	for _, m := range input {
		if _, ok := pending[m.Dst]; ok && used[m.Dst] == 0 {
			queue = append(queue, m.Dst)
		}
	}
	drain := func() {
		for len(queue) > 0 {
			dst := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			src := pending[dst]
			out = append(out, Move{dst, src})
			delete(pending, dst)
			used[src]--
			if used[src] == 0 {
				if _, ok := pending[src]; ok {
					queue = append(queue, src)
				}
			}
		}
	}
	drain()
	// What remains consists only of cycles: every destination is also a source.
	// Walk each cycle backwards, replacing its saved endpoint by scratch.
	for _, m := range input {
		if _, ok := pending[m.Dst]; !ok {
			continue
		}
		start := m.Dst
		tmp := scratch(start)
		out = append(out, Move{tmp, start})
		dst := start
		for {
			src := pending[dst]
			delete(pending, dst)
			if src == start {
				out = append(out, Move{dst, tmp})
				break
			}
			out = append(out, Move{dst, src})
			dst = src
		}
	}
	return out
}
