package canonical

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type snapshotMemory struct {
	Memory
	read, write error
	cancel      context.CancelFunc
	short       bool
	reads       int
	writes      int
}

func (m *snapshotMemory) ReadString(at, n uint64) (string, error) {
	m.reads++
	if m.read != nil {
		return "", m.read
	}
	data := make([]byte, int(n))
	if err := m.Memory.Read(at, data); err != nil {
		return "", err
	}
	if m.cancel != nil {
		m.cancel()
	}
	if m.short && n != 0 {
		data = data[:n-1]
	}
	return string(data), nil
}
func (m *snapshotMemory) WriteString(at uint64, data string) error {
	m.writes++
	if m.write != nil {
		return m.write
	}
	if m.cancel != nil {
		m.cancel()
	}
	return m.Memory.Write(at, []byte(data))
}

func TestPackedBytesWithImmutableMemory(t *testing.T) {
	for _, wide := range []bool{false, true} {
		for _, data := range []string{"", "\x00\xff\x80hello"} {
			for _, fixed := range []bool{false, true} {
				c := valueCodec(wide)
				m := &snapshotMemory{Memory: c.Transfer.Memory}
				c.Transfer.Memory = m
				list := &model.Type{Kind: model.List, Element: typ(model.U8)}
				if fixed {
					list.Kind, list.Length = model.FixedList, uint32(len(data))
				}
				if fixed && data == "" {
					if err := c.Store(list, 0, Value{}); !errors.Is(err, wasm.ErrInvalid) {
						t.Fatal("empty fixed list accepted", err)
					}
					continue
				}
				if err := c.Store(list, 0, Value{Text: data}); err != nil {
					t.Fatal(err)
				}
				got, err := c.Load(list, 0)
				if err != nil || got.Text != data || m.reads != 1 || m.writes != 1 {
					t.Fatal(got, err, m.reads, m.writes)
				}
				// Logical transfer accounting matches the slice-only adapter.
				plain := valueCodec(wide)
				if err := plain.Store(list, 0, Value{Text: data}); err != nil {
					t.Fatal(err)
				}
				if _, err := plain.Load(list, 0); err != nil || plain.Transfer.used != c.Transfer.used {
					t.Fatal("changed transfer budget", err, plain.Transfer.used, c.Transfer.used)
				}
			}
		}
	}
}

func TestImmutableByteMemoryFailures(t *testing.T) {
	sentinel := errors.New("snapshot memory failure")
	for _, mode := range []string{"read", "write", "read-cancel", "write-cancel", "short", "limit", "bounds"} {
		t.Run(mode, func(t *testing.T) {
			c := valueCodec(false)
			m := &snapshotMemory{Memory: c.Transfer.Memory}
			c.Transfer.Memory = m
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			c.Transfer.Context = ctx
			want := sentinel
			switch mode {
			case "read":
				m.read = sentinel
			case "write":
				m.write = sentinel
			case "read-cancel", "write-cancel":
				m.cancel, want = cancel, context.Canceled
			case "short":
				m.short, want = true, wasm.ErrInvalid
			case "limit":
				c.Transfer.MaxBytes, want = 1, wasm.ErrLimit
			case "bounds":
				want = wasm.ErrTrap
			}
			var err error
			if mode == "write" || mode == "write-cancel" {
				err = c.storeBytes(64, "xx")
			} else if mode == "bounds" {
				_, err = c.loadBytes(4096, 2, 0)
			} else {
				_, err = c.loadBytes(64, 2, 0)
			}
			if !errors.Is(err, want) {
				t.Fatal(err, want)
			}
			if (mode == "limit" || mode == "bounds") && m.reads != 0 {
				t.Fatal("allocated before range/limit check")
			}
		})
	}
}

func TestUnicodeStringsWithImmutableMemory(t *testing.T) {
	for _, wide := range []bool{false, true} {
		for _, encoding := range []Encoding{UTF8, UTF16, Latin1UTF16} {
			for _, data := range []string{"", "ascii\x00end", "éÿ", "中文😀"} {
				plain, _ := textTransfer(encoding, wide)
				fast, _ := textTransfer(encoding, wide)
				memory := &snapshotMemory{Memory: fast.Memory}
				fast.Memory = memory
				if err := plain.StoreString(0, data); err != nil {
					t.Fatal(err)
				}
				if err := fast.StoreString(0, data); err != nil {
					t.Fatal(err)
				}
				got, err := fast.LoadString(0)
				if err != nil || got != data {
					t.Fatal(wide, encoding, data, got, err)
				}
				if _, err := plain.LoadString(0); err != nil || fast.used != plain.used {
					t.Fatal("changed logical byte accounting", err, fast.used, plain.used)
				}
				want := 0
				if encoding == UTF8 {
					want = 1
				}
				if memory.reads != want || memory.writes != want {
					t.Fatal("unexpected encoding path", encoding, memory.reads, memory.writes)
				}
				// Later guest writes cannot change the lifted immutable snapshot.
				if err := memory.Memory.Write(64, []byte{0, 0, 0, 0}); err != nil || got != data {
					t.Fatal("snapshot aliases guest memory", err, got)
				}
			}
		}
	}
}

func TestImmutableUnicodeMemoryFailures(t *testing.T) {
	sentinel := errors.New("Unicode memory failure")
	for _, wide := range []bool{false, true} {
		for _, mode := range []string{"read", "write", "read-cancel", "write-cancel", "short", "utf8", "host-utf8", "read-limit", "write-limit", "bounds", "realloc-bounds"} {
			c, _ := textTransfer(UTF8, wide)
			memory := &snapshotMemory{Memory: c.Memory}
			c.Memory = memory
			ctx, cancel := context.WithCancel(context.Background())
			c.Context = ctx
			want := sentinel
			switch mode {
			case "read":
				memory.read = sentinel
			case "write":
				memory.write = sentinel
			case "read-cancel", "write-cancel":
				memory.cancel, want = cancel, context.Canceled
			case "short":
				memory.short, want = true, wasm.ErrInvalid
			case "utf8":
				memory.Memory.Write(64, []byte{0xc0, 0x80})
				want = wasm.ErrTrap
			case "host-utf8":
				want = wasm.ErrInvalid
			case "read-limit":
				c.MaxBytes, want = 3, wasm.ErrLimit
			case "write-limit":
				c.MaxBytes, want = 1, wasm.ErrLimit
			case "bounds", "realloc-bounds":
				want = wasm.ErrTrap
			}
			var err error
			switch mode {
			case "write", "write-cancel", "write-limit":
				_, _, err = c.LowerString("xx")
			case "host-utf8":
				_, _, err = c.LowerString("\xff")
			case "bounds":
				_, err = c.LiftString(4096, 2)
			case "realloc-bounds":
				c.Realloc = func(uint64, uint64, uint64, uint64) (uint64, error) { return 4096, nil }
				_, _, err = c.LowerString("xx")
			default:
				_, err = c.LiftString(64, 2)
			}
			cancel()
			if !errors.Is(err, want) {
				t.Fatal(wide, mode, err, want)
			}
			if mode == "read-limit" || mode == "write-limit" || mode == "bounds" || mode == "realloc-bounds" || mode == "host-utf8" {
				if memory.reads != 0 || memory.writes != 0 {
					t.Fatal("transfer before validation", mode, memory.reads, memory.writes)
				}
			}
		}
	}
}
