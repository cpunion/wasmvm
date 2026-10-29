package native

import (
	"bytes"
	"testing"
	"time"
)

func TestMemoryOwnershipAndGrowth(t *testing.T) {
	if !Available {
		if _, err := NewMemory(0, 0); err != ErrUnavailable {
			t.Fatal(err)
		}
		return
	}
	for _, size := range []uint64{0, 1, 32} {
		m, err := NewMemory(size, 64)
		if err != nil {
			t.Fatal(err)
		}
		data := make([]byte, size)
		if err = m.Read(0, data); err != nil || !bytes.Equal(data, make([]byte, size)) {
			t.Fatal(data, err)
		}
		if size > 0 {
			data[0] = 42
			if err = m.Write(0, data); err != nil {
				t.Fatal(err)
			}
		}
		if err = m.Grow(size); err != nil {
			t.Fatal(err)
		}
		if err = m.Grow(64); err != nil {
			t.Fatal(err)
		}
		got := make([]byte, 64)
		if err = m.Read(0, got); err != nil {
			t.Fatal(err)
		}
		want := make([]byte, 64)
		copy(want, data)
		if !bytes.Equal(got, want) {
			t.Fatal(got, want)
		}
		if err = m.Grow(63); err != ErrInvalid {
			t.Fatal(err)
		}
		for _, bad := range []uint64{65, ^uint64(0)} {
			if err = m.Grow(bad); err != ErrLimit {
				t.Fatal(err)
			}
		}
		if n, err := m.Size(); err != nil || n != 64 {
			t.Fatal(n, err)
		}
		for _, at := range []uint64{64, 65, ^uint64(0)} {
			out := []byte{99}
			if err = m.Read(at, out); err != ErrLimit || out[0] != 99 {
				t.Fatal(err, out)
			}
			if err = m.Write(at, []byte{9}); err != ErrLimit {
				t.Fatal(err)
			}
		}
		if err = m.Read(64, nil); err != nil {
			t.Fatal(err)
		}
		if err = m.Write(64, nil); err != nil {
			t.Fatal(err)
		}
		c, err := NewContext(0, 0, 1)
		if err != nil {
			t.Fatal(err)
		}
		if err = c.BindMemory(m); err != nil {
			t.Fatal(err)
		}
		if err = c.BindMemory(m); err != nil || m.bindings != 1 {
			t.Fatal(err, m.bindings)
		}
		m.Close()
		m.Close()
		if m.data == nil {
			t.Fatal("freed a bound memory")
		}
		if _, err = m.Size(); err != ErrClosed {
			t.Fatal(err)
		}
		if err = m.Read(0, nil); err != ErrClosed {
			t.Fatal(err)
		}
		if err = m.Write(0, nil); err != ErrClosed {
			t.Fatal(err)
		}
		if err = m.Grow(64); err != ErrClosed {
			t.Fatal(err)
		}
		other, err := NewContext(0, 0, 1)
		if err != nil {
			t.Fatal(err)
		}
		if err = other.BindMemory(m); err != ErrClosed {
			t.Fatal(err)
		}
		if err = c.BindMemory(nil); err != nil || m.data != nil {
			t.Fatal("retirement did not drain", err)
		}
		c.Close()
		other.Close()
		if err = c.BindMemory(nil); err != ErrClosed {
			t.Fatal(err)
		}
	}
	if _, err := NewMemory(2, 1); err != ErrLimit {
		t.Fatal(err)
	}
	if _, err := NewMemory(^uint64(0), ^uint64(0)); err != ErrLimit {
		t.Fatal(err)
	}
}

func TestMemoryExcludesConcurrentNativeAccess(t *testing.T) {
	if !Available {
		return
	}
	_, lease, c := makeProbe(t, cancelProbe())
	m, err := NewMemory(16, 32)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err = c.BindMemory(m); err != nil {
		t.Fatal(err)
	}
	other, err := NewContext(0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if err = other.BindMemory(m); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		exit, err := c.Enter(lease, 0)
		if err == nil && exit != Poll {
			err = ErrInvalid
		}
		done <- err
	}()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err = c.State(); err == ErrBusy {
			break
		}
		if time.Now().After(deadline) {
			c.Cancel()
			t.Fatal("native entry did not start")
		}
		time.Sleep(time.Millisecond)
	}
	if err = m.Read(0, nil); err != ErrBusy {
		t.Fatal(err)
	}
	if err = m.Write(0, nil); err != ErrBusy {
		t.Fatal(err)
	}
	if err = m.Grow(32); err != ErrBusy {
		t.Fatal(err)
	}
	if err = c.BindMemory(nil); err != ErrBusy {
		t.Fatal(err)
	}
	if _, err = other.Enter(lease, 0); err != ErrBusy {
		t.Fatal(err)
	}
	// Retiring owner/context during execution must not reclaim the active buffer.
	m.Close()
	c.Close()
	other.Close()
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	if m.data != nil {
		t.Fatal("closed memory retained after all bindings drained")
	}
}
