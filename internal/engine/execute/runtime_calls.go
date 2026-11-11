package execute

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (e *Execution) runCall() Status {
	fi := len(e.frames) - 1
	f := &e.frames[fi]
	fn := f.function
	code := fn.Code
	pc := f.pc
	in := &code.Code[pc]
	base := f.base
	values := e.values[base : base+len(code.Slots)]
	instance := fn.Instance
	op := in.Op
	im := &code.Sources[in.Source].Wasm
	tail := module.IsTailCall(op)
	count := in.Count
	inputs := code.Args[in.Extra : in.Extra+count]
	var target uint64
	if op == 0x10 || op == 0x12 {
		target = instance.Functions[im.A]
	} else if op == 0x14 || op == 0x15 {
		count--
		target = values[int(inputs[count])].Lo
		if target == 0 {
			return e.failTrap(fn, in.Source, wasm.TrapNullReference)
		}
	} else {
		count--
		table := instance.Tables[im.B]
		index := table.Type.Limits.Address(values[int(inputs[count])].Lo)
		if index >= uint64(len(table.Values)) {
			return e.failTrap(fn, in.Source, wasm.TrapTableBounds)
		}
		target = table.Values[index]
		if target == 0 {
			return e.failTrap(fn, in.Source, wasm.TrapUninitializedElement)
		}
	}
	callee, ok := e.store.Function(target)
	if !ok {
		return e.fail(state.ErrReference)
	}
	if (op == 0x11 || op == 0x13 || op == 0x14 || op == 0x15) && !e.store.FunctionSubtype(callee.TypeID, instance.Types[im.A]) {
		return e.failTrap(fn, in.Source, wasm.TrapIndirectCallType)
	}
	var dest []slots.Slot
	if tail {
		dest = f.dest
	} else {
		dest = code.Args[in.Extra+in.Count : in.Extra+in.Count+uint32(len(callee.Signature.Results))]
	}
	if err := e.poll(); err != nil {
		return e.fail(err)
	}
	if access, ok := callee.Host.(*state.ScalarAccess); ok && !tail {
		var bits uint64
		if access.Write {
			bits = values[int(inputs[0])].Lo
		}
		bits, err := access.Access(bits)
		if err != nil {
			return e.fail(err)
		}
		if !access.Write {
			values[int(dest[0])] = value.Value{Lo: bits}
		}
		f.pc = pc + 1
		return Running
	}
	if tail {
		// Preserve the tuple before retiring/reusing the current frame.
		// This supports overlapping/permuted arguments and a different
		// callee signature without charging an additional guest frame.
		if callee.Code != nil && uint64(base)+uint64(len(callee.Code.Slots)) > uint64(e.config.Limits.MaxSlots) {
			return e.fail(&wasm.Error{Kind: wasm.ErrLimit, Detail: "guest frame slots"})
		}
		if int(count) > cap(e.tailArgs) {
			e.tailArgs = make([]value.Value, count)
		} else {
			e.tailArgs = e.tailArgs[:count]
		}
		for i, src := range inputs[:count] {
			e.tailArgs[i] = values[int(src)]
		}
		e.values = e.values[:base]
		e.frames[fi] = frame{}
		e.frames = e.frames[:fi]
		if callee.Code == nil {
			copy(e.prepareHost(callee, instance, len(e.tailArgs)), e.tailArgs)
			e.pendingDest, e.status = dest, HostRequest
			clear(e.tailArgs)
			e.tailArgs = e.tailArgs[:0]
			return HostRequest
		}
		if err := e.push(callee, dest); err != nil {
			return e.fail(err)
		}
		copy(e.values[base:], e.tailArgs)
		clear(e.tailArgs)
		e.tailArgs = e.tailArgs[:0]
		return Running
	}
	// Publish the caller continuation before a host suspension or
	// frame push. Straight-line execution keeps its PC local.
	f.pc = pc + 1
	f.callPC, f.suspended = pc, true
	if callee.Code == nil {
		args := e.prepareHost(callee, instance, int(count))
		for i := range args {
			args[i] = e.values[base+int(inputs[i])]
		}
		e.pendingDest = dest
		e.status = HostRequest
		return HostRequest
	}
	if err := e.push(callee, dest); err != nil {
		return e.fail(err)
	}
	calleeBase := e.frames[len(e.frames)-1].base
	for i := uint32(0); i < count; i++ {
		e.values[calleeBase+int(i)] = e.values[base+int(inputs[i])]
	}
	return Running
}

func (e *Execution) runThrow() Status {
	f := &e.frames[len(e.frames)-1]
	fn := f.function
	code := fn.Code
	pc := f.pc
	in := &code.Code[pc]
	base := f.base
	values := e.values[base : base+len(code.Slots)]
	instance := fn.Instance
	op := in.Op
	f.pc = pc
	var ex *state.Exception
	if op == 8 {
		im := &code.Sources[in.Source].Wasm
		var local [8]value.Value
		operands := local[:0]
		if int(in.Count) > cap(operands) {
			operands = make([]value.Value, in.Count)
		} else {
			operands = operands[:in.Count]
		}
		for i, s := range code.Args[in.Extra : in.Extra+in.Count] {
			operands[i] = values[int(s)]
		}
		var err error
		ex, err = e.store.NewException(instance.Tags[im.A], operands, e.RetainedValues)
		if err != nil {
			return e.fail(err)
		}
	} else {
		ref := values[int(code.Args[in.Extra])].Lo
		if ref == 0 {
			return e.failTrap(fn, in.Source, wasm.TrapNullReference)
		}
		ex = &state.Exception{Owner: e.store, Ref: ref}
	}
	if e.raise(ex) {
		return Running
	}
	return Failed
}

func (e *Execution) returnFrame() Status {
	fi := len(e.frames) - 1
	f := &e.frames[fi]
	fn := f.function
	code := fn.Code
	pc := f.pc
	in := &code.Code[pc]
	base := f.base
	values := e.values[base : base+len(code.Slots)]
	e.trimPrompts(fi)
	outputs := code.Args[in.Extra : in.Extra+in.Count]
	if fi == 0 {
		if len(outputs) > cap(e.result) {
			e.result = make([]value.Value, len(outputs))
		} else {
			e.result = e.result[:len(outputs)]
		}
		for i, src := range outputs {
			e.result[i] = values[int(src)]
		}
		clear(e.frames)
		e.frames = e.frames[:0]
		e.values = e.values[:0]
		e.status = Done
		return Done
	}
	parentBase := e.frames[fi-1].base
	for i, src := range outputs {
		e.values[parentBase+int(f.dest[i])] = values[int(src)]
	}
	e.frames[fi] = frame{}
	e.frames = e.frames[:fi]
	e.values = e.values[:base]
	return Running
}
