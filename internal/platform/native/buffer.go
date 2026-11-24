package native

import "unsafe"

// Buffer exposes pointer-free storage to the trusted, serialized Store runtime.
// The view must not reach embedding callers, outlive the owner, or be accessed
// while native code executes. Moving private growth invalidates previous views;
// stable private owners retain their base but require a fresh length. Shared
// addresses use SharedView instead. This is an internal ownership
// contract, not a public guest-memory borrow API.
func (m *Memory) Buffer() ([]byte, error) {
	m.mu.Lock()
	if err := m.idle(); err != nil {
		m.mu.Unlock()
		return nil, err
	}
	buf := unsafe.Slice((*byte)(m.data), int(m.size))
	m.mu.Unlock()
	return buf, nil
}

// ResizeArena grows the private frame arena while idle and returns its current
// pointer-free storage. Growth preserves bytes and zeroes the extension. The
// caller drops the old view before further use and never exposes either view
// outside the serialized execution runtime. A failed growth changes nothing.
func (c *Context) ResizeArena(size, maximum uint64) ([]byte, error) {
	c.mu.Lock()
	if err := c.idle(); err != nil {
		c.mu.Unlock()
		return nil, err
	}
	if size > maximum || size > uint64(^uint(0)>>1) {
		c.mu.Unlock()
		return nil, ErrLimit
	}
	if size > c.size {
		p := allocate(size)
		if p == nil {
			c.mu.Unlock()
			return nil, ErrLimit
		}
		copy(unsafe.Slice((*byte)(p), int(size)), unsafe.Slice((*byte)(c.arena), int(c.size)))
		deallocate(c.arena)
		c.arena, c.size = p, size
		c.record.Arena, c.record.Capacity = uint64(uintptr(p)), size
	}
	buf := unsafe.Slice((*byte)(c.arena), int(c.size))
	c.mu.Unlock()
	return buf, nil
}
