package memory

import (
	"fmt"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNativeSharedReservationKeepsLogicalBoundsAndStableGrowth(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, pageSize := range []uint64{1, PageSize} {
		for _, index64 := range []bool{false, true} {
			t.Run(fmt.Sprint(pageSize, "/", index64), func(t *testing.T) {
				declared := module.Limits{Min: 1, Max: 3, HasMax: true, Shared: true, PageSize: pageSize, Index64: index64}
				m, err := NewShared(declared, 3, NewDomain(1, true))
				if err != nil {
					t.Fatal(err)
				}
				defer m.Release()
				base := unsafe.SliceData(m.bytes)
				if len(m.bytes) != int(3*pageSize) || m.Pages() != 1 || m.sharedSize == nil || atomic.LoadUint64(m.sharedSize) != pageSize {
					t.Fatal("reservation and accessible size were conflated")
				}
				if m.Write(pageSize, []byte{9}) || m.Fill(pageSize, 9, 1) || m.Store(0x3a, pageSize, 0, 0, value.Value{Lo: 9}) != wasm.TrapMemoryBounds {
					t.Fatal("reservation was guest-accessible before growth")
				}
				if !m.Write(0, []byte{42}) {
					t.Fatal("initial write")
				}
				if old, ok := m.Grow(2); !ok || old != 1 || m.Pages() != 3 || atomic.LoadUint64(m.sharedSize) != 3*pageSize || unsafe.SliceData(m.bytes) != base {
					t.Fatal("growth moved storage or failed size publication", old, ok)
				}
				if m.bytes[0] != 42 {
					t.Fatal("growth lost old data")
				}
				for i, b := range m.bytes[pageSize:] {
					if b != 0 {
						t.Fatal("growth did not expose zero bytes", i, b)
					}
				}
				if old, ok := m.Grow(1); ok || old != 3 || atomic.LoadUint64(m.sharedSize) != 3*pageSize {
					t.Fatal("failed growth changed logical size")
				}
				if !m.Read(3*pageSize, nil) || m.Write(3*pageSize, []byte{1}) {
					t.Fatal("logical end bounds")
				}
			})
		}
	}
}

func TestSharedAtomicCheckedAddressAcrossGrowth(t *testing.T) {
	for _, offheap := range []bool{false, true} {
		if offheap && !native.Available {
			continue
		}
		for _, index64 := range []bool{false, true} {
			t.Run(fmt.Sprint(offheap, "/", index64), func(t *testing.T) {
				domain := NewDomain(1, offheap)
				limits := module.Limits{Min: 16, Max: 32, HasMax: true, Shared: true, PageSize: 1, Index64: index64}
				m, err := NewShared(limits, 1, domain)
				if err != nil {
					t.Fatal(err)
				}
				defer m.Release()
				check := func(op uint32, address, offset uint64, expected wasm.TrapCode) {
					t.Helper()
					got, trap, err := m.Atomic(nil, domain, op, address, offset, 0, 0)
					if trap != expected || err != nil || got != (value.Value{}) {
						t.Fatalf("op %#x address %#x offset %#x = %+v, %v, %v", op, address, offset, got, trap, err)
					}
				}
				widths := [...]uint64{4, 8, 1, 2}
				for i, width := range widths {
					op := uint32(0x10 + i)
					check(op, 16-width, 0, wasm.NoTrap)
					check(op, 17-width, 0, wasm.TrapMemoryBounds)
					if width > 1 {
						check(op, 1, 0, wasm.TrapAtomicUnaligned)
					}
					check(op, 1, ^uint64(0), wasm.TrapMemoryBounds)
					if !index64 {
						check(op, 1<<32, 0, wasm.NoTrap)
					} else {
						check(op, 1<<32, 0, wasm.TrapMemoryBounds)
					}
				}
				if old, ok := m.Grow(8); !ok || old != 16 {
					t.Fatal("grow", old, ok)
				}
				for i := range widths {
					check(uint32(0x10+i), 16, 0, wasm.NoTrap)
				}
			})
		}
	}
}

func TestLocalNativeSharedGrowthUpdatesBothLogicalViews(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, pageSize := range []uint64{1, PageSize} {
		t.Run(fmt.Sprint(pageSize), func(t *testing.T) {
			declared := module.Limits{Max: 3, HasMax: true, Shared: true, PageSize: pageSize, Index64: true}
			m, err := NewStoreLocalShared(declared, 3, NewDomain(1, true))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Release()
			owner := m.NativeShared()
			if owner == nil || m.Shareable() || m.Retain() || m.sharedSize != nil {
				t.Fatal("Store-local backing escaped or lost its native owner")
			}
			base, published, maximum, err := owner.SharedView()
			if err != nil || maximum != 3*pageSize || len(m.bytes) != 0 || cap(m.bytes) != int(maximum) {
				t.Fatal("wrong initial logical view", maximum, len(m.bytes), cap(m.bytes), err)
			}
			if m.Write(0, []byte{9}) || m.Range(0, 1) {
				t.Fatal("empty memory accessed before growth")
			}
			if old, ok := m.Grow(1); !ok || old != 0 {
				t.Fatal(old, ok)
			}
			if m.Pages() != 1 || atomic.LoadUint64((*uint64)(published)) != pageSize ||
				unsafe.Pointer(unsafe.SliceData(m.bytes)) != base || len(m.bytes) != int(pageSize) {
				t.Fatal("growth did not update both views in place")
			}
			if !m.Write(0, []byte{42}) || m.Write(pageSize, []byte{9}) {
				t.Fatal("wrong logical write bounds after growth")
			}
			if old, ok := m.Grow(2); !ok || old != 1 {
				t.Fatal(old, ok)
			}
			if got, trap := m.Load(0x2d, 0, 0, 0, value.Value{}); got.Lo != 42 || trap != wasm.NoTrap {
				t.Fatal("ordinary access lost old bytes", got, trap)
			}
			for i, v := range m.bytes[pageSize:] {
				if v != 0 {
					t.Fatal("growth did not expose zero bytes", i, v)
				}
			}
			if old, ok := m.Grow(1); ok || old != 3 || len(m.bytes) != int(maximum) ||
				atomic.LoadUint64((*uint64)(published)) != maximum {
				t.Fatal("failed growth altered a logical bound", old, ok)
			}
		})
	}
}
