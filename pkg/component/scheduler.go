package component

import (
	"errors"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type taskArguments func(*vm.Caller) ([]Value, error)
type taskResolution func(*vm.Caller, *Value, bool) error

// Callback continuations retain component state, never a Go stack or an expired
// Caller. Every turn executes under the currently active Core session capability.
type callbackTask struct {
	in               *Instance
	n                *namespace
	task             callTask
	entry            vm.Extern
	args             taskArguments
	onResolve        taskResolution
	previous, next   *callbackTask
	waiting          *waitableSet
	started, running bool
	exited           bool
	cancelRequested  bool
	cancelDelivered  bool
}

type taskScheduler struct {
	head, tail *callbackTask
	live       uint32
	subtasks   *componentSubtask
	sets       *waitableSet
}

func (in *Instance) startCallback(c *vm.Caller, n *namespace, plan *validate.CanonicalFunction, entry vm.Extern, args taskArguments, resolve taskResolution) (*callbackTask, error) {
	if n.failed || n.blocked {
		return nil, wasm.ErrTrap
	}
	if in.tasks == nil {
		in.tasks = &taskScheduler{}
	}
	s := in.tasks
	if s.live >= in.config.MaxTasks {
		return nil, wasm.ErrLimit
	}
	t := &callbackTask{in: in, n: n, entry: entry, args: args, onResolve: resolve}
	t.task = callTask{plan: plan, callback: t}
	t.previous = s.tail
	if s.tail != nil {
		s.tail.next = t
	} else {
		s.head = t
	}
	s.tail = t
	s.live++
	if t.ready() {
		if err := t.resume(c); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (t *callbackTask) ready() bool {
	if t.running || t.exited {
		return false
	}
	if !t.started && t.cancelRequested {
		return true
	}
	// Stackful async threads have independent shadow stacks. Other active
	// turns retain their instance's exclusive execution until they return.
	if t.n.entered && (t.n.task == nil || t.n.task.plan == nil || !t.n.task.plan.Async || t.n.task.plan.Callback != nil) {
		return false
	}
	return !t.started || t.cancelRequested && !t.cancelDelivered || t.waiting == nil || t.waiting.ready != nil
}

func bindResourceCaller(codec *canonical.Codec, c *vm.Caller) {
	if codec != nil {
		if r, ok := codec.Resources.(*resourceCodec); ok {
			r.core = c
		}
	}
}

func (t *callbackTask) resolve(c *vm.Caller, value *Value, cancelled bool) error {
	if t.onResolve != nil {
		return t.onResolve(c, value, cancelled)
	}
	return nil
}

func (t *callbackTask) leaveWait() {
	if t.waiting != nil {
		t.waiting.waiters--
		t.waiting = nil
	}
}

func (t *callbackTask) finish(c *vm.Caller, failed error) error {
	if t.exited {
		return failed
	}
	t.leaveWait()
	t.exited = true
	s := t.in.tasks
	if t.previous != nil {
		t.previous.next = t.next
	} else {
		s.head = t.next
	}
	if t.next != nil {
		t.next.previous = t.previous
	} else {
		s.tail = t.previous
	}
	s.live--
	t.previous, t.next, t.args, t.onResolve = nil, nil, nil, nil
	if failed != nil && t.started {
		t.n.failed = true
	}
	if t.task.returned != nil {
		bindResourceCaller(t.task.returned, c)
		failed = finishResources(t.task.returned, failed == nil, failed)
	}
	if t.task.parameters != nil {
		bindResourceCaller(t.task.parameters, c)
		failed = finishResources(t.task.parameters, failed == nil, failed)
	}
	t.task.parameters, t.task.returned = nil, nil
	return failed
}

func (t *callbackTask) resume(c *vm.Caller) (err error) {
	t.running = true
	previous, entered := t.n.task, t.n.entered
	t.n.task, t.n.entered = &t.task, true
	complete := false
	defer func() {
		t.running = false
		t.n.task, t.n.entered = previous, entered
		if !complete {
			err = t.finish(c, errors.Join(wasm.ErrTrap, err))
		}
	}()
	var called vm.CallResult
	if !t.started {
		if t.cancelRequested {
			t.task.resolved = true
			if err = t.resolve(c, nil, true); err != nil {
				return err
			}
			complete = true
			return t.finish(c, nil)
		}
		args, failed := t.args(c)
		if failed != nil {
			return failed
		}
		t.args = nil
		t.started = true
		codec := t.in.codec(c, t.n, t.task.plan)
		t.task.parameters = codec
		for i, p := range t.task.plan.Type.Parameters {
			if err = codec.Validate(p.Type, args[i]); err != nil {
				return err
			}
		}
		if codec.Resources != nil {
			if err = codec.PrepareParameters(t.task.plan.Type.Parameters, args); err != nil {
				return err
			}
		}
		slots, failed := lowerParameters(codec, t.task.plan, args)
		if failed != nil {
			return failed
		}
		called, err = c.Call(t.entry, vmValues(slots))
	} else {
		var event [3]uint32
		if t.cancelRequested && !t.cancelDelivered {
			t.cancelDelivered = true
			event[0] = 6
		} else if t.waiting != nil {
			event, err = t.waiting.poll(c)
			if err != nil {
				return err
			}
		}
		t.leaveWait()
		called, err = c.Call(t.n.items[model.CoreFunctionSort][*t.task.plan.Callback].core, []vm.Value{vm.I32(event[0]), vm.I32(event[1]), vm.I32(event[2])})
	}
	if err != nil {
		return err
	}
	bits, _, _ := called.Values[0].Bits()
	packed := uint32(bits)
	switch packed & 15 {
	case 0:
		if !t.task.resolved {
			return wasm.ErrTrap
		}
		complete = true
		return t.finish(c, nil)
	case 1:
	case 2:
		t.waiting, err = t.in.waitableSet(t.n, packed>>4)
		if err != nil {
			return err
		}
		t.waiting.waiters++
	default:
		return wasm.ErrTrap
	}
	complete = true
	return nil
}

func (in *Instance) stepTasks(c *vm.Caller) (bool, error) {
	if in.tasks == nil {
		return false, nil
	}
	for t := in.tasks.head; t != nil; t = t.next {
		if t.ready() {
			if t != in.tasks.tail {
				if t.previous != nil {
					t.previous.next = t.next
				} else {
					in.tasks.head = t.next
				}
				t.next.previous = t.previous
				t.previous, t.next = in.tasks.tail, nil
				in.tasks.tail.next, in.tasks.tail = t, t
			}
			return true, t.resume(c)
		}
	}
	return false, nil
}

func (in *Instance) awaitTasks(c *vm.Caller, ready func() bool) error {
	for !ready() {
		if err := c.Context().Err(); err != nil {
			return err
		}
		progress, err := in.stepTasks(c)
		if err != nil {
			return err
		}
		if !progress {
			// No external producers are admitted in this profile. A deadlocked
			// guest waits without spinning until its session is cancelled.
			<-c.Context().Done()
			return c.Context().Err()
		}
	}
	return nil
}

func (in *Instance) finishTasks(c *vm.Caller, failed *error, complete *bool) {
	if in.tasks == nil {
		return
	}
	if !*complete && *failed == nil {
		*failed = wasm.ErrTrap
	}
	if *failed == nil {
		*failed = in.awaitTasks(c, func() bool { return in.tasks.live == 0 })
	}
	for in.tasks.head != nil {
		in.tasks.head.finish(c, wasm.ErrTrap)
	}
	if *failed != nil {
		for in.tasks.subtasks != nil {
			in.tasks.subtasks.abort(c)
		}
		for in.tasks.sets != nil {
			s := in.tasks.sets
			for s.head != nil {
				s.head.join(nil)
			}
			s.unlink()
			s.n.handles.RemoveOpaque(s.id)
		}
	}
	in.tasks = nil
}
