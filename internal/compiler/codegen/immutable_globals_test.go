package codegen

import (
	"bytes"
	"runtime"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func immutableGlobalFixture(typ wasm.ValueType, pressure bool) *slots.Function {
	f := globalFixture(typ)
	f.Code[1].Op, f.Sources[1].Wasm.Op = 1, 1
	if pressure {
		ret := f.Code[len(f.Code)-1]
		f.Code = f.Code[:len(f.Code)-1]
		for i := range 20 {
			dst := slots.Slot(len(f.Slots))
			f.Slots = append(f.Slots, typ)
			source := uint32(len(f.Sources))
			f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x23, A: uint32(i % 2)}})
			f.Code = append(f.Code, slots.Instruction{Op: 0x23, Source: source, Dst: dst, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
		}
		f.Code = append(f.Code, ret)
	}
	f.LocalCount = uint32(len(f.Slots))
	return f
}

func TestImmutableGlobalValueResidencyEligibility(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, typ := range []wasm.ValueType{wasm.I64, wasm.F32, wasm.V128, wasm.AnyRef} {
			for _, mode := range []string{"unmetered", "precise", "mutable", "uncached"} {
				f := immutableGlobalFixture(typ, false)
				cfg := Config{Globals: []module.GlobalType{{Value: typ}, {Value: typ}}}
				if mode == "mutable" {
					cfg.Globals[0].Mutable = true
					cfg.Globals[1].Mutable = true
				}
				e := functionEmitter{a: &assembler{arch: arch}, f: f, unmetered: mode != "precise"}
				e.selectRegisters(cfg, nil)
				if mode == "uncached" {
					e.runtimeRegisters = nil
				}
				e.prepareRuntimeRegisters(cfg)
				want := 0
				if arch == "arm64" && mode == "unmetered" && typ != wasm.V128 {
					want = 2
				}
				if len(e.runtimeGlobalValues) != want {
					t.Fatal(arch, typ, mode, e.runtimeGlobalValues, want)
				}
			}
		}
	}
}

func TestImmutableGlobalsAllColdEntriesRefreshAndRetainBits(t *testing.T) {
	for _, typ := range []wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64, wasm.V128, wasm.AnyRef} {
		for _, unmetered := range []bool{false, true} {
			for _, pressure := range []bool{false, true} {
				f := immutableGlobalFixture(typ, pressure)
				cfg := Config{Unmetered: unmetered, FunctionCount: 3, TableCount: 3000, Globals: []module.GlobalType{{Value: typ}, {Value: typ}}}
				for _, arch := range []string{"amd64", "arm64"} {
					first, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), cfg)
					if err != nil || first.NativeGlobals != uint32(len(f.Code)-2) || first.Helpers != 0 {
						t.Fatal(arch, typ, first, err)
					}
					next, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), cfg)
					if err != nil || !bytes.Equal(first.Entry.Code, next.Entry.Code) {
						t.Fatal("unstable global entry image", arch, typ, err)
					}
				}
				if !native.Available {
					continue
				}
				_, lease, ctx := configuredContext(t, f, cfg)
				offset, _ := nativeabi.GlobalPointerOffset(3, 3000, 0)
				storage, err := native.NewMemory(uint64(offset)+32, uint64(offset)+32)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(storage.Close)
				buffer, _ := storage.Buffer()
				address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
				words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
				words[offset/8], words[offset/8+1] = address+uint64(offset)+16, address+uint64(offset)+16
				for pass := range 2 {
					current := value.Value{Lo: 0xfff8123456789abc + uint64(pass), Hi: 0xfedcba9876543210 + uint64(pass)}
					if typ == wasm.I32 || typ == wasm.F32 {
						current.Lo = 0xffc01234 + uint64(pass)
					}
					if unmetered && typ != wasm.V128 {
						current.Hi = 0
					}
					// A new entry must reload even after the same bound cell changes outside
					// native execution. Guest/API mutation of an immutable global is denied.
					words[offset/8+2], words[offset/8+3] = current.Lo, current.Hi
					for pc := range f.Code {
						want := make([]value.Value, len(f.Slots))
						for i := range want {
							writeSlot(t, ctx, 16, slots.Slot(i), want[i])
						}
						for _, in := range f.Code[pc:] {
							if in.Op == 0x23 {
								want[in.Dst] = current
							}
						}
						state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 64, Fuel: 1000}
						state.Payload[2] = address
						ctx.SetState(state)
						if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
							t.Fatal(runtime.GOARCH, typ, unmetered, pressure, pass, pc, exit, err)
						}
						for i, v := range want {
							if got := readSlot(t, ctx, 16, slots.Slot(i)); got != v {
								t.Fatal("global entry bits", typ, unmetered, pressure, pass, pc, i, got, v)
							}
						}
					}
				}
			}
		}
	}
}
