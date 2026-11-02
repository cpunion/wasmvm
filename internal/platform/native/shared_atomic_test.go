package native

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"unsafe"
)

func TestSharedAtomicWidthsBitPatternsAndCompareExchange(t *testing.T) {
	if !Available {
		if _, err := SharedAtomic(nil, 1, AtomicLoad, 0, 0); !errors.Is(err, ErrUnavailable) {
			t.Fatal(err)
		}
		return
	}
	m, err := NewSharedMemory(64, 64)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := SharedAtomicContext(ctx, m.data, 8, AtomicAdd, 1, 0); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled atomic committed", err)
	}
	if old, err := SharedAtomic(m.data, 8, AtomicLoad, 0, 0); err != nil || old != 0 {
		t.Fatal("canceled atomic changed data", old, err)
	}
	rng := rand.New(rand.NewSource(64))
	for _, width := range []uint32{1, 2, 4, 8} {
		mask := ^uint64(0) >> (64 - width*8)
		for i := 0; i < 100; i++ {
			old, a, b := rng.Uint64()&mask, rng.Uint64(), rng.Uint64()
			for operation := AtomicLoad; operation <= AtomicCompareExchange; operation++ {
				if _, err := SharedAtomic(m.data, width, AtomicStore, old, 0); err != nil {
					t.Fatal(err)
				}
				x := a
				if operation == AtomicCompareExchange && i%2 == 0 {
					x = old | (a & ^mask)
				}
				want, next := old, old
				switch operation {
				case AtomicStore:
					want, next = 0, x
				case AtomicAdd:
					next = old + x
				case AtomicSub:
					next = old - x
				case AtomicAnd:
					next = old & x
				case AtomicOr:
					next = old | x
				case AtomicXor:
					next = old ^ x
				case AtomicExchange:
					next = x
				case AtomicCompareExchange:
					if old == x&mask {
						next = b
					}
				}
				got, err := SharedAtomic(m.data, width, operation, x, b)
				stored, loadErr := SharedAtomic(m.data, width, AtomicLoad, 0, 0)
				if err != nil || loadErr != nil || got != want || stored != next&mask {
					t.Fatalf("width=%d op=%d old=%x a=%x b=%x: got=%x stored=%x err=%v/%v", width, operation, old, x, b, got, stored, err, loadErr)
				}
			}
		}
	}
	for _, args := range []struct {
		address unsafe.Pointer
		width   uint32
		op      AtomicOperation
	}{{nil, 1, AtomicLoad}, {m.data, 0, AtomicLoad}, {m.data, 3, AtomicLoad}, {unsafe.Add(m.data, 1), 4, AtomicLoad}, {m.data, 4, AtomicCompareExchange + 1}} {
		if _, err := SharedAtomic(args.address, args.width, args.op, 0, 0); !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid access", args, err)
		}
	}
}

func TestSharedAtomicConcurrentRMWReturnsUniqueOldValues(t *testing.T) {
	if !Available {
		t.Skip("native platform unavailable")
	}
	m, err := NewSharedMemory(8, 8)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	const workers, iterations = 8, 1000
	results := make(chan []uint64, workers)
	for range workers {
		go func() {
			old := make([]uint64, iterations)
			for i := range old {
				old[i], _ = SharedAtomic(m.data, 8, AtomicAdd, 1, 0)
			}
			results <- old
		}()
	}
	seen := make([]bool, workers*iterations)
	for range workers {
		for _, old := range <-results {
			if old >= uint64(len(seen)) || seen[old] {
				t.Fatal("RMW was not indivisible", old)
			}
			seen[old] = true
		}
	}
	if got, err := SharedAtomic(m.data, 8, AtomicLoad, 0, 0); err != nil || got != workers*iterations {
		t.Fatal(got, err)
	}
}

func TestSharedAtomicOverlappingSubwordAndWideAccesses(t *testing.T) {
	if !Available {
		t.Skip("native platform unavailable")
	}
	m, err := NewSharedMemory(8, 8)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	done := make(chan struct{}, 8)
	for lane := 0; lane < 8; lane++ {
		go func(lane int) {
			for i := 0; i < 1001; i++ {
				_, _ = SharedAtomic(unsafe.Add(m.data, lane), 1, AtomicAdd, 1, 0)
			}
			done <- struct{}{}
		}(lane)
	}
	for range 8 {
		<-done
	}
	const want = uint64(1001 % 256 * 0x0101010101010101)
	if got, err := SharedAtomic(m.data, 8, AtomicLoad, 0, 0); err != nil || got != want {
		t.Fatalf("subword isolation: got %x want %x err %v", got, want, err)
	}
}
