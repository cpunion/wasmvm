package canonical

import (
	"bytes"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

type descriptorMemory struct {
	Memory
	reads, writes int
	failure       error
}

func (m *descriptorMemory) Read(at uint64, dst []byte) error {
	m.reads++
	if m.failure != nil {
		return m.failure
	}
	return m.Memory.Read(at, dst)
}
func (m *descriptorMemory) Write(at uint64, src []byte) error {
	m.writes++
	if m.failure != nil {
		return m.failure
	}
	return m.Memory.Write(at, src)
}

func TestPointerLengthDescriptorsKeepFullWidthAndCheckedCopies(t *testing.T) {
	for _, wide := range []bool{false, true} {
		c := valueCodec(wide)
		storage := testMemory(bytes.Repeat([]byte{0xa5}, 32))
		m := &descriptorMemory{Memory: storage}
		c.Transfer.Memory = m
		ptr, n := uint64(0xfedcba98), uint64(0x89abcdef)
		if wide {
			ptr, n = ptr<<32|0x12345678, n<<32|0x76543210
		}
		if err := c.storePair(8, ptr, n); err != nil {
			t.Fatal(err)
		}
		a, b, err := c.loadPair(8)
		if err != nil || a != ptr || b != n || m.reads != 1 || m.writes != 1 {
			t.Fatal(wide, a, b, err, m.reads, m.writes)
		}
		before := append([]byte(nil), storage...)
		if err := c.storePair(32-c.Transfer.pointerSize(), 1, 2); !errors.Is(err, wasm.ErrTrap) || !bytes.Equal(storage, before) {
			t.Fatal("short descriptor changed guest bytes", wide, err)
		}
		if _, _, err := c.loadPair(^uint64(0)); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("overflowing descriptor accepted", err)
		}
		m.failure = errors.New("descriptor transfer failure")
		if _, _, err := c.loadPair(8); err != m.failure {
			t.Fatal("read failure lost", err)
		}
		if err := c.storePair(8, ptr, n); err != m.failure {
			t.Fatal("write failure lost", err)
		}
	}
}
