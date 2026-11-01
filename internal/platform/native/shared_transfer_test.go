package native

import (
	"bytes"
	"errors"
	"testing"
	"unsafe"
)

func TestSharedTransfersUnalignedValuesOverlapAndCanaries(t *testing.T) {
	if !Available {
		for _, err := range []error{SharedStore(nil, 1, 0), SharedTransfer(nil, nil, 0, true, true), SharedFill(nil, 0, 0), SharedFence()} {
			if !errors.Is(err, ErrUnavailable) {
				t.Fatal(err)
			}
		}
		if _, err := SharedLoad(nil, 1); !errors.Is(err, ErrUnavailable) {
			t.Fatal(err)
		}
		return
	}
	m, err := NewSharedMemory(128, 128)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	initial := make([]byte, 128)
	for i := range initial {
		initial[i] = byte(i*17 + 31)
	}
	actual := make([]byte, len(initial))
	want := make([]byte, len(initial))
	reset := func() {
		t.Helper()
		if err := m.Write(0, initial); err != nil {
			t.Fatal(err)
		}
		copy(want, initial)
	}
	check := func() {
		t.Helper()
		if err := m.Read(0, actual); err != nil || !bytes.Equal(actual, want) {
			t.Fatal("transfer changed bytes/canaries", err, actual, want)
		}
	}
	for offset := 0; offset < 16; offset++ {
		for _, width := range []uint32{1, 2, 4, 8} {
			reset()
			const value = uint64(0xfedcba9876543210)
			address := unsafe.Add(m.data, offset)
			if err := SharedStore(address, width, value); err != nil {
				t.Fatal(err)
			}
			for i := uint32(0); i < width; i++ {
				want[offset+int(i)] = byte(value >> (i * 8))
			}
			mask := ^uint64(0) >> (64 - width*8)
			if got, err := SharedLoad(address, width); err != nil || got != value&mask {
				t.Fatal("unaligned bits", offset, width, got, err)
			}
			check()
		}
	}
	for _, count := range []int{0, 1, 7, 8, 9, 15, 16, 17, 31, 64, 65} {
		for _, src := range []int{0, 1, 7, 8, 31} {
			for _, dst := range []int{0, 1, 7, 8, 31} {
				reset()
				copy(want[dst:dst+count], want[src:src+count])
				if err := SharedTransfer(unsafe.Add(m.data, dst), unsafe.Add(m.data, src), uint64(count), true, true); err != nil {
					t.Fatal(err)
				}
				check()
			}
			for _, fill := range []byte{0, 0x9b} {
				reset()
				for i := src; i < src+count; i++ {
					want[i] = fill
				}
				if err := SharedFill(unsafe.Add(m.data, src), uint64(count), fill); err != nil {
					t.Fatal(err)
				}
				check()
			}
		}
	}
	if err := SharedFence(); err != nil {
		t.Fatal(err)
	}
	for _, err := range []error{SharedStore(nil, 1, 0), SharedStore(m.data, 3, 0), SharedTransfer(nil, m.data, 1, true, true), SharedTransfer(m.data, m.data, ^uint64(0), true, true), SharedFill(nil, 1, 0), SharedFill(m.data, ^uint64(0), 0)} {
		if !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid transfer", err)
		}
	}
	if _, err := SharedLoad(nil, 1); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := SharedLoad(m.data, 16); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}
