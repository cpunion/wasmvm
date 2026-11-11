package component

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (in *Instance) asyncValueBuiltin(caller *vm.Caller, n *namespace, p *validate.CanonicalFunction) error {
	in.resourceTable(n)
	op := p.Definition.Op
	if op >= 0x15 && op <= 0x1b {
		op -= 7
	}
	fn, err := newScalarBuiltin(caller, p.Signature, !p.Async && op != 0x0e && op != 0x13 && op != 0x14 && op != 0x2e && op != 0x2f, func(c *vm.Caller, args scalarOperands) (scalarStep, error) {
		if n.failed || n.blocked {
			return scalarStep{}, wasm.ErrTrap
		}
		if op == 0x0e {
			r, w, err := in.newEnds(n, p.Type.Value)
			if err != nil {
				return scalarStep{}, err
			}
			return scalarStep{bits: uint64(r.id) | uint64(w.id)<<32}, nil
		}
		id := args.at(0)
		read := op == 0x0f || op == 0x11 || op == 0x13 || op == 0x2e || op == 0x2f
		e, err := in.end(n, uint32(id), p.Type.Value, read)
		if err != nil {
			return scalarStep{}, err
		}
		if op == 0x2e || op == 0x2f {
			wid := args.at(1)
			w, err := in.end(n, uint32(wid), p.Type.Value, false)
			if err != nil {
				return scalarStep{}, err
			}
			if e.state != endIdle || w.state != endIdle || e.waitable.set != nil || w.waitable.set != nil {
				return scalarStep{}, wasm.ErrTrap
			}
			if err := forwardEnds(c, e, w); err != nil {
				return scalarStep{}, err
			}
			return scalarStep{}, nil
		}
		if op == 0x13 || op == 0x14 {
			if e.state == endCopying || e.state == endCancelling || !e.read && e.typ.Kind == model.Future && e.state != endDone {
				return scalarStep{}, wasm.ErrTrap
			}
			e.detach()
			return scalarStep{}, n.handles.RemoveOpaque(e.id)
		}
		if !p.Async && e.waitable.set != nil {
			return scalarStep{}, wasm.ErrTrap
		}
		if op == 0x11 || op == 0x12 {
			if e.state != endCopying || e.waitable.syncWaiter {
				return scalarStep{}, wasm.ErrTrap
			}
			e.state = endCancelling
			if e.sourceJob != nil {
				if err := e.sourceJob.requestCancel(c); err != nil {
					return scalarStep{}, err
				}
			} else if !e.waitable.pending {
				e.notify(0)
			}
		} else {
			if e.state != endIdle {
				return scalarStep{}, wasm.ErrTrap
			}
			ptr := args.at(1)
			length := uint64(1)
			if e.typ.Kind == model.Stream {
				length = args.at(2)
			}
			b, err := in.guestBuffer(c, n, p, ptr, length)
			if err != nil {
				return scalarStep{}, err
			}
			if err := e.copy(c, b); err != nil {
				return scalarStep{}, err
			}
		}
		result := func(current *vm.Caller) (uint64, error) {
			e.waitable.syncWaiter = false
			event, err := e.waitable.poll(current)
			return uint64(event[2]), err
		}
		if e.waitable.pending {
			bits, err := result(c)
			return scalarStep{bits: bits}, err
		}
		if p.Async {
			bits := uint64(^uint32(0))
			if e.typ.Kind == model.Stream && p.Memory64 {
				bits = ^uint64(0)
			}
			return scalarStep{bits: bits}, nil
		}
		if n.task == nil || n.task.runtime == nil || n.task.plan.Type.Kind != model.AsyncFunction {
			return scalarStep{}, wasm.ErrTrap
		}
		task := n.task.runtime
		e.waitable.syncWaiter = true
		task.pendingCleanup = func() { e.waitable.syncWaiter = false }
		return scalarStep{pending: &vm.HostSuspension{Ready: func() bool { return e.waitable.pending }, Resume: func(current *vm.Caller) (vm.HostResult, error) {
			task.pendingCleanup = nil
			bits, err := result(current)
			return scalarHostResult(p.Signature, scalarStep{bits: bits}), err
		}}}, nil
	})
	if err == nil {
		n.add(item{sort: model.CoreFunctionSort, core: fn})
	}
	return err
}
func forwardEnds(c *vm.Caller, src, dst *asyncEnd) error {
	if src.other == dst || src.other == nil || dst.other == nil {
		src.detach()
		dst.detach()
	} else {
		writer, reader := src.other, dst.other
		writer.other, reader.other = reader, writer
		src.other, dst.other = nil, nil
		if writer.source != nil && reader.buffer != nil && reader.sourceJob == nil {
			if err := writer.in.startSourceRead(c, reader, writer, reader.buffer); err != nil {
				return err
			}
		}
		if writer.buffer != nil && reader.buffer != nil {
			smaller := reader
			if reader.buffer.remain() > writer.buffer.remain() {
				smaller = writer
			}
			buffer := smaller.buffer
			smaller.buffer = nil
			if err := smaller.copy(c, buffer); err != nil {
				return err
			}
		}
		src.detach()
		dst.detach()
	}
	src.n.handles.RemoveOpaque(src.id)
	return dst.n.handles.RemoveOpaque(dst.id)
}
