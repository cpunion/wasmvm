package component

import (
	"encoding/binary"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type eventProducer interface {
	event(*vm.Caller) ([3]uint32, error)
}

// Membership and ready queues are intrusive: joining, notifying and polling
// take constant time and a producer occupies at most one slot in each queue.
type waitable struct {
	set                      *waitableSet
	previous, next           *waitable
	readyPrevious, readyNext *waitable
	pending, syncWaiter      bool
	producer                 eventProducer
}

type waitableSet struct {
	head, ready, tail *waitable
	waiters           uint32
	n                 *namespace
	id                uint32
	group             *taskScheduler
	previous, next    *waitableSet
}

func (s *waitableSet) unlink() {
	if s.group == nil {
		return
	}
	if s.previous != nil {
		s.previous.next = s.next
	} else {
		s.group.sets = s.next
	}
	if s.next != nil {
		s.next.previous = s.previous
	}
	s.previous, s.next, s.group = nil, nil, nil
}

func (w *waitable) unlinkReady() {
	s := w.set
	if w.readyPrevious != nil {
		w.readyPrevious.readyNext = w.readyNext
	} else {
		s.ready = w.readyNext
	}
	if w.readyNext != nil {
		w.readyNext.readyPrevious = w.readyPrevious
	} else {
		s.tail = w.readyPrevious
	}
	w.readyPrevious, w.readyNext = nil, nil
}

func (w *waitable) notify() {
	if w.pending {
		return
	}
	w.pending = true
	if w.set != nil {
		w.enqueue()
	}
}

func (w *waitable) enqueue() {
	s := w.set
	w.readyPrevious = s.tail
	if s.tail != nil {
		s.tail.readyNext = w
	} else {
		s.ready = w
	}
	s.tail = w
}

func (w *waitable) join(s *waitableSet) {
	if old := w.set; old != nil {
		if w.pending {
			w.unlinkReady()
		}
		if w.previous != nil {
			w.previous.next = w.next
		} else {
			old.head = w.next
		}
		if w.next != nil {
			w.next.previous = w.previous
		}
	}
	w.previous, w.next, w.set = nil, nil, s
	if s != nil {
		w.next = s.head
		if s.head != nil {
			s.head.previous = w
		}
		s.head = w
		if w.pending {
			w.enqueue()
		}
	}
}

func (w *waitable) poll(c *vm.Caller) ([3]uint32, error) {
	if w.set != nil {
		w.unlinkReady()
	}
	w.pending = false
	return w.producer.event(c)
}

func (s *waitableSet) poll(c *vm.Caller) ([3]uint32, error) {
	if s.ready == nil {
		return [3]uint32{}, nil
	}
	return s.ready.poll(c)
}

func (in *Instance) waitableSet(n *namespace, id uint32) (*waitableSet, error) {
	if n.handles == nil {
		return nil, wasm.ErrTrap
	}
	v, err := n.handles.Opaque(id)
	if err != nil {
		return nil, err
	}
	s, ok := v.(*waitableSet)
	if !ok {
		return nil, wasm.ErrTrap
	}
	return s, nil
}

func (in *Instance) subtask(n *namespace, id uint32) (*componentSubtask, error) {
	if n.handles == nil {
		return nil, wasm.ErrTrap
	}
	v, err := n.handles.Opaque(id)
	if err != nil {
		return nil, err
	}
	s, ok := v.(*componentSubtask)
	if !ok {
		return nil, wasm.ErrTrap
	}
	return s, nil
}

func (in *Instance) waitableBuiltin(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) error {
	in.resourceTable(n)
	n.taskOps = true
	fn, err := caller.NewFunction(plan.Signature, func(c *vm.Caller, args []vm.Value) ([]vm.Value, error) {
		if n.failed || n.blocked {
			return nil, wasm.ErrTrap
		}
		op := plan.Definition.Op
		if op == 5 {
			t := n.task
			if t == nil || t.callback == nil || t.resolved || !t.callback.cancelDelivered {
				return nil, wasm.ErrTrap
			}
			if err := returnBorrows(t.parameters); err != nil {
				return nil, err
			}
			t.resolved = true
			return nil, t.callback.resolve(c, nil, true)
		}
		if op == 0x1f {
			s := &waitableSet{n: n}
			id, err := n.handles.AddOpaque(s)
			if err != nil {
				return nil, err
			}
			if in.tasks == nil {
				in.tasks = &taskScheduler{}
			}
			s.id, s.group, s.next = id, in.tasks, in.tasks.sets
			if s.next != nil {
				s.next.previous = s
			}
			in.tasks.sets = s
			return []vm.Value{vm.I32(id)}, nil
		}
		bits, _, _ := args[0].Bits()
		id := uint32(bits)
		if op == 0x20 || op == 0x21 || op == 0x22 {
			s, err := in.waitableSet(n, id)
			if err != nil {
				return nil, err
			}
			if op == 0x22 {
				if s.head != nil || s.waiters != 0 {
					return nil, wasm.ErrTrap
				}
				s.unlink()
				return nil, n.handles.RemoveOpaque(id)
			}
			if op == 0x20 {
				if s.ready == nil && n.task != nil && n.task.plan != nil && n.task.plan.Type.Kind != model.AsyncFunction {
					return nil, wasm.ErrTrap
				}
				s.waiters++
				err = in.awaitTasks(c, func() bool { return s.ready != nil })
				s.waiters--
				if err != nil {
					return nil, err
				}
			}
			event, err := s.poll(c)
			if err != nil {
				return nil, err
			}
			ptr, _, _ := args[1].Bits()
			if ptr&3 != 0 {
				return nil, wasm.ErrTrap
			}
			var bytes [8]byte
			binary.LittleEndian.PutUint32(bytes[:4], event[1])
			binary.LittleEndian.PutUint32(bytes[4:], event[2])
			err = c.WriteExternalMemory(n.items[model.CoreMemorySort][*plan.Memory].core, ptr, bytes[:])
			if err != nil {
				return nil, err
			}
			return []vm.Value{vm.I32(event[0])}, nil
		}
		s, err := in.subtask(n, id)
		if err != nil {
			return nil, err
		}
		switch op {
		case 0x23:
			if s.waitable.syncWaiter {
				return nil, wasm.ErrTrap
			}
			var set *waitableSet
			other, _, _ := args[1].Bits()
			if other != 0 {
				set, err = in.waitableSet(n, uint32(other))
				if err != nil {
					return nil, err
				}
			}
			s.waitable.join(set)
			return nil, nil
		case 0x0d:
			if !s.delivered || s.waitable.syncWaiter {
				return nil, wasm.ErrTrap
			}
			s.waitable.join(nil)
			return nil, n.handles.RemoveOpaque(id)
		default:
			if s.delivered || s.cancellationRequested || s.waitable.set != nil {
				return nil, wasm.ErrTrap
			}
			if s.state < 2 {
				s.cancellationRequested, s.waitable.syncWaiter = true, true
				defer func() { s.waitable.syncWaiter = false }()
				s.task.cancelRequested = true
				if s.task.ready() {
					if err := s.task.resume(c); err != nil {
						return nil, err
					}
				}
				if !plan.Async {
					if s.state < 2 && n.task != nil && n.task.plan != nil && n.task.plan.Type.Kind != model.AsyncFunction {
						return nil, wasm.ErrTrap
					}
					if err := in.awaitTasks(c, func() bool { return s.state >= 2 }); err != nil {
						return nil, err
					}
				}
				if s.state < 2 {
					return []vm.Value{vm.I32(^uint32(0))}, nil
				}
			}
			_, err := s.waitable.poll(c)
			if err != nil {
				return nil, err
			}
			return []vm.Value{vm.I32(s.state)}, nil
		}
	})
	if err == nil {
		n.add(item{sort: model.CoreFunctionSort, core: fn})
	}
	return err
}
