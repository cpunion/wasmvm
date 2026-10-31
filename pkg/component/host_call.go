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
	codec         *canonical.Codec
	args          []Value
	created       []*Resource
	ctx           context.Context
	cancel        context.CancelFunc
	completion    *HostCompletion
	closed        bool
	values, bytes uint64
}

func (f *Function) hostCall(core *vm.Caller, args []Value) (h *hostCall, err error) {
	h = &hostCall{function: f, codec: f.instance.codec(core, f.scope, nil)}
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
	h.args = make([]Value, len(args))
	for i, p := range f.typ.Parameters {
		h.args[i], err = h.codec.Clone(p.Type, args[i])
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
	return &Caller{callerState: &callerState{active: true, core: core, instance: h.function.instance, function: h.function}}
}
func (h *hostCall) expire(c *Caller) {
	c.mu.Lock()
	c.active = false
	c.mu.Unlock()
	h.created = append(h.created, c.created...)
	c.created = nil
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
	h.limit()
	defer h.account()
	if (result == nil) != (h.function.typ.Result == nil) {
		return nil, vm.ErrType
	}
	if result == nil {
		return nil, nil
	}
	v, err := h.codec.Clone(h.function.typ.Result, *result)
	if err != nil {
		return nil, err
	}
	if h.codec.Resources != nil {
		if err = h.codec.AcceptResult(h.function.typ.Result, &v); err != nil {
			return nil, err
		}
	}
	return &v, nil
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
	bindResourceCaller(h.codec, core)
	guard.run(func() { err = finishResources(h.codec, complete, err) })
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
