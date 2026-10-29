package canonical

import (
	"context"
	"encoding/binary"
	"math"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

type Encoding byte

const (
	UTF8 Encoding = iota
	UTF16
	Latin1UTF16
	MaxStringBytes = (1 << 28) - 1
)

// Transfer is scoped to a synchronous canonical call. Memory and Realloc use
// checked capabilities; allocation never returns a directly borrowed slice.
// MaxBytes bounds aggregate scratch/output, including transcoding expansion.
type Transfer struct {
	Context  context.Context
	Memory   Memory
	Memory64 bool
	Encoding Encoding
	MaxBytes uint64
	Realloc  func(old, oldSize, alignment, newSize uint64) (uint64, error)
	used     uint64
}

func (c *Transfer) check() error {
	if c.Context != nil {
		if err := c.Context.Err(); err != nil {
			return err
		}
	}
	if c.Encoding > Latin1UTF16 {
		return invalid("invalid string encoding")
	}
	return nil
}

func (c *Transfer) charge(n uint64) error {
	if n > c.MaxBytes-c.used {
		return exhausted()
	}
	c.used += n
	return nil
}

func (c *Transfer) pointerSize() uint64 {
	if c.Memory64 {
		return 8
	}
	return 4
}
func (c *Transfer) maxPointer() uint64 {
	if c.Memory64 {
		return math.MaxUint64
	}
	return math.MaxUint32
}

func (c *Transfer) rangeCheck(at, n, align uint64) error {
	if err := c.check(); err != nil {
		return err
	}
	if c.Memory == nil {
		return invalid("missing canonical memory")
	}
	endLimit := c.maxPointer()
	if !c.Memory64 {
		endLimit++
	}
	if at&(align-1) != 0 || at > c.maxPointer() || n > endLimit-at {
		return scalarTrap("invalid string range")
	}
	// Check the whole range before allocating host scratch, even for empty strings.
	return c.Memory.Read(at+n, nil)
}

func (c *Transfer) LoadString(at uint64) (string, error) {
	width := c.pointerSize()
	if err := c.rangeCheck(at, 2*width, width); err != nil {
		return "", err
	}
	var pair [16]byte
	if err := c.Memory.Read(at, pair[:2*width]); err != nil {
		return "", err
	}
	ptr, length := uint64(binary.LittleEndian.Uint32(pair[:])), uint64(binary.LittleEndian.Uint32(pair[4:]))
	if c.Memory64 {
		ptr, length = binary.LittleEndian.Uint64(pair[:]), binary.LittleEndian.Uint64(pair[8:])
	}
	return c.LiftString(ptr, length)
}

func (c *Transfer) LiftString(at, length uint64) (string, error) {
	if err := c.check(); err != nil {
		return "", err
	}
	if length > c.maxPointer() {
		return "", scalarTrap("string length overflow")
	}
	encoding, align, n := c.Encoding, uint64(1), length
	if encoding == Latin1UTF16 {
		align = 2
		tag := uint64(1) << (c.pointerSize()*8 - 1)
		if length&tag != 0 {
			encoding = UTF16
			n = length &^ tag
		}
	}
	if encoding == UTF16 {
		align = 2
		if n > MaxStringBytes/2 {
			return "", scalarTrap("string length exceeds ABI limit")
		}
		n *= 2
	}
	if n > MaxStringBytes {
		return "", scalarTrap("string length exceeds ABI limit")
	}
	if err := c.rangeCheck(at, n, align); err != nil {
		return "", err
	}
	// UTF-16 and latin1 can expand in host UTF-8. Charge a safe upper bound.
	output := n
	if encoding != UTF8 {
		output = 3 * n
	}
	if err := c.charge(n + output); err != nil {
		return "", err
	}
	b := make([]byte, int(n))
	if err := c.Memory.Read(at, b); err != nil {
		return "", err
	}
	if encoding == UTF8 {
		if !utf8.Valid(b) {
			return "", scalarTrap("invalid UTF-8")
		}
		return string(b), nil
	}
	var out strings.Builder
	out.Grow(int(output))
	for i := 0; i < len(b); {
		if i&4095 == 0 {
			if err := c.check(); err != nil {
				return "", err
			}
		}
		if encoding == Latin1UTF16 {
			out.WriteRune(rune(b[i]))
			i++
			continue
		}
		v := binary.LittleEndian.Uint16(b[i:])
		i += 2
		r := rune(v)
		if v >= 0xd800 && v <= 0xdfff {
			if v > 0xdbff || i == len(b) {
				return "", scalarTrap("unpaired UTF-16 surrogate")
			}
			low := binary.LittleEndian.Uint16(b[i:])
			i += 2
			if low < 0xdc00 || low > 0xdfff {
				return "", scalarTrap("unpaired UTF-16 surrogate")
			}
			r = utf16.DecodeRune(r, rune(low))
		}
		out.WriteRune(r)
	}
	return out.String(), nil
}

// LowerString allocates an exact-size encoding of a host Unicode string. No
// source-encoding optimization is implied: this is the host-value entry path.
func (c *Transfer) LowerString(value string) (ptr, length uint64, err error) {
	if err = c.check(); err != nil {
		return
	}
	if c.Memory == nil {
		return 0, 0, invalid("missing canonical memory")
	}
	if err = c.validateString(value); err != nil {
		return
	}
	encoding, align := c.Encoding, uint64(1)
	units, latin := uint64(0), true
	for index, r := range value {
		if index&4095 == 0 {
			if err = c.check(); err != nil {
				return
			}
		}
		units++
		if r > 0xffff {
			units++
		}
		if r > 255 {
			latin = false
		}
	}
	n := uint64(len(value))
	length = n
	if encoding == UTF16 || encoding == Latin1UTF16 && !latin {
		encoding, align, n, length = UTF16, 2, 2*units, units
		if c.Encoding == Latin1UTF16 {
			length |= uint64(1) << (c.pointerSize()*8 - 1)
		}
	} else if encoding == Latin1UTF16 {
		align, n, length = 2, uint64(utf8.RuneCountInString(value)), uint64(utf8.RuneCountInString(value))
	}
	if n > MaxStringBytes {
		return 0, 0, scalarTrap("string length exceeds ABI limit")
	}
	if err = c.charge(n); err != nil {
		return
	}
	if c.Realloc == nil {
		return 0, 0, invalid("missing realloc")
	}
	ptr, err = c.Realloc(0, 0, align, n)
	if err != nil {
		return
	}
	if err = c.rangeCheck(ptr, n, align); err != nil {
		return
	}
	buf := make([]byte, int(n))
	if encoding == UTF8 {
		copy(buf, value)
	} else {
		i := 0
		for _, r := range value {
			if i&4095 == 0 {
				if err = c.check(); err != nil {
					return
				}
			}
			if encoding == Latin1UTF16 {
				buf[i] = byte(r)
				i++
				continue
			}
			if r <= 0xffff {
				binary.LittleEndian.PutUint16(buf[i:], uint16(r))
				i += 2
			} else {
				hi, lo := utf16.EncodeRune(r)
				binary.LittleEndian.PutUint16(buf[i:], uint16(hi))
				binary.LittleEndian.PutUint16(buf[i+2:], uint16(lo))
				i += 4
			}
		}
	}
	err = c.Memory.Write(ptr, buf)
	return
}

func (c *Transfer) validateString(value string) error {
	if !utf8.ValidString(value) {
		return invalid("host UTF-8 string")
	}
	if uint64(len(value)) > MaxStringBytes {
		return scalarTrap("string length exceeds ABI limit")
	}
	return nil
}

func (c *Transfer) StoreString(at uint64, value string) error {
	width := c.pointerSize()
	if err := c.rangeCheck(at, 2*width, width); err != nil {
		return err
	}
	ptr, length, err := c.LowerString(value)
	if err != nil {
		return err
	}
	var pair [16]byte
	if c.Memory64 {
		binary.LittleEndian.PutUint64(pair[:], ptr)
		binary.LittleEndian.PutUint64(pair[8:], length)
	} else {
		binary.LittleEndian.PutUint32(pair[:], uint32(ptr))
		binary.LittleEndian.PutUint32(pair[4:], uint32(length))
	}
	return c.Memory.Write(at, pair[:2*width])
}
