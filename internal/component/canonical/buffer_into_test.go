package canonical

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestReadBufferIntoPreservesBytesBoundsAndBudgets(t *testing.T) {
	for _, wide := range []bool{false, true} {
		c := valueCodec(wide)
		source := []byte{0, 255, 128, 42}
		c.Transfer.Memory.Write(64, source)
		dst := make([]byte, len(source))
		if err := c.ReadBufferInto(64, uint64(len(dst)), func() ([]byte, error) { return dst, nil }); err != nil || !bytes.Equal(dst, source) || c.Transfer.used != 8 || c.values != 4 {
			t.Fatal(wide, dst, c.Transfer.used, c.values, err)
		}
		c.Transfer.Memory.Write(64, []byte{1})
		if dst[0] != 0 {
			t.Fatal("destination aliases guest")
		}
		for _, mode := range []string{"values", "bytes", "depth", "bounds", "read", "cancel"} {
			c = valueCodec(wide)
			want := error(wasm.ErrLimit)
			switch mode {
			case "values":
				c.MaxValues = 3
			case "bytes":
				c.Transfer.MaxBytes = 7
			case "depth":
				c.Layouts = NewBuilder(wide, Limits{MaxNodes: 100, MaxDepth: 1, MaxBytes: 1024})
			case "bounds":
				want = wasm.ErrTrap
			case "read":
				want = errors.New("read failure")
				c.Transfer.Memory = byteFaultMemory{Memory: c.Transfer.Memory, read: want}
			case "cancel":
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				c.Transfer.Context = ctx
				c.Transfer.Memory = byteFaultMemory{Memory: c.Transfer.Memory, cancel: cancel}
				want = context.Canceled
			}
			ptr := uint64(64)
			if mode == "bounds" {
				ptr = 4095
			}
			if err := c.ReadBufferInto(ptr, 4, func() ([]byte, error) { return make([]byte, 4), nil }); !errors.Is(err, want) {
				t.Fatal(mode, err)
			}
		}
	}
}

func TestReadBufferIntoChecksBeforeDestinationEffects(t *testing.T) {
	c := valueCodec(false)
	calls := 0
	destination := func() ([]byte, error) { calls++; return make([]byte, 4), nil }
	c.Transfer.MaxBytes = 7
	if err := c.ReadBufferInto(64, 4, destination); !errors.Is(err, wasm.ErrLimit) || calls != 0 {
		t.Fatal(calls, err)
	}
	c = valueCodec(false)
	if err := c.ReadBufferInto(4095, 4, destination); !errors.Is(err, wasm.ErrTrap) || calls != 0 {
		t.Fatal(calls, err)
	}
	sentinel := errors.New("destination failed")
	if err := valueCodec(false).ReadBufferInto(64, 4, func() ([]byte, error) { return nil, sentinel }); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	if err := valueCodec(false).ReadBufferInto(64, 4, func() ([]byte, error) { return nil, nil }); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestReserveByteBufferMatchesPackedCloneLimits(t *testing.T) {
	list := &model.Type{Kind: model.List, Element: typ(model.U8)}
	for _, wide := range []bool{false, true} {
		for _, n := range []uint64{0, 1, 8} {
			actual, oracle := valueCodec(wide), valueCodec(wide)
			if err := actual.ReserveByteBuffer(list, n); err != nil {
				t.Fatal(err)
			}
			if _, err := oracle.Clone(list, Value{Text: string(make([]byte, n))}); err != nil {
				t.Fatal(err)
			}
			if actual.values != oracle.values || actual.Transfer.used != oracle.Transfer.used {
				t.Fatal(actual.values, oracle.values, actual.Transfer.used, oracle.Transfer.used)
			}
		}
	}
	for _, bad := range []*model.Type{nil, typ(model.U8), {Kind: model.List}, {Kind: model.List, Element: typ(model.U16)}} {
		if err := valueCodec(false).ReserveByteBuffer(bad, 1); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	for _, mode := range []string{"context", "layout", "values", "bytes"} {
		c := valueCodec(false)
		want := error(wasm.ErrLimit)
		switch mode {
		case "context":
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			c.Transfer.Context = ctx
			want = context.Canceled
		case "layout":
			c.Layouts = NewBuilder(false, Limits{MaxNodes: 1, MaxDepth: 1, MaxBytes: 1024})
		case "values":
			c.MaxValues = 1
		case "bytes":
			c.Transfer.MaxBytes = 0
		}
		if err := c.ReserveByteBuffer(list, 1); !errors.Is(err, want) {
			t.Fatal(mode, err)
		}
	}
}
