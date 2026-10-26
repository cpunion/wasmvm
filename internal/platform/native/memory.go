package native

import (
	"sync"
	"unsafe"
)

// Memory owns zeroed native storage. Context bindings retain it across exits;
// Close retires the owner but does not invalidate already bound executions.
// The Store driver supplies guest page limits and aggregate allocation quotas.
// No caller receives a raw pointer or a growth-sensitive backing slice.
type Memory struct {
	mu                      sync.Mutex
	data                    unsafe.Pointer
	size, maximum, bindings uint64
	active, retired         bool
}

func NewMemory(size, maximum uint64) (*Memory, error) {
	if !Available {
		return nil, ErrUnavailable
	}
	if size > maximum || size > uint64(^uint(0)>>1) {
		return nil, ErrLimit
	}
	allocation := size
	if allocation == 0 {
		allocation = 1
	}
	data := allocate(allocation)
	if data == nil {
		return nil, ErrLimit
	}
	return &Memory{data: data, size: size, maximum: maximum}, nil
}

func (m *Memory) idle() error {
	if m.retired {
		return ErrClosed
	}
	if m.active {
		return ErrBusy
	}
	return nil
}
func (m *Memory) Size() (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.retired {
		return 0, ErrClosed
	}
	return m.size, nil
}
func (m *Memory) Read(offset uint64, dst []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.idle(); err != nil {
		return err
	}
	if offset > m.size || uint64(len(dst)) > m.size-offset {
		return ErrLimit
	}
	copy(dst, unsafe.Slice((*byte)(m.data), int(m.size))[offset:offset+uint64(len(dst))])
	return nil
}
func (m *Memory) Write(offset uint64, src []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.idle(); err != nil {
		return err
	}
	if offset > m.size || uint64(len(src)) > m.size-offset {
		return ErrLimit
	}
	copy(unsafe.Slice((*byte)(m.data), int(m.size))[offset:offset+uint64(len(src))], src)
	return nil
}

// Grow changes the byte length while idle. Failure leaves storage untouched;
// success preserves old bytes and zeroes the extension. Old and new allocations
// coexist during the copy, which the Store's transient allocation budget must
// account for. A size equal to the current size performs no allocation.
func (m *Memory) Grow(size uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.idle(); err != nil {
		return err
	}
	if size < m.size {
		return ErrInvalid
	}
	if size > m.maximum || size > uint64(^uint(0)>>1) {
		return ErrLimit
	}
	if size == m.size {
		return nil
	}
	data := allocate(size)
	if data == nil {
		return ErrLimit
	}
	copy(unsafe.Slice((*byte)(data), int(size)), unsafe.Slice((*byte)(m.data), int(m.size)))
	deallocate(m.data)
	m.data, m.size = data, size
	return nil
}
func (m *Memory) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retired = true
	m.release()
}
func (m *Memory) release() {
	if m.retired && m.bindings == 0 && !m.active && m.data != nil {
		deallocate(m.data)
		m.data = nil
	}
}
func (m *Memory) retain() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.retired {
		return ErrClosed
	}
	m.bindings++
	return nil
}
func (m *Memory) drop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bindings--
	m.release()
}
func (m *Memory) begin() (uint64, uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		return 0, 0, ErrClosed
	}
	if m.active {
		return 0, 0, ErrBusy
	}
	m.active = true
	return uint64(uintptr(m.data)), m.size, nil
}
func (m *Memory) end() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.active = false
	m.release()
}

// BindMemory replaces the current Core 2.0 memory while idle. Multiple contexts
// can bind the same owner, but only one may execute against it at a time. Pass
// nil to release a binding. Entry refreshes the record after any intervening
// growth or host reentry; no stale native address survives that handoff.
func (c *Context) BindMemory(m *Memory) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.idle(); err != nil {
		return err
	}
	if c.memory == m {
		return nil
	}
	if m != nil {
		if err := m.retain(); err != nil {
			return err
		}
	}
	if c.memory != nil {
		c.memory.drop()
	}
	c.memory = m
	c.record.MemoryBase, c.record.MemorySize = 0, 0
	return nil
}
