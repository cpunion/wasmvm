package execute

import (
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Keep packet construction, validation and cleanup out of the portable opcode
// dispatcher. Inlining this native-only bridge enlarges the shared hot loop.
//
//go:noinline
func (e *Execution) runBackend(backend Backend, fn *state.Function, base int, pc uint32) (uint32, error) {
	resume := e.frames[len(e.frames)-1].nativeResume
	e.frames[len(e.frames)-1].nativeResume = 0
	e.nativeFrames = state.NativeFrames{Function: fn, Base: base, Depth: len(e.frames), PC: pc, Fuel: e.fuel,
		MaxFrames: e.config.Limits.MaxFrames, MaxSlots: e.config.Limits.MaxSlots, Values: e.values, Resume: resume}
	e.nativeResume = 0
	err := backend.RunFrames(&e.nativeFrames)
	if err == nil {
		err = e.adoptNativeFrames(&e.nativeFrames)
	}
	if err != nil && len(e.nativeFrames.Values) >= len(e.values) {
		// Preserve the current owned storage after an error or relocation.
		e.values = e.nativeFrames.Values[:len(e.values)]
	}
	pc, e.fuel = e.nativeFrames.PC, e.nativeFrames.Fuel
	if err == nil {
		e.nativeResume = e.nativeFrames.Resume
	}
	e.nativeFrames = state.NativeFrames{}
	return pc, err
}

// adoptNativeFrames validates the whole transfer before changing the shared
// stack. Native execution coalesces ordinary same-instance direct, indirect and reference
// calls; imports, cross-instance targets and tail calls retain runtime boundaries. The
// existing root destination belongs to its earlier caller and must survive.
func (e *Execution) adoptNativeFrames(batch *state.NativeFrames) error {
	if len(e.frames) == 0 {
		return wasm.ErrInvalid
	}
	if uint64(batch.Parents) >= uint64(len(e.frames)) {
		return wasm.ErrInvalid
	}
	rootIndex := len(e.frames) - 1 - int(batch.Parents)
	root := e.frames[rootIndex]
	frames := batch.Frames
	if len(frames) == 0 {
		if batch.Parents != 0 || len(batch.Values) != len(e.values) || uint64(batch.PC) >= uint64(len(root.function.Code.Code)) {
			return wasm.ErrInvalid
		}
		e.values = batch.Values
		return nil
	}
	if uint64(rootIndex)+uint64(len(frames)) > uint64(e.config.Limits.MaxFrames) {
		return &wasm.Error{Kind: wasm.ErrLimit, Detail: "guest call stack exhausted"}
	}
	if frames[0].Function != root.function || frames[0].Base != root.base || frames[len(frames)-1].PC != batch.PC {
		return wasm.ErrInvalid
	}
	base := uint64(root.base)
	for i, f := range frames {
		if f.Function == nil || f.Function.Code == nil || f.Function.Instance != root.function.Instance ||
			f.Base < 0 || uint64(f.Base) != base || uint64(f.PC) >= uint64(len(f.Function.Code.Code)) {
			return wasm.ErrInvalid
		}
		base += uint64(len(f.Function.Code.Slots))
		if base > uint64(e.config.Limits.MaxSlots) {
			return &wasm.Error{Kind: wasm.ErrLimit, Detail: "guest frame slots"}
		}
		if i+1 == len(frames) {
			continue
		}
		code := f.Function.Code
		if uint64(f.CallPC) >= uint64(len(code.Code)) || f.PC != f.CallPC+1 {
			return wasm.ErrInvalid
		}
		call := code.Code[f.CallPC]
		if uint64(call.Source) >= uint64(code.SourceCount()) {
			return wasm.ErrInvalid
		}
		index, _ := code.SourceIndices(call.Source)
		instance := f.Function.Instance
		callee := frames[i+1].Function
		if callee == nil || callee.Instance != instance {
			return wasm.ErrInvalid
		}
		count := call.Count
		switch call.Op {
		case 0x10:
			if uint64(index) >= uint64(len(instance.Functions)) {
				return wasm.ErrInvalid
			}
			target, ok := e.store.Function(instance.Functions[index])
			if !ok || target != callee {
				return wasm.ErrInvalid
			}
		case 0x11, 0x14:
			if count == 0 || uint64(index) >= uint64(len(instance.Types)) ||
				!e.store.FunctionSubtype(callee.TypeID, instance.Types[index]) {
				return wasm.ErrInvalid
			}
			count--
		default:
			return wasm.ErrInvalid
		}
		if uint64(count) != uint64(len(callee.Signature.Params)) ||
			uint64(call.Extra)+uint64(call.Count)+uint64(len(callee.Signature.Results)) > uint64(len(code.Args)) {
			return wasm.ErrInvalid
		}
	}
	if base != uint64(len(batch.Values)) {
		return wasm.ErrInvalid
	}
	needed := rootIndex + len(frames)
	if needed < len(e.frames) {
		e.trimPrompts(needed)
		clear(e.frames[needed:])
	}
	if needed > cap(e.frames) {
		grown := make([]frame, needed)
		copy(grown, e.frames[:rootIndex])
		e.frames = grown
	} else {
		e.frames = e.frames[:needed]
	}
	for i, f := range frames {
		dest := root.dest
		if i != 0 {
			parent := frames[i-1]
			code := parent.Function.Code
			call := code.Code[parent.CallPC]
			start := call.Extra + call.Count
			dest = code.Args[start : start+uint32(len(f.Function.Signature.Results))]
		}
		e.frames[rootIndex+i] = frame{function: f.Function, base: f.Base, pc: f.PC,
			callPC: f.CallPC, suspended: i+1 != len(frames), dest: dest}
	}
	e.values = batch.Values
	return nil
}
