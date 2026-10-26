// Package binary implements bounded primitives for the WebAssembly binary
// format. Decoders must separately bound input ownership and aggregate module
// allocations. A Reader borrows its input; only Name allocates a result.
package binary

import (
	stdbinary "encoding/binary"
	"errors"
	"fmt"
	"unicode/utf8"
)

var (
	ErrUnexpectedEOF   = errors.New("unexpected end of input")
	ErrIntegerTooLong  = errors.New("integer representation too long")
	ErrIntegerOverflow = errors.New("integer too large")
	ErrInvalidUTF8     = errors.New("invalid UTF-8 name")
	ErrTrailingBytes   = errors.New("unexpected trailing bytes")
	ErrLimit           = errors.New("binary resource limit exceeded")
)

// Error locates a decoding failure relative to the original input, including
// when it occurs in a section or function body's bounded subreader.
type Error struct {
	Offset uint64
	Reason error
}

func (e *Error) Error() string { return fmt.Sprintf("wasm byte %#x: %v", e.Offset, e.Reason) }
func (e *Error) Unwrap() error { return e.Reason }

// Reader is a cursor over a borrowed byte slice. A failed fixed-length read
// leaves its cursor unchanged; a failed LEB read consumes the inspected bytes.
// The zero value is an empty reader.
type Reader struct {
	data []byte
	base uint64
	pos  int
}

func New(data []byte) Reader { return Reader{data: data} }

func (r *Reader) Offset() uint64 { return r.base + uint64(r.pos) }
func (r *Reader) Remaining() int { return len(r.data) - r.pos }

// Bytes checks the length before converting it to a host index. Its result
// aliases the input and cannot extend beyond this reader's boundary.
func (r *Reader) Bytes(n uint64) ([]byte, error) {
	if n > uint64(r.Remaining()) {
		return nil, &Error{r.Offset(), ErrUnexpectedEOF}
	}
	start := r.pos
	r.pos += int(n)
	return r.data[start:r.pos:r.pos], nil
}

func (r *Reader) Byte() (byte, error) {
	if r.Remaining() == 0 {
		return 0, &Error{r.Offset(), ErrUnexpectedEOF}
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

// Sub advances the parent by n bytes and returns a reader limited to that
// range. Finish must be checked when the section or body has been decoded.
func (r *Reader) Sub(n uint64) (Reader, error) {
	base := r.Offset()
	data, err := r.Bytes(n)
	if err != nil {
		return Reader{}, err
	}
	return Reader{data: data, base: base}, nil
}

func (r *Reader) Finish() error {
	if r.Remaining() != 0 {
		return &Error{r.Offset(), ErrTrailingBytes}
	}
	return nil
}

func (r *Reader) U32() (uint32, error) {
	v, err := r.leb(32, false)
	return uint32(v), err
}

func (r *Reader) U64() (uint64, error) { return r.leb(64, false) }

func (r *Reader) S32() (int32, error) {
	v, err := r.leb(32, true)
	return int32(v), err
}

// S33 reads a signed block type: a negative type code or a u32 type index.
func (r *Reader) S33() (int64, error) {
	v, err := r.leb(33, true)
	return int64(v), err
}

func (r *Reader) S64() (int64, error) {
	v, err := r.leb(64, true)
	return int64(v), err
}

// WebAssembly permits padded encodings within ceil(bits/7) bytes, but requires
// the terminal byte's unused bits to be zero or the correct sign extension.
func (r *Reader) leb(bits uint, signed bool) (uint64, error) {
	var value uint64
	for shift := uint(0); ; shift += 7 {
		b, err := r.Byte()
		if err != nil {
			return 0, err
		}
		payload := uint64(b & 0x7f)
		remaining := bits - shift
		if remaining <= 7 {
			if b&0x80 != 0 {
				return 0, &Error{r.Offset() - 1, ErrIntegerTooLong}
			}
			unused := uint64(0x7f) & (^uint64(0) << remaining)
			want := uint64(0)
			if signed && payload&(1<<(remaining-1)) != 0 {
				want = unused
			}
			if payload&unused != want {
				return 0, &Error{r.Offset() - 1, ErrIntegerOverflow}
			}
		}
		value |= payload << shift
		if b&0x80 == 0 {
			if signed && b&0x40 != 0 && shift+7 < 64 {
				value |= ^uint64(0) << (shift + 7)
			}
			return value, nil
		}
	}
}

// F32Bits and F64Bits preserve all floating-point bits, including NaN payloads
// and signed zero, without applying host floating-point operations.
func (r *Reader) F32Bits() (uint32, error) {
	b, err := r.Bytes(4)
	if err != nil {
		return 0, err
	}
	return stdbinary.LittleEndian.Uint32(b), nil
}

func (r *Reader) F64Bits() (uint64, error) {
	b, err := r.Bytes(8)
	if err != nil {
		return 0, err
	}
	return stdbinary.LittleEndian.Uint64(b), nil
}

// Name rejects a length above maxBytes before allocating. A zero limit permits
// only an empty name. Names are UTF-8 byte vectors, not zero-terminated strings.
func (r *Reader) Name(maxBytes uint32) (string, error) {
	start := r.Offset()
	n, err := r.U32()
	if err != nil {
		return "", err
	}
	if n > maxBytes {
		return "", &Error{start, ErrLimit}
	}
	start = r.Offset()
	b, err := r.Bytes(uint64(n))
	if err != nil {
		return "", err
	}
	if !utf8.Valid(b) {
		return "", &Error{start, ErrInvalidUTF8}
	}
	return string(b), nil
}
