package slots

type movePlanner struct {
	pending map[Slot]Slot
	used    map[Slot]int
	queue   []Slot
	out     []Move
}

// parallelMoves orders simultaneous assignments in linear expected time.
// Destinations are unique and types agree (guaranteed by verified SSA edges).
// scratch returns a disjoint slot of the destination's type. Cycles use one
// temporary each; callers may reuse the same typed scratch between cycles.
func parallelMoves(input []Move, scratch func(Slot) Slot) []Move {
	var planner movePlanner
	ordered := planner.order(input, scratch)
	return append([]Move(nil), ordered...)
}

func (p *movePlanner) order(input []Move, scratch func(Slot) Slot) []Move {
	if len(input) == 0 {
		return nil
	}
	if len(input) == 1 {
		if input[0].Dst == input[0].Src {
			return nil
		}
		p.out = append(p.out[:0], input[0])
		return p.out
	}
	if p.pending == nil {
		p.pending = make(map[Slot]Slot, len(input))
		p.used = make(map[Slot]int, len(input))
	} else {
		clear(p.pending)
		clear(p.used)
	}
	pending, used := p.pending, p.used
	for _, m := range input {
		if m.Dst != m.Src {
			pending[m.Dst] = m.Src
			used[m.Src]++
		}
	}
	if len(pending) == 0 {
		return nil
	}
	queue := p.queue[:0]
	out := p.out[:0]
	for _, m := range input {
		if _, ok := pending[m.Dst]; ok && used[m.Dst] == 0 {
			queue = append(queue, m.Dst)
		}
	}
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
	p.queue, p.out = queue, out
	return out
}
