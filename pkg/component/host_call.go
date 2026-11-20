package component

import (
	"context"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Both immediate and delayed host functions share argument/result ownership.
type hostCall struct {
	function      *Function
	codec         canonical.Codec
	args          []Value
	created       []*Resource
	reads         []*pendingHostRead
	ctx           context.Context
	cancel        context.CancelFunc
	completion    *HostCompletion
	closed        bool
	prompt        bool
	values, bytes uint64
}

type resourceHostCall struct {
	hostCall
	resources resourceCodec
}

// Non-nil storage is zeroed, private synchronous canonical storage. It cannot
// belong to a pending task or escape through a public callback capability.
func (in *Instance) newHostCall(core *vm.Caller, f *Function, n *namespace, storage *resourceHostCall) *hostCall {
	var h *hostCall
	if storage != nil {
		h = &storage.hostCall
	}
	if n != nil && (len(n.resources) != 0 || n.handles != nil) {
		if storage == nil {
			storage = &resourceHostCall{}
			h = &storage.hostCall
		}
		storage.resources = resourceCodec{in: in, n: n, core: core}
		h.codec.Resources = &storage.resources
	} else if h == nil {
		h = &hostCall{}
	}
	h.function = f
	in.configureCodec(&h.codec, core, n, nil, true)
	return h
}

func (f *Function) hostCall(core *vm.Caller, args []Value) (h *hostCall, err error) {
	return f.hostParameters(core, args, false, nil)
}

// Only the canonical lower's normalized invocation-owned arguments may be
// borrowed. Public Function.Call and pending calls keep owned snapshots.
func (f *Function) hostParameters(core *vm.Caller, args []Value, borrowed bool, storage *resourceHostCall) (h *hostCall, err error) {
	h = f.instance.newHostCall(core, f, f.scope, storage)
	if len(args) != len(f.typ.Parameters) {
		return h, vm.ErrType
	}
	complete := false
	defer func() {
		h.account()
		if !complete {
			err = h.finish(core, false, err)
		}
	}()
	h.limit()
	if borrowed {
		h.args = args
	} else {
		h.args = make([]Value, len(args))
	}
	plans := f.valuePlans()
	for i := range f.typ.Parameters {
		if borrowed {
			err = plans.parameters[i].RetainBorrowed(&h.codec, args[i])
		} else {
			h.args[i], err = plans.parameters[i].Clone(&h.codec, args[i])
		}
		if err != nil {
			return h, err
		}
	}
	if h.codec.Resources != nil {
		if err = h.codec.AcceptParameters(f.typ.Parameters, h.args); err != nil {
			return h, err
		}
	}
	complete = true
	return h, nil
}
func (h *hostCall) caller(core *vm.Caller) *Caller {
	return newCallbackCaller(core, h.function.instance, h.function, h.prompt)
}
func (h *hostCall) expire(c *Caller) {
	c.mu.Lock()
	c.active = false
	c.mu.Unlock()
	h.created = append(h.created, c.created...)
	c.created = nil
	h.reads = append(h.reads, c.reads...)
	c.reads = nil
}
func (h *hostCall) invoke(core *vm.Caller, fn func(*Caller) (*Value, error)) (*Value, error) {
	c := h.caller(core)
	defer h.expire(c)
	return fn(c)
}
func (h *hostCall) start(core *vm.Caller) (p *HostCompletion, err error) {
	c := h.caller(core)
	c.ctx = h.ctx
	defer h.expire(c)
	return h.function.asyncHost(c, h.args)
}
func (h *hostCall) accept(result *Value) (*Value, error) {
	var accepted Value
	if err := h.acceptInto(result, &accepted); err != nil || result == nil {
		return nil, err
	}
	return &accepted, nil
}

func (h *hostCall) acceptInto(result, accepted *Value) error {
	h.limit()
	defer h.account()
	if (result == nil) != (h.function.typ.Result == nil) {
		return vm.ErrType
	}
	if result == nil {
		return nil
	}
	v, err := h.function.valuePlans().result.Clone(&h.codec, *result)
	if err != nil {
		return err
	}
	*accepted = v
	if h.codec.Resources != nil {
		return h.codec.AcceptResult(h.function.typ.Result, accepted)
	}
	return nil
}
func (h *hostCall) finish(core *vm.Caller, complete bool, err error) error {
	if h.closed {
		return err
	}
	h.closed = true
	defer func() {
		in := h.function.instance
		in.heldValues -= h.values
		in.heldBytes -= h.bytes
		h.values, h.bytes = 0, 0
	}()
	if h.cancel != nil {
		h.cancel()
	}

	created := h.created
	h.created = nil
	h.args = nil
	guard := cleanupGuard{}
	for _, read := range h.reads {
		guard.run(func() { read.abort(core) })
	}
	h.reads = nil
	if h.completion != nil {
		guard.run(h.completion.close)
	}
	if guard.panicked != nil {
		complete = false
		if err == nil {
			err = wasm.ErrTrap
		}
	}

	if !complete {
		for _, ref := range created {
			guard.run(func() { h.function.instance.cleanupResource(core, ref) })
		}
	} else {
		for _, ref := range created {
			h.function.instance.publishReference(ref)
		}
	}
	bindResourceCaller(&h.codec, core)
	guard.run(func() { err = finishResources(&h.codec, complete, err) })
	guard.finish()
	return err

}

func (h *hostCall) limit() {
	in := h.function.instance
	v, b := h.codec.Usage()
	h.codec.MaxValues = v + in.config.MaxValues - in.heldValues
	h.codec.Transfer.MaxBytes = b + in.config.MaxBytes - in.heldBytes
}
func (h *hostCall) account() {
	if h.closed {
		return
	}
	in := h.function.instance
	v, b := h.codec.Usage()
	in.heldValues += v - h.values
	in.heldBytes += b - h.bytes
	h.values, h.bytes = v, b
}

func (h *hostCall) finishDeferred(core *vm.Caller, complete *bool, err *error) {
	guard := cleanupGuard{panicked: recover()}
	guard.run(func() { *err = h.finish(core, *complete, *err) })
	guard.finish()
}
