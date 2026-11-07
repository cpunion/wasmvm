package execute

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/simd"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (e *Execution) runSIMD() Status {
	f := &e.frames[len(e.frames)-1]
	fn := f.function
	code := fn.Code
	pc := f.pc
	in := &code.Code[pc]
	base := f.base
	values := e.values[base : base+len(code.Slots)]
	instance := fn.Instance
	op := in.Op
	var result value.Value
	trap := wasm.NoTrap
	var a, b, c value.Value
	if in.A != slots.NoSlot {
		a = values[int(in.A)]
	}
	if in.B != slots.NoSlot {
		b = values[int(in.B)]
	}
	if in.C != slots.NoSlot {
		c = values[int(in.C)]
	}
	im := &code.Sources[in.Source].Wasm
	sub := op & 0xffff
	switch {
	case sub <= 10 || sub >= 84 && sub <= 87 || sub == 92 || sub == 93:
		result, trap = instance.Memories[im.Memory].Load(op, a.Lo, im.MemoryOffset(), im.Lane, b)
	case sub == 11 || sub >= 88 && sub <= 91:
		trap = instance.Memories[im.Memory].Store(op, a.Lo, im.MemoryOffset(), im.Lane, b)
	default:
		result = simd.Eval(im, a, b, c)
	}
	if trap != wasm.NoTrap {
		return e.failTrap(fn, in.Source, trap)
	}
	if in.Dst != slots.NoSlot {
		values[int(in.Dst)] = result
	}
	f.pc = pc + 1
	return Running
}
