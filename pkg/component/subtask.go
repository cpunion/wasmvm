package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type componentSubtask struct {
	waitable              waitable
	task                  *componentTask
	n                     *namespace
	codec                 *canonical.Codec
	id, state             uint32
	values                []vm.Value
	delivered             bool
	cancellationRequested bool
	group                 *taskScheduler
	previous, next        *componentSubtask
}

func (s *componentSubtask) unlink() {
	if s.group == nil {
		return
	}
	if s.previous != nil {
		s.previous.next = s.next
	} else {
		s.group.subtasks = s.next
	}
	if s.next != nil {
		s.next.previous = s.previous
	}
	s.group, s.previous, s.next = nil, nil, nil
}

func (s *componentSubtask) event(c *vm.Caller) ([3]uint32, error) {
	if s.state >= 2 && !s.delivered {
		s.delivered = true
		s.unlink()
		if s.codec != nil {
			bindResourceCaller(s.codec, c)
			if err := finishResources(s.codec, true, nil); err != nil {
				return [3]uint32{}, err
			}
			s.codec = nil
		}
	}
	return [3]uint32{1, s.id, s.state}, nil
}

func (s *componentSubtask) abort(c *vm.Caller) {
	s.unlink()
	s.waitable.join(nil)
	if s.codec != nil {
		bindResourceCaller(s.codec, c)
		finishResources(s.codec, false, wasm.ErrTrap)
		s.codec = nil
	}
	if s.id != 0 {
		s.n.handles.RemoveOpaque(s.id)
	}
}

func (in *Instance) lowerTask(c *vm.Caller, n *namespace, plan *validate.CanonicalFunction, flat []vm.Value, callee *Function) (result vm.HostResult, err error) {
	s := &componentSubtask{n: n}
	s.waitable.producer = s
	complete := false
	defer func() {
		if !complete {
			n.failed = true
			if s.codec != nil {
				bindResourceCaller(s.codec, c)
				err = finishResources(s.codec, false, err)
				s.codec = nil
			}
		}
	}()
	t, err := callee.start(c, func(current *vm.Caller) ([]Value, error) {
		s.codec = in.codec(current, n, plan)
		args, err := liftArguments(s.codec, plan, flat)
		if err == nil {
			s.state = 1
			if s.id != 0 {
				s.waitable.notify()
			}
		}
		return args, err
	}, func(current *vm.Caller, value *Value, cancelled bool) error {
		if cancelled {
			if s.state == 0 {
				s.state = 3
			} else {
				s.state = 4
			}
		} else {
			s.codec.Transfer = in.codec(current, n, plan).Transfer
			bindResourceCaller(s.codec, current)
			var err error
			s.values, err = lowerResult(s.codec, plan, flat, value)
			if err != nil {
				return err
			}
			s.state = 2
		}
		if s.id != 0 {
			s.waitable.notify()
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	s.task = t
	finish := func(current *vm.Caller) (vm.HostResult, error) {
		_, err := s.event(current)
		if err != nil {
			n.failed = true
			return vm.HostResult{}, err
		}
		values := s.values
		if plan.Async {
			values = []vm.Value{vm.I32(2)}
		}
		return vm.HostResult{Values: values}, nil
	}
	if s.state >= 2 {
		result, err = finish(c)
		complete = err == nil
		return result, err
	}
	if n.task != nil && n.task.plan != nil && n.task.plan.Type.Kind != model.AsyncFunction {
		return result, wasm.ErrTrap
	}
	if plan.Async {
		in.resourceTable(n)
		s.id, err = n.handles.AddOpaque(s)
		if err != nil {
			return result, err
		}
	} else if n.task == nil || n.task.runtime == nil {
		return result, wasm.ErrTrap
	}
	s.group = in.scheduler()
	s.next = s.group.subtasks
	if s.next != nil {
		s.next.previous = s
	}
	s.group.subtasks = s
	complete = true
	if plan.Async {
		return vm.HostResult{Values: []vm.Value{vm.I32(s.state | s.id<<4)}}, nil
	}
	return vm.HostResult{Pending: &vm.HostSuspension{Ready: func() bool { return s.state >= 2 }, Resume: finish}}, nil
}

func liftArguments(codec *canonical.Codec, plan *validate.CanonicalFunction, flat []vm.Value) ([]Value, error) {
	if plan.ABI.ParameterMemory != nil {
		ptr, _, _ := flat[0].Bits()
		v, err := codec.Load(plan.ABI.ParameterMemory, ptr)
		return v.Elements, err
	}
	slots, err := flatValues(flat)
	if err != nil {
		return nil, err
	}
	var args []Value
	for _, p := range plan.Type.Parameters {
		layout, err := codec.Layouts.Build(p.Type)
		if err != nil {
			return nil, err
		}
		v, err := codec.LiftFlat(p.Type, slots[:layout.FlatCount])
		if err != nil {
			return nil, err
		}
		args = append(args, v)
		slots = slots[layout.FlatCount:]
	}
	return args, nil
}

func lowerResult(codec *canonical.Codec, plan *validate.CanonicalFunction, flat []vm.Value, v *Value) ([]vm.Value, error) {
	if v == nil {
		return nil, nil
	}
	if codec.Resources != nil {
		if err := codec.AcceptResult(plan.Type.Result, v); err != nil {
			return nil, err
		}
	}
	if plan.ABI.ResultMemory != nil {
		ptr, _, _ := flat[len(flat)-1].Bits()
		return nil, codec.Store(plan.Type.Result, ptr, *v)
	}
	slots, err := codec.LowerFlat(plan.Type.Result, *v)
	return vmValues(slots), err
}
