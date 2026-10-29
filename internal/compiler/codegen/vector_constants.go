package codegen

import (
	"encoding/binary"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

// Legalization masks share the immutable guest-literal register budget and
// reload on every independent entry. No mask occupies a live guest register.
func (e *functionEmitter) vectorLegalizationConstants(in slots.Instruction) [][2]uint64 {
	if in.Op == 0xfd0062 && e.x86Features&X86AVX512BITALG != 0 {
		return nil // VPOPCNTB needs neither nibble mask nor lookup table.
	}
	if e.x86Features&(X86SSE41|X86AVX2) != 0 {
		if in.Op == 0xfd00e8 || in.Op == 0xfd00e9 {
			return [][2]uint64{{0x7fc000007fc00000, 0x7fc000007fc00000}}
		}
		if in.Op == 0xfd00f4 || in.Op == 0xfd00f5 {
			return [][2]uint64{{0x7ff8000000000000, 0x7ff8000000000000}}
		}
	}
	if e.x86Features&(X86SSSE3|X86AVX2) == 0 {
		return nil
	}
	switch in.Op {
	case 0xfd000d:
		var values [][2]uint64
		lanes := e.f.Sources[in.Source].Wasm.Vector
		for source := byte(0); source < 2; source++ {
			var mask [16]byte
			used := false
			for i, n := range lanes {
				mask[i] = 0x80
				if n/16 == source {
					mask[i], used = n%16, true
				}
			}
			if used {
				values = append(values, [2]uint64{binary.LittleEndian.Uint64(mask[:8]), binary.LittleEndian.Uint64(mask[8:])})
			}
		}
		return values
	case 0xfd000e:
		return [][2]uint64{{0x7070707070707070, 0x7070707070707070}}
	case 0xfd0062:
		return [][2]uint64{{0x0f0f0f0f0f0f0f0f, 0x0f0f0f0f0f0f0f0f}, {0x0302020102010100, 0x0403030203020201}}
	case 0xfd0082:
		return [][2]uint64{{0x8000800080008000, 0x8000800080008000}}
	}
	return nil
}
