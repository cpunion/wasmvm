package canonical

import (
	"errors"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"sync"
	"testing"
)

func TestOpaqueMovePreservesQuotaAndAuthority(t *testing.T) {
	budget := NewResourceBudget(1)
	a, b := NewResourceTable(budget), NewResourceTable(budget)
	defer a.Close()
	defer b.Close()
	id, _ := a.AddOpaque("value")
	next, err := a.MoveOpaque(id, b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Opaque(id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("source remained valid", err)
	}
	if v, err := b.Opaque(next); err != nil || v != "value" {
		t.Fatal(v, err)
	}
	if budget.live.Load() != 1 {
		t.Fatal("transfer exceeded quota")
	}
	for _, dst := range []*Resources{nil, b, NewResources(1)} {
		if _, err := b.MoveOpaque(next, dst); err == nil {
			t.Fatal("foreign or identical domain")
		}
	}
	a.next = 1 << 28
	if _, err := b.MoveOpaque(next, a); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := b.Opaque(next); err != nil {
		t.Fatal("failed move consumed source", err)
	}
	a.Close()
	if _, err := b.MoveOpaque(next, a); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if _, err := b.MoveOpaque(next+1, a); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	var absent *Resources
	if _, err := absent.MoveOpaque(0, b); err == nil {
		t.Fatal("nil source")
	}
}
func TestOpaqueOppositeTransfersDoNotDeadlock(t *testing.T) {
	budget := NewResourceBudget(2)
	a, b := NewResourceTable(budget), NewResourceTable(budget)
	defer a.Close()
	defer b.Close()
	first, _ := a.AddOpaque(1)
	second, _ := b.AddOpaque(2)
	var wg sync.WaitGroup
	for _, work := range []struct {
		src, dst *Resources
		id       uint32
	}{{a, b, first}, {b, a, second}} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			src, dst, id := work.src, work.dst, work.id
			for range 1000 {
				var err error
				id, err = src.MoveOpaque(id, dst)
				if err != nil {
					t.Error(err)
					return
				}
				src, dst = dst, src
			}
		}()
	}
	wg.Wait()
	if budget.live.Load() != 2 {
		t.Fatal("transfer lost live quota")
	}
}
