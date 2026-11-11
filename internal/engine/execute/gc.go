package execute

import (
	"errors"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (e *Execution) runGC() Status {
	f := &e.frames[len(e.frames)-1]
	fn := f.function
	code := fn.Code
	pc := f.pc
	in := &code.Code[pc]
	base := f.base
	values := e.values[base : base+len(code.Slots)]
	instance := fn.Instance
	var result value.Value
	var a, b, c value.Value
	if in.A != slots.NoSlot {
		a = values[int(in.A)]
	}
	if in.B != slots.NoSlot {
		b = values[int(in.B)]
	}
	im := &code.ReadSource(in.Source, &e.source).Wasm
	var err error
	if n := state.GCAccessArity(in.Op); n != 0 {
		// Slot lowering already supplies the first two operands.
		if n > 2 {
			c = values[int(code.Args[in.Extra+2])]
		}
		result, err = e.store.GCAccess(instance, im, a, b, c)
	} else {
		if int(in.Count) > cap(e.gcArgs) {
			e.gcArgs = make([]value.Value, in.Count)
		}
		e.gcArgs = e.gcArgs[:in.Count]
		operands := e.gcArgs
		for i, slot := range code.Args[in.Extra : in.Extra+in.Count] {
			operands[i] = values[int(slot)]
		}
		if e.store.HeapSafepoint != nil {
			result, err = e.store.GC(e.context, instance, im, operands)
		} else {
			result, err = e.store.GC(e.context, instance, im, operands, e.RetainedValues)
		}
		e.clearGCArgs()
	}
	if err != nil {
		var t *wasm.Trap
		if errors.As(err, &t) {
			t.Function = fn.Index
		}
		return e.fail(err)
	}
	if in.Dst != slots.NoSlot {
		values[int(in.Dst)] = result
	}
	pc++
	f.pc = pc
	return Running
}
