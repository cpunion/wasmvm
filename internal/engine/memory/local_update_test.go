package memory

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLocalWordUpdateNeverUsesCrossStoreBacking(t *testing.T) {
	for _, offheap := range []bool{false, true} {
		if offheap && !native.Available {
			continue
		}
		d := module.Limits{Min: 1, Max: 1, HasMax: true, Shared: true, Index64: true}
		domain := NewDomain(1, offheap)
		shared, err := NewShared(d, 1, domain)
		if err != nil {
			t.Fatal(err)
		}
		defer shared.Release()
		local, err := NewStoreLocalShared(d, 1, domain)
		if err != nil {
			t.Fatal(err)
		}
		defer local.Release()
		for _, m := range []*Memory{local, shared} {
			if m.Store(0x36, 1, 0, 0, value.Value{Lo: 0xffffffff}) != wasm.NoTrap {
				t.Fatal("initialize unaligned word")
			}
		}
		if _, _, trap, ok := shared.UpdateLocalI32(1, 0, 0x6a, 3); ok || trap != wasm.NoTrap {
			t.Fatal("combined update admitted a shared capability", ok, trap)
		}
		if got, trap := shared.Load(0x28, 1, 0, 0, value.Value{}); got.Lo != 0xffffffff || trap != wasm.NoTrap {
			t.Fatal("declined update changed its backing", got, trap)
		}
		if old, next, trap, ok := local.UpdateLocalI32(1, 0, 0x6a, 3); !ok || trap != wasm.NoTrap || old != 0xffffffff || next != 2 {
			t.Fatal("local integer did not wrap", old, next, trap, ok)
		}
		if _, _, trap, ok := local.UpdateLocalI32(1, ^uint64(0), 0x6a, 3); !ok || trap != wasm.TrapMemoryBounds {
			t.Fatal("overflow in a memory64 addition escaped", ok, trap)
		}
	}
}
