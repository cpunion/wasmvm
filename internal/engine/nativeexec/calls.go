package nativeexec

import (
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (e *Execution) resetCalls() {
	e.resumeToken, e.resumeDepth = 0, 0
	clear(e.callFrames)
	e.callFrames = e.callFrames[:0]
	e.callEnabled = false
	if len(e.callWords) != 0 {
		e.callWords[0] = 0
	}
}

func (e *Execution) resumeCalls(in *state.NativeFrames) bool {
	if in.Resume == 0 || in.Resume != e.resumeToken || len(e.callFrames) < 2 ||
		in.Depth != e.resumeDepth || len(e.callWords) == 0 || e.callWords[0] != uint64(len(e.callFrames)-1) {
		return false
	}
	top := e.callFrames[len(e.callFrames)-1]
	return in.Function == top.Function && in.Base == top.Base && in.PC == top.PC+1
}

func (e *Execution) prepareCalls(maxSlots, maxFrames uint64) (uint64, error) {
	capacity := uint64(0)
	if len(e.callWords) != 0 {
		capacity = e.callWords[1]
	}
	wanted := capacity
	if wanted == 0 {
		wanted = min(uint64(16), maxFrames)
	} else if e.callGrow {
		wanted = min(capacity*2, maxFrames)
	}
	if wanted > capacity {
		bytes := uint64(nativeabi.CallHeaderBytes) + wanted*nativeabi.CallFrameBytes
		data, err := native.NewMemory(bytes, bytes)
		if err != nil {
			return 0, err
		}
		buffer, err := data.Buffer()
		if err != nil {
			data.Close()
			return 0, err
		}
		words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
		copy(words, e.callWords)
		if e.callData != nil {
			e.callData.Close()
		}
		e.callData = data
		e.callWords = words
		e.callWords[1] = wanted
	}
	if len(e.callWords) == 0 {
		return 0, wasm.ErrLimit
	}
	e.callWords[2], e.callWords[3] = maxSlots*16, maxFrames
	e.callWords[nativeabi.CallChangedOffset/8] = e.callWords[0]
	e.callGrow = false
	return uint64(uintptr(unsafe.Pointer(unsafe.SliceData(e.callWords)))), nil
}

func linkedFunction(in *state.Instance, index uint64) (*state.Function, bool) {
	if in == nil || in.Owner == nil || index >= uint64(len(in.Functions)) {
		return nil, false
	}
	fn, ok := in.Owner.Function(in.Functions[index])
	return fn, ok && fn.Code != nil && fn.Instance == in
}

func (e *Execution) publishCallFrames(out *state.NativeFrames) error {
	top := e.callWords[0]
	changed := e.callWords[nativeabi.CallChangedOffset/8]
	if out.Depth < 1 || top > e.callWords[1] || top > uint64(out.MaxFrames) ||
		uint64(out.Depth) > uint64(out.MaxFrames)-top ||
		top > (uint64(len(e.callWords))*8-nativeabi.CallHeaderBytes)/nativeabi.CallFrameBytes ||
		e.entryState.Depth != uint64(out.Depth)+top || changed > top || changed > uint64(out.Parents) ||
		changed != 0 && changed >= uint64(len(e.callFrames)) {
		return wasm.ErrInvalid
	}
	count := int(top) + 1
	if cap(e.callFrames) < count {
		frames := make([]state.NativeFrame, count)
		copy(frames, e.callFrames[:int(changed)])
		e.callFrames = frames
	} else {
		if count < len(e.callFrames) {
			clear(e.callFrames[count:])
		}
		e.callFrames = e.callFrames[:count]
	}
	for i := changed; i < top; i++ {
		word := (nativeabi.CallHeaderBytes + i*nativeabi.CallFrameBytes) / 8
		fn, ok := linkedFunction(out.Function.Instance, e.callWords[word+nativeabi.CallFunctionOffset/8])
		used, pc := e.callWords[word+nativeabi.CallUsedOffset/8], e.callWords[word+nativeabi.CallPCOffset/8]
		if !ok || used%16 != 0 || used/16 > uint64(len(e.values)) || pc >= uint64(len(fn.Code.Code)) {
			return wasm.ErrInvalid
		}
		e.callFrames[i] = state.NativeFrame{Function: fn, Base: int(used / 16), PC: uint32(pc + 1), CallPC: uint32(pc)}
	}
	fn, ok := linkedFunction(out.Function.Instance, e.entryState.Function)
	used := e.entryState.Used
	if !ok || used%16 != 0 {
		return wasm.ErrInvalid
	}
	end := used/16 + uint64(len(fn.Code.Slots))
	if end > uint64(len(e.values)) || end > uint64(out.MaxSlots) {
		return wasm.ErrInvalid
	}
	e.callFrames[top] = state.NativeFrame{Function: fn, Base: int(used / 16), PC: out.PC}
	out.Frames, out.Values = e.callFrames[changed:], e.values[:int(end)]
	out.Parents -= uint32(changed)
	out.Function, out.Base = out.Frames[0].Function, out.Frames[0].Base
	out.Depth += int(changed)
	return nil
}
