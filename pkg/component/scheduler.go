package component

import (
	"context"
	"errors"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type taskArguments func(*vm.Caller) ([]Value, error)
type taskResolution func(*vm.Caller, *Value, bool) error

// Component tasks retain explicit Core continuations and canonical transactions,
// never an expired Caller or a host Go stack. Callback and stackful ABIs share
// the same task, budget, cancellation and cleanup state machine.
type componentTask struct {
	in                               *Instance
	n                                *namespace
	task                             callTask
	entry                            vm.Extern
	args                             taskArguments
	onResolve                        taskResolution
	previous, next                   *componentTask
	waiting                          *waitableSet
	core                             *vm.Continuation
	hostFunction                     *Function
	host                             *hostCall
	work                             *workState
	started, running, exited         bool
	cancelRequested, cancelDelivered bool
	waitingEnter                     bool
	pendingCleanup                   func()
}

type taskScheduler struct {
	head, tail *componentTask
	live       uint32
	ends       *asyncEnd
	subtasks   *componentSubtask
	sets       *waitableSet
	wake       chan struct{}
}

func (in *Instance) scheduler() *taskScheduler {
	if in.tasks == nil {
		in.tasks = &taskScheduler{}
	}
	if in.tasks.wake == nil {
		in.tasks.wake = make(chan struct{}, 1)
	}
	return in.tasks
}

func (t *componentTask) needsExclusive() bool {
	return t.task.plan != nil && (!t.task.plan.Async || t.task.plan.Callback != nil)
}
func (t *componentTask) canEnter() bool {
	if t.n.backpressure != 0 {
		return false
	}
	if t.needsExclusive() {
		if t.n.exclusive != nil && t.n.exclusive != t {
			return false
		}
		if t.n.entered && (t.n.task == nil || t.n.task.runtime == nil) {
			return false
		}
	}
	return true
}

func (in *Instance) makeTask(c *vm.Caller, n *namespace, plan *validate.CanonicalFunction, entry vm.Extern, args taskArguments, resolve taskResolution) (*componentTask, error) {
	if n.failed || n.blocked {
		return nil, wasm.ErrTrap
	}
	s := in.scheduler()
	if s.live >= in.config.MaxTasks {
		return nil, wasm.ErrLimit
	}
	t := &componentTask{in: in, n: n, entry: entry, args: args, onResolve: resolve}
	t.task = callTask{plan: plan, runtime: t}
	if !t.canEnter() || n.waitingEnter != 0 {
		t.waitingEnter = true
		n.waitingEnter++
	}
	t.previous = s.tail
	if s.tail != nil {
		s.tail.next = t
	} else {
		s.head = t
	}
	s.tail = t
	s.live++
	return t, nil
}

func (in *Instance) startTask(c *vm.Caller, n *namespace, plan *validate.CanonicalFunction, entry vm.Extern, args taskArguments, resolve taskResolution) (*componentTask, error) {
	t, err := in.makeTask(c, n, plan, entry, args, resolve)
	if err != nil {
		return nil, err
	}
	if !t.waitingEnter && t.ready() {
		if err := t.resume(c); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (t *componentTask) ready() bool {
	if t.running || t.exited {
		return false
	}
	if !t.started {
		return t.cancelRequested || t.canEnter()
	}
	if t.needsExclusive() && t.n.exclusive != nil && t.n.exclusive != t {
		return false
	}
	if t.work != nil && t.work.control.requested.Load() && !t.cancelRequested {
		return true
	}
	if t.host != nil {
		return t.host.completion != nil && t.host.completion.ready()
	}
	if t.core != nil {
		return t.core.Ready()
	}
	return t.cancelRequested && !t.cancelDelivered || t.waiting == nil || t.waiting.ready != nil
}

func bindResourceCaller(codec *canonical.Codec, c *vm.Caller) {
	if codec != nil {
		if r, ok := codec.Resources.(*resourceCodec); ok {
			r.core = c
		}
	}
}
func (t *componentTask) resolve(c *vm.Caller, value *Value, cancelled bool) error {
	if t.onResolve != nil {
		return t.onResolve(c, value, cancelled)
	}
	return nil
}
func (t *componentTask) leaveWait() {
	if t.waiting != nil {
		t.waiting.waiters--
		t.waiting = nil
	}
}
func (t *componentTask) releaseExclusive() {
	if t.n.exclusive == t {
		t.n.exclusive = nil
	}
}

func (t *componentTask) finish(c *vm.Caller, failed error) error {
	if t.exited {
		return failed
	}
	t.leaveWait()
	t.releaseExclusive()
	pending := t.pendingCleanup
	t.pendingCleanup = nil
	if t.waitingEnter {
		t.n.waitingEnter--
		t.waitingEnter = false
	}
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
	core, host, returned, parameters := t.core, t.host, t.task.returned, t.task.parameters
	t.core, t.host, t.task.returned, t.task.parameters = nil, nil, nil, nil
	if failed != nil && t.started && t.task.plan != nil {
		t.n.failed = true
	}
	if host != nil && host.cancel != nil {
		host.cancel()
	}
	guard := cleanupGuard{}
	if core != nil {
		guard.run(func() { c.Abort(core) })
	}
	if pending != nil {
		guard.run(pending)
	}
	if host != nil {
		guard.run(func() { failed = host.finish(c, failed == nil, failed) })
	}
	for _, codec := range []*canonical.Codec{returned, parameters} {
		if codec != nil {
			guard.run(func() { bindResourceCaller(codec, c); failed = finishResources(codec, failed == nil, failed) })
		}
	}
	if (failed != nil || guard.panicked != nil) && t.started && t.task.plan != nil {
		t.n.failed = true
	}
	if t.work != nil {
		state := t.work
		t.work = nil
		guard.run(func() { state.finish(c, t.in) })
	}
	guard.finish()
	return failed
}

func (t *componentTask) requestCancel(c *vm.Caller) error {
	t.cancelRequested = true
	if !t.started {
		return t.resume(c)
	}
	if t.host != nil {
		t.host.cancel()
		if t.host.completion != nil && t.host.completion.ready() {
			return t.resume(c)
		}
		return nil
	}
	if !t.started || t.task.plan != nil && t.task.plan.Callback != nil && t.core == nil && t.ready() {
		return t.resume(c)
	}
	return nil
}

func (t *componentTask) resume(c *vm.Caller) (err error) {
	t.syncCancellation()
	t.running = true
	previous, entered := t.n.task, t.n.entered
	t.n.task, t.n.entered = &t.task, true
	complete := false
	defer func() {
		panicked := recover()
		t.running = false
		t.n.task, t.n.entered = previous, entered
		if !complete {
			guard := cleanupGuard{panicked: panicked}
			guard.run(func() { err = t.finish(c, errors.Join(wasm.ErrTrap, err)) })
			guard.finish()
		} else if panicked != nil {
			panic(panicked)
		}
	}()
	if !t.started {
		if t.waitingEnter {
			t.waitingEnter = false
			t.n.waitingEnter--
		}
		if t.cancelRequested {
			t.task.resolved = true
			if err = t.resolve(c, nil, true); err != nil {
				return err
			}
			complete = true
			return t.finish(c, nil)
		}
		if t.needsExclusive() {
			t.n.exclusive = t
		}
	}
	if t.hostFunction != nil {
		if err = t.resumeHost(c); err != nil {
			return err
		}
		complete = true
		if t.task.resolved {
			return t.finish(c, nil)
		}
		return nil
	}
	plan := t.task.plan
	if !t.started {
		args, failed := t.args(c)
		if failed != nil {
			return failed
		}
		t.args = nil
		t.started = true
		codec := t.in.codec(c, t.n, plan)
		t.task.parameters = codec
		for i, p := range plan.Type.Parameters {
			if err = codec.Validate(p.Type, args[i]); err != nil {
				return err
			}
		}
		if codec.Resources != nil {
			if err = codec.PrepareParameters(plan.Type.Parameters, args); err != nil {
				return err
			}
		}
		slots, failed := lowerParameters(codec, plan, args)
		if failed != nil {
			return failed
		}
		t.core, err = c.Start(t.entry, vmValues(slots))
		if err != nil {
			return err
		}
	} else if t.core == nil {
		if t.needsExclusive() {
			t.n.exclusive = t
		}
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
		t.core, err = c.Start(t.n.items[model.CoreFunctionSort][*plan.Callback].core, []vm.Value{vm.I32(event[0]), vm.I32(event[1]), vm.I32(event[2])})
		if err != nil {
			return err
		}
	}
	called, done, err := c.Poll(t.core)
	if err != nil {
		return err
	}
	if !done {
		complete = true
		return nil
	}
	t.core = nil
	if !plan.Async {
		codec := t.task.parameters
		codec.Transfer = t.in.codec(c, t.n, plan).Transfer
		bindResourceCaller(codec, c)
		t.task.result, err = liftResult(codec, plan, called.Values)
		if err != nil {
			return err
		}
		if err = returnBorrows(codec); err != nil {
			return err
		}
		t.task.resolved = true
		if err = t.resolve(c, t.task.result, false); err != nil {
			return err
		}
		if plan.PostReturn != nil {
			t.n.blocked = true
			_, err = c.Call(t.n.items[model.CoreFunctionSort][*plan.PostReturn].core, called.Values)
			t.n.blocked = false
			if err != nil {
				return err
			}
		}
		complete = true
		return t.finish(c, nil)
	}
	if plan.Callback == nil {
		if !t.task.resolved {
			return wasm.ErrTrap
		}
		complete = true
		return t.finish(c, nil)
	}
	t.releaseExclusive()
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

func liftResult(codec *canonical.Codec, plan *validate.CanonicalFunction, values []vm.Value) (*Value, error) {
	if plan.Type.Result == nil {
		return nil, nil
	}
	var v Value
	var err error
	if plan.ABI.ResultMemory != nil {
		ptr, _, _ := values[0].Bits()
		v, err = codec.Load(plan.Type.Result, ptr)
	} else {
		var slots []canonical.Slot
		slots, err = flatValues(values)
		if err == nil {
			v, err = codec.LiftFlat(plan.Type.Result, slots)
		}
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
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
	return in.awaitTasksContext(c, c.Context(), ready)
}
func (in *Instance) awaitTasksContext(c *vm.Caller, ctx context.Context, ready func() bool) error {
	for !ready() {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := c.Context().Err(); err != nil {
			return err
		}
		progress, err := in.stepTasks(c)
		if err != nil {
			return err
		}
		if !progress {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c.Context().Done():
				return c.Context().Err()
			case <-in.scheduler().wake:
			}
		}
	}
	return nil
}
func (in *Instance) finishTasks(c *vm.Caller, failed *error, complete *bool) {
	guard := cleanupGuard{panicked: recover()}
	if in.tasks == nil {
		guard.finish()
		return
	}
	if (!*complete || guard.panicked != nil) && *failed == nil {
		*failed = wasm.ErrTrap
	}
	if *failed == nil {
		guard.run(func() { *failed = in.awaitTasks(c, func() bool { return in.tasks.live == 0 }) })
	}
	if guard.panicked != nil && *failed == nil {
		*failed = wasm.ErrTrap
	}
	for in.tasks.head != nil {
		task := in.tasks.head
		guard.run(func() { task.finish(c, wasm.ErrTrap) })
	}
	for in.tasks.ends != nil {
		end := in.tasks.ends
		if *failed != nil {
			guard.run(end.abort)
		} else {
			end.unlink()
		}
	}
	if *failed != nil {
		for in.tasks.subtasks != nil {
			sub := in.tasks.subtasks
			guard.run(func() { sub.abort(c) })
		}
		for in.tasks.sets != nil {
			set := in.tasks.sets
			for set.head != nil {
				set.head.join(nil)
			}
			set.unlink()
			set.n.handles.RemoveOpaque(set.id)
		}
	}
	in.tasks = nil
	guard.finish()
}
