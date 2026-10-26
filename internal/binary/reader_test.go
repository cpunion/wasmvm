package binary_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/cpunion/wasmvm/internal/binary"
)

func checkError(t *testing.T, err, reason error, offset uint64) {
	t.Helper()
	var detail *binary.Error
	if !errors.Is(err, reason) || !errors.As(err, &detail) || detail.Offset != offset {
		t.Fatalf("error = %v, want %v at byte %d", err, reason, offset)
	}
	if !strings.Contains(detail.Error(), reason.Error()) {
		t.Fatalf("diagnostic omits reason: %v", detail)
	}
}

func TestReaderBoundaries(t *testing.T) {
	r := binary.New([]byte{0x11, 0x22, 0x33, 0x44})
	if r.Offset() != 0 || r.Remaining() != 4 {
		t.Fatal("incorrect initial cursor")
	}
	b, err := r.Byte()
	if b != 0x11 || err != nil {
		t.Fatalf("Byte = %x, %v", b, err)
	}
	child, err := r.Sub(2)
	if err != nil || r.Offset() != 3 || child.Offset() != 1 {
		t.Fatalf("Sub = %v; parent=%d, child=%d", err, r.Offset(), child.Offset())
	}
	checkError(t, child.Finish(), binary.ErrTrailingBytes, 1)
	grandchild, err := child.Sub(1)
	if err != nil || grandchild.Offset() != 1 {
		t.Fatal("nested reader lost its absolute offset")
	}
	data, err := grandchild.Bytes(1)
	if err != nil || !bytes.Equal(data, []byte{0x22}) || cap(data) != 1 {
		t.Fatalf("Bytes = %x (cap %d), %v", data, cap(data), err)
	}
	if err := grandchild.Finish(); err != nil {
		t.Fatal(err)
	}
	_, err = grandchild.Byte()
	checkError(t, err, binary.ErrUnexpectedEOF, 2)
	_, err = child.Bytes(^uint64(0))
	checkError(t, err, binary.ErrUnexpectedEOF, 2)
	if child.Remaining() != 1 {
		t.Fatal("failed read advanced the cursor")
	}
	_, err = child.Sub(2)
	checkError(t, err, binary.ErrUnexpectedEOF, 2)
	b, err = child.Byte()
	if b != 0x33 || err != nil || child.Finish() != nil {
		t.Fatal("child crossed its boundary or lost data after a failed read")
	}
	b, err = r.Byte()
	if b != 0x44 || err != nil || r.Finish() != nil {
		t.Fatal("child reads changed its parent's cursor")
	}
	var empty binary.Reader
	data, err = empty.Bytes(0)
	if len(data) != 0 || err != nil || empty.Finish() != nil {
		t.Fatal("zero-value reader should accept an empty range")
	}
	_, err = empty.Byte()
	checkError(t, err, binary.ErrUnexpectedEOF, 0)
}

func readInteger(r *binary.Reader, kind string) (uint64, error) {
	switch kind {
	case "u32":
		v, err := r.U32()
		return uint64(v), err
	case "u64":
		return r.U64()
	case "s32":
		v, err := r.S32()
		return uint64(int64(v)), err
	case "s33":
		v, err := r.S33()
		return uint64(v), err
	case "s64":
		v, err := r.S64()
		return uint64(v), err
	default:
		panic("unknown test integer kind")
	}
}

func TestIntegerExamples(t *testing.T) {
	for _, test := range []struct {
		kind  string
		bytes []byte
		want  uint64
	}{
		{"u32", []byte{0xe5, 0x8e, 0x26}, 624485},
		{"u32", []byte{0xff, 0xff, 0xff, 0xff, 0x0f}, 0xffffffff},
		{"u32", []byte{0x83, 0x80, 0x80, 0x80, 0x00}, 3},
		{"u64", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}, ^uint64(0)},
		{"s32", []byte{0x9b, 0xf1, 0x59}, ^uint64(624484)},
		{"s32", []byte{0xff, 0xff, 0xff, 0xff, 0x07}, 0x7fffffff},
		{"s32", []byte{0x80, 0x80, 0x80, 0x80, 0x78}, ^uint64(0x7fffffff)},
		{"s32", []byte{0xff, 0xff, 0xff, 0xff, 0x7f}, ^uint64(0)},
		{"s33", []byte{0xff, 0xff, 0xff, 0xff, 0x0f}, 0xffffffff},
		{"s33", []byte{0x80, 0x80, 0x80, 0x80, 0x70}, ^uint64(0xffffffff)},
		{"s33", []byte{0x40}, ^uint64(63)},
		{"s64", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}, 0x7fffffffffffffff},
		{"s64", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x7f}, 0x8000000000000000},
		{"s64", []byte{0xfe, 0xff, 0x7f}, ^uint64(1)},
	} {
		r := binary.New(append(append([]byte(nil), test.bytes...), 0x55))
		got, err := readInteger(&r, test.kind)
		if err != nil || got != test.want || r.Remaining() != 1 {
			t.Fatalf("%s(%x) = %#x, %v; remaining=%d, want %#x and one byte", test.kind, test.bytes, got, err, r.Remaining(), test.want)
		}
	}
}

func TestMalformedIntegers(t *testing.T) {
	for _, test := range []struct {
		kind   string
		bytes  []byte
		reason error
		offset uint64
	}{
		{"u32", nil, binary.ErrUnexpectedEOF, 0},
		{"u32", []byte{0x80}, binary.ErrUnexpectedEOF, 1},
		{"u32", []byte{0x80, 0x80, 0x80, 0x80, 0x10}, binary.ErrIntegerOverflow, 4},
		{"u32", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0}, binary.ErrIntegerTooLong, 4},
		{"u64", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 2}, binary.ErrIntegerOverflow, 9},
		{"u64", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0}, binary.ErrIntegerTooLong, 9},
		{"s32", []byte{0x80, 0x80, 0x80, 0x80, 0x08}, binary.ErrIntegerOverflow, 4},
		{"s32", []byte{0xff, 0xff, 0xff, 0xff, 0x77}, binary.ErrIntegerOverflow, 4},
		{"s33", []byte{0x80, 0x80, 0x80, 0x80, 0x10}, binary.ErrIntegerOverflow, 4},
		{"s33", []byte{0xff, 0xff, 0xff, 0xff, 0x6f}, binary.ErrIntegerOverflow, 4},
		{"s64", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 1}, binary.ErrIntegerOverflow, 9},
		{"s64", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7e}, binary.ErrIntegerOverflow, 9},
	} {
		r := binary.New(test.bytes)
		_, err := readInteger(&r, test.kind)
		checkError(t, err, test.reason, test.offset)
	}
}

func TestEveryTerminalIntegerByte(t *testing.T) {
	// The specification constrains the last byte independently of the lower
	// payload bits. Exhaust its complete byte domain for every supported width.
	for _, kind := range []string{"u32", "u64", "s32", "s33", "s64"} {
		width := 5
		if kind == "u64" || kind == "s64" {
			width = 10
		}
		for last := 0; last < 256; last++ {
			valid := false
			switch kind {
			case "u32":
				valid = last <= 15
			case "u64":
				valid = last <= 1
			case "s32":
				valid = last <= 7 || last >= 120 && last <= 127
			case "s33":
				valid = last <= 15 || last >= 112 && last <= 127
			case "s64":
				valid = last == 0 || last == 127
			}
			data := append(bytes.Repeat([]byte{0x80}, width-1), byte(last))
			r := binary.New(data)
			_, err := readInteger(&r, kind)
			if (err == nil) != valid || r.Remaining() != 0 {
				t.Fatalf("%s last=%#x: error=%v; want valid=%v, consumed=%d/%d", kind, last, err, valid, r.Offset(), width)
			}
		}
	}
}

func encodeUnsigned(v uint64) []byte {
	var out []byte
	for v >= 128 {
		out = append(out, byte(v)|0x80)
		v >>= 7
	}
	return append(out, byte(v))
}

func encodeSigned(v int64) []byte {
	var out []byte
	for {
		b := byte(v) & 0x7f
		v >>= 7
		if v == 0 && b&0x40 == 0 || v == -1 && b&0x40 != 0 {
			return append(out, b)
		}
		out = append(out, b|0x80)
	}
}

func TestIntegerRoundTripsAndPadding(t *testing.T) {
	states := []uint64{0, 1, 2, 63, 64, 127, 128, 1<<31 - 1, 1 << 31,
		1<<32 - 1, 1 << 32, 1<<63 - 1, 1 << 63, ^uint64(0), ^uint64(1), ^uint64(63), ^uint64(64)}
	state := uint64(1)
	for i := 0; i < 256; i++ {
		state = state*6364136223846793005 + 1442695040888963407
		states = append(states, state)
	}
	for _, state := range states {
		for _, kind := range []string{"u32", "u64", "s32", "s33", "s64"} {
			value := state
			width := 10
			switch kind {
			case "u32":
				value, width = uint64(uint32(state)), 5
			case "s32":
				value, width = uint64(int64(int32(state))), 5
			case "s33":
				value, width = uint64(int64(state<<31)>>31), 5
			}
			encoded := encodeUnsigned(value)
			pad := byte(0)
			if kind[0] == 's' {
				encoded = encodeSigned(int64(value))
				if int64(value) < 0 {
					pad = 0x7f
				}
			}
			for {
				r := binary.New(encoded)
				got, err := readInteger(&r, kind)
				if got != value || err != nil || r.Finish() != nil {
					t.Fatalf("%s(%x) = %#x, %v; want %#x", kind, encoded, got, err, value)
				}
				if len(encoded) == width {
					break
				}
				encoded[len(encoded)-1] |= 0x80
				encoded = append(encoded, pad)
			}
		}
	}
}

func TestFloatingPointBits(t *testing.T) {
	r := binary.New([]byte{0, 0, 0, 0x80, 1, 0, 0, 0, 0, 0, 0xf8, 0x7f})
	f32, err := r.F32Bits()
	if f32 != 0x80000000 || err != nil {
		t.Fatalf("f32 negative zero = %#x, %v", f32, err)
	}
	f64, err := r.F64Bits()
	if f64 != 0x7ff8000000000001 || err != nil || r.Finish() != nil {
		t.Fatalf("f64 NaN payload = %#x, %v", f64, err)
	}
	_, err = r.F32Bits()
	checkError(t, err, binary.ErrUnexpectedEOF, 12)
	_, err = r.F64Bits()
	checkError(t, err, binary.ErrUnexpectedEOF, 12)
}

func TestNames(t *testing.T) {
	for _, name := range []string{"", "hello", "a\x00b", "\u4f60\u597d", "\U0001f600"} {
		r := binary.New(append(encodeUnsigned(uint64(len(name))), []byte(name)...))
		got, err := r.Name(uint32(len(name)))
		if got != name || err != nil || r.Finish() != nil {
			t.Fatalf("Name(%q) = %q, %v", name, got, err)
		}
	}
	for _, bad := range [][]byte{{0xc0, 0x80}, {0xed, 0xa0, 0x80}, {0xf4, 0x90, 0x80, 0x80}, {0xff}, {0xc2}} {
		r := binary.New(append([]byte{byte(len(bad))}, bad...))
		_, err := r.Name(16)
		checkError(t, err, binary.ErrInvalidUTF8, 1)
	}
	r := binary.New([]byte{0xff, 0xff, 0xff, 0xff, 0x0f})
	_, err := r.Name(1024)
	checkError(t, err, binary.ErrLimit, 0)
	r = binary.New([]byte{2, 'x'})
	_, err = r.Name(2)
	checkError(t, err, binary.ErrUnexpectedEOF, 1)
	r = binary.New([]byte{0x80})
	_, err = r.Name(2)
	checkError(t, err, binary.ErrUnexpectedEOF, 1)
}

func FuzzReader(f *testing.F) {
	for _, seed := range [][]byte{nil, {0}, {0x7f}, {0x80}, {0xff, 0xff, 0xff, 0xff, 0x7f}, {0xff, 0xff, 0xff, 0xff, 0x0f}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		for _, kind := range []string{"u32", "u64", "s32", "s33", "s64"} {
			r := binary.New(data)
			_, _ = readInteger(&r, kind)
			if r.Offset() > uint64(len(data)) || r.Offset() > 10 || r.Remaining() < 0 {
				t.Fatalf("%s escaped its bounded input", kind)
			}
		}
		r := binary.New(data)
		_, _ = r.Name(1024)
		if r.Offset() > uint64(len(data)) {
			t.Fatal("name escaped its bounded input")
		}
	})
}
