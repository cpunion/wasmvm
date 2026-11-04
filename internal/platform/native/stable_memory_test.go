package native

import (
	"bytes"
	"testing"
	"unsafe"
)

func TestStableMemoryCommitsZeroPagesWithoutMovingAndRetiresBindings(t *testing.T) {
	if !Available {
		if _, err := NewStableMemory(0, 0); err != ErrUnavailable {
			t.Fatal(err)
		}
		return
	}
	for _, initial := range []uint64{0, 1, 16383} {
		const maximum = 65539
		m, err := NewStableMemory(initial, maximum)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(m.Close)
		base := m.data
		want := make([]byte, initial)
		for _, size := range []uint64{initial, max(initial, 16385), maximum} {
			previous := len(want)
			if err := m.Grow(size); err != nil || m.data != base {
				t.Fatal("growth changed stable address", size, err)
			}
			want = append(want, make([]byte, int(size)-previous)...)
			buffer, err := m.Buffer()
			if err != nil || !bytes.Equal(buffer, want) || unsafe.Pointer(unsafe.SliceData(buffer)) != base {
				t.Fatal("growth did not preserve bytes/zero its extension", size, err)
			}
			for i := range buffer {
				buffer[i] = byte(i*31 + int(size))
			}
			copy(want, buffer)
		}
		for _, size := range []uint64{maximum - 1, maximum + 1, ^uint64(0)} {
			if err := m.Grow(size); err == nil || m.size != maximum || m.data != base {
				t.Fatal("rejected growth changed storage", size, err)
			}
		}
		ctx, err := NewContext(0, 0, 1)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(ctx.Close)
		if err := ctx.BindMemory(m); err != nil {
			t.Fatal(err)
		}
		if _, _, err := m.begin(); err != nil {
			t.Fatal(err)
		}
		if err := m.Grow(maximum); err != ErrBusy {
			t.Fatal("active private storage grew", err)
		}
		if _, err := m.Buffer(); err != ErrBusy {
			t.Fatal("active storage exposed", err)
		}
		m.Close()
		m.end()
		if m.data != base || m.reserved == 0 {
			t.Fatal("bound retired storage released early")
		}
		if err := m.Grow(maximum); err != ErrClosed {
			t.Fatal(err)
		}
		if err := ctx.BindMemory(nil); err != nil || m.data != nil || m.reserved != 0 {
			t.Fatal("last binding did not release reservation", err)
		}
		m.Close()
	}
	zero, err := NewStableMemory(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if buffer, err := zero.Buffer(); err != nil || len(buffer) != 0 {
		t.Fatal(buffer, err)
	}
	zero.Close()
	for _, bounds := range [][2]uint64{{2, 1}, {0, ^uint64(0)}, {0, uint64(^uint(0) >> 1)}} {
		if m, err := NewStableMemory(bounds[0], bounds[1]); err != ErrLimit || m != nil {
			t.Fatal("invalid/unreservable extent accepted", bounds, m, err)
		}
	}
}
