package component

import (
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
	typ              *model.Type
	ptr, stride      uint64
	length, progress uint32
}

func (b *guestBuffer) remain() uint32 { return b.length - b.progress }
func (in *Instance) guestBuffer(c *vm.Caller, n *namespace, p *validate.CanonicalFunction, ptr, length uint64) (*guestBuffer, error) {
	if length >= 1<<28 {
		return nil, wasm.ErrTrap
	}
	b := &guestBuffer{in: in, n: n, plan: p, typ: p.Type.Value.Element, ptr: ptr, length: uint32(length)}
	if b.typ != nil && length > 0 {
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
	write := dst.in.codec(c, dst.n, dst.plan)
	complete := false
	defer func() { err = finishResources(write, complete, err); err = finishResources(read, complete, err) }()

	values, err := read.LoadBuffer(src.typ, src.ptr, uint64(count))
	if err != nil {
		return err
	}
	if err := write.StoreBuffer(dst.typ, dst.ptr, values); err != nil {
		return err
	}
	src.ptr += uint64(count) * src.stride
	dst.ptr += uint64(count) * dst.stride
	src.progress += count
	dst.progress += count

	complete = true
	return nil
}
