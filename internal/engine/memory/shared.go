package memory

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"unsafe"
)

// Domain coordinates wait queues, ownership and growth. Native-backed shared
// bytes use atomic accesses; portable backing serializes access under the domain
// lock. Wait registration and notification share that lock in both modes.
type Domain struct {
	mu         sync.Mutex
	offheap    bool
	maxWaiters uint32
	waiting    uint32
	queues     map[waitKey]*waitQueue
}
type waitKey struct {
	memory  *Memory
	address uint64
}
type waiter struct {
	ready          chan struct{}
	notified       bool
	previous, next *waiter
	linked         bool
}

func NewDomain(maxWaiters uint32, offheap bool) *Domain {
	return &Domain{maxWaiters: maxWaiters, offheap: offheap}
}
func (d *Domain) Fence() {
	if d.offheap && native.Available {
		checkedShared(native.SharedFence())
		return
	}
	d.mu.Lock()
	d.mu.Unlock()
}
func (d *Domain) Waiters() uint32 { d.mu.Lock(); defer d.mu.Unlock(); return d.waiting }

func NewShared(declared module.Limits, quota uint32, domain *Domain) (*Memory, error) {
	return newShared(declared, quota, domain, true)
}

// NewStoreLocalShared preserves all shared-memory instructions and quota
// accounting. Its backing is never lent to a different Store.
func NewStoreLocalShared(declared module.Limits, quota uint32, domain *Domain) (*Memory, error) {
	return newShared(declared, quota, domain, false)
}

func newShared(declared module.Limits, quota uint32, domain *Domain, shareable bool) (*Memory, error) {
	if domain == nil || !declared.Shared || !declared.HasMax {
		return nil, wasm.ErrInvalid
	}
	m, err := newMemory(declared, quota, domain.offheap)
	if err != nil {
		return nil, err
	}
	m.domain, m.refs, m.reservation = domain, 1, QuotaPages(declared, m.maximum)
	m.shareable = shareable
	if shareable {
		m.lockedShared = m.sharedSize == nil
	} else if m.sharedSize != nil {
		// Keep the reserved capacity for growth, but bound ordinary accesses
		// by the slice length. The Store serializes every user of this view.
		m.bytes = m.bytes[:declared.Min*declared.PageBytes()]
		m.sharedSize = nil
	}
	return m, nil
}
func (m *Memory) Shared() bool             { return m.domain != nil }
func (m *Memory) Domain() *Domain          { return m.domain }
func (m *Memory) ReservationPages() uint64 { return m.reservation }
func (m *Memory) Shareable() bool          { return m.shareable }
func (m *Memory) Retain() bool {
	if !m.Shareable() {
		return false
	}
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	if m.refs == 0 || m.refs == math.MaxUint64 {
		return false
	}
	m.refs++
	return true
}
func (m *Memory) Release() {
	if m.domain == nil {
		m.release()
		return
	}
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	if m.refs == 0 {
		return
	}
	m.refs--
	if m.refs == 0 {
		m.release()
	}
}
func (m *Memory) Pages() uint64 {
	if !m.lockedShared {
		return m.pages()
	}
	return m.pagesLocked()
}
func (m *Memory) pagesLocked() uint64 {
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.pages()
}
func (m *Memory) Grow(n uint64) (uint64, bool) {
	if m.domain == nil {
		return m.grow(n)
	}
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	if m.refs == 0 {
		return 0, false
	}
	return m.grow(n)
}
func (m *Memory) Range(at, n uint64) bool {
	if !m.lockedShared {
		return m.rangeBytes(at, n)
	}
	return m.rangeLocked(at, n)
}
func (m *Memory) rangeLocked(at, n uint64) bool {
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.rangeBytes(at, n)
}
func (m *Memory) Read(at uint64, p []byte) bool {
	if !m.lockedShared {
		return m.readBytes(at, p)
	}
	return m.readLocked(at, p)
}
func (m *Memory) readLocked(at uint64, p []byte) bool {
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.readBytes(at, p)
}
func (m *Memory) Write(at uint64, p []byte) bool {
	if !m.lockedShared {
		return m.writeBytes(at, p)
	}
	return m.writeLocked(at, p)
}
func (m *Memory) writeLocked(at uint64, p []byte) bool {
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.writeBytes(at, p)
}

// ReadString returns an immutable snapshot, never a view of guest storage.
// Cross-Store native memory retains its byte-wise synchronized transfer path.
func (m *Memory) ReadString(at, n uint64) (string, bool) {
	if m.lockedShared {
		m.domain.mu.Lock()
		defer m.domain.mu.Unlock()
	}
	if !m.rangeBytes(at, n) {
		return "", false
	}
	if m.sharedSize != nil {
		data := make([]byte, int(n))
		m.readBytes(at, data)
		return string(data), true
	}
	return string(m.bytes[int(at):int(at+n)]), true
}

// WriteString copies immutable host bytes without an intermediate byte slice
// for Store-local memory. Rejected ranges leave memory unchanged.
func (m *Memory) WriteString(at uint64, data string) bool {
	if m.lockedShared {
		m.domain.mu.Lock()
		defer m.domain.mu.Unlock()
	}
	if !m.rangeBytes(at, uint64(len(data))) {
		return false
	}
	if m.sharedSize != nil {
		return m.writeBytes(at, []byte(data))
	}
	copy(m.bytes[int(at):], data)
	return true
}
func (m *Memory) Fill(at uint64, b byte, n uint64) bool {
	if !m.lockedShared {
		return m.fill(at, b, n)
	}
	return m.fillLocked(at, b, n)
}
func (m *Memory) fillLocked(at uint64, b byte, n uint64) bool {
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.fill(at, b, n)
}
func (m *Memory) Init(dst, src, n uint64, p []byte) bool {
	if !m.lockedShared {
		return m.initBytes(dst, src, n, p)
	}
	return m.initLocked(dst, src, n, p)
}
func (m *Memory) initLocked(dst, src, n uint64, p []byte) bool {
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.initBytes(dst, src, n, p)
}
func (m *Memory) Copy(dst, src, n uint64) bool { return m.CopyFrom(m, dst, src, n) }
func (m *Memory) CopyFrom(source *Memory, dst, src, n uint64) bool {
	d := m.domain
	if d == nil {
		d = source.domain
	}
	if m.domain != nil && source.domain != nil && m.domain != source.domain {
		return false
	}
	if d != nil && (m.lockedShared || source.lockedShared) {
		d.mu.Lock()
		defer d.mu.Unlock()
	}
	return m.copyFrom(source, dst, src, n)
}
func (m *Memory) Load(op uint32, address, offset uint64, lane byte, previous value.Value) (value.Value, wasm.TrapCode) {
	if !m.lockedShared {
		return m.load(op, address, offset, lane, previous)
	}
	return m.loadLocked(op, address, offset, lane, previous)
}
func (m *Memory) loadLocked(op uint32, address, offset uint64, lane byte, previous value.Value) (value.Value, wasm.TrapCode) {
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.load(op, address, offset, lane, previous)
}
func (m *Memory) storeWord(address, offset uint64, width uint, x uint64) wasm.TrapCode {
	if !m.lockedShared {
		return m.storeWordChecked(address, offset, width, x)
	}
	return m.storeWordLocked(address, offset, width, x)
}
func (m *Memory) storeWordLocked(address, offset uint64, width uint, x uint64) wasm.TrapCode {
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.storeWordChecked(address, offset, width, x)
}
func (m *Memory) storeVector(address, offset uint64, v value.Value) wasm.TrapCode {
	if !m.lockedShared {
		return m.storeVectorChecked(address, offset, v)
	}
	return m.storeVectorLocked(address, offset, v)
}
func (m *Memory) storeVectorLocked(address, offset uint64, v value.Value) wasm.TrapCode {
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.storeVectorChecked(address, offset, v)
}

type atomicShape struct {
	width     uint32
	operation native.AtomicOperation
}

// Wait and notify use the opcode directly. Ordinary operations share the
// width sequence for each load, store, and read-modify-write family.
var atomicShapes = func() [0x4f]atomicShape {
	var shapes [0x4f]atomicShape
	shapes[0].width, shapes[1].width, shapes[2].width = 4, 4, 8
	widths := [...]uint32{4, 8, 1, 2, 1, 2, 4}
	for i := 0x10; i < len(shapes); i++ {
		shapes[i] = atomicShape{width: widths[(i-0x10)%len(widths)],
			operation: native.AtomicOperation((i - 0x10) / len(widths))}
	}
	return shapes
}()

// Atomic executes one validated atomic opcode. Both execution backends use
// these typed, individual helpers; no Go pointer is stored in guest frames.
func (m *Memory) Atomic(ctx context.Context, d *Domain, sub uint32, address, offset uint64, a, b uint64) (value.Value, wasm.TrapCode, error) {
	if sub >= uint32(len(atomicShapes)) || atomicShapes[sub].width == 0 {
		return value.Value{}, wasm.NoTrap, wasm.ErrInvalid
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return value.Value{}, wasm.NoTrap, err
	}
	if d == nil || m.domain != nil && m.domain != d {
		return value.Value{}, wasm.NoTrap, wasm.ErrInvalid
	}
	shape := atomicShapes[sub]
	n := uint(shape.width)
	if m.owner != nil {
		bytes, ok := m.access(address, offset, n)
		if !ok {
			return value.Value{}, wasm.TrapMemoryBounds, nil
		}
		at := m.declared.Address(address) + offset
		if at%uint64(n) != 0 {
			return value.Value{}, wasm.TrapAtomicUnaligned, nil
		}
		if sub >= 0x10 {
			old, err := native.SharedAtomicInRange(ctx, unsafe.Pointer(unsafe.SliceData(bytes)), shape.width, shape.operation, a, b)
			return value.Value{Lo: old}, wasm.NoTrap, err
		}
		if sub == 0 {
			if uint32(a) == 0 || m.domain == nil || atomic.LoadUint64(m.sharedWaiters) == 0 {
				return value.Value{}, wasm.NoTrap, nil
			}
		} else {
			if m.domain == nil {
				return value.Value{}, wasm.TrapWaitUnshared, nil
			}
			if err := ctx.Err(); err != nil {
				return value.Value{}, wasm.NoTrap, err
			}
			observed := loadNativeWaitWord(bytes, n)
			mask := ^uint64(0) >> (64 - n*8)
			if observed != a&mask {
				return value.Value{Lo: 1}, wasm.NoTrap, nil
			}
			if int64(b) == 0 {
				return value.Value{Lo: 2}, wasm.NoTrap, nil
			}
		}
	}
	return m.atomicSlow(ctx, d, sub, address, offset, a, b, n)
}

// Wait compares a checked, naturally aligned i32 or i64 word in a retained
// native mapping. The comparison is sequentially consistent.
func loadNativeWaitWord(bytes []byte, width uint) uint64 {
	p := unsafe.Pointer(unsafe.SliceData(bytes))
	if width == 4 {
		return uint64(atomic.LoadUint32((*uint32)(p)))
	}
	return atomic.LoadUint64((*uint64)(p))
}

// The queue path owns timers and deferred cleanup; scalar fast paths never
// construct those captures under either host compiler.
func (m *Memory) atomicSlow(ctx context.Context, d *Domain, sub uint32, address, offset, a, b uint64, n uint) (value.Value, wasm.TrapCode, error) {
	d.mu.Lock()
	bytes, ok := m.access(address, offset, n)
	if !ok {
		d.mu.Unlock()
		return value.Value{}, wasm.TrapMemoryBounds, nil
	}
	at := m.declared.Address(address) + offset
	if at%uint64(n) != 0 {
		d.mu.Unlock()
		return value.Value{}, wasm.TrapAtomicUnaligned, nil
	}
	if sub == 0 {
		key := waitKey{m, at}
		count := uint32(0)
		queue := d.queues[key]
		for queue != nil && queue.first != nil && count < uint32(a) {
			w := queue.first
			queue.remove(w)
			m.releaseNativeWaiter()
			w.notified = true
			close(w.ready)
			count++
		}
		if queue != nil && queue.first == nil {
			delete(d.queues, key)
		}
		d.mu.Unlock()
		return value.Value{Lo: uint64(count)}, wasm.NoTrap, nil
	}
	if sub <= 2 {
		if m.domain == nil {
			d.mu.Unlock()
			return value.Value{}, wasm.TrapWaitUnshared, nil
		}
		// Publish a pending registration before the SC comparison. A native
		// zero-waiter notify may linearize before this reservation, but must
		// fall back through the ordering lock after observing it.
		if m.sharedWaiters != nil {
			atomic.AddUint64(m.sharedWaiters, 1)
		}
		mask := ^uint64(0) >> (64 - n*8)
		var observed uint64
		if m.owner != nil {
			observed = loadNativeWaitWord(bytes, n)
		} else {
			observed = read(bytes, n)
		}
		if observed != a&mask {
			m.releaseNativeWaiter()
			d.mu.Unlock()
			return value.Value{Lo: 1}, wasm.NoTrap, nil
		}
		if int64(b) == 0 {
			m.releaseNativeWaiter()
			d.mu.Unlock()
			return value.Value{Lo: 2}, wasm.NoTrap, nil
		}
		if d.waiting >= d.maxWaiters {
			m.releaseNativeWaiter()
			d.mu.Unlock()
			return value.Value{}, wasm.NoTrap, &wasm.Error{Kind: wasm.ErrLimit, Detail: "shared memory waiters"}
		}
		key := waitKey{m, at}
		w := &waiter{ready: make(chan struct{})}
		if d.queues == nil {
			d.queues = make(map[waitKey]*waitQueue)
		}
		queue := d.queues[key]
		if queue == nil {
			queue = &waitQueue{}
			d.queues[key] = queue
		}
		queue.append(w)
		d.waiting++
		d.mu.Unlock()
		var timer *time.Timer
		var deadline <-chan time.Time
		if int64(b) > 0 {
			timer = time.NewTimer(time.Duration(b))
			deadline = timer.C
			defer timer.Stop()
		}
		select {
		case <-w.ready:
		case <-ctx.Done():
		case <-deadline:
		}
		d.mu.Lock()
		// Notification wins if it acquired the ordering lock before completion.
		result := uint64(2)
		var err error
		if w.notified {
			result = 0
		} else {
			err = ctx.Err()
		}
		if w.linked {
			queue.remove(w)
			m.releaseNativeWaiter()
			if queue.first == nil {
				delete(d.queues, key)
			}
		}
		d.waiting--
		d.mu.Unlock()
		return value.Value{Lo: result}, wasm.NoTrap, err
	}
	old := read(bytes, n)
	next := a
	group := (sub - 0x10) / 7
	switch group {
	case 0:
	case 1:
		write(bytes, n, a)
	case 2:
		next = old + a
	case 3:
		next = old - a
	case 4:
		next = old & a
	case 5:
		next = old | a
	case 6:
		next = old ^ a
	case 7:
	case 8:
		mask := ^uint64(0) >> (64 - n*8)
		if old == a&mask {
			write(bytes, n, b)
		}
	}
	if group >= 2 && group <= 7 {
		write(bytes, n, next)
	}
	d.mu.Unlock()
	return value.Value{Lo: old}, wasm.NoTrap, nil
}

// Intrusive queues keep registration, timeout/cancel and notified removal O(1).
// Waking a group is linear only in the number of waiters actually woken.
type waitQueue struct{ first, last *waiter }

func (q *waitQueue) append(w *waiter) {
	w.linked = true
	w.previous = q.last
	if q.last != nil {
		q.last.next = w
	} else {
		q.first = w
	}
	q.last = w
}
func (q *waitQueue) remove(w *waiter) {
	if w.previous != nil {
		w.previous.next = w.next
	} else {
		q.first = w.next
	}
	if w.next != nil {
		w.next.previous = w.previous
	} else {
		q.last = w.previous
	}
	w.previous = nil
	w.next = nil
	w.linked = false
}

// The metadata counts both pending SC comparisons and enqueued waiters.
func (m *Memory) releaseNativeWaiter() {
	if m.sharedWaiters != nil {
		atomic.AddUint64(m.sharedWaiters, ^uint64(0))
	}
}
