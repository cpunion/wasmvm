package execute

import (
	"errors"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (e *Execution) failTrap(fn *state.Function, source uint32, trap wasm.TrapCode) Status {
	return e.fail(&wasm.Trap{Code: trap, Function: fn.Index, Offset: fn.Code.Sources[source].Wasm.Offset})
}

// runRuntime services one already-charged boundary. It publishes the complete
// next PC or frame transition, so either driver can select its next instruction.
// Generated native execution validates the boundary kind before reaching here.
func (e *Execution) runRuntime() Status {
	f := &e.frames[len(e.frames)-1]
	fn := f.function
	code := fn.Code
	pc := f.pc
	in := &code.Code[pc]
	instance := fn.Instance
	values := e.values[f.base : f.base+len(code.Slots)]
	var a, b, c, result value.Value
	trap := wasm.NoTrap
	if in.A != slots.NoSlot {
		a = values[int(in.A)]
	}
	if in.B != slots.NoSlot {
		b = values[int(in.B)]
	}
	switch op := in.Op; op {
	case slots.ControlBase + uint32(ir.IfCast):
		index := uint32(0)
		im := &code.Sources[in.Source].Wasm
		target := module.ResolveType(im.Types[1], instance.Types)
		matched := false
		if im.Op == 0xfb0025 || im.Op == 0xfb0026 {
			var err error
			matched, err = e.store.DescriptorEqual(a.Lo, b.Lo, target)
			if err != nil {
				return e.fail(&wasm.Trap{Code: wasm.TrapNullReference, Offset: im.Offset, Function: fn.Index})
			}
		} else {
			matched = e.store.TestReference(a.Lo, target)
		}
		if !matched {
			index = 1
		}
		edge := code.Edges[in.Extra+index]
		if edge.PC <= pc {
			if err := e.poll(); err != nil {
				return e.fail(err)
			}
		}
		for _, move := range code.Moves[edge.Start : edge.Start+edge.Count] {
			values[int(move.Dst)] = values[int(move.Src)]
		}
		f.pc = edge.PC
		return Running
	case 0x23:
		result = instance.Globals[in.Extra].Value
	case 0x24:
		instance.Globals[in.Extra].Value = a
	case 0x25:
		table := instance.Tables[in.Extra]
		index := table.Type.Limits.Address(a.Lo)
		if index >= uint64(len(table.Values)) {
			trap = wasm.TrapTableBounds
		} else {
			result.Lo = table.Values[index]
		}
	case 0x26:
		table := instance.Tables[in.Extra]
		index := table.Type.Limits.Address(a.Lo)
		if index >= uint64(len(table.Values)) {
			trap = wasm.TrapTableBounds
		} else {
			table.Values[index] = b.Lo
		}
	case 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35:
		im := &code.Sources[in.Source].Wasm
		result, trap = instance.Memories[im.Memory].Load(op, a.Lo, im.MemoryOffset(), 0, value.Value{})
	case 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d,
		0x3e:
		im := &code.Sources[in.Source].Wasm
		trap = instance.Memories[im.Memory].Store(op, a.Lo, im.MemoryOffset(), 0, b)
	case 0x3f:
		result.Lo = uint64(instance.Memories[in.Extra].Pages())
	case 0x40:
		mem := instance.Memories[in.Extra]
		old, ok := e.store.GrowMemory(mem, mem.Limits().Address(a.Lo))
		result.Lo = uint64(old)
		if !ok {
			result.Lo = mem.Limits().Failure()
		}
	case 0xd3:
		if a.Lo == b.Lo {
			result.Lo = 1
		}
	case 0xd2:
		result.Lo = instance.Functions[in.Extra]
	case 0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5, 0xe6:
		im := &code.Sources[in.Source].Wasm
		f.pc = pc
		if err := e.stackInstruction(*in, im); err != nil {
			var t *wasm.Trap
			if errors.As(err, &t) {
				t.Function, t.Offset = fn.Index, im.Offset
			}
			return e.fail(err)
		}
		if e.status != Running {
			return e.status
		}
		return Running
	default:
		if in.C != slots.NoSlot {
			c = values[int(in.C)]
		}
		switch {

		case op>>16 == 0xfe:
			if op&0xffff == 3 {
				e.store.Domain.Fence()
				break
			}
			im := &code.Sources[in.Source].Wasm
			var err error
			result, trap, err = instance.Memories[im.Memory].Atomic(e.context, e.store.Domain, op&0xffff, a.Lo, im.MemoryOffset(), b.Lo, c.Lo)
			if err != nil {
				return e.fail(err)
			}
		case op >= 0xfc0008 && op <= 0xfc0011:
			im := &code.Sources[in.Source].Wasm
			var err error
			result, trap, err = e.bulk(instance, *im, a, b, c)
			if err != nil {
				return e.fail(err)
			}
		default:
			return e.fail(wasm.ErrInvalid)
		}
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
