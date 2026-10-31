package component

import (
	"context"
	"github.com/cpunion/wasmvm/pkg/vm"
)

func (in *Instance) startHostTask(c *vm.Caller, f *Function, args taskArguments, resolve taskResolution) (*componentTask, error) {
	t, err := in.makeTask(c, f.scope, nil, vm.Extern{}, args, resolve)
	if err != nil {
		return nil, err
	}
	t.hostFunction = f
	if !t.waitingEnter && t.ready() {
		if err := t.resume(c); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (t *componentTask) resumeHost(c *vm.Caller) error {
	if err := c.Context().Err(); err != nil {
		return err
	}

	if !t.started {
		h := t.host
		if h == nil {
			args, err := t.args(c)
			if err != nil {
				return err
			}
			t.args = nil
			h, err = t.hostFunction.hostCall(c, args)
			t.host = h
			if err != nil {
				return err
			}
		}
		t.started = true
		h.ctx, h.cancel = context.WithCancel(c.Context())
		if t.work != nil {
			t.work.control.bind(h.cancel)
		}

		p, err := h.start(c)
		if err != nil {
			return err
		}
		t.syncCancellation()
		if p == nil {
			return vm.ErrType
		}
		if err := p.bind(t.in.scheduler().wake); err != nil {
			return err
		}
		h.completion = p
	}
	h := t.host
	if !h.completion.ready() {
		return nil
	}
	resolve, cancelled, err := h.completion.take()
	if err != nil {
		return err
	}
	if cancelled {
		if !t.cancelRequested {
			return vm.ErrType
		}
		t.task.resolved = true
		if err := t.resolve(c, nil, true); err != nil {
			return err
		}
		h.finish(c, false, nil)
		return nil
	}
	value, err := h.invoke(c, resolve)
	if err != nil {
		return err
	}
	value, err = h.accept(value)
	if err != nil {
		return err
	}
	t.task.result, t.task.resolved = value, true
	if err := t.resolve(c, value, false); err != nil {
		return err
	}
	return h.finish(c, true, nil)
}
