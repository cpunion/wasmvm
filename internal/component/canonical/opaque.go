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
