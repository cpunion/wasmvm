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
	state  resourceCodecState
	args   []Value
	result [canonical.MaxFlatResults]canonical.Slot
	memory memory
	// Only the synchronous private transaction uses this lazy storage. Public
	// Callers and parameter/result snapshots never point into this object.
	host    *resourceHostCall
	caller  *vm.Caller
	n       *namespace
	plan    *validate.CanonicalFunction
	realloc func(uint64, uint64, uint64, uint64) (uint64, error)
	next    *canonicalFrame
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

func (frame *canonicalFrame) hostStorage() *resourceHostCall {
	if frame.host == nil {
		frame.host = &resourceHostCall{}
	}
	return frame.host
}

func (in *Instance) acquireCanonicalFrame(c *vm.Caller, n *namespace, plan *validate.CanonicalFunction) *canonicalFrame {
	frame := in.acquireFrame()
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
	if frame.host != nil {
		*frame.host = resourceHostCall{}
	}
	clear(frame.args)
	frame.args = frame.args[:0]
	if cap(frame.args) > canonical.MaxFlatParams {
		frame.args = nil
	}
	frame.state = resourceCodecState{}
	frame.memory = memory{}
	frame.caller, frame.n, frame.plan = nil, nil, nil
	frame.next = in.canonicalFrames
	in.canonicalFrames = frame
}
