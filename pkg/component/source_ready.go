package component

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

var errSourceNotReady = errors.New("component source probe is not ready")

// A prompt typed transfer needs the ordinary callback/ownership transaction,
// but no external-work task, completion or cancellation context. Byte probes
// retain their smaller resource-free path.
func (in *Instance) readReadySource(core *vm.Caller, r, w *asyncEnd, b *guestBuffer, batch *model.Type) (handled bool, err error) {
	source := w.source
	if source.source.ReadReady == nil {
		return false, nil
	}
	if err := core.Context().Err(); err != nil {
		return true, err
	}
	if w.scope.failed || w.scope.blocked {
		return true, wasm.ErrTrap
	}
	scheduler := in.scheduler()
	if scheduler.live >= in.config.MaxTasks || in.callDepth >= in.config.Store.MaxReentry {
		return true, wasm.ErrLimit
	}
	if source.closed.Load() {
		return true, wasm.ErrClosed
	}
	// The callback may schedule other work. Reserve the same logical task slot
	// until its transfer/cleanup ends, without allocating a scheduler node.
	scheduler.live++
	defer func() { scheduler.live-- }()
	plan := source.ensureReadPlan(batch)
	frame := in.acquireFrame()
	defer in.releaseCanonicalFrame(frame)
	frame.readFn = Function{instance: in, scope: w.scope, typ: &plan.typ, values: &plan.values}
	h, err := frame.readFn.hostParameters(core, nil, false, frame.hostStorage())
	if err != nil {
		return true, err
	}
	h.prompt = true
	complete := false
	defer h.finishDeferred(core, &complete, &err)
	handled = true
	var output *canonical.Codec
	defer func() {
		guard := cleanupGuard{panicked: recover()}
		if output != nil {
			err = guard.finishCodec(output, complete, err)
		}
		if handled && !complete {
			guard.run(func() { r.buffer, r.state = nil, endDone; w.closeSource() })
		}
		guard.finish()
	}()
	var meta ReadResult
	var ready bool
	_, err = h.invoke(core, func(c *Caller) (*Value, error) {
		var err error
		meta, ready, err = source.readReady(c, b.remain())
		return nil, err
	})
	if err != nil {
		return true, err
	}
	if !ready {
		if meta.Values != nil || meta.Bytes != "" || meta.Count != 0 || meta.Closed {
			return true, vm.ErrType
		}
		// Roll back any speculative creations before starting the fallback.
		err = h.finish(core, false, errSourceNotReady)
		if err == errSourceNotReady {
			err = nil
		}
		return false, err
	}
	unit := r.typ.Element == nil
	count := uint64(len(meta.Values))
	encoded := Value{Elements: meta.Values, Text: meta.Bytes}
	if unit {
		count = uint64(meta.Count)
		encoded = Value{Bits: count}
	} else if r.typ.Element.Kind == model.U8 && meta.Values == nil {
		count = uint64(len(meta.Bytes))
	}
	if meta.Values != nil && meta.Bytes != "" || count > uint64(b.remain()) || meta.Count != 0 && uint64(meta.Count) != count || unit && (len(meta.Values) != 0 || meta.Bytes != "") || r.typ.Kind == model.Future && count != 1 || count == 0 && !meta.Closed && b.remain() != 0 {
		return true, vm.ErrType
	}
	frame.cloneBuf = frame.cloneValues[:0]
	if err = h.acceptIntoScratch(&encoded, &frame.hostValue, &frame.cloneBuf); err != nil {
		return true, err
	}
	if !unit && count != 0 {
		output, err = b.storeCodec(core, frame.hostValue, frame.aux())
		if err != nil {
			if !r.rejectCopy(err) {
				return true, err
			}
			complete = true
			return true, nil
		}
	}
	b.progress += uint32(count)
	r.notify(b.progress)
	if meta.Closed || r.typ.Kind == model.Future {
		w.closeSource()
	}
	complete = true
	return true, nil
}

func (s *sourceState) readReady(c *Caller, limit uint32) (ReadResult, bool, error) {
	s.in.callDepth++
	defer func() { s.in.callDepth-- }()
	return s.source.ReadReady(c, limit)
}
