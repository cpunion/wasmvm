package canonical

import "github.com/cpunion/wasmvm/internal/component/model"

// LoadBuffer snapshots all source elements before a destination is modified.
// It shares list bounds, packed bytes, ownership and aggregate copy budgets.
func (c *Codec) LoadBuffer(t *model.Type, ptr, length uint64) (Value, error) {
	return c.loadList(t, ptr, length, 0)
}

// StoreBuffer writes a sequence into the caller-provided fixed-size region;
// nested strings/lists still allocate through this buffer's canonical options.
func (c *Codec) StoreBuffer(t *model.Type, ptr uint64, v Value) error {
	l, err := c.Layouts.Build(t)
	if err != nil {
		return err
	}
	if t.Kind == model.U8 && v.Elements == nil {
		n := uint64(len(v.Text))
		if err := c.byteValues(n, 0); err != nil {
			return err
		}
		return c.storeBytes(ptr, v.Text)
	}
	n := uint64(len(v.Elements))
	if n > c.MaxValues-c.values || n > MaxStringBytes/l.Size {
		return exhausted()
	}
	if err := c.Transfer.rangeCheck(ptr, n*l.Size, l.Align); err != nil {
		return err
	}
	for _, element := range v.Elements {
		if err := c.validate(t, element, 0); err != nil {
			return err
		}
	}
	for i, element := range v.Elements {
		if err := c.store(t, ptr+uint64(i)*l.Size, element, 0); err != nil {
			return err
		}
	}
	return nil
}

// ReadBufferInto copies packed bytes into caller-owned storage. Source bounds
// and the same logical charges as LoadBuffer are checked before destination is
// invoked, so an invalid read cannot cause destination allocation or effects.
func (c *Codec) ReadBufferInto(ptr, length uint64, destination func() ([]byte, error)) error {
	if err := c.prepareReadBytes(ptr, length, 0); err != nil {
		return err
	}
	dst, err := destination()
	if err != nil {
		return err
	}
	if uint64(len(dst)) < length {
		return invalid("short byte buffer")
	}
	if err := c.Transfer.Memory.Read(ptr, dst[:int(length)]); err != nil {
		return err
	}
	return c.Transfer.check()
}

// ReserveByteBuffer charges a host list<u8> snapshot without materializing an
// intermediate string. The reservation includes all earlier writes in a batch.
func (c *Codec) ReserveByteBuffer(t *model.Type, length uint64) error {
	if err := c.visit(0); err != nil {
		return err
	}
	if t == nil || t.Kind != model.List || t.Element == nil || t.Element.Kind != model.U8 {
		return invalid("byte buffer type")
	}
	if _, err := c.Layouts.Build(t); err != nil {
		return err
	}
	if err := c.byteValues(length, 0); err != nil {
		return err
	}
	return c.Transfer.charge(length)
}
