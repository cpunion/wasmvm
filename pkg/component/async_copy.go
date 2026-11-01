package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"math"
)

// A parked buffer retains immutable options and a memory capability. Every
// actual transfer rebuilds Caller-bound codecs in the currently active session.
type guestBuffer struct {
	in               *Instance
	n                *namespace
	plan             *validate.CanonicalFunction
	scope            *namespace
	batch            *model.Type
	data             Value
	typ              *model.Type
	ptr, stride      uint64
	length, progress uint32
	check            func(ReadResult) error
	rejected         error
}

func (b *guestBuffer) remain() uint32 { return b.length - b.progress }
func (in *Instance) guestBuffer(c *vm.Caller, n *namespace, p *validate.CanonicalFunction, ptr, length uint64) (*guestBuffer, error) {
	if length >= 1<<28 {
		return nil, wasm.ErrTrap
	}
	b := &guestBuffer{in: in, n: n, plan: p, typ: p.Type.Value.Element, ptr: ptr, length: uint32(length)}
	if b.typ != nil {
		codec := in.codec(c, n, p)
		l, err := codec.Layouts.Build(b.typ)
		if err != nil {
			return nil, err
		}
		if ptr&(l.Align-1) != 0 || l.Size != 0 && length > (math.MaxUint64-ptr)/l.Size {
			return nil, wasm.ErrTrap
		}
		b.stride = l.Size
		if err := c.ReadExternalMemory(n.items[model.CoreMemorySort][*p.Memory].core, ptr+length*l.Size, nil); err != nil {
			return nil, err
		}
	}
	return b, nil
}
func copyGuestBuffers(c *vm.Caller, src, dst *guestBuffer, count uint32) (err error) {
	if src.typ == nil {
		src.progress += count
		dst.progress += count
		return nil
	}
	read := src.in.codec(c, src.n, src.plan)
	complete := false
	var write *canonical.Codec
	defer func() {
		guard := cleanupGuard{panicked: recover()}
		if write != nil {
			guard.run(func() { err = finishResources(write, complete, err) })
		}
		guard.run(func() { err = finishResources(read, complete, err) })
		guard.finish()
	}()

	values, err := read.LoadBuffer(src.typ, src.ptr, uint64(count))
	if err != nil {
		return err
	}
	write, err = dst.store(c, values)
	if err != nil {
		return err
	}
	src.ptr += uint64(count) * src.stride
	dst.ptr += uint64(count) * dst.stride
	src.progress += count
	dst.progress += count

	complete = true
	return nil
}

type bufferRejection struct{ error }

func (b *guestBuffer) store(c *vm.Caller, values Value) (*canonical.Codec, error) {
	if b.check != nil {
		data := values.Text
		if values.Elements != nil {
			var err error
			data, err = values.ByteString()
			if err != nil {
				return nil, err
			}
		}
		result := ReadResult{Bytes: data, Count: uint32(len(data))}
		if err := b.check(result); err != nil {
			b.rejected = err
			return nil, &bufferRejection{err}
		}
	}
	n, plan := b.n, b.plan
	if n == nil {
		n = b.scope
	}
	codec := b.in.codec(c, n, plan)
	var err error
	if b.n != nil {
		err = codec.StoreBuffer(b.typ, b.ptr, values)
	} else {
		// A producer may complete several writes before the consumer observes its
		// event. Preserve every copied chunk and the cumulative budget/ownership.
		if b.progress != 0 {
			if values.Elements == nil && b.typ.Kind == model.U8 {
				if uint64(len(b.data.Text))+uint64(len(values.Text)) > b.in.config.MaxBytes-b.in.heldBytes {
					return codec, wasm.ErrLimit
				}
				values.Text = b.data.Text + values.Text
			} else {
				if uint64(len(b.data.Elements))+uint64(len(values.Elements))+1 > b.in.config.MaxValues-b.in.heldValues {
					return codec, wasm.ErrLimit
				}
				values.Elements = append(append([]Value(nil), b.data.Elements...), values.Elements...)
			}
		}
		b.data, err = codec.Clone(b.batch, values)
		if err == nil {
			err = codec.AcceptResult(b.batch, &b.data)
		}
	}
	return codec, err
}
