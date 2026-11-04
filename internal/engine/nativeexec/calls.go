package nativeexec

import (
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (e *Execution) resetCalls() {
	clear(e.callFrames)
	e.callFrames = e.callFrames[:0]
	e.callEnabled = false
	if len(e.callWords) != 0 {
		e.callWords[0] = 0
	}
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
		if e.callData != nil {
			e.callData.Close()
		}
		e.callData = data
		e.callWords = unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
		e.callWords[1] = wanted
	}
	if len(e.callWords) == 0 {
		return 0, wasm.ErrLimit
	}
	e.callWords[0], e.callWords[2], e.callWords[3] = 0, maxSlots*16, maxFrames
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
	if top > e.callWords[1] || top+uint64(out.Depth) > uint64(out.MaxFrames) ||
		top*nativeabi.CallFrameBytes+nativeabi.CallHeaderBytes > uint64(len(e.callWords))*8 || e.entryState.Depth != uint64(out.Depth)+top {
		return wasm.ErrInvalid
	}
	count := int(top) + 1
	if cap(e.callFrames) < count {
		e.callFrames = make([]state.NativeFrame, count)
	} else {
		e.callFrames = e.callFrames[:count]
	}
	for i := uint64(0); i < top; i++ {
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
	out.Frames, out.Values = e.callFrames, e.values[:int(end)]
	return nil
}
