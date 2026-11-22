package memory

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNativeSharedWordArithmetic(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	d := NewDomain(1, true)
	m, err := NewShared(module.Limits{Min: 8, Max: 16, HasMax: true, Shared: true, PageSize: 1, Index64: true}, 1, d)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Release()
	active := context.Background()
	canceled, cancel := context.WithCancel(active)
	cancel()
	other := NewDomain(1, true)

	for _, check := range []struct {
		name            string
		ctx             context.Context
		domain          *Domain
		address, offset uint64
		trap            wasm.TrapCode
		err             error
	}{
		{"offset wrap", active, d, math.MaxUint64, 1, wasm.TrapMemoryBounds, nil},
		{"end wrap", active, d, math.MaxUint64 - 1, 0, wasm.TrapMemoryBounds, nil},
		{"reserved page", active, d, 8, 0, wasm.TrapMemoryBounds, nil},
		{"crosses logical end", active, d, 6, 0, wasm.TrapMemoryBounds, nil},
		{"misaligned", active, d, 2, 0, wasm.TrapAtomicUnaligned, nil},
		{"canceled", canceled, d, 0, 0, wasm.NoTrap, context.Canceled},
		{"cancellation before trap", canceled, d, math.MaxUint64, 1, wasm.NoTrap, context.Canceled},
		{"cross-domain", active, other, 0, 0, wasm.NoTrap, wasm.ErrInvalid},
	} {
		t.Run(check.name, func(t *testing.T) {
			_, trap, err := m.Atomic(check.ctx, check.domain, 0x1e, check.address, check.offset, 99, 0)
			if trap != check.trap || !errors.Is(err, check.err) {
				t.Fatalf("trap = %v, err = %v; want %v, %v", trap, err, check.trap, check.err)
			}
		})
	}
	check := func(op uint32, offset, delta, old uint64) {
		t.Helper()
		got, trap, err := m.Atomic(active, d, op, 0, offset, delta, 0)
		if got.Lo != old || trap != wasm.NoTrap || err != nil {
			t.Fatalf("opcode %#x: got %#x, %v, %v; want %#x", op, got.Lo, trap, err, old)
		}
	}
	check(0x17, 0, uint64(math.MaxUint32)+2, 0)
	check(0x10, 0, 0, 1)
	check(0x17, 0, 0, 0)
	check(0x1e, 0, math.MaxUint32, 0)
	check(0x24, 0, 3, math.MaxUint32)
	check(0x25, 0, 4, 2)
	check(0x2b, 0, math.MaxUint32-1, math.MaxUint32-1)
	if old, ok := m.Grow(8); old != 8 || !ok {
		t.Fatal(old, ok)
	}
	check(0x18, 8, 0xaabbccdd11223344, 0)
	check(0x11, 8, 0, 0xaabbccdd11223344)
	check(0x18, 8, 0, 0)
	check(0x1f, 8, math.MaxUint64, 0)
	check(0x26, 8, math.MaxUint64, math.MaxUint64)
	check(0x11, 8, 0, 0)
}
