package execute

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// A prompt delimits guest frames in the same execution arena. Capturing one
// copies only its guest state; it never captures a Go/C stack or starts a worker.
type prompt struct {
	base, owner int
	instance    *state.Instance
	handlers    []slots.Catch
}

type continuationBody struct {
	function *state.Function
	frames   []frame
	values   []value.Value
	prompts  []prompt
	inputs   []wasm.ValueType
	bound    []value.Value
	dest     []slots.Slot
	throwPC  uint32
}

func (c *continuationBody) HeapBytes() uint64 {
	return 96 + uint64(len(c.frames))*64 + uint64(len(c.values)+len(c.bound))*16 + uint64(len(c.prompts))*64
}
func (c *continuationBody) Visit(visit func(wasm.ValueType, value.Value), mark func(*state.Instance)) {
	if c.function != nil {
		mark(c.function.Instance)
	}
	for i, v := range c.bound {
		if c.inputs[i].IsReference() {
			visit(c.inputs[i], v)
		}
	}
	for _, f := range c.frames {
		mark(f.function.Instance)
		for i, typ := range f.function.Code.Slots {
			if typ.IsReference() {
				visit(typ, c.values[f.base+i])
			}
		}
	}
	for _, p := range c.prompts {
		mark(p.instance)
	}
}

func (e *Execution) trimPrompts(frames int) {
	for len(e.prompts) != 0 && e.prompts[len(e.prompts)-1].base >= frames {
		e.prompts[len(e.prompts)-1] = prompt{}
		e.prompts = e.prompts[:len(e.prompts)-1]
	}
}

func (e *Execution) takeContinuation(ref uint64) (*continuationBody, error) {
	body, err := e.store.Continuation(ref, true)
	if err != nil {
		return nil, err
	}
	c, ok := body.(*continuationBody)
	if !ok {
		return nil, wasm.ErrInvalid
	}
	e.transferring = c
	return c, nil
}

func (e *Execution) findPrompt(tag *state.Tag, kind byte) (int, slots.Catch, error) {
	for i := len(e.prompts) - 1; i >= 0; i-- {
		p := &e.prompts[i]
		for _, h := range p.handlers {
			if h.Kind == kind && p.instance.Tags[h.Tag] == tag {
				return i, h, nil
			}
		}
	}
	return 0, slots.Catch{}, &state.Suspension{Owner: e.store, Tag: tag, Switch: kind == 1}
}

func (e *Execution) capture(pi int, typ uint32, inputs []wasm.ValueType, dest []slots.Slot, pc uint32) (uint64, error) {
	p := e.prompts[pi]
	base := e.frames[p.base].base
	c := &continuationBody{frames: append([]frame(nil), e.frames[p.base:]...), values: append([]value.Value(nil), e.values[base:]...),
		prompts: append([]prompt(nil), e.prompts[pi+1:]...), inputs: inputs, dest: dest, throwPC: pc}
	for i := range c.frames {
		c.frames[i].base -= base
	}
	for i := range c.prompts {
		c.prompts[i].base -= p.base
		c.prompts[i].owner -= p.base
	}
	return e.store.NewContinuation(typ, c)
}

func (e *Execution) detach(pi int, keep bool) {
	p := e.prompts[pi]
	base := e.frames[p.base].base
	clear(e.frames[p.base:])
	e.frames = e.frames[:p.base]
	e.values = e.values[:base]
	n := pi
	if keep {
		n++
	}
	clear(e.prompts[n:])
	e.prompts = e.prompts[:n]
}

// Restore into the selected backend's arena. Every native entry reloads the
// same slots/PCs, so suspension never retains a machine address or native lease.
func (e *Execution) restore(c *continuationBody, args []value.Value, dest []slots.Slot, injected *state.Exception) error {
	all := append(append([]value.Value(nil), c.bound...), args...)
	if injected == nil && len(all) != len(c.inputs) {
		return wasm.ErrInvalid
	}
	if c.function != nil {
		if injected != nil {
			e.raise(injected) // Injection precedes invocation of a fresh function.
			return nil
		}
		if c.function.Code == nil {
			caller := e.frames[len(e.frames)-1].function.Instance
			copy(e.prepareHost(c.function, caller, len(all)), all)
			e.pendingDest, e.status = dest, HostRequest
			return nil
		}
		if err := e.push(c.function, dest); err != nil {
			return err
		}
		copy(e.values[e.frames[len(e.frames)-1].base:], all)
		return nil
	}
	if len(c.frames) == 0 || uint64(len(e.frames)+len(c.frames)) > uint64(e.config.Limits.MaxFrames) || uint64(len(e.values)+len(c.values)) > uint64(e.config.Limits.MaxSlots) {
		return &wasm.Error{Kind: wasm.ErrLimit, Detail: "restored continuation frames/slots"}
	}
	frameBase, valueBase := len(e.frames), len(e.values)
	needed := valueBase + len(c.values)
	if needed > cap(e.values) {
		if e.config.Backend != nil {
			grown, err := e.config.Backend.Resize(needed, needed)
			if err != nil {
				return err
			}
			e.values = grown
		} else {
			grown := make([]value.Value, needed)
			copy(grown, e.values)
			e.values = grown
		}
	} else {
		e.values = e.values[:needed]
	}
	copy(e.values[valueBase:], c.values)
	for _, f := range c.frames {
		f.base += valueBase
		e.frames = append(e.frames, f)
	}
	e.frames[frameBase].dest = dest
	for _, p := range c.prompts {
		p.base += frameBase
		p.owner += frameBase
		e.prompts = append(e.prompts, p)
	}
	top := &e.frames[len(e.frames)-1]
	if injected != nil {
		top.callPC, top.suspended = c.throwPC, true
		e.raise(injected)
	} else {
		for i, slot := range c.dest {
			e.values[top.base+int(slot)] = all[i]
		}
		top.suspended = false
	}
	return nil
}

func (e *Execution) stackInstruction(in slots.Instruction, im *module.Instr) error {
	if e.store.HeapSafepoint == nil {
		e.store.HeapSafepoint = func() {
			var roots []state.ReferenceValue
			var active []*state.Instance
			e.RetainedValues(func(t wasm.ValueType, v value.Value) { roots = append(roots, state.ReferenceValue{Type: t, Value: v}) })
			e.RetainedInstances(func(in *state.Instance) { active = append(active, in) })
			e.store.CollectReferences(roots, active, nil)
		}
		defer func() { e.store.HeapSafepoint = nil }()
	}
	fi := len(e.frames) - 1
	f := &e.frames[fi]
	owner, code, pc, base := f.function.Instance, f.function.Code, f.pc, f.base
	outputs := len(code.Sources[in.Source].Outputs)
	tuple := code.Args[in.Extra : in.Extra+in.Count+uint32(outputs)]
	args := make([]value.Value, in.Count)
	for i, slot := range tuple[:in.Count] {
		args[i] = e.values[base+int(slot)]
	}
	dest := tuple[in.Count:]
	f.pc++
	defer func() { e.transferring = nil }()
	switch in.Op {
	case 0xe0:
		if args[0].Lo == 0 {
			return &wasm.Trap{Code: wasm.TrapNullReference}
		}
		fn, ok := e.store.Function(args[0].Lo)
		if !ok {
			return state.ErrReference
		}
		ref, err := e.store.NewContinuation(owner.Types[im.A], &continuationBody{function: fn, inputs: fn.Signature.Params})
		if err != nil {
			return err
		}
		e.values[base+int(dest[0])] = value.Value{Lo: ref}
	case 0xe1:
		c, err := e.takeContinuation(args[len(args)-1].Lo)
		if err != nil {
			return err
		}
		next := *c
		next.bound = append(append([]value.Value(nil), c.bound...), args[:len(args)-1]...)
		ref, err := e.store.NewContinuation(owner.Types[im.B], &next)
		if err != nil {
			return err
		}
		e.values[base+int(dest[0])] = value.Value{Lo: ref}
	case 0xe2:
		tag := owner.Tags[im.A]
		pi, h, err := e.findPrompt(tag, 0)
		if err != nil {
			if suspension, ok := err.(*state.Suspension); ok {
				suspension.Args = args
			}
			return err
		}
		p := e.prompts[pi]
		ref, err := e.capture(pi, p.instance.Types[h.Type], tag.Signature.Results, dest, pc)
		if err != nil {
			return err
		}
		e.detach(pi, false)
		caller := &e.frames[p.owner]
		for i, v := range args {
			e.values[caller.base+int(h.Dest[i])] = v
		}
		e.values[caller.base+int(h.Dest[len(h.Dest)-1])] = value.Value{Lo: ref}
		caller.pc, caller.suspended = h.PC, false
	case 0xe3, 0xe4, 0xe5:
		_, err := e.store.Continuation(args[len(args)-1].Lo, false)
		if err != nil {
			return err
		}
		var injected *state.Exception
		if in.Op == 0xe4 {
			injected, err = e.store.NewException(owner.Tags[im.B], args[:len(args)-1], e.RetainedValues)
		} else if in.Op == 0xe5 {
			if args[0].Lo == 0 {
				return &wasm.Trap{Code: wasm.TrapNullReference}
			}
			if _, _, err = e.store.Exception(args[0].Lo); err == nil {
				injected = &state.Exception{Owner: e.store, Ref: args[0].Lo}
			}
		}
		if err != nil {
			return err
		}
		c, err := e.takeContinuation(args[len(args)-1].Lo)
		if err != nil {
			return err
		}
		f.callPC, f.suspended = pc, true
		// Host continuations cannot suspend through this host boundary.
		if c.function == nil || c.function.Code != nil {
			e.prompts = append(e.prompts, prompt{base: len(e.frames), owner: fi, instance: owner, handlers: code.Prompts[in.Source]})
		}
		if in.Op != 0xe3 {
			args = nil
		} else {
			args = args[:len(args)-1]
		}
		return e.restore(c, args, dest, injected)
	case 0xe6:
		_, err := e.store.Continuation(args[len(args)-1].Lo, false)
		if err != nil {
			return err
		}
		tag := owner.Tags[im.B]
		pi, _, err := e.findPrompt(tag, 1)
		if err != nil {
			return err
		}
		c, err := e.takeContinuation(args[len(args)-1].Lo)
		if err != nil {
			return err
		}
		sig, ok := e.store.ContinuationSignature(owner.Types[im.A])
		if !ok || len(sig.Params) == 0 {
			return wasm.ErrInvalid
		}
		backType, _ := sig.Params[len(sig.Params)-1].TypeIndex()
		back, ok := e.store.ContinuationSignature(backType)
		if !ok {
			return wasm.ErrInvalid
		}
		ref, err := e.capture(pi, backType, back.Params, dest, pc)
		if err != nil {
			return err
		}
		p := e.prompts[pi]
		e.detach(pi, true)
		caller := &e.frames[p.owner]
		resume := caller.function.Code.Code[caller.callPC]
		resumeOutputs := len(caller.function.Code.Sources[resume.Source].Outputs)
		resumeDest := caller.function.Code.Args[resume.Extra+resume.Count : resume.Extra+resume.Count+uint32(resumeOutputs)]
		args[len(args)-1] = value.Value{Lo: ref}
		return e.restore(c, args, resumeDest, nil)
	}
	return nil
}
