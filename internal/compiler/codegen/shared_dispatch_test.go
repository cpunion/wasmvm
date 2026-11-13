package codegen

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSharedDispatchTableAndAddressRelocations(t *testing.T) {
	callee := leafFixture(0x6a, nil)
	config := Config{FunctionCount: 8, LinkedSignatures: []uint32{31, 47},
		LinkedCallees: map[uint32]*slots.Function{6: callee, 7: callee}}
	for _, arch := range []string{"amd64", "arm64"} {
		table, calls, err := LinkedDispatchTable(arch, config)
		if err != nil || len(table) != 64 || len(calls) != 2 {
			t.Fatal(arch, len(table), calls, err)
		}
		for i, signature := range config.LinkedSignatures {
			row := table[i*32:]
			if binary.LittleEndian.Uint32(row) != uint32(len(callee.Slots)) || binary.LittleEndian.Uint32(row[4:]) != 6+uint32(i) || binary.LittleEndian.Uint32(row[8:]) != signature || calls[i].Offset != uint32(i*32+16) {
				t.Fatal("invalid shared row", arch, i, row, calls)
			}
		}
		if err := LinkCalls(arch, table, calls, map[uint32]uint32{6: 0, 7: 32}); err != nil {
			t.Fatal(arch, err)
		}
		a := &assembler{arch: arch}
		e := &functionEmitter{a: a}
		e.sharedDispatchAddress()
		width := len(a.code)
		// Exercise arm64's former +/-1 MiB local ADR range and an unaligned
		// publication base: the result depends only on the image-relative delta.
		const target = (2 << 20) + 64
		image := make([]byte, target+len(table))
		copy(image, a.code)
		copy(image[target:], table)
		if err := LinkDispatch(arch, image, []uint32{0}, target); err != nil {
			t.Fatal(arch, err)
		}
		var delta uint32
		if arch == "amd64" {
			delta = binary.LittleEndian.Uint32(image[3:]) + 7
		} else {
			delta = (binary.LittleEndian.Uint32(image[4:])>>5)&0xffff | ((binary.LittleEndian.Uint32(image[8:])>>5)&0xffff)<<16
		}
		if delta != target {
			t.Fatal(arch, delta)
		}
		for _, bad := range []struct {
			site, target uint32
		}{{0, 0}, {0, uint32(len(image))}, {uint32(len(image) - 4), uint32(len(image) - 1)}, {1, 64}} {
			copy(image, a.code)
			if err := LinkDispatch(arch, image, []uint32{bad.site}, bad.target); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(arch, bad, err)
			}
		}
		for b := range width {
			copy(image, a.code)
			image[b] ^= 1
			before := append([]byte(nil), image...)
			// AMD64's displacement is rewritten, rather than a template opcode.
			if arch == "amd64" && b >= 3 {
				continue
			}
			if err := LinkDispatch(arch, image, []uint32{0}, target); !errors.Is(err, wasm.ErrInvalid) || !bytes.Equal(image, before) {
				t.Fatal("malformed template changed image", arch, b, err)
			}
		}
		copy(image, a.code)
		before := append([]byte(nil), image...)
		if err := LinkDispatch(arch, image, []uint32{0, 32}, target); !errors.Is(err, wasm.ErrInvalid) || !bytes.Equal(image, before) {
			t.Fatal("partial address link", arch, err)
		}
		if err := LinkDispatch("unsupported", image, []uint32{0}, target); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal(err)
		}
	}
	for _, config := range []Config{{}, {FunctionCount: 1, LinkedSignatures: []uint32{1, 2}}, {FunctionCount: 1, LinkedSignatures: []uint32{1}}} {
		if _, _, err := LinkedDispatchTable("arm64", config); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(config, err)
		}
	}
	if _, _, err := LinkedDispatchTable("unsupported", config); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
}
