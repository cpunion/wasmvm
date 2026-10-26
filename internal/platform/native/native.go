package native

import (
	"runtime"
	"sync"
	"unsafe"
)

// Code owns immutable executable bytes. Close retires publication; existing
// leases can still enter, and storage is freed only after leases and entries
// drain. No finalizer participates in code lifetime.
type Code struct {
	mu             sync.Mutex
	address        unsafe.Pointer
	size           uint64
	entries        []uint32
	retired        bool
	leases, active uint64
}

// Publish copies code, seals it executable, and records approved C-ABI entry
// offsets. maxBytes is a literal budget. Interior continuation labels are not
// entries. The compiler must validate instructions, relocations and targets.
func Publish(code []byte, entries []uint32, maxBytes uint64) (*Code, error) {
	if !Available {
		return nil, ErrUnavailable
	}
	if len(code) == 0 || len(entries) == 0 {
		return nil, ErrInvalid
	}
	if uint64(len(code)) > maxBytes {
		return nil, ErrLimit
	}
	for _, off := range entries {
		if uint64(off) >= uint64(len(code)) || runtime.GOARCH == "arm64" && off%4 != 0 {
			return nil, ErrInvalid
		}
	}
	if runtime.GOARCH == "arm64" && len(code)%4 != 0 {
		return nil, ErrInvalid
	}
	p := publish(code)
	if p == nil {
		return nil, ErrUnavailable
	}
	return &Code{address: p, size: uint64(len(code)), entries: append([]uint32(nil), entries...)}, nil
}

type Lease struct {
	mu   sync.Mutex
	code *Code
}

func (c *Code) Acquire() (*Lease, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.retired || c.address == nil {
		return nil, ErrClosed
	}
	c.leases++
	return &Lease{code: c}, nil
}

func (c *Code) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.retired = true
	c.release()
}

func (c *Code) release() {
	if c.retired && c.leases == 0 && c.active == 0 && c.address != nil {
		unpublish(c.address, c.size)
		c.address = nil
		c.entries = nil
	}
}

func (l *Lease) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.code != nil {
		c := l.code
		c.mu.Lock()
		c.leases--
		c.release()
		c.mu.Unlock()
		l.code = nil
	}
}

func (l *Lease) begin(entry uint32) (*Code, unsafe.Pointer, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.code == nil {
		return nil, nil, ErrClosed
	}
	c := l.code
	c.mu.Lock()
	defer c.mu.Unlock()
	if uint64(entry) >= uint64(len(c.entries)) {
		return nil, nil, ErrInvalid
	}
	c.active++
	return c, unsafe.Add(c.address, uintptr(c.entries[entry])), nil
}

func (c *Code) end() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.active--
	c.release()
}

// Context owns a zeroed off-heap record and frame arena. Every access is either
// by one native entry or by the idle driver. Cancellation is the only concurrent
// native-visible mutation and uses the adapter's release/acquire atomic word.
type Context struct {
	mu             sync.Mutex
	record         *record
	arena          unsafe.Pointer
	size           uint64
	active, closed bool
}

func NewContext(arenaBytes, maxBytes, maxDepth uint64) (*Context, error) {
	if !Available {
		return nil, ErrUnavailable
	}
	if arenaBytes > maxBytes || arenaBytes > uint64(^uint(0)>>1) {
		return nil, ErrLimit
	}
	if maxDepth == 0 {
		return nil, ErrLimit
	}
	r := (*record)(allocate(uint64(RecordSize)))
	if r == nil {
		return nil, ErrLimit
	}
	// Even an empty frame arena has an owned, non-nil base.
	n := arenaBytes
	if n == 0 {
		n = 1
	}
	a := allocate(n)
	if a == nil {
		deallocate(unsafe.Pointer(r))
		return nil, ErrLimit
	}
	r.Version, r.Arena, r.Capacity, r.MaxDepth = ABIVersion, uint64(uintptr(a)), arenaBytes, maxDepth
	return &Context{record: r, arena: a, size: arenaBytes}, nil
}

func (c *Context) idle() error {
	if c.closed {
		return ErrClosed
	}
	if c.active {
		return ErrBusy
	}
	return nil
}

func (c *Context) State() (State, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.idle(); err != nil {
		return State{}, err
	}
	return c.record.State, nil
}

func (c *Context) SetState(s State) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.idle(); err != nil {
		return err
	}
	if s.Used > c.size || s.Depth > c.record.MaxDepth {
		return ErrLimit
	}
	c.record.State = s
	return nil
}

// Read and Write copy bytes; no native storage or Go pointer is retained by the
// caller or generated code. Reference slots contain handles, not host pointers.
func (c *Context) Read(offset uint64, dst []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.idle(); err != nil {
		return err
	}
	if offset > c.size || uint64(len(dst)) > c.size-offset {
		return ErrLimit
	}
	copy(dst, unsafe.Slice((*byte)(c.arena), int(c.size))[offset:offset+uint64(len(dst))])
	return nil
}

func (c *Context) Write(offset uint64, src []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.idle(); err != nil {
		return err
	}
	if offset > c.size || uint64(len(src)) > c.size-offset {
		return ErrLimit
	}
	copy(unsafe.Slice((*byte)(c.arena), int(c.size))[offset:offset+uint64(len(src))], src)
	return nil
}

// Cancel is sticky and never waits for native execution. A new invocation needs
// a new context after cancellation. Generated code must poll ControlOffset with
// acquire semantics and return through its regular exit path.
func (c *Context) Cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.record != nil {
		cancel(unsafe.Pointer(c.record))
	}
}

// Enter calls an approved C entry, never an interior continuation address. The
// generated entry validates its continuation ID and restores spilled state.
func (c *Context) Enter(l *Lease, entry uint32) (Exit, error) {
	if l == nil {
		return 0, ErrInvalid
	}
	c.mu.Lock()
	if err := c.idle(); err != nil {
		c.mu.Unlock()
		return 0, err
	}
	code, address, err := l.begin(entry)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	c.active = true
	r := c.record
	c.mu.Unlock()
	status := enter(address, unsafe.Pointer(r))
	runtime.KeepAlive(c)
	runtime.KeepAlive(l)
	code.end()
	c.mu.Lock()
	c.active = false
	c.release()
	c.mu.Unlock()
	if status < Return || status > Move {
		return 0, ErrInvalid
	}
	return status, nil
}

// Close requests cancellation and defers freeing until the active entry exits.
func (c *Context) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	if c.record != nil {
		cancel(unsafe.Pointer(c.record))
	}
	c.release()
}

func (c *Context) release() {
	if c.closed && !c.active && c.record != nil {
		deallocate(c.arena)
		deallocate(unsafe.Pointer(c.record))
		c.arena, c.record = nil, nil
	}
}
