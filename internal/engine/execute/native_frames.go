package execute

import (
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// FrameBackend extends the native boundary with a batch of explicit guest
// frames. It shares the existing frame machine for all observable runtime work;
// it does not add a second host, exception, GC or continuation driver.
type FrameBackend interface {
	RunFrames(*state.NativeFrames) error
}

// adoptNativeFrames validates the whole transfer before changing the shared
// stack. Native execution may only coalesce ordinary same-instance direct calls;
// imported, indirect and tail calls remain explicit runtime boundaries. The
// existing root destination belongs to its earlier caller and must survive.
func (e *Execution) adoptNativeFrames(batch *state.NativeFrames) error {
	if len(e.frames) == 0 {
		return wasm.ErrInvalid
	}
	rootIndex := len(e.frames) - 1
	root := e.frames[rootIndex]
	frames := batch.Frames
	if len(frames) == 0 {
		if len(batch.Values) != len(e.values) || uint64(batch.PC) >= uint64(len(root.function.Code.Code)) {
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
		if call.Op != 0x10 || uint64(call.Source) >= uint64(len(code.Sources)) {
			return wasm.ErrInvalid
		}
		index := code.Sources[call.Source].Wasm.A
		instance := f.Function.Instance
		if uint64(index) >= uint64(len(instance.Functions)) {
			return wasm.ErrInvalid
		}
		callee, ok := e.store.Function(instance.Functions[index])
		if !ok || callee != frames[i+1].Function || uint64(call.Count) != uint64(len(callee.Signature.Params)) ||
			uint64(call.Extra)+uint64(call.Count)+uint64(len(callee.Signature.Results)) > uint64(len(code.Args)) {
			return wasm.ErrInvalid
		}
	}
	if base != uint64(len(batch.Values)) {
		return wasm.ErrInvalid
	}
	needed := rootIndex + len(frames)
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
