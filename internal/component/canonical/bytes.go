package canonical

import "github.com/cpunion/wasmvm/internal/component/model"

// ByteList constructs an owned packed list<u8>. Text holds immutable raw bytes,
// including invalid UTF-8; this reuses Value storage without enlarging each slot.
func ByteList(data []byte) Value { return Value{Text: string(data)} }

// ByteString reads either the packed or element-wise representation of a
// list<u8>/fixed-list<u8>. Its immutable result never aliases mutable host data.
// The caller must already know the declared type; a Value carries no type tag.
func (v Value) ByteString() (string, error) {
	if v.Elements == nil {
		return v.Text, nil
	}
	if v.Text != "" {
		return "", invalid("mixed byte-list representations")
	}
	data := make([]byte, len(v.Elements))
	for i, element := range v.Elements {
		if element.Bits > 255 {
			return "", invalid("host byte-list element")
		}
		data[i] = byte(element.Bits)
	}
	return string(data), nil
}

func byteSequence(t *model.Type) bool {
	return (t.Kind == model.List || t.Kind == model.FixedList) && t.Element.Kind == model.U8
}

func (c *Codec) byteValues(n uint64, depth uint32) error {
	if n > c.MaxValues-c.values || n > MaxStringBytes {
		return exhausted()
	}
	if n != 0 && depth+1 >= c.Layouts.limits.MaxDepth {
		return exhausted()
	}
	c.values += n
	return nil
}

func (c *Codec) loadBytes(at, n uint64, depth uint32) (Value, error) {
	if err := c.byteValues(n, depth); err != nil {
		return Value{}, err
	}
	if err := c.Transfer.rangeCheck(at, n, 1); err != nil {
		return Value{}, err
	}
	// Account for scratch and the immutable result, independently of value count.
	if err := c.Transfer.charge(2 * n); err != nil {
		return Value{}, err
	}
	data := make([]byte, int(n))
	if err := c.Transfer.Memory.Read(at, data); err != nil {
		return Value{}, err
	}
	if err := c.Transfer.check(); err != nil {
		return Value{}, err
	}
	return ByteList(data), nil
}

func (c *Codec) storeBytes(at uint64, data string) error {
	if err := c.Transfer.rangeCheck(at, uint64(len(data)), 1); err != nil {
		return err
	}
	if err := c.Transfer.charge(uint64(len(data))); err != nil {
		return err
	}
	if err := c.Transfer.Memory.Write(at, []byte(data)); err != nil {
		return err
	}
	return c.Transfer.check()
}

func (c *Codec) lowerBytes(data string) (uint64, error) {
	n := uint64(len(data))
	if n > MaxStringBytes {
		return 0, exhausted()
	}
	if err := c.Transfer.charge(n); err != nil {
		return 0, err
	}
	if c.Transfer.Realloc == nil {
		return 0, invalid("missing realloc")
	}
	ptr, err := c.Transfer.Realloc(0, 0, 1, n)
	if err != nil {
		return 0, err
	}
	return ptr, c.storeBytes(ptr, data)
}
