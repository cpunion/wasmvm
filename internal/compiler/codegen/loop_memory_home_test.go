package codegen

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestPrivateWordHomePreservesInterveningTemporaryDefinitions(t *testing.T) {
	variants := []struct {
		name  string
		extra module.Expr
	}{
		{"literal-drop", module.Expr{{Op: 0x41, Bits: ^uint64(0)}, {Op: 0x1a}}},
		{"shift-mask", module.Expr{{Op: 0x41, Bits: 7}, {Op: 0x41, Bits: 3}, {Op: 0x74}, {Op: 0x41, Bits: 15}, {Op: 0x71}, {Op: 0x1a}}},
		{"add-local", module.Expr{{Op: 0x20}, {Op: 0x41, Bits: 19}, {Op: 0x6a}, {Op: 0x22, A: 3}, {Op: 0x1a}}},
		{"multiply-local", module.Expr{{Op: 0x20}, {Op: 0x41, Bits: 17}, {Op: 0x6c}, {Op: 0x22, A: 3}, {Op: 0x1a}}},
		{"assigned-local-read", module.Expr{{Op: 0x41, Bits: 5}, {Op: 0x21, A: 3}, {Op: 0x20, A: 3}, {Op: 0x1a}}},
		{"zero-shift-local", module.Expr{{Op: 0x20}, {Op: 0x41}, {Op: 0x74}, {Op: 0x22, A: 3}, {Op: 0x1a}}},
	}
	for _, tc := range variants {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile("../../../benchmarks/modern/threads-private-memory.wasm")
			if err != nil {
				t.Fatal(err)
			}
			m, err := decode.Decode(data, decode.DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			body := m.Functions[0].Body
			combined := append(module.Expr{}, body[:10]...)
			combined = append(combined, tc.extra...)
			combined = append(combined, body[10:]...)
			m.Functions[0].Body = combined
			f := functionModule(t, m).Functions[0]
			for _, arch := range []string{"arm64", "amd64"} {
				if _, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1}); err != nil {
					t.Fatal(arch, err)
				}
			}
			if !native.Available {
				return
			}
			compiled, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
			_, reference, ref := functionContext(t, f, false)
			memory, refMemory := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
			initial := bytes.Repeat([]byte{0xa5}, 65536)
			for _, seed := range []uint64{0, 17, 0xffffffff, ^uint64(0)} {
				for _, count := range []uint64{1, 3, 31} {
					if err := memory.Write(0, initial); err != nil {
						t.Fatal(err)
					}
					if err := refMemory.Write(0, initial); err != nil {
						t.Fatal(err)
					}
					for s := range f.Slots {
						// Scalar numeric inputs have a zero high word; retain raw upper Lo bits.
						bits := value.Value{Lo: 0xdeadbeef00000100 + uint64(s), Hi: 0}
						switch s {
						case 0:
							bits.Lo = seed
						case 1:
							bits.Lo = count
						case 2:
							bits.Lo = 0
						}
						writeSlot(t, ctx, 16, slots.Slot(s), bits)
						writeSlot(t, ref, 16, slots.Slot(s), bits)
					}
					state := native.State{Used: 16, Depth: 1, Continuation: 0, Budget: 100000}
					if err := ctx.SetState(state); err != nil {
						t.Fatal(err)
					}
					if err := ref.SetState(state); err != nil {
						t.Fatal(err)
					}
					got, err := ctx.Enter(lease, 0)
					want, werr := ref.Enter(reference, 0)
					if err != nil || werr != nil || got != want {
						t.Fatal(runtime.GOARCH, compiled.NativeLoops, seed, count, got, want, err, werr)
					}
					for s := range f.Slots {
						if a, b := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s)); a != b {
							t.Fatal(fmt.Sprintf("seed=%x/count=%d/frame=%d", seed, count, s), a, b)
						}
					}
					actual, expected := make([]byte, 65536), make([]byte, 65536)
					if err := memory.Read(0, actual); err != nil {
						t.Fatal(err)
					}
					if err := refMemory.Read(0, expected); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(actual, expected) {
						t.Fatal(seed, count, "memory effects")
					}
				}
			}
		})
	}
}
