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
	fn, err := caller.NewSuspendableFunction(p.Signature, func(c *vm.Caller, args []vm.Value) (vm.HostResult, error) {
		if n.failed || n.blocked {
			return vm.HostResult{}, wasm.ErrTrap
		}
		if op == 0x0e {
			r, w, err := in.newEnds(n, p.Type.Value)
			if err != nil {
				return vm.HostResult{}, err
			}
			return vm.HostResult{Values: []vm.Value{vm.I64(uint64(r.id) | uint64(w.id)<<32)}}, nil
		}
		id, _, _ := args[0].Bits()
		read := op == 0x0f || op == 0x11 || op == 0x13 || op == 0x2e || op == 0x2f
		e, err := in.end(n, uint32(id), p.Type.Value, read)
		if err != nil {
			return vm.HostResult{}, err
		}
		if op == 0x2e || op == 0x2f {
			wid, _, _ := args[1].Bits()
			w, err := in.end(n, uint32(wid), p.Type.Value, false)
			if err != nil {
				return vm.HostResult{}, err
			}
			if e.state != endIdle || w.state != endIdle || e.waitable.set != nil || w.waitable.set != nil {
				return vm.HostResult{}, wasm.ErrTrap
			}
			if err := forwardEnds(c, e, w); err != nil {
				return vm.HostResult{}, err
			}
			return vm.HostResult{}, nil
		}
		if op == 0x13 || op == 0x14 {
			if e.state == endCopying || e.state == endCancelling || !e.read && e.typ.Kind == model.Future && e.state != endDone {
				return vm.HostResult{}, wasm.ErrTrap
			}
			e.detach()
			return vm.HostResult{}, n.handles.RemoveOpaque(e.id)
		}
		if !p.Async && e.waitable.set != nil {
			return vm.HostResult{}, wasm.ErrTrap
		}
		if op == 0x11 || op == 0x12 {
			if e.state != endCopying || e.waitable.syncWaiter {
				return vm.HostResult{}, wasm.ErrTrap
			}
			e.state = endCancelling
			if e.sourceJob != nil {
				if err := e.sourceJob.requestCancel(c); err != nil {
					return vm.HostResult{}, err
				}
			} else if !e.waitable.pending {
				e.notify(0)
			}
		} else {
			if e.state != endIdle {
				return vm.HostResult{}, wasm.ErrTrap
			}
			ptr, _, _ := args[1].Bits()
			length := uint64(1)
			if e.typ.Kind == model.Stream {
				length, _, _ = args[2].Bits()
			}
			b, err := in.guestBuffer(c, n, p, ptr, length)
			if err != nil {
				return vm.HostResult{}, err
			}
			if err := e.copy(c, b); err != nil {
				return vm.HostResult{}, err
			}
		}
		result := func(current *vm.Caller) (vm.HostResult, error) {
			e.waitable.syncWaiter = false
			event, err := e.waitable.poll(current)
			if err != nil {
				return vm.HostResult{}, err
			}
			value := vm.I32(event[2])
			if e.typ.Kind == model.Stream && p.Memory64 {
				value = vm.I64(uint64(event[2]))
			}
			return vm.HostResult{Values: []vm.Value{value}}, nil
		}
		if e.waitable.pending {
			return result(c)
		}
		if p.Async {
			value := vm.I32(^uint32(0))
			if e.typ.Kind == model.Stream && p.Memory64 {
				value = vm.I64(^uint64(0))
			}
			return vm.HostResult{Values: []vm.Value{value}}, nil
		}
		if n.task == nil || n.task.runtime == nil || n.task.plan.Type.Kind != model.AsyncFunction {
			return vm.HostResult{}, wasm.ErrTrap
		}
		task := n.task.runtime
		e.waitable.syncWaiter = true
		task.pendingCleanup = func() { e.waitable.syncWaiter = false }
		return vm.HostResult{Pending: &vm.HostSuspension{Ready: func() bool { return e.waitable.pending }, Resume: func(current *vm.Caller) (vm.HostResult, error) { task.pendingCleanup = nil; return result(current) }}}, nil
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
