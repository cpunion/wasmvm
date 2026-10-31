package canonical

// AddOpaque allocates a non-resource entry in the same bounded, tagged handle
// domain. Async waitables and sets cannot alias resources or error contexts.
func (r *Resources) AddOpaque(value any) (uint32, error) {
	if value == nil {
		return 0, invalid("missing opaque handle value")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.addLocked(&resourceEntry{opaque: value, owned: true})
}

func (r *Resources) Opaque(id uint32) (any, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e := r.entries[id]
	if r.closed || e == nil || e.opaque == nil {
		return nil, scalarTrap("invalid async handle")
	}
	return e.opaque, nil
}

func (r *Resources) RemoveOpaque(id uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e := r.entries[id]
	if r.closed || e == nil || e.opaque == nil {
		return scalarTrap("invalid async handle")
	}
	delete(r.entries, id)
	r.budget.release(1)
	return nil
}

// MoveOpaque transfers a handle between same-budget domains without transiently
// consuming another live slot. Table creation order gives all transfers a
// consistent lock order, including opposite-direction concurrent callers.
func (r *Resources) MoveOpaque(id uint32, dst *Resources) (uint32, error) {
	if r == nil || dst == nil || r == dst || r.budget != dst.budget {
		return 0, invalid("opaque transfer domain")
	}
	first, second := r, dst
	if first.order > second.order {
		first, second = second, first
	}
	first.mu.Lock()
	defer first.mu.Unlock()
	second.mu.Lock()
	defer second.mu.Unlock()
	e := r.entries[id]
	if r.closed || dst.closed || e == nil || e.opaque == nil {
		return 0, scalarTrap("invalid async handle")
	}
	if dst.next >= 1<<28 {
		return 0, exhausted()
	}
	next := dst.next
	dst.next++
	e.id = next
	dst.entries[next] = e
	delete(r.entries, id)
	return next, nil
}
