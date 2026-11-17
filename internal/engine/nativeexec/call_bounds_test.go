package nativeexec

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPrivateBoundsTrackCapacityGrowthAndPreserveDepthTopInvariant(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	n, err := New(context.Background(), nil, 256, 128, false)
	if err != nil {
		t.Fatal(err)
	}
	defer n.Close()
	if _, err := n.Resize(0, 64); err != nil {
		t.Fatal(err)
	}
	check := func(slots, frames, wantSlots, wantFrames uint64) {
		t.Helper()
		if _, err := n.prepareCalls(slots, frames); err != nil || n.callWords[2] != wantSlots*16 || n.callWords[3] != wantFrames {
			t.Fatal(err, n.callWords, wantSlots, wantFrames)
		}
	}
	n.entryState.Depth = 7
	check(200, 100, 64, 23)
	for top := uint64(0); top <= 16; top++ {
		n.callWords[0], n.entryState.Depth = top, 7+top
		check(200, 100, 64, 23)
	}
	// Growing a suspended full chain preserves its unchanged prefix and
	// recomputes both limits, rather than retaining an earlier clamp.
	n.callGrow = true
	if _, err := n.Resize(0, 128); err != nil {
		t.Fatal(err)
	}
	check(200, 100, 128, 39)
	check(8, 24, 8, 24)
	check(^uint64(0), 100, 128, 39)
	n.entryState.Depth = ^uint64(0) - 1
	check(200, ^uint64(0), 128, ^uint64(0))
	// Metered bodies retain independent logical and physical checks.
	n.metered = true
	check(200, 100, 200, 100)
	if _, err := n.prepareCalls(^uint64(0), 100); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("overflowing logical slot bytes accepted", err)
	}
	capacity := n.callWords[1]
	n.callWords[1]++
	if _, err := n.prepareCalls(200, 100); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("unauthenticated physical capacity accepted", err)
	}
	n.callWords[1], n.callWords[0] = capacity, capacity+1
	if _, err := n.prepareCalls(200, 100); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("invalid parent count accepted", err)
	}
	words := n.callWords
	n.callWords = words[:1]
	if _, err := n.prepareCalls(200, 100); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("short call header accepted", err)
	}
	n.callWords = words
}
