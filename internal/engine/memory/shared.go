package memory

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Domain provides sequentially consistent ordering for every shared byte and
// atomic operation, across all memories and Stores in one Engine. Private
// memory keeps its existing direct access path. Waits release the domain lock.
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
func (d *Domain) Fence()          { d.mu.Lock(); d.mu.Unlock() }
func (d *Domain) Waiters() uint32 { d.mu.Lock(); defer d.mu.Unlock(); return d.waiting }

func NewShared(declared module.Limits, quota uint32, domain *Domain) (*Memory, error) {
	if domain == nil || !declared.Shared || !declared.HasMax {
		return nil, wasm.ErrInvalid
	}
	m, err := newMemory(declared, quota, domain.offheap)
	if err != nil {
		return nil, err
	}
	m.domain, m.refs, m.reservation = domain, 1, m.maximum
	return m, nil
}
func (m *Memory) Shared() bool             { return m.domain != nil }
func (m *Memory) Domain() *Domain          { return m.domain }
func (m *Memory) ReservationPages() uint64 { return m.reservation }
func (m *Memory) Retain() bool {
	if m.domain == nil {
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
	if m.domain == nil {
		return m.pages()
	}
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
	if m.domain == nil {
		return m.rangeBytes(at, n)
	}
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.rangeBytes(at, n)
}
func (m *Memory) Read(at uint64, p []byte) bool {
	if m.domain == nil {
		return m.readBytes(at, p)
	}
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.readBytes(at, p)
}
func (m *Memory) Write(at uint64, p []byte) bool {
	if m.domain == nil {
		return m.writeBytes(at, p)
	}
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.writeBytes(at, p)
}
func (m *Memory) Fill(at uint64, b byte, n uint64) bool {
	if m.domain == nil {
		return m.fill(at, b, n)
	}
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.fill(at, b, n)
}
func (m *Memory) Init(dst, src, n uint64, p []byte) bool {
	if m.domain == nil {
		return m.initBytes(dst, src, n, p)
	}
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
	if d != nil {
		d.mu.Lock()
		defer d.mu.Unlock()
	}
	return m.copyFrom(source, dst, src, n)
}
func (m *Memory) Load(op uint32, address, offset uint64, lane byte, previous value.Value) (value.Value, wasm.TrapCode) {
	if m.domain == nil {
		return m.load(op, address, offset, lane, previous)
	}
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.load(op, address, offset, lane, previous)
}
func (m *Memory) Store(op uint32, address, offset uint64, lane byte, v value.Value) wasm.TrapCode {
	if m.domain == nil {
		return m.store(op, address, offset, lane, v)
	}
	m.domain.mu.Lock()
	defer m.domain.mu.Unlock()
	return m.store(op, address, offset, lane, v)
}

// Atomic executes one validated atomic opcode. Both execution backends use
// these typed, individual helpers; no Go pointer is stored in guest frames.
func (m *Memory) Atomic(ctx context.Context, d *Domain, sub uint32, address, offset uint64, a, b uint64) (value.Value, wasm.TrapCode, error) {
	if sub == 3 || sub > 0x4e || sub > 2 && sub < 0x10 {
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
	d.mu.Lock()
	n := uint(4)
	if sub == 2 {
		n = 8
	} else if sub >= 0x10 {
		n = [7]uint{4, 8, 1, 2, 1, 2, 4}[(sub-0x10)%7]
	}
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
		mask := ^uint64(0) >> (64 - n*8)
		if read(bytes, n) != a&mask {
			d.mu.Unlock()
			return value.Value{Lo: 1}, wasm.NoTrap, nil
		}
		if int64(b) == 0 {
			d.mu.Unlock()
			return value.Value{Lo: 2}, wasm.NoTrap, nil
		}
		if d.waiting >= d.maxWaiters {
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
