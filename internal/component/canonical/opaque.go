package canonical

import "github.com/cpunion/wasmvm/pkg/wasm"

// AddOpaque allocates a non-resource entry in the same bounded, tagged handle
// domain. Async waitables and sets cannot alias resources or error contexts.
func (r *Resources) AddOpaque(value any) (uint32, error) {
	if value == nil {
		return 0, invalid("missing opaque handle value")
	}
	r.mu.Lock()
	e, id, err := r.addLocked()
	if err == nil {
		e.opaque, e.owned = value, true
	}
	r.mu.Unlock()
	return id, err
}

func (r *Resources) Opaque(id uint32) (any, error) {
	r.mu.Lock()
	// Closing the owner can race an already admitted canonical operation.
	// Match resource/error-context domains: shutdown is not a guest trap.
	if r.closed {
		r.mu.Unlock()
		return nil, wasm.ErrClosed
	}
	e := r.entryLocked(id)
	if e == nil || e.opaque == nil {
		r.mu.Unlock()
		return nil, scalarTrap("invalid async handle")
	}
	v := e.opaque
	r.mu.Unlock()
	return v, nil
}

func (r *Resources) RemoveOpaque(id uint32) error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return wasm.ErrClosed
	}
	e := r.entryLocked(id)
	if e == nil || e.opaque == nil {
		r.mu.Unlock()
		return scalarTrap("invalid async handle")
	}
	r.removeLocked(id)
	r.budget.release(1)
	r.mu.Unlock()
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
	second.mu.Lock()
	if r.closed || dst.closed {
		second.mu.Unlock()
		first.mu.Unlock()
		return 0, wasm.ErrClosed
	}
	e := r.entryLocked(id)
	if e == nil || e.opaque == nil {
		second.mu.Unlock()
		first.mu.Unlock()
		return 0, scalarTrap("invalid async handle")
	}
	if !dst.availableLocked() {
		second.mu.Unlock()
		first.mu.Unlock()
		return 0, exhausted()
	}
	next := dst.indexLocked()
	e.id, e.generation = next, dst.generation
	dst.insertLocked(next, e)
	r.removeLocked(id)
	second.mu.Unlock()
	first.mu.Unlock()
	return next, nil
}
