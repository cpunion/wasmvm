package canonical

import (
	"sync"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Reference is an opaque host resource capability. Copies share one state;
// moving any copy invalidates every alias of that capability.
type Reference struct{ state *referenceState }

type referenceState struct {
	mu     sync.Mutex
	valid  bool
	table  *Resources
	typeID *ResourceType
	id     uint32
	owner  any
	scope  *BorrowScope
}

func NewReference(table *Resources, typeID *ResourceType, id uint32, owner any, scope *BorrowScope) *Reference {
	return &Reference{state: &referenceState{valid: true, table: table, typeID: typeID, id: id, owner: owner, scope: scope}}
}

// ReferenceInfo is internal binding machinery; it is deliberately not a method
// exposed by the public opaque capability alias.
func ReferenceInfo(r *Reference) (*Resources, *ResourceType, uint32, any, bool, error) {
	if r == nil || r.state == nil {
		return nil, nil, 0, nil, false, wasm.ErrInvalid
	}
	s := r.state
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.valid || s.scope != nil && !s.scope.Active() {
		return nil, nil, 0, nil, false, scalarTrap("expired resource capability")
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
	if r == nil || r.state == nil {
		return nil, wasm.ErrInvalid
	}
	s := r.state
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.valid || s.scope != nil {
		return nil, scalarTrap("cannot move expired or borrowed resource")
	}
	s.valid = false
	return NewReference(s.table, s.typeID, s.id, s.owner, nil), nil
}

func InvalidateReference(r *Reference) {
	if r != nil && r.state != nil {
		r.state.mu.Lock()
		r.state.valid = false
		r.state.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	return !s.ended
}
