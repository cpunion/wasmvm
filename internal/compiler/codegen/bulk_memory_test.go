package codegen

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func bulkFunction(t *testing.T, copying, wide bool) (*slots.Function, int) {
	t.Helper()
	address, second, op := wasm.I32, wasm.I32, uint32(0xfc000b)
	if wide {
		address = wasm.I64
	}
	if copying {
		second, op = address, 0xfc000a
	}
	types := []wasm.ValueType{address, second, address, wasm.I64, wasm.F64, wasm.V128}
	body := module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x20, A: 2}, {Op: op, Offset: 123}}
	for i := range types {
		body = append(body, module.Instr{Op: 0x20, A: uint32(i)})
	}
	body = append(body, module.Instr{Op: 0xb})
	m := functionFixture(module.FuncType{Params: types, Results: types}, 0, body)
	m.Memories = []module.Limits{{Min: 1, Max: 1, HasMax: true, Index64: wide}}
	checked, err := validate.CheckWithFeatures(m, validate.DefaultLimits(), wasm.SupportedFeatures)
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Compile(checked, lower.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := slots.Compile(lowered, slots.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	f := compiled.Functions[0]
	for pc, in := range f.Code {
		if in.Op == op {
			return f, pc
		}
	}
	t.Fatal("missing bulk instruction")
	return nil, 0
}

func TestBulkMemoryTargetsAndFallbacks(t *testing.T) {
	for _, copying := range []bool{false, true} {
		for _, wide := range []bool{false, true} {
			f, _ := bulkFunction(t, copying, wide)
			for _, arch := range []string{"amd64", "arm64"} {
				config := Config{Unmetered: true, MemoryCount: 1}
				compiled, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
				if err != nil || compiled.NativeBulk != 1 || compiled.Helpers != 0 || !compiled.UsesMemory {
					t.Fatal(arch, copying, wide, compiled, err)
				}
				limit := DefaultLimits()
				limit.MaxCodeBytes = uint64(len(compiled.Entry.Code))
				if _, err := CompileFunctionWithConfig(arch, f, limit, config); err != nil {
					t.Fatal(err)
				}
				limit.MaxCodeBytes--
				if _, err := CompileFunctionWithConfig(arch, f, limit, config); !errors.Is(err, wasm.ErrLimit) {
					t.Fatal("code limit", err)
				}
				for _, fallback := range []Config{{MemoryCount: 1}, {Unmetered: true}, {Unmetered: true, MemoryCount: 2}, {Unmetered: true, MemoryCount: 1, SharedMemories: []bool{true}}} {
					compiled, err = CompileFunctionWithConfig(arch, f, DefaultLimits(), fallback)
					if err != nil || compiled.NativeBulk != 0 || compiled.Helpers != 1 {
						t.Fatal("missing helper fallback", arch, fallback, err)
					}
				}
			}
		}
	}
}

func TestNativeBulkMemoryMatchesCompleteByteOracle(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	initial := make([]byte, 65536)
	for i := range initial {
		initial[i] = byte(i*193+127) ^ byte(i>>8)
	}
	for _, copying := range []bool{false, true} {
		for _, wide := range []bool{false, true} {
			f, pc := bulkFunction(t, copying, wide)
			compiled, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryCount: 1})
			mem := bindMemory(t, ctx, 65536, 65536)
			cases := [][3]uint64{{65536, 65536, 0}, {65536, 0, 1}, {65535, 0, 2}, {0, 65536, 1}, {65537, 0, 0}, {math.MaxUint64, 0, 1}, {1, 0, math.MaxUint64}, {0, math.MaxUint64, 2}, {0, 65537, 0}, {1 << 32, 0, 1}}
			for _, n := range []uint64{0, 1, 2, 3, 7, 15, 16, 17, 31, 32, 33, 63, 64, 65, 127, 128, 129, 255, 256, 257, 4095, 4096, 4097, 8191, 8192, 8193, 32767} {
				for _, pair := range [][2]uint64{{0, 1}, {1, 0}, {5, 31}, {31, 5}, {16, 32}, {32, 16}, {257, 65}, {65, 257}, {0, 0}, {10001, 40000}} {
					cases = append(cases, [3]uint64{pair[0], pair[1], n})
				}
			}
			for _, operands := range cases {
				if err := mem.Write(0, initial); err != nil {
					t.Fatal(err)
				}
				dst, src, count := operands[0], operands[1], operands[2]
				if !wide {
					dst, src, count = uint64(uint32(dst)), uint64(uint32(src)), uint64(uint32(count))
				}
				valid := dst <= 65536 && count <= 65536-dst
				if copying {
					valid = valid && src <= 65536 && count <= 65536-src
				}
				want := append([]byte(nil), initial...)
				if valid {
					if copying {
						copy(want[dst:dst+count], want[src:src+count])
					} else {
						for i := uint64(0); i < count; i++ {
							want[dst+i] = byte(src)
						}
					}
				}
				args := []value.Value{{Lo: operands[0]}, {Lo: operands[1]}, {Lo: operands[2]}, {Lo: 0xff123456deadbeef}, {Lo: 0x7ff012345678abcd}, {Lo: 0x1122334455667788, Hi: 0xffeeddccbbaa0099}}
				for i := range args {
					if f.Signature.Params[i] == wasm.I32 {
						args[i].Lo = uint64(uint32(args[i].Lo))
					}
				}
				out, exit, state := runFunction(t, compiled, lease, ctx, args, 100, false, 64)
				if valid {
					if exit != native.Return || len(out) != len(args) {
						t.Fatal(copying, wide, operands, exit, out, state)
					}
					for i, got := range out {
						want := args[i]
						if f.Signature.Params[i] == wasm.I32 {
							want.Lo = uint64(uint32(want.Lo))
						}
						if got != want {
							t.Fatal("bulk clobbered live value", copying, wide, operands, i, got, want)
						}
					}
				} else if exit != native.Trap || state.Payload[0] != uint64(wasm.TrapMemoryBounds) || state.Source != f.Sources[f.Code[pc].Source].Wasm.Offset {
					t.Fatal("wrong bounds exit", copying, wide, operands, exit, state)
				}
				got := make([]byte, len(initial))
				if err := mem.Read(0, got); err != nil || !bytes.Equal(got, want) {
					t.Fatal("memory differs", copying, wide, operands, err)
				}
			}
		}
	}
}

func TestNativeBulkPollRestoresStateBeforeAnyChunk(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, copying := range []bool{false, true} {
		for _, backwards := range []bool{false, true} {
			f, pc := bulkFunction(t, copying, true)
			compiled, _, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryCount: 1})
			// Set control after the adapter check, then enter directly at bulk's
			// original PC so the chunk poll must preserve all cached slot values.
			var prefix []byte
			if runtime.GOARCH == "arm64" {
				for _, word := range []uint32{0xd2800029, 0x9100000a | uint32(nativeabi.ControlOffset)<<10, 0xc89ffd49} {
					prefix = binary.LittleEndian.AppendUint32(prefix, word)
				}
			} else {
				prefix = []byte{0x48, 0xc7, 0x87}
				prefix = binary.LittleEndian.AppendUint32(prefix, uint32(nativeabi.ControlOffset))
				prefix = binary.LittleEndian.AppendUint32(prefix, 1)
			}
			code, err := native.Publish(append(prefix, compiled.Entry.Code...), []uint32{0}, DefaultLimits().MaxCodeBytes)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(code.Close)
			lease, err := code.Acquire()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(lease.Close)
			mem := bindMemory(t, ctx, 65536, 65536)
			before := bytes.Repeat([]byte{0x57}, 65536)
			if err := mem.Write(0, before); err != nil {
				t.Fatal(err)
			}
			values := []value.Value{{Lo: 1}, {Lo: 0}, {Lo: 8193}, {Lo: 0xfedcba9876543210}, {Lo: 0x7ff0123456789abc}, {Lo: 0x1234, Hi: 0xfedcba9876543210}}
			if !backwards {
				values[0].Lo, values[1].Lo = 0, 1
			}
			for i, v := range values {
				writeSlot(t, ctx, 16, slots.Slot(i), v)
			}
			in := f.Code[pc]
			for i, slot := range []slots.Slot{in.A, in.B, in.C} {
				writeSlot(t, ctx, 16, slot, values[i])
			}
			if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}); err != nil {
				t.Fatal(err)
			}
			exit, err := ctx.Enter(lease, 0)
			state, stateErr := ctx.State()
			if err != nil || stateErr != nil || exit != native.Poll || state.Continuation != uint64(pc) {
				t.Fatal("bulk poll lost original continuation", copying, backwards, exit, state, err, stateErr)
			}
			for i, want := range values {
				if got := readSlot(t, ctx, 16, slots.Slot(i)); got != want {
					t.Fatal("bulk poll clobbered live slot", copying, backwards, i, got, want)
				}
			}
			got := make([]byte, len(before))
			if err := mem.Read(0, got); err != nil || !bytes.Equal(before, got) {
				t.Fatal("canceled chunk wrote memory", err)
			}
		}
	}
}
