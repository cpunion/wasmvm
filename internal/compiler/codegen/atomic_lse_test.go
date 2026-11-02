package codegen

import (
	"encoding/binary"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestNativeLSEWidthsAndAliasedResults(t *testing.T) {
	for sub := uint32(0x1e); sub <= 0x4e; sub++ {
		f, inputs := atomicFunction(sub, false)
		cfg := Config{Unmetered: true, NativeThreads: true, MemoryCount: 1, ARM64Features: ARM64LSE}
		compiled, err := CompileFunctionWithConfig("arm64", f, DefaultLimits(), cfg)
		if err != nil || compiled.NativeAtomics != 1 || compiled.Helpers != 0 {
			t.Fatal(sub, compiled, err)
		}
		if !native.Available || runtime.GOARCH != "arm64" || native.ARM64CPUFeatures()&uint32(ARM64LSE) == 0 {
			continue
		}
		_, lease, ctx := configuredContext(t, f, cfg)
		mem := bindMemory(t, ctx, 8, 8)
		ref, err := native.NewSharedMemory(8, 8)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(ref.Close)
		address, _, _, err := ref.SharedView()
		if err != nil {
			t.Fatal(err)
		}
		width := [...]uint32{4, 8, 1, 2, 1, 2, 4}[(sub-0x10)%7]
		const initial, replacement = uint64(0x89abcdef11223344), uint64(0xabcdef9876543210)
		for _, argument := range []uint64{initial, 0xffffffffff223344, ^uint64(0)} {
			var raw [8]byte
			binary.LittleEndian.PutUint64(raw[:], initial)
			mem.Write(0, raw[:])
			ref.Write(0, raw[:])
			want, err := native.SharedAtomic(address, width, native.AtomicOperation((sub-0x10)/7), argument, replacement)
			if err != nil {
				t.Fatal(err)
			}
			for i, input := range []uint64{0, argument, replacement} {
				writeSlot(t, ctx, 16, inputs[i], value.Value{Lo: input})
			}
			if err := ctx.SetState(native.State{Used: 16, Depth: 1, Budget: 64}); err != nil {
				t.Fatal(err)
			}
			if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return || readSlot(t, ctx, 16, f.Code[0].Dst) != (value.Value{Lo: want}) {
				t.Fatal("LSE old bits", sub, exit, err, readSlot(t, ctx, 16, f.Code[0].Dst), want)
			}
			var actual, expected [8]byte
			mem.Read(0, actual[:])
			ref.Read(0, expected[:])
			if actual != expected {
				t.Fatal("LSE write", sub, actual, expected)
			}
		}
	}
}
