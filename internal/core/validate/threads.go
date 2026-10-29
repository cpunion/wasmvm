package validate

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (v *checker) atomic(in module.Instr) {
	if v.features&wasm.Threads == 0 {
		v.fail("Threads are disabled")
	}
	sub := in.Op & 0xffff
	if sub == 3 {
		return
	}
	if sub > 2 && (sub < 0x10 || sub > 0x4e) {
		v.fail("unknown atomic instruction")
	}
	align := uint32(2)
	if sub == 2 {
		align = 3
	}
	shape := uint32(0)
	if sub >= 0x10 {
		shape = (sub - 0x10) % 7
		align = []uint32{2, 3, 0, 1, 0, 1, 2}[shape]
	}
	if in.A != align {
		v.fail("atomic alignment must be natural")
	}
	v.alignment(in, align)
	address := v.memories[in.Memory].AddressType()
	if sub == 0 {
		v.pop(wasm.I32)
		v.pop(address)
		v.push(wasm.I32)
		return
	}
	if sub <= 2 {
		v.pop(wasm.I64)
		expected := wasm.I32
		if sub == 2 {
			expected = wasm.I64
		}
		v.pop(expected)
		v.pop(address)
		v.push(wasm.I32)
		return
	}
	typ := wasm.I64
	if shape == 0 || shape == 2 || shape == 3 {
		typ = wasm.I32
	}
	group := (sub - 0x10) / 7
	if group >= 1 {
		v.pop(typ)
	}
	if group == 8 {
		v.pop(typ)
	}
	v.pop(address)
	if group != 1 {
		v.push(typ)
	}
}
