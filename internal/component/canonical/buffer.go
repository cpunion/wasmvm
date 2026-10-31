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
		if err := c.Transfer.rangeCheck(ptr, n, 1); err != nil {
			return err
		}
		if err := c.Transfer.charge(n); err != nil {
			return err
		}
		return c.Transfer.Memory.Write(ptr, []byte(v.Text))
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
