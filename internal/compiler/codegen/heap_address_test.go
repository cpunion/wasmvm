package codegen

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Real checked field accesses exercise packed/aligned displacement limits,
// spills, both cache inventories and every original continuation. Sentinel
// bytes detect an incorrect size or address independently of the loaded result.
func TestNativeHeapFieldDisplacements(t *testing.T) {
	fields := []module.FieldType{{Value: wasm.I32, Packed: 8}, {Value: wasm.I32, Packed: 16},
		{Value: wasm.I32}, {Value: wasm.I64}, {Value: wasm.V128}}
	for _, field := range fields {
		width := field.HeapBytes()
		for _, offset := range []int{0, 1, 127, 128, 255, 256, 4095 * width, 4096 * width, 4096*width + 1} {
			for _, pressure := range []bool{false, true} {
				f, config := heapFixture(field, false, false)
				typ := &config.HeapTypes[0]
				typ.Fields = nil
				for remaining := offset; remaining > 0; {
					prefix := module.FieldType{Value: wasm.V128}
					if remaining < 16 {
						prefix = module.FieldType{Value: wasm.I32, Packed: 8}
					}
					typ.Fields = append(typ.Fields, prefix)
					remaining -= prefix.HeapBytes()
				}
				index := uint32(len(typ.Fields))
				typ.Fields = append(typ.Fields, field)
				for pc := 0; pc < 3; pc++ {
					f.Sources[pc].Wasm.B = index
				}
				if pressure {
					addHeapFixturePressure(f)
				}
				addHeapFixtureLoop(f, 0)
				config.Unmetered = true
				for _, arch := range []string{"amd64", "arm64"} {
					c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
					if err != nil || c.HeapCaches != 1 {
						t.Fatal(field, offset, pressure, arch, c, err)
					}
				}
				if !native.Available {
					continue
				}
				_, lease, ctx := configuredContext(t, f, config)
				size := uint64(256 + offset + width)
				storage, err := native.NewMemory(size, size)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(storage.Close)
				buffer, _ := storage.Buffer()
				address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
				words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
				ref := uint64(0x8000000100000001)
				words[0], words[1] = 2, address+16
				words[6], words[7], words[8] = ref, address+128, 0
				payload := buffer[128:]
				for _, pc := range []int{0, 1, 2, 3} {
					for i := range payload {
						payload[i] = 0xa5
					}
					initial := value.Value{Lo: 0x89abcdef87658081, Hi: 0x123456789abcdef0}
					argument := value.Value{Lo: 0xfedcba987654def0, Hi: 0xf0e1d2c3b4a59687}
					original := make([]byte, 16)
					binary.LittleEndian.PutUint64(original, initial.Lo)
					binary.LittleEndian.PutUint64(original[8:], initial.Hi)
					copy(payload[offset:], original[:width])
					wantMemory := bytes.Clone(payload)
					want := initial
					if pc <= 1 {
						binary.LittleEndian.PutUint64(original, argument.Lo)
						binary.LittleEndian.PutUint64(original[8:], argument.Hi)
						copy(wantMemory[offset:], original[:width])
						want = argument
					}
					if width < 8 {
						want.Lo &= uint64(1)<<(width*8) - 1
					}
					if width < 16 {
						want.Hi = 0
					}
					if pc > 2 {
						want = value.Value{}
					}
					for s := range f.Slots {
						writeSlot(t, ctx, 16, slots.Slot(s), value.Value{})
					}
					writeSlot(t, ctx, 16, 0, value.Value{Lo: ref})
					writeSlot(t, ctx, 16, 2, argument)
					state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Fuel: 42, Budget: 37}
					state.Payload[2] = address
					ctx.SetState(state)
					exit, err := ctx.Enter(lease, 0)
					got, _ := ctx.State()
					if err != nil || exit != native.Return || got.Fuel != 42 || got.Budget != 37 || !bytes.Equal(payload, wantMemory) || readSlot(t, ctx, 16, 4) != want {
						t.Fatal(field, offset, pressure, pc, exit, err, got, readSlot(t, ctx, 16, 4), want)
					}
				}
				storage.Close()
			}
		}
	}
}
