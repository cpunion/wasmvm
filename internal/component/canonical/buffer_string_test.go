package canonical

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type cancelBufferMemory struct {
	Memory
	cancel context.CancelFunc
}

func (m cancelBufferMemory) Write(at uint64, data []byte) error {
	err := m.Memory.Write(at, data)
	m.cancel()
	return err
}

func TestPackedBufferUsesImmutableMemoryAndSameBudget(t *testing.T) {
	c := valueCodec(false)
	m := &snapshotMemory{Memory: c.Transfer.Memory}
	c.Transfer.Memory = m
	data := "\x00\xff\x80hello"
	if err := c.StoreBuffer(typ(model.U8), 64, Value{Text: data}); err != nil {
		t.Fatal(err)
	}
	if m.writes != 1 || c.Transfer.used != uint64(len(data)) || c.values != uint64(len(data)) {
		t.Fatal(m.writes, c.Transfer.used, c.values)
	}
	got := make([]byte, len(data))
	if err := m.Read(64, got); err != nil || string(got) != data {
		t.Fatal(got, err)
	}
	c = valueCodec(false)
	c.Transfer.MaxBytes = uint64(len(data) - 1)
	if err := c.StoreBuffer(typ(model.U8), 64, Value{Text: data}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}

func TestPackedBufferRejectsCancellationDuringMemoryWrite(t *testing.T) {
	for _, immutable := range []bool{false, true} {
		c := valueCodec(false)
		ctx, cancel := context.WithCancel(context.Background())
		c.Transfer.Context = ctx
		if immutable {
			c.Transfer.Memory = &snapshotMemory{Memory: c.Transfer.Memory, cancel: cancel}
		} else {
			c.Transfer.Memory = cancelBufferMemory{Memory: c.Transfer.Memory, cancel: cancel}
		}
		err := c.StoreBuffer(typ(model.U8), 64, Value{Text: "data"})
		cancel()
		if !errors.Is(err, context.Canceled) {
			t.Fatal(immutable, err)
		}
	}
}
