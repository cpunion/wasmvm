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
