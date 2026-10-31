package canonical

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// ResourceType is a nominal identity created for one instantiated resource
// declaration. Two declarations with the same representation remain distinct.
type ResourceType struct {
	Drop     func(context.Context, uint64) error
	identity byte
}

type resourceEntry struct {
	id           uint32
	typeID       *ResourceType
	rep          uint64
	owned        bool
	scope        *BorrowScope
	lends        uint32
	errorContext *errorContext
}

// ResourceBudget bounds live handles across all domains of a Store. Table IDs
// remain local, while allocation in one component consumes the shared quota.
type ResourceBudget struct {
	live                   atomic.Uint32
	maximum                uint32
	errorMu                sync.Mutex
	errorLimit, errorBytes uint64
	errors                 map[*errorContext]uint32
}

func NewResourceBudget(maximum uint32) *ResourceBudget {
	return NewResourceBudgetWithErrors(maximum, ^uint64(0))
}

// NewResourceBudgetWithErrors additionally bounds unique error-message bytes
// retained by guest handles across all namespaces in a Store.
func NewResourceBudgetWithErrors(maximum uint32, bytes uint64) *ResourceBudget {
	return &ResourceBudget{maximum: maximum, errorLimit: bytes}
}

func (b *ResourceBudget) acquire() error {
	for {
		live := b.live.Load()
		if live >= b.maximum {
			return exhausted()
		}
		if b.live.CompareAndSwap(live, live+1) {
			return nil
		}
	}
}

func (b *ResourceBudget) release(count uint32) { b.live.Add(0 - count) }

// Resources is an instance's handle domain. IDs are never reused, so stale
// handles cannot gain authority over a later resource. It is bounded by both
// live entries and the pinned ABI's 28-bit handle space.
type Resources struct {
	mu         sync.Mutex
	entries    map[uint32]*resourceEntry
	next       uint32
	budget     *ResourceBudget
	closed     bool
	errorCount uint32
}

func NewResources(maximum uint32) *Resources {
	return NewResourceTable(NewResourceBudget(maximum))
}

// NewResourceTable creates an independent handle domain with a shared quota.
func NewResourceTable(budget *ResourceBudget) *Resources {
	return &Resources{entries: map[uint32]*resourceEntry{}, next: 1, budget: budget}
}

func (r *Resources) addLocked(e *resourceEntry) (uint32, error) {
	if r.closed {
		return 0, wasm.ErrClosed
	}
	if r.next >= (1 << 28) {
		return 0, exhausted()
	}
	if err := r.budget.acquire(); err != nil {
		return 0, err
	}
	i := r.next
	r.next++
	e.id = i
	r.entries[i] = e
	return i, nil
}

func (r *Resources) New(t *ResourceType, rep uint64) (uint32, error) {
	if t == nil {
		return 0, invalid("missing resource type")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.addLocked(&resourceEntry{typeID: t, rep: rep, owned: true})
}

func (r *Resources) getLocked(t *ResourceType, id uint32) (*resourceEntry, error) {
	if r.closed {
		return nil, wasm.ErrClosed
	}
	e := r.entries[id]
	if e == nil || t == nil || e.typeID != t {
		return nil, scalarTrap("invalid resource handle or type")
	}
	return e, nil
}

func (r *Resources) Rep(t *ResourceType, id uint32) (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, err := r.getLocked(t, id)
	if err != nil {
		return 0, err
	}
	return e.rep, nil
}

// Take consumes an own handle without running its destructor. Publication in
// another instance is performed by the call's ownership-transfer transaction.
func (r *Resources) Take(t *ResourceType, id uint32) (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, err := r.getLocked(t, id)
	if err != nil {
		return 0, err
	}
	if !e.owned || e.lends != 0 {
		return 0, scalarTrap("resource is borrowed")
	}
	delete(r.entries, id)
	r.budget.release(1)
	return e.rep, nil
}

func (r *Resources) Drop(ctx context.Context, t *ResourceType, id uint32) error {
	var drop func(context.Context, uint64) error
	if t != nil {
		drop = t.Drop
	}
	return r.DropWith(ctx, t, id, drop)
}

// DropWith runs a destructor using the caller's existing execution session.
// Ownership is consumed before the callback; a trap cannot resurrect it.
func (r *Resources) DropWith(ctx context.Context, t *ResourceType, id uint32, drop func(context.Context, uint64) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	e, err := r.getLocked(t, id)
	if err == nil && e.lends != 0 {
		err = scalarTrap("resource has outstanding lends")
	}
	if err != nil {
		r.mu.Unlock()
		return err
	}
	delete(r.entries, id)
	r.budget.release(1)
	r.mu.Unlock()
	if !e.owned {
		e.scope.mu.Lock()
		delete(e.scope.borrows, e)
		e.scope.mu.Unlock()
		return nil
	}
	// Remove the handle before invoking user/guest code. Reentry cannot drop it
	// twice, and destructor failures do not resurrect consumed ownership.
	if drop != nil {
		return drop(ctx, e.rep)
	}
	return nil
}

type lender struct {
	table *Resources
	entry *resourceEntry
}

// BorrowScope follows one synchronous call. Lenders remain pinned until End;
// lowered borrowed handles must be explicitly dropped before normal return.
// End reports escaped borrows and invalidates them even on an error path.
type BorrowScope struct {
	mu      sync.Mutex
	maximum uint32
	ended   bool
	lenders []lender
	borrows map[*resourceEntry]lender
}

func NewBorrowScope(maximum uint32) *BorrowScope {
	return &BorrowScope{maximum: maximum, borrows: map[*resourceEntry]lender{}}
}

func (s *BorrowScope) chargeLocked() error {
	if s.ended {
		return scalarTrap("borrow scope ended")
	}
	if uint64(len(s.lenders))+uint64(len(s.borrows)) >= uint64(s.maximum) {
		return exhausted()
	}
	return nil
}

func (s *BorrowScope) Lift(r *Resources, t *ResourceType, id uint32) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.chargeLocked(); err != nil {
		return 0, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	e, err := r.getLocked(t, id)
	if err != nil {
		return 0, err
	}
	if e.lends == ^uint32(0) {
		return 0, exhausted()
	}
	e.lends++
	s.lenders = append(s.lenders, lender{r, e})
	return e.rep, nil
}

// Lower publishes a borrowed representation in a foreign instance's handle
// table. The defining instance passes its representation directly and bypasses
// this operation, as required by the canonical ABI.
func (s *BorrowScope) Lower(r *Resources, t *ResourceType, rep uint64) (uint32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.chargeLocked(); err != nil {
		return 0, err
	}
	if t == nil {
		return 0, invalid("missing resource type")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	e := &resourceEntry{typeID: t, rep: rep, scope: s}
	id, err := r.addLocked(e)
	if err != nil {
		return 0, err
	}
	s.borrows[e] = lender{r, e}
	return id, nil
}

func (s *BorrowScope) End() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ended {
		return nil
	}
	s.ended = true
	escaped := false
	for e, borrow := range s.borrows {
		borrow.table.mu.Lock()
		if borrow.table.entries[e.id] == e {
			delete(borrow.table.entries, e.id)
			borrow.table.budget.release(1)
			escaped = true
		}
		borrow.table.mu.Unlock()
	}
	for _, l := range s.lenders {
		l.table.mu.Lock()
		l.entry.lends--
		l.table.mu.Unlock()
	}
	s.lenders = nil
	s.borrows = nil
	if escaped {
		return scalarTrap("borrowed handle escaped its call")
	}
	return nil
}

// Drain destroys remaining owned handles after all call scopes have ended.
// The component lifecycle serializes Drain with calls and creation. It invokes
// every destructor even if one fails and returns the first error.
func (r *Resources) Drain(ctx context.Context) error {
	r.mu.Lock()
	for _, e := range r.entries {
		if e.lends != 0 || !e.owned {
			r.mu.Unlock()
			return scalarTrap("cannot drain active resource borrows")
		}
	}
	entries := r.entries
	r.budget.release(uint32(len(entries)))
	r.entries = map[uint32]*resourceEntry{}
	r.mu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}
	var first error
	for _, e := range entries {
		if e.typeID.Drop != nil {
			if err := e.typeID.Drop(ctx, e.rep); err != nil && first == nil {
				first = err
			}
		}
	}
	return first
}

// Close invalidates the handle domain without executing guest code. Instance
// cancellation and teardown do not synthesize resource.drop calls.
func (r *Resources) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	if r.errorCount != 0 {
		for _, e := range r.entries {
			if e.errorContext != nil {
				r.budget.releaseError(e.errorContext)
			}
		}
		r.errorCount = 0
	}
	r.budget.release(uint32(len(r.entries)))
	r.entries = nil
}
