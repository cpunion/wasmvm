package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const (
	endIdle       = 1
	endCopying    = 2
	endCancelling = 3
	endDone       = 4
)

// One state machine serves all readable/writable future/stream endpoints.
// State transitions occur on event delivery, preserving buffer ownership until
// the consumer actually observes completion, cancellation or peer closure.
type asyncEnd struct {
	in             *Instance
	typ            *model.Type
	scope, n       *namespace
	id             uint32
	read           bool
	state          uint8
	other          *asyncEnd
	buffer         *guestBuffer
	waitable       waitable
	progress       uint32
	group          *taskScheduler
	previous, next *asyncEnd
}

func (e *asyncEnd) event(*vm.Caller) ([3]uint32, error) {
	code := uint32(2)
	if !e.read {
		code++
	}
	if e.typ.Kind == model.Future {
		code += 2
	}
	result := uint32(0)
	if e.typ.Kind == model.Future && e.progress == 1 {
		e.state = endDone
		e.buffer = nil
	} else if e.other == nil {
		result = 1
		e.state = endDone
		e.buffer = nil
	} else if e.state == endCancelling {
		result = 2
		e.state = endIdle
		e.buffer = nil
	} else if e.typ.Kind == model.Stream && e.state == endCopying {
		e.state = endIdle
		e.buffer = nil
	} else {
		return [3]uint32{}, wasm.ErrTrap
	}
	if e.typ.Kind == model.Stream {
		result |= e.progress << 4
	}
	return [3]uint32{code, e.id, result}, nil
}
func (e *asyncEnd) notify(progress uint32) { e.progress = progress; e.waitable.notify() }
func (e *asyncEnd) unlink() {
	if e.group == nil {
		return
	}
	if e.previous != nil {
		e.previous.next = e.next
	} else {
		e.group.ends = e.next
	}
	if e.next != nil {
		e.next.previous = e.previous
	}
	e.group, e.previous, e.next = nil, nil, nil
}
func (e *asyncEnd) detach() {
	e.waitable.join(nil)
	e.unlink()
	if e.other != nil {
		other := e.other
		other.other = nil
		e.other = nil
		if other.state != endDone && !other.waitable.pending {
			other.notify(0)
		}
	}
	e.buffer = nil
}
func (e *asyncEnd) abort() {
	e.detach()
	e.table().RemoveOpaque(e.id)
}
func (in *Instance) waitable(n *namespace, id uint32) (*waitable, error) {
	if n.handles == nil {
		return nil, wasm.ErrTrap
	}
	v, err := n.handles.Opaque(id)
	if err != nil {
		return nil, err
	}
	switch x := v.(type) {
	case *componentSubtask:
		return &x.waitable, nil
	case *asyncEnd:
		return &x.waitable, nil
	}
	return nil, wasm.ErrTrap
}
func (in *Instance) end(n *namespace, id uint32, t *model.Type, read bool) (*asyncEnd, error) {
	if n.handles == nil {
		return nil, wasm.ErrTrap
	}
	v, err := n.handles.Opaque(id)
	if err != nil {
		return nil, err
	}
	e, ok := v.(*asyncEnd)
	if !ok || e.n != n || e.id != id || e.read != read {
		return nil, wasm.ErrTrap
	}
	equal, err := validate.EqualValue(e.typ, t, in.config.Validate.MaxNodes, func(a, b uint64) bool {
		left, right := e.scope.resources[a], n.resources[b]
		return left != nil && right != nil && left.identity == right.identity
	})
	if err != nil {
		return nil, err
	}
	if !equal {
		return nil, wasm.ErrTrap
	}
	return e, nil
}
func (in *Instance) newEnds(n *namespace, t *model.Type) (*asyncEnd, *asyncEnd, error) {
	in.resourceTable(n)
	r := &asyncEnd{in: in, typ: t, scope: n, n: n, read: true, state: endIdle}
	w := &asyncEnd{in: in, typ: t, scope: n, n: n, state: endIdle}
	r.other, w.other = w, r
	r.waitable.producer = r
	w.waitable.producer = w
	for _, e := range []*asyncEnd{r, w} {
		id, err := n.handles.AddOpaque(e)
		if err != nil {
			if r.id != 0 {
				n.handles.RemoveOpaque(r.id)
			}
			return nil, nil, err
		}
		e.id = id
	}
	s := in.scheduler()
	for _, e := range []*asyncEnd{r, w} {
		e.group = s
		e.next = s.ends
		if e.next != nil {
			e.next.previous = e
		}
		s.ends = e
	}
	return r, w, nil
}
func (e *asyncEnd) copy(c *vm.Caller, b *guestBuffer) error {
	e.state = endCopying
	other := e.other
	if other == nil {
		if !e.waitable.pending {
			return wasm.ErrTrap
		}
		return nil
	}
	if other.buffer == nil {
		e.buffer = b
		return nil
	}
	ob := other.buffer
	if b.remain() > 0 && ob.remain() > 0 {
		if e.n != nil && e.n == other.n && !numberOrUnit(e.typ.Element) {
			return wasm.ErrTrap
		}
		count := min(b.remain(), ob.remain())
		src, dst := b, ob
		if e.read {
			src, dst = ob, b
		}
		if err := copyGuestBuffers(c, src, dst, count); err != nil {
			return err
		}
		e.notify(b.progress)
		other.notify(ob.progress)
		if ob.remain() == 0 {
			other.buffer = nil
		}
	} else if b.remain() > 0 || e.read && ob.remain() == 0 {
		other.notify(0)
		other.buffer = nil
		e.buffer = b
	} else {
		e.notify(0)
	}
	return nil
}
func numberOrUnit(t *model.Type) bool { return t == nil || t.Kind >= model.F64 && t.Kind <= model.S8 }

func (e *asyncEnd) table() *canonical.Resources {
	if e.n == nil {
		return e.in.hostResources
	}
	return e.n.handles
}
