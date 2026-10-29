package canonical

import (
	"context"
	"encoding/binary"
	"errors"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func textTransfer(encoding Encoding, memory64 bool) (*Transfer, *uint64) {
	next := uint64(64)
	c := &Transfer{Memory: make(testMemory, 4096), Memory64: memory64, Encoding: encoding, MaxBytes: 1 << 20}
	c.Realloc = func(old, size, align, n uint64) (uint64, error) {
		if old != 0 || size != 0 {
			panic("not a host allocation")
		}
		ptr := (next + align - 1) &^ (align - 1)
		next = ptr + n
		return ptr, nil
	}
	return c, &next
}

func TestStringEncodingsAndPointerWidths(t *testing.T) {
	for _, wide := range []bool{false, true} {
		for _, encoding := range []Encoding{UTF8, UTF16, Latin1UTF16} {
			for _, text := range []string{"", "ascii\x00end", "éÿ", "中文", "a😀z", "\ufffd", "\U0010ffff"} {
				c, _ := textTransfer(encoding, wide)
				if err := c.StoreString(0, text); err != nil {
					t.Fatal(wide, encoding, text, err)
				}
				got, err := c.LoadString(0)
				if err != nil || got != text {
					t.Fatal(wide, encoding, text, got, err)
				}
				ptr, length, err := c.LowerString(text)
				if err != nil {
					t.Fatal(err)
				}
				got, err = c.LiftString(ptr, length)
				if err != nil || got != text {
					t.Fatal(text, got, err)
				}
				if encoding == Latin1UTF16 {
					tag := uint64(1) << (c.pointerSize()*8 - 1)
					wantUTF16 := text == "中文" || text == "a😀z" || text == "\ufffd" || text == "\U0010ffff"
					if (length&tag != 0) != wantUTF16 || ptr%2 != 0 {
						t.Fatal("compact encoding", text, ptr, length)
					}
				}
			}
		}
	}
	c, _ := textTransfer(UTF16, false)
	ptr, n, err := c.LowerString("a😀z")
	if err != nil || n != 4 {
		t.Fatal(ptr, n, err)
	}
	buf := make([]byte, 8)
	if err := c.Memory.Read(ptr, buf); err != nil {
		t.Fatal(err)
	}
	if binary.LittleEndian.Uint16(buf[2:]) != 0xd83d || binary.LittleEndian.Uint16(buf[4:]) != 0xde00 {
		t.Fatal(buf)
	}
}

func TestStringRejectsMalformedGuestRepresentations(t *testing.T) {
	for _, row := range []struct {
		encoding Encoding
		data     []byte
		length   uint64
	}{
		{UTF8, []byte{0xff}, 1}, {UTF8, []byte{0xc0, 0x80}, 2}, {UTF8, []byte{0xed, 0xa0, 0x80}, 3},
		{UTF16, []byte{0, 0xd8}, 1}, {UTF16, []byte{0, 0xdc}, 1}, {UTF16, []byte{0, 0xd8, 0, 0}, 2},
		{Latin1UTF16, []byte{0, 0xd8}, 0x80000001},
	} {
		c, _ := textTransfer(row.encoding, false)
		c.Memory.Write(64, row.data)
		if _, err := c.LiftString(64, row.length); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(row, err)
		}
	}
	c, _ := textTransfer(UTF8, false)
	if _, err := c.LiftString(0, 1<<32); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if _, err := c.LiftString(0, MaxStringBytes+1); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	c.Encoding = UTF16
	if _, err := c.LiftString(0, MaxStringBytes/2+1); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	for _, row := range [][2]uint64{{65, 0}, {4096, 1}, {math.MaxUint64, 1}} {
		if _, err := c.LiftString(row[0], row[1]); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(row, err)
		}
	}
	if _, err := c.LoadString(1); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if err := c.StoreString(4096, "x"); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if _, _, err := c.LowerString(string([]byte{0xff})); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

type sparseMemory struct{ end uint64 }

func (m sparseMemory) Read(at uint64, b []byte) error {
	if at > m.end || uint64(len(b)) > m.end-at {
		return wasm.ErrTrap
	}
	clear(b)
	return nil
}
func (m sparseMemory) Write(at uint64, b []byte) error { return m.Read(at, b) }

func TestStringTransferQuotasAndAllocatorFailures(t *testing.T) {
	c, _ := textTransfer(UTF8, false)
	c.MaxBytes = 1
	if _, _, err := c.LowerString("ab"); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c, _ = textTransfer(UTF8, false)
	c.MaxBytes = 1
	if _, err := c.LiftString(64, 1); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c, _ = textTransfer(UTF8, false)
	c.Realloc = nil
	if _, _, err := c.LowerString(""); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	c, _ = textTransfer(UTF8, false)
	c.Memory = nil
	if _, err := c.LoadString(0); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	c, _ = textTransfer(Encoding(255), false)
	if _, err := c.LiftString(0, 0); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c, _ = textTransfer(UTF8, false)
	c.Context = ctx
	if _, _, err := c.LowerString("x"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := c.LiftString(0, 0); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for _, ptr := range []uint64{65, 4096, 1 << 32, math.MaxUint64} {
		c, _ := textTransfer(UTF16, false)
		c.Realloc = func(uint64, uint64, uint64, uint64) (uint64, error) { return ptr, nil }
		if _, _, err := c.LowerString("x"); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(ptr, err)
		}
	}
	c, _ = textTransfer(UTF8, false)
	c.Realloc = func(uint64, uint64, uint64, uint64) (uint64, error) { return 0, context.Canceled }
	if err := c.StoreString(0, "x"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	// A Memory32 range may end exactly at 2^32, but its pointer must fit i32.
	c, _ = textTransfer(UTF8, false)
	c.Memory = sparseMemory{1 << 32}
	if _, err := c.LiftString(math.MaxUint32, 1); err != nil {
		t.Fatal("last address", err)
	}
	c.Memory64 = true
	c.Memory = sparseMemory{math.MaxUint64}
	if _, err := c.LiftString(math.MaxUint64-1, 1); err != nil {
		t.Fatal("memory64 end", err)
	}
}
