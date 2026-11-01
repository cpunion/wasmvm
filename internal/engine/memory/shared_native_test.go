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
