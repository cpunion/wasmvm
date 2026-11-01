package native

import (
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"
)

func TestSharedMemoryAtomicSizePublicationAndBindingReclamation(t *testing.T) {
	if !Available {
		t.Skip("native platform unavailable")
	}
	m, err := NewSharedMemory(0, 8192)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	c, err := NewContext(0, 16, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.BindMemory(m); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.begin(); err != nil {
		t.Fatal(err)
	}
	size := m.sharedSize
	done := make(chan error, 1)
	go func() {
		for i := uint64(1); i <= 8192; i++ {
			if err := m.Grow(i); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	previous := uint64(0)
	for {
		next := atomic.LoadUint64(size)
		if next < previous || next > 8192 {
			t.Fatal("size publication regressed", previous, next)
		}
		previous = next
		select {
		case err := <-done:
			if err != nil || atomic.LoadUint64(size) != 8192 {
				t.Fatal(err)
			}
			m.Close()
			m.end()
			if m.data == nil || m.sharedSize == nil {
				t.Fatal("binding lost retired storage")
			}
			if err := c.BindMemory(nil); err != nil || m.data != nil || m.sharedSize != nil {
				t.Fatal("last unbind retained storage", err)
			}
			return
		default:
			runtime.Gosched()
		}
	}
}

func TestSharedMemoryRejectsInvalidViewsAndSizes(t *testing.T) {
	if !Available {
		t.Skip("native platform unavailable")
	}
	for _, bounds := range [][2]uint64{{2, 1}, {0, ^uint64(0)}} {
		if _, err := NewSharedMemory(bounds[0], bounds[1]); !errors.Is(err, ErrLimit) {
			t.Fatal(bounds, err)
		}
	}
	private, err := NewMemory(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer private.Close()
	if _, _, _, err := private.SharedView(); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	m, err := NewSharedMemory(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.Grow(0); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	m.active = ^uint64(0)
	if _, _, err := m.begin(); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	m.active = 0
	m.Close()
	if _, _, _, err := m.SharedView(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}

func TestSharedMemoryStableGrowthAndConcurrentEntryLifetime(t *testing.T) {
	if !Available {
		t.Skip("native platform unavailable")
	}
	m, err := NewSharedMemory(0, 4096)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	base, sizeAddress, maximum, err := m.SharedView()
	if err != nil || base == 0 || sizeAddress == 0 || maximum != 4096 || atomic.LoadUint64(m.sharedSize) != 0 {
		t.Fatal("shared descriptor", base, sizeAddress, maximum, err)
	}
	for i := 0; i < 2; i++ {
		address, size, err := m.begin()
		if err != nil || address != base || size != 0 {
			t.Fatal(address, size, err)
		}
	}
	if err := m.Grow(4096); err != nil || atomic.LoadUint64(m.sharedSize) != 4096 {
		t.Fatal("concurrent growth", err)
	}
	if err := m.Grow(4097); !errors.Is(err, ErrLimit) || atomic.LoadUint64(m.sharedSize) != 4096 {
		t.Fatal("growth limit changed size", err)
	}
	if nextBase, nextSize, _, err := m.SharedView(); err != nil || nextBase != base || nextSize != sizeAddress {
		t.Fatal("shared view relocated", err)
	}
	if _, err := m.Buffer(); !errors.Is(err, ErrBusy) {
		t.Fatal("borrow during execution", err)
	}
	bytes := unsafe.Slice((*byte)(m.data), 4096)
	for i, b := range bytes {
		if b != 0 {
			t.Fatal("growth not zeroed", i)
		}
	}
	m.Close()
	if m.data == nil || m.sharedSize == nil {
		t.Fatal("active entries lost storage")
	}
	m.end()
	if m.data == nil || m.active != 1 {
		t.Fatal("first exit reclaimed shared owner")
	}
	m.end()
	if m.data != nil || m.sharedSize != nil || m.active != 0 {
		t.Fatal("final exit did not reclaim shared owner")
	}
	if _, _, err := m.begin(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}
