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

	var ready bool
	var resolve func(*Caller) (*Value, error)
	var cancelled bool
	var err error
	if !t.started {
		h := t.host
		if h == nil {
			var args []Value
			if t.args != nil {
				args, err = t.args(c)
				if err != nil {
					return err
				}
				t.args = nil
			}
			var storage *resourceHostCall
			if t.sourceRead != nil {
				storage = &t.sourceRead.host
			}
			h, err = t.hostFunction.hostParameters(c, args, false, storage)
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

		var p *HostCompletion
		if t.work != nil && t.work.start != nil {
			p, err = h.startWith(c, t.work.startWork)
		} else if t.sourceRead != nil {
			p, err = h.startWith(c, t.sourceRead.start)
		} else {
			p, err = h.start(c)
		}
		if err != nil {
			return err
		}
		t.syncCancellation()
		if p == nil {
			return vm.ErrType
		}
		if resolve, cancelled, ready, err = p.poll(t.in.scheduler().wake, true); err != nil {
			return err
		}
		h.completion = p
	} else {
		resolve, cancelled, ready, err = t.host.completion.poll(nil, false)
		if err != nil {
			return err
		}
	}
	h := t.host
	if !ready {
		return nil
	}
	if cancelled {
		// Cancellation can arrive after resume's initial snapshot but before the
		// producer publishes this acknowledgement. Synchronize after the handoff.
		t.syncCancellation()
		if err := c.Context().Err(); err != nil {
			return err
		}
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
	if err = h.acceptInto(value, &t.task.resultValue); err != nil {
		return err
	}
	if value != nil {
		value = &t.task.resultValue
	}
	t.task.result, t.task.resolved = value, true
	if err := t.resolve(c, value, false); err != nil {
		return err
	}
	return h.finish(c, true, nil)
}
