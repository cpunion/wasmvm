package canonical

import (
	"sync/atomic"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Reference is an opaque value reference. Resource capabilities share one
// affine state; moving any alias invalidates all copies. Error contexts instead
// carry immutable, copyable diagnostics. Value retains one reference slot.
type Reference struct {
	state        *referenceState
	errorContext *errorContext
}

type referenceState struct {
	// The binding never changes. Moving or invalidating any alias only clears
	// this shared affine bit; a successful move publishes a fresh binding.
	valid  atomic.Bool
	table  *Resources
	typeID *ResourceType
	id     uint32
	owner  any
	scope  *BorrowScope
}

// The header and its affine state share storage. Copied headers retain the
// interior state pointer, so every alias still observes the same lifetime.
type referenceAllocation struct {
	reference Reference
	state     referenceState
}

func NewReference(table *Resources, typeID *ResourceType, id uint32, owner any, scope *BorrowScope) *Reference {
	return initializeReference(&referenceAllocation{}, table, typeID, id, owner, scope)
}

func initializeReference(allocation *referenceAllocation, table *Resources, typeID *ResourceType, id uint32, owner any, scope *BorrowScope) *Reference {
	allocation.state = referenceState{table: table, typeID: typeID, id: id, owner: owner, scope: scope}
	allocation.state.valid.Store(true)
	allocation.reference.state = &allocation.state
	return &allocation.reference
}

// ReferenceAllocator groups fresh capabilities for one serialized owner. Slots
// are consumed once: moving, invalidating or clearing the allocator never
// rebinds a published header or affine state. Retained header copies keep their
// own block alive. This is allocation grouping, not capability pooling.
// Blocks stay <= 368 B so BDWGC serves them from lock-free thread-local freelists.
type ReferenceAllocator struct {
	_     referenceAllocatorNoCopy
	block *[5]referenceAllocation
	next  uint8
}

type referenceAllocatorNoCopy struct{}

func (*referenceAllocatorNoCopy) Lock()   {}
func (*referenceAllocatorNoCopy) Unlock() {}

func (a *ReferenceAllocator) New(table *Resources, typeID *ResourceType, id uint32, owner any, scope *BorrowScope) *Reference {
	if a.block == nil && a.next == 0 {
		// An owner creating only one capability needs no partially used block.
		a.next = 5
		return NewReference(table, typeID, id, owner, scope)
	}
	if a.block == nil || a.next == uint8(len(a.block)) {
		a.block = new([5]referenceAllocation)
		a.next = 0
	}
	slot := &a.block[a.next]
	a.next++
	return initializeReference(slot, table, typeID, id, owner, scope)
}

// Clear drops only the owner's allocation cursor, after its calls have drained.
// Already published capabilities, including copied headers, remain independent.
func (a *ReferenceAllocator) Clear() { a.block, a.next = nil, 0 }

// Affine cleanup commonly inspects aliases invalidated by a successful move.
// This immutable diagnostic carries no per-reference or per-call state.
var errExpiredReference = scalarTrap("expired resource capability")

// ReferenceInfo is internal binding machinery; it is deliberately not a method
// exposed by the public opaque capability alias.
func ReferenceInfo(r *Reference) (*Resources, *ResourceType, uint32, any, bool, error) {
	if r == nil || r.state == nil {
		return nil, nil, 0, nil, false, wasm.ErrInvalid
	}
	s := r.state
	if !s.valid.Load() || s.scope != nil && !s.scope.Active() {
		return nil, nil, 0, nil, false, errExpiredReference
	}
	return s.table, s.typeID, s.id, s.owner, s.scope != nil, nil
}

func ReferenceKey(r *Reference) any {
	if r == nil {
		return nil
	}
	return r.state
}

func MoveReference(r *Reference) (*Reference, error) {
	return moveReference(r, nil)
}

func (a *ReferenceAllocator) Move(r *Reference) (*Reference, error) {
	return moveReference(r, a)
}

func moveReference(r *Reference, allocator *ReferenceAllocator) (*Reference, error) {
	if r == nil || r.state == nil {
		return nil, wasm.ErrInvalid
	}
	s := r.state
	if s.scope != nil || !s.valid.CompareAndSwap(true, false) {
		return nil, scalarTrap("cannot move expired or borrowed resource")
	}
	if allocator != nil {
		return allocator.New(s.table, s.typeID, s.id, s.owner, nil), nil
	}
	return NewReference(s.table, s.typeID, s.id, s.owner, nil), nil
}

func InvalidateReference(r *Reference) {
	if r != nil && r.state != nil {
		r.state.valid.Store(false)
	}
}

// CheckOwn validates a pending transfer without consuming it. Calls serialize
// movement with execution; outstanding lends prohibit both movement and drop.
func (r *Resources) CheckOwn(t *ResourceType, id uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, err := r.getLocked(t, id)
	if err != nil {
		return err
	}
	if !e.owned || e.lends != 0 {
		return scalarTrap("resource is borrowed")
	}
	return nil
}

func (s *BorrowScope) Active() bool {
	return !s.ended
}
