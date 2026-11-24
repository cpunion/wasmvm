package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
)

// A frame is private scratch for one synchronous canonical call. Ordinary host
// callbacks receive independent snapshots; explicitly annotated callbacks borrow
// these arguments only until return. Capabilities remain fresh. Suspended calls
// use their existing task-owned state instead of these frames.
type canonicalFrame struct {
	state    resourceCodecState
	auxState resourceCodecState
	args     []Value
	result   [canonical.MaxFlatResults]canonical.Slot
	memory   memory
	// Only the synchronous private transaction uses this lazy storage. Public
	// Callers and parameter/result snapshots never point into this object.
	host        *resourceHostCall
	hostRead    hostReadBuffer
	readFn      Function
	hostValue   Value
	cloneValues [8]Value
	cloneBuf    []Value
	caller       *vm.Caller
	n            *namespace
	plan         *validate.CanonicalFunction
	realloc      func(uint64, uint64, uint64, uint64) (uint64, error)
	next         *canonicalFrame
	stateUsed    bool
	auxUsed      bool
	hostUsed     bool
	hostReadUsed bool
}

func (in *Instance) acquireFrame() *canonicalFrame {
	frame := in.canonicalFrames
	if frame == nil {
		frame = &canonicalFrame{}
	} else {
		in.canonicalFrames = frame.next
		frame.next = nil
	}
	return frame
}

func (frame *canonicalFrame) primaryState() *resourceCodecState {
	frame.stateUsed = true
	return &frame.state
}

func (frame *canonicalFrame) aux() *resourceCodecState {
	frame.auxUsed = true
	return &frame.auxState
}

func (frame *canonicalFrame) hostReadStorage() *hostReadBuffer {
	frame.hostReadUsed = true
	return &frame.hostRead
}

func (frame *canonicalFrame) hostStorage() *resourceHostCall {
	if frame.host == nil {
		frame.host = &resourceHostCall{}
	}
	frame.hostUsed = true
	return frame.host
}

func (in *Instance) acquireCanonicalFrame(c *vm.Caller, n *namespace, plan *validate.CanonicalFunction) *canonicalFrame {
	frame := in.acquireFrame()
	frame.stateUsed = true
	frame.caller, frame.n, frame.plan = c, n, plan
	if n != nil && (len(n.resources) != 0 || n.handles != nil) {
		// Release already zeroed the state. Rebind only the active owners.
		r := &frame.state.resources
		r.in, r.n, r.core = in, n, c
		frame.state.codec.Resources = &frame.state.resources
	}
	in.configureCanonicalCodec(&frame.state.codec, c, n, plan, true, frame)
	return frame
}

func (frame *canonicalFrame) reallocate(old, size, alignment, next uint64) (uint64, error) {
	return canonicalReallocate(frame.caller, frame.n, frame.plan, old, size, alignment, next)
}

// Release runs after canonical resource cleanup, also when cleanup panics.
// Never retain a Caller, Context, memory capability, value or resource between
// invocations. Reentrant calls acquire a different frame until this one returns.
func (in *Instance) releaseCanonicalFrame(frame *canonicalFrame) {
	if frame.hostUsed {
		*frame.host = resourceHostCall{}
		frame.hostUsed = false
	}
	if len(frame.args) != 0 {
		clear(frame.args)
		frame.args = frame.args[:0]
	}
	if cap(frame.args) > canonical.MaxFlatParams {
		frame.args = nil
	}
	if frame.stateUsed {
		frame.state = resourceCodecState{}
		frame.stateUsed = false
	}
	if frame.auxUsed {
		frame.auxState = resourceCodecState{}
		frame.auxUsed = false
	}
	if frame.hostReadUsed {
		frame.hostRead = hostReadBuffer{}
		frame.hostReadUsed = false
	}
	if frame.readFn.instance != nil {
		frame.readFn = Function{}
	}
	if frame.cloneBuf != nil {
		frame.cloneValues = [8]Value{}
		frame.cloneBuf = nil
	}
	frame.hostValue = Value{}
	if frame.memory.caller != nil {
		frame.memory = memory{}
	}
	frame.caller, frame.n, frame.plan = nil, nil, nil
	frame.next = in.canonicalFrames
	in.canonicalFrames = frame
}
