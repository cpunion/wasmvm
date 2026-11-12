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
	generation   uint64
	typeID       *ResourceType
	rep          uint64
	owned        bool
	scope        *BorrowScope
	lends        uint32
	errorContext *errorContext
	opaque       any
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

// Resources is a bounded handle domain. Guest domains reuse freed indices in
// ABI-mandated LIFO order; private host capability domains never reuse IDs.
type Resources struct {
	mu           sync.Mutex
	order        uint64
	entries      map[uint32]*resourceEntry
	guestEntries []*resourceEntry
	count        uint32
	next         uint32
	generation   uint64
	free         []uint32
	reuse        bool
	budget       *ResourceBudget
	closed       bool
	errorCount   uint32
}

func NewResources(maximum uint32) *Resources {
	return NewResourceTable(NewResourceBudget(maximum))
}

// NewResourceTable creates an independent handle domain with a shared quota.
var resourceTableOrder atomic.Uint64

func NewResourceTable(budget *ResourceBudget) *Resources {
	return &Resources{order: resourceTableOrder.Add(1), entries: map[uint32]*resourceEntry{}, next: 1, budget: budget}
}

// NewGuestResourceTable implements the canonical per-component handles table.
// Host references use a separate nonreusing domain or a pinned borrow scope.
func NewGuestResourceTable(budget *ResourceBudget) *Resources {
	r := NewResourceTable(budget)
	r.reuse = true
	r.entries = nil
	return r
}

func (r *Resources) availableLocked() bool {
	return r.generation != ^uint64(0) && (len(r.free) != 0 || r.next < 1<<28)
}

func (r *Resources) indexLocked() uint32 {
	r.generation++
	if n := len(r.free); n != 0 {
		i := r.free[n-1]
		r.free = r.free[:n-1]
		return i
	}
	i := r.next
	r.next++
	return i
}

func (r *Resources) entryLocked(id uint32) *resourceEntry {
	if r.reuse {
		if uint64(id) >= uint64(len(r.guestEntries)) {
			return nil
		}
		return r.guestEntries[id]
	}
	return r.entries[id]
}
func (r *Resources) insertLocked(id uint32, e *resourceEntry) {
	if r.reuse {
		if int(id) >= len(r.guestEntries) {
			wanted := max(16, len(r.guestEntries)*2, int(id)+1)
			r.guestEntries = append(r.guestEntries, make([]*resourceEntry, wanted-len(r.guestEntries))...)
		}
		r.guestEntries[id] = e
	} else {
		r.entries[id] = e
	}
	r.count++
}
func (r *Resources) removeLocked(id uint32) {
	if r.reuse {
		r.guestEntries[id] = nil
		r.free = append(r.free, id)
	} else {
		delete(r.entries, id)
	}
	r.count--
}

// Guest indices are densely allocated and reused in the ABI's LIFO order.
// Host IDs remain nonreusing sparse keys. Entry objects retain stable identity
// across growth, borrowing and cross-domain opaque transfers.
func visitResourceEntries(entries map[uint32]*resourceEntry, guest []*resourceEntry, visit func(*resourceEntry) error) error {
	for _, e := range entries {
		if err := visit(e); err != nil {
			return err
		}
	}
	for _, e := range guest {
		if e != nil {
			if err := visit(e); err != nil {
				return err
			}
		}
	}
	return nil
}

// EntryGeneration identifies one publication across index reuse and transfers.
// Rollback must not consume a later publication at the same guest index.
func (r *Resources) EntryGeneration(id uint32) uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		if entry := r.entryLocked(id); entry != nil {
			return entry.generation
		}
	}
	return 0
}

func (r *Resources) addLocked(e *resourceEntry) (uint32, error) {
	if r.closed {
		return 0, wasm.ErrClosed
	}
	if !r.availableLocked() {
		return 0, exhausted()
	}
	if err := r.budget.acquire(); err != nil {
		return 0, err
	}
	i := r.indexLocked()
	e.id, e.generation = i, r.generation
	r.insertLocked(i, e)
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
	e := r.entryLocked(id)
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
	r.removeLocked(id)
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
	r.removeLocked(id)
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
		if borrow.table.entryLocked(e.id) == e {
			borrow.table.removeLocked(e.id)
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

// CheckReturned checks the task.return boundary without releasing its lenders.
// Guest code must drop foreign borrowed handles before publishing a result.
func (s *BorrowScope) CheckReturned() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ended || len(s.borrows) != 0 {
		return scalarTrap("task returned with outstanding borrows")
	}
	return nil
}

// Drain destroys remaining owned handles after all call scopes have ended.
// The component lifecycle serializes Drain with calls and creation. It invokes
// every destructor even if one fails and returns the first error.
func (r *Resources) Drain(ctx context.Context) error {
	r.mu.Lock()
	if err := visitResourceEntries(r.entries, r.guestEntries, func(e *resourceEntry) error {
		if e.lends != 0 || !e.owned {
			return scalarTrap("cannot drain active resource borrows")
		}
		return nil
	}); err != nil {
		r.mu.Unlock()
		return err
	}
	entries, guest := r.entries, r.guestEntries
	r.budget.release(r.count)
	r.count = 0
	r.guestEntries = nil
	if r.reuse {
		r.next = 1
		r.free = nil
	} else {
		r.entries = map[uint32]*resourceEntry{}
	}

	r.mu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}
	var first error
	visitResourceEntries(entries, guest, func(e *resourceEntry) error {
		if e.typeID != nil && e.typeID.Drop != nil {
			if err := e.typeID.Drop(ctx, e.rep); err != nil && first == nil {
				first = err
			}
		}
		return nil
	})
	return first
}

// Close invalidates the handle domain without executing guest code. Instance
// cancellation and teardown do not synthesize resource.drop calls.
func (r *Resources) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	if r.errorCount != 0 {
		visitResourceEntries(r.entries, r.guestEntries, func(e *resourceEntry) error {
			if e.errorContext != nil {
				r.budget.releaseError(e.errorContext)
			}
			return nil
		})
		r.errorCount = 0
	}
	r.budget.release(r.count)
	r.count = 0
	r.guestEntries = nil
	r.entries = nil
	r.free = nil
}
