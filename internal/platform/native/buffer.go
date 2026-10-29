package native

import "unsafe"

// Buffer exposes pointer-free storage to the trusted, serialized Store runtime.
// The view must not reach embedding callers, outlive the owner, or be accessed
// while native code executes. Grow invalidates all previous views. This is an
// internal ownership contract, not a public guest-memory borrow API.
func (m *Memory) Buffer() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.idle(); err != nil {
		return nil, err
	}
	return unsafe.Slice((*byte)(m.data), int(m.size)), nil
}

// ResizeArena grows the private frame arena while idle and returns its current
// pointer-free storage. Growth preserves bytes and zeroes the extension. The
// caller drops the old view before further use and never exposes either view
// outside the serialized execution runtime. A failed growth changes nothing.
func (c *Context) ResizeArena(size, maximum uint64) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.idle(); err != nil {
		return nil, err
	}
	if size > maximum || size > uint64(^uint(0)>>1) {
		return nil, ErrLimit
	}
	if size > c.size {
		p := allocate(size)
		if p == nil {
			return nil, ErrLimit
		}
		copy(unsafe.Slice((*byte)(p), int(size)), unsafe.Slice((*byte)(c.arena), int(c.size)))
		deallocate(c.arena)
		c.arena, c.size = p, size
		c.record.Arena, c.record.Capacity = uint64(uintptr(p)), size
	}
	return unsafe.Slice((*byte)(c.arena), int(c.size)), nil
}
