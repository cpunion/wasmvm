package codegen

import (
	"errors"
	"slices"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNativeFrameRegistersRequireUnmeteredLinkedBodies(t *testing.T) {
	f := leafFixture(0x6a, nil)
	for _, arch := range []string{"amd64", "arm64"} {
		for _, config := range []Config{
			{NativeFrameRegisters: true},
			{NativeFrameRegisters: true, Unmetered: true},
			{NativeFrameRegisters: true, LinkedCallees: map[uint32]*slots.Function{7: f}},
		} {
			if _, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(arch, config, err)
			}
		}
	}
}

func TestNativeFrameStateIsExcludedFromGuestAllocation(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, memory := range []bool{false, true} {
			e := functionEmitter{a: &assembler{arch: arch}, usesMemory: memory}
			ordinary := slices.Clone(e.integerRegisterPool())
			e.frameRegisters = true
			reserved := []byte{e.callReg(callUsed), e.callReg(callDepth)}
			want := slices.DeleteFunc(ordinary, func(r byte) bool { return slices.Contains(reserved, r) })
			if got := e.integerRegisterPool(); !slices.Equal(got, want) {
				t.Fatal(arch, memory, got, want)
			}
			for _, r := range e.spareCallerRegisters() {
				if slices.Contains(reserved, r) {
					t.Fatal("frame state overlaps leaf scratch", arch, r)
				}
			}
		}
	}
}
