package memory

import (
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestImmutableMemoryTransfers(t *testing.T) {
	for _, shared := range []bool{false, true} {
		for _, offheap := range []bool{false, true} {
			if offheap && !native.Available {
				continue
			}
			t.Run(fmt.Sprint(shared, "/", offheap), func(t *testing.T) {
				limits := module.Limits{Min: 1, Max: 2, HasMax: true}
				var m *Memory
				var err error
				if shared {
					limits.Shared = true
					m, err = NewShared(limits, 2, NewDomain(1, offheap))
				} else if offheap {
					m, err = NewNative(limits, 2)
				} else {
					m, err = New(limits, 2)
				}
				if err != nil {
					t.Fatal(err)
				}
				data := "\x00\xff\x80hello"
				if !m.WriteString(PageSize-uint64(len(data)), data) {
					t.Fatal("valid write failed")
				}
				got, ok := m.ReadString(PageSize-uint64(len(data)), uint64(len(data)))
				if !ok || got != data {
					t.Fatal(got, ok)
				}
				for _, at := range []uint64{PageSize + 1, math.MaxUint64} {
					if s, ok := m.ReadString(at, 0); ok || s != "" || m.WriteString(at, "") {
						t.Fatal("invalid empty range accepted", at)
					}
				}
				if s, ok := m.ReadString(PageSize, 0); !ok || s != "" || !m.WriteString(PageSize, "") {
					t.Fatal("valid empty range rejected")
				}
				if _, ok := m.ReadString(PageSize-1, 2); ok || m.WriteString(PageSize-1, "xx") {
					t.Fatal("partial out-of-range access")
				}
				if _, ok := m.ReadString(1, math.MaxUint64); ok {
					t.Fatal("overflow accepted")
				}
				if last, _ := m.ReadString(PageSize-1, 1); last != "o" {
					t.Fatal("rejected write mutated memory", last)
				}
				if !m.Fill(0, 7, PageSize) {
					t.Fatal("mutation failed")
				}
				if _, ok := m.Grow(1); !ok {
					t.Fatal("grow failed")
				}
				if !m.WriteString(PageSize, data) {
					t.Fatal("grown memory rejected")
				}
				m.Release()
				if got != data {
					t.Fatal("snapshot changed after mutation/grow/release", got)
				}
			})
		}
	}
}

func TestSharedStringTransfersSerializeWithSliceAccess(t *testing.T) {
	m, err := NewShared(module.Limits{Min: 1, Max: 1, HasMax: true, Shared: true}, 1, NewDomain(1, false))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Release()
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Go(func() {
			for i := 0; i < 100; i++ {
				if !m.WriteString(0, "abcdefgh") {
					t.Error("write")
				}
				if s, ok := m.ReadString(0, 8); !ok || s != "abcdefgh" {
					t.Error(s, ok)
				}
				if !m.Write(0, []byte("abcdefgh")) {
					t.Error("slice write")
				}
			}
		})
	}
	wg.Wait()
}
