package memory

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLimitsGrowthAndHostAccess(t *testing.T) {
	for _, tc := range []struct {
		limits module.Limits
		quota  uint32
		kind   error
	}{
		{module.Limits{Min: MaxPages + 1}, MaxPages, wasm.ErrInvalid},
		{module.Limits{Min: 2, Max: 1, HasMax: true}, 2, wasm.ErrInvalid},
		{module.Limits{Max: MaxPages + 1, HasMax: true}, MaxPages, wasm.ErrInvalid},
		{module.Limits{Min: 1}, 0, wasm.ErrLimit},
	} {
		if got, err := New(tc.limits, tc.quota); got != nil || !errors.Is(err, tc.kind) {
			t.Fatal(got, err)
		}
	}
	declared := module.Limits{Min: 1, Max: 3, HasMax: true}
	m, err := New(declared, 2)
	if err != nil {
		t.Fatal(err)
	}
	if m.Pages() != 1 || m.Limits() != declared {
		t.Fatal("memory metadata")
	}
	input := []byte{1, 2, 3, 4}
	if !m.Write(PageSize-4, input) {
		t.Fatal("write")
	}
	input[0] = 9
	if old, ok := m.Grow(0); !ok || old != 1 {
		t.Fatal(old, ok)
	}
	if old, ok := m.Grow(1); !ok || old != 1 || m.Pages() != 2 {
		t.Fatal(old, ok)
	}
	got := make([]byte, 8)
	if !m.Read(PageSize-4, got) || !bytes.Equal(got, []byte{1, 2, 3, 4, 0, 0, 0, 0}) {
		t.Fatal(got)
	}
	for _, delta := range []uint64{1, MaxPages, math.MaxUint32} {
		if old, ok := m.Grow(delta); ok || old != 2 || m.Pages() != 2 {
			t.Fatal(old, ok)
		}
	}
	for _, at := range []uint64{2*PageSize - 3, 2 * PageSize, math.MaxUint64} {
		dst := []byte{9, 9, 9, 9}
		if m.Read(at, dst) || !bytes.Equal(dst, []byte{9, 9, 9, 9}) || m.Write(at, []byte{1, 2, 3, 4}) {
			t.Fatal("bad host range")
		}
	}
	if !m.Read(2*PageSize, nil) || !m.Write(2*PageSize, nil) || m.Read(2*PageSize+1, nil) {
		t.Fatal("empty host range")
	}
	m, err = New(module.Limits{}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if old, ok := m.Grow(1); ok || old != 0 {
		t.Fatal(old, ok)
	}
	m, err = New(module.Limits{Max: 0, HasMax: true}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Grow(1); ok {
		t.Fatal("declared max ignored")
	}
}

func TestRelease(t *testing.T) {
	m, err := New(module.Limits{Min: 1}, 2)
	if err != nil {
		t.Fatal(err)
	}
	m.Release()
	m.Release()
	if m.Pages() != 0 || m.bytes != nil || m.Write(0, []byte{1}) {
		t.Fatal("retained released storage")
	}
	if _, ok := m.Grow(1); ok {
		t.Fatal("released memory grew")
	}
}

func TestBulkOperations(t *testing.T) {
	for _, pair := range [][2]uint64{{2, 0}, {0, 2}, {0, 0}} {
		initial := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
		want := append([]byte(nil), initial...)
		saved := append([]byte(nil), initial[pair[1]:pair[1]+8]...)
		for i, b := range saved {
			want[int(pair[0])+i] = b
		}
		m := &Memory{bytes: append([]byte(nil), initial...)}
		if !m.Copy(pair[0], pair[1], 8) || !bytes.Equal(m.bytes, want) {
			t.Fatal(pair, m.bytes, want)
		}
	}
	m := &Memory{bytes: make([]byte, 12)}
	if !m.Fill(2, 0xab, 8) || !bytes.Equal(m.bytes, []byte{0, 0, 0xab, 0xab, 0xab, 0xab, 0xab, 0xab, 0xab, 0xab, 0, 0}) {
		t.Fatal(m.bytes)
	}
	if !m.Init(0, 1, 3, []byte{4, 3, 2, 1}) || !bytes.Equal(m.bytes[:3], []byte{3, 2, 1}) {
		t.Fatal(m.bytes)
	}
	before := append([]byte(nil), m.bytes...)
	for _, r := range [][3]uint64{{10, 0, 3}, {0, 10, 3}, {math.MaxUint64, 0, 1}, {0, math.MaxUint64, 1}, {0, 0, math.MaxUint64}} {
		if m.Copy(r[0], r[1], r[2]) || m.Init(r[0], r[1], r[2], before) || !bytes.Equal(m.bytes, before) {
			t.Fatal(r, m.bytes)
		}
	}
	for _, r := range [][2]uint64{{13, 0}, {11, 2}, {math.MaxUint64, 1}, {1, math.MaxUint64}} {
		if m.Fill(r[0], 0, r[1]) || !bytes.Equal(m.bytes, before) {
			t.Fatal(r)
		}
	}
	if !m.Fill(12, 0, 0) || !m.Copy(12, 12, 0) || !m.Init(12, 0, 0, nil) || m.Init(0, 1, 0, nil) {
		t.Fatal("zero length bounds")
	}
}

type loadCase struct {
	op                   uint32
	size, width, lanes   int
	signed               bool
	scalar32, zero, lane bool
}

func loads() []loadCase {
	result := []loadCase{
		{op: 0x28, size: 4, scalar32: true}, {op: 0x29, size: 8}, {op: 0x2a, size: 4, scalar32: true}, {op: 0x2b, size: 8},
		{op: 0x2c, size: 1, signed: true, scalar32: true}, {op: 0x2d, size: 1, scalar32: true}, {op: 0x2e, size: 2, signed: true, scalar32: true}, {op: 0x2f, size: 2, scalar32: true},
		{op: 0x30, size: 1, signed: true}, {op: 0x31, size: 1}, {op: 0x32, size: 2, signed: true}, {op: 0x33, size: 2}, {op: 0x34, size: 4, signed: true}, {op: 0x35, size: 4},
		{op: 0xfd0000, size: 16},
		{op: 0xfd0001, size: 8, width: 1, lanes: 8, signed: true}, {op: 0xfd0002, size: 8, width: 1, lanes: 8},
		{op: 0xfd0003, size: 8, width: 2, lanes: 4, signed: true}, {op: 0xfd0004, size: 8, width: 2, lanes: 4},
		{op: 0xfd0005, size: 8, width: 4, lanes: 2, signed: true}, {op: 0xfd0006, size: 8, width: 4, lanes: 2},
		{op: 0xfd0007, size: 1, width: 1, lanes: 16}, {op: 0xfd0008, size: 2, width: 2, lanes: 8}, {op: 0xfd0009, size: 4, width: 4, lanes: 4}, {op: 0xfd000a, size: 8, width: 8, lanes: 2},
		{op: 0xfd005c, size: 4, zero: true}, {op: 0xfd005d, size: 8, zero: true},
	}
	for n := 0; n < 4; n++ {
		result = append(result, loadCase{op: 0xfd0054 + uint32(n), size: 1 << n, lane: true})
	}
	return result
}
func vectorBytes(v value.Value) []byte {
	b := make([]byte, 16)
	for i := 0; i < 8; i++ {
		b[i] = byte(v.Lo >> uint(8*i))
		b[i+8] = byte(v.Hi >> uint(8*i))
	}
	return b
}
func expectedLoad(tc loadCase, input []byte, previous []byte, laneIndex int) []byte {
	out := make([]byte, 16)
	switch {
	case tc.op < 0xfd0000:
		copy(out, input)
		end := 8
		if tc.scalar32 {
			end = 4
		}
		if tc.signed && input[len(input)-1]&0x80 != 0 {
			for i := len(input); i < end; i++ {
				out[i] = 255
			}
		}
	case tc.lane:
		copy(out, previous)
		copy(out[laneIndex*tc.size:], input)
	case tc.zero || tc.op == 0xfd0000:
		copy(out, input)
	case tc.op <= 0xfd0006:
		for i := 0; i < tc.lanes; i++ {
			src := input[i*tc.width : (i+1)*tc.width]
			dst := out[i*2*tc.width : (i+1)*2*tc.width]
			copy(dst, src)
			if tc.signed && src[tc.width-1]&0x80 != 0 {
				for j := tc.width; j < 2*tc.width; j++ {
					dst[j] = 255
				}
			}
		}
	default:
		for i := 0; i < tc.lanes; i++ {
			copy(out[i*tc.width:], input)
		}
	}
	return out
}
func TestEveryLoadAndBounds(t *testing.T) {
	m := &Memory{bytes: make([]byte, 64)}
	previous := value.Value{Lo: 0x123456789abcdef0, Hi: 0xfedcba9876543210}
	for _, tc := range loads() {
		for _, high := range []byte{0, 0x80} {
			for i := range m.bytes {
				m.bytes[i] = byte(i*13)&0x7f | high
			}
			lanes := 1
			if tc.lane {
				lanes = 16 / tc.size
			}
			for l := 0; l < lanes; l++ {
				for _, start := range []int{0, 1, 64 - tc.size} {
					expected := expectedLoad(tc, m.bytes[start:start+tc.size], vectorBytes(previous), l)
					for _, addr := range [][2]uint64{{uint64(start), 0}, {0, uint64(start)}} {
						got, trap := m.Load(tc.op, addr[0], addr[1], byte(l), previous)
						if trap != wasm.NoTrap || !bytes.Equal(vectorBytes(got), expected) {
							t.Fatalf("op=%x at=%v lane=%d got=%x want=%x trap=%v", tc.op, addr, l, vectorBytes(got), expected, trap)
						}
					}
				}
			}
			for _, addr := range [][2]uint64{{uint64(65 - tc.size), 0}, {64, 0}, {math.MaxUint32, 1}, {1, math.MaxUint32}, {math.MaxUint32, math.MaxUint32}} {
				if got, trap := m.Load(tc.op, addr[0], addr[1], 0, previous); trap != wasm.TrapMemoryBounds || got != (value.Value{}) {
					t.Fatal(tc, addr, got, trap)
				}
			}
		}
	}
}
func TestEveryStoreAndBounds(t *testing.T) {
	v := value.Value{Lo: 0x123456789abcdef0, Hi: 0xfedcba9876543210}
	raw := vectorBytes(v)
	cases := []struct {
		op   uint32
		size int
		lane bool
	}{{0x36, 4, false}, {0x37, 8, false}, {0x38, 4, false}, {0x39, 8, false}, {0x3a, 1, false}, {0x3b, 2, false}, {0x3c, 1, false}, {0x3d, 2, false}, {0x3e, 4, false}, {0xfd000b, 16, false}, {0xfd0058, 1, true}, {0xfd0059, 2, true}, {0xfd005a, 4, true}, {0xfd005b, 8, true}}
	for _, tc := range cases {
		lanes := 1
		if tc.lane {
			lanes = 16 / tc.size
		}
		for l := 0; l < lanes; l++ {
			for _, at := range []int{0, 1, 64 - tc.size} {
				m := &Memory{bytes: bytes.Repeat([]byte{0xaa}, 64)}
				want := append([]byte(nil), m.bytes...)
				start := 0
				if tc.lane {
					start = l * tc.size
				}
				copy(want[at:at+tc.size], raw[start:start+tc.size])
				if trap := m.Store(tc.op, 0, uint64(at), byte(l), v); trap != wasm.NoTrap || !bytes.Equal(m.bytes, want) {
					t.Fatal(tc, at, l, m.bytes, want, trap)
				}
			}
		}
		for _, addr := range [][2]uint64{{uint64(65 - tc.size), 0}, {64, 0}, {math.MaxUint32, 1}, {1, math.MaxUint32}, {math.MaxUint32, math.MaxUint32}} {
			m := &Memory{bytes: bytes.Repeat([]byte{0xaa}, 64)}
			before := append([]byte(nil), m.bytes...)
			if trap := m.Store(tc.op, addr[0], addr[1], 0, v); trap != wasm.TrapMemoryBounds || !bytes.Equal(m.bytes, before) {
				t.Fatal(tc, addr, trap)
			}
		}
	}
}
func TestInvalidInternalAccess(t *testing.T) {
	for _, call := range []func(){func() { read(nil, 3) }, func() { write(nil, 3, 0) }, func() { new(Memory).Load(0, 0, 0, 0, value.Value{}) }, func() { new(Memory).Load(0xfd000b, 0, 0, 0, value.Value{}) }, func() { new(Memory).Store(0, 0, 0, 0, value.Value{}) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("accepted internal invalid opcode/width")
				}
			}()
			call()
		}()
	}
}
func FuzzMemoryRanges(f *testing.F) {
	f.Add(uint64(0), uint64(64), uint64(1))
	f.Add(uint64(math.MaxUint64), uint64(0), uint64(math.MaxUint64))
	f.Fuzz(func(t *testing.T, dst, src, count uint64) {
		m := &Memory{bytes: bytes.Repeat([]byte{0xab}, 64)}
		for i := range m.bytes {
			m.bytes[i] = byte(i) ^ 0xab
		}
		before := append([]byte(nil), m.bytes...)
		valid := dst <= 64 && count <= 64-dst && src <= 64 && count <= 64-src
		if got := m.Copy(dst, src, count); got != valid {
			t.Fatal(got, valid)
		}
		if !valid && !bytes.Equal(m.bytes, before) {
			t.Fatal("partial out-of-bounds write")
		}
		if valid {
			want := append([]byte(nil), before...)
			for i := uint64(0); i < count; i++ {
				want[dst+i] = before[src+i]
			}
			if !bytes.Equal(m.bytes, want) {
				t.Fatal("overlap copy changed input before reading it")
			}
		}
		for _, tc := range loads() {
			m.Load(tc.op, uint64(uint32(dst)), uint64(uint32(src)), 0, value.Value{})
		}
	})
}
