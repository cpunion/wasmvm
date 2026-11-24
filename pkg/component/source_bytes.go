package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Already available bytes need a checked copy, not a host invocation and its
// external-work completion. Keep admission and value reservations identical to
// the ordinary source path, including while a destination callback runs.
func (in *Instance) readSourceBytes(c *vm.Caller, r, w *asyncEnd, b *guestBuffer, batch *model.Type) (bool, error) {
	source := w.source
	if source.source.ReadBytes == nil {
		return false, nil
	}
	if err := c.Context().Err(); err != nil {
		return true, err
	}
	if w.scope.failed || w.scope.blocked {
		return true, wasm.ErrTrap
	}
	if in.scheduler().live >= in.config.MaxTasks || in.callDepth >= in.config.Store.MaxReentry {
		return true, wasm.ErrLimit
	}
	if source.closed.Load() {
		return true, wasm.ErrClosed
	}
	// Both codecs belong to this read, including across copy cleanup.
	frame := in.acquireFrame()
	defer in.releaseCanonicalFrame(frame)
	return in.copySourceBytes(c, r, w, b, batch, frame.primaryState(), frame.aux())
}

func (in *Instance) copySourceBytes(c *vm.Caller, r, w *asyncEnd, b *guestBuffer, batch *model.Type, inputState, outputState *resourceCodecState) (handled bool, err error) {
	source := w.source
	complete := false
	var codec, output *canonical.Codec
	var values, bytes uint64
	defer func() {
		guard := cleanupGuard{panicked: recover()}
		if output != nil {
			err = guard.finishCodec(output, complete, err)
		}
		if !complete {
			guard.run(func() {
				r.buffer, r.state = nil, endDone
				w.closeSource()
			})
		}
		if codec != nil {
			err = guard.finishCodec(codec, complete, err)
		}
		in.heldValues -= values
		in.heldBytes -= bytes
		guard.finish()
	}()
	result, ready, err := source.readBytes(b.remain())
	if err != nil {
		return true, err
	}
	if !ready {
		if result.Values != nil || result.Bytes != "" || result.Count != 0 || result.Closed {
			return true, vm.ErrType
		}
		complete = true
		return false, nil
	}
	count := uint64(len(result.Bytes))
	if result.Values != nil || count > uint64(b.remain()) || result.Count != 0 && uint64(result.Count) != count || count == 0 && !result.Closed && b.remain() != 0 {
		return true, vm.ErrType
	}
	codec = in.initBufferCodec(inputState, c, w.scope, nil, r.typ.Element)
	codec.MaxValues = in.config.MaxValues - in.heldValues
	codec.Transfer.MaxBytes = in.config.MaxBytes - in.heldBytes
	value, err := codec.Clone(batch, Value{Text: result.Bytes})
	values, bytes = codec.Usage()
	in.heldValues += values
	in.heldBytes += bytes
	if err != nil {
		return true, err
	}
	if count != 0 {
		output, err = b.storeCodec(c, value, outputState)
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
	if result.Closed {
		w.closeSource()
	}
	complete = true
	return true, nil
}

func (s *sourceState) readBytes(limit uint32) (ReadResult, bool, error) {
	s.in.callDepth++
	defer func() { s.in.callDepth-- }()
	return s.source.ReadBytes(limit)
}
