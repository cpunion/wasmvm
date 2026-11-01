package native

import (
	"sync/atomic"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
)

// NewSharedMemory reserves zeroed storage up to the host-bounded maximum.
// Neither its base nor its native-owned size/waiter cells move on growth.
// Simultaneous entries must use the shared-memory atomic access contract;
// this owner supplies lifetime and size publication, not access synchronization.
func NewSharedMemory(size, maximum uint64) (*Memory, error) {
	if size > maximum {
		return nil, ErrLimit
	}
	m, err := NewMemory(maximum, maximum)
	if err != nil {
		return nil, err
	}
	p := allocate(nativeabi.SharedMemoryControlBytes)
	if p == nil {
		m.Close()
		return nil, ErrLimit
	}
	m.sharedSize = (*uint64)(p)
	m.size = size
	atomic.StoreUint64(m.sharedSize, size)
	return m, nil
}

// SharedView borrows stable native addresses and the reserved extent. A caller
// must retain the owner or a bound context until every use ends. The size cell
// must be read atomically on every bounds observation; cached entry lengths do
// not establish visibility of concurrent memory growth.
func (m *Memory) SharedView() (base, sizeAddress unsafe.Pointer, maximum uint64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.retired {
		return nil, nil, 0, ErrClosed
	}
	if m.sharedSize == nil {
		return nil, nil, 0, ErrInvalid
	}
	return m.data, unsafe.Pointer(m.sharedSize), m.maximum, nil
}
