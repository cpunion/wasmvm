package native

import (
	"bytes"
	"errors"
	"testing"
)

func TestArenaGrowthPreservesFramesAndLimits(t *testing.T) {
	if !Available {
		t.Skip("native platform unavailable")
	}
	c, err := NewContext(16, 128, 4)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	input := bytes.Repeat([]byte{0xa5}, 16)
	if err = c.Write(0, input); err != nil {
		t.Fatal(err)
	}
	view, err := c.ResizeArena(128, 128)
	if err != nil || len(view) != 128 || !bytes.Equal(view[:16], input) || !bytes.Equal(view[16:], make([]byte, 112)) {
		t.Fatal(view, err)
	}
	view[64] = 99
	if _, err = c.ResizeArena(129, 128); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	var read [1]byte
	if err = c.Read(64, read[:]); err != nil || read[0] != 99 {
		t.Fatal(read, err)
	}
	if _, err = c.ResizeArena(0, 128); err != nil {
		t.Fatal(err)
	}
	c.Close()
	if _, err = c.ResizeArena(1, 128); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}

func TestMemoryViewOwnership(t *testing.T) {
	if !Available {
		t.Skip("native platform unavailable")
	}
	m, err := NewMemory(16, 32)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	view, err := m.Buffer()
	if err != nil || len(view) != 16 {
		t.Fatal(err)
	}
	view[3] = 42
	if _, _, err = m.begin(); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Buffer(); !errors.Is(err, ErrBusy) {
		t.Fatal(err)
	}
	m.end()
	if err = m.Grow(32); err != nil {
		t.Fatal(err)
	}
	view, err = m.Buffer()
	if err != nil || view[3] != 42 || !bytes.Equal(view[16:], make([]byte, 16)) {
		t.Fatal(view, err)
	}
	m.Close()
	if _, err = m.Buffer(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}
