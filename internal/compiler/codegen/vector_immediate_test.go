package codegen

import (
	"encoding/binary"
	"math"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestVectorImmediateLLVMEncodings(t *testing.T) {
	for _, test := range []struct {
		bits uint64
		dst  byte
		word uint32
	}{
		{0x7070707070707070, 2, 0x4f03e602},
		{0x1200120012001200, 19, 0x4f00a653},
		{0x7fffffff7fffffff, 2, 0x6f046402},
		{0x007fffff007fffff, 2, 0x4f03d7e2},
		{0x3f8000003f800000, 2, 0x4f03f602},
		{0xc008000000000000, 19, 0x6f04f513},
	} {
		e := functionEmitter{a: &assembler{arch: "arm64"}}
		e.vectorConstant(test.dst, test.bits, test.bits)
		if len(e.a.code) != 4 || binary.LittleEndian.Uint32(e.a.code) != test.word {
			t.Fatalf("%+v: %x", test, e.a.code)
		}
	}
}

func TestVectorImmediateExpansionAndRawFallback(t *testing.T) {
	type sample struct {
		bits   [2]uint64
		single bool
	}
	var cases []sample
	repeat := func(lane uint64, width uint, single bool) {
		var data [16]byte
		for i := uint(0); i < 128/width; i++ {
			for b := uint(0); b < width/8; b++ {
				data[i*width/8+b] = byte(lane >> (b * 8))
			}
		}
		cases = append(cases, sample{[2]uint64{binary.LittleEndian.Uint64(data[:8]), binary.LittleEndian.Uint64(data[8:])}, single})
	}
	for imm := uint64(0); imm < 256; imm++ {
		repeat(imm, 8, true)
		for _, width := range []uint{16, 32} {
			mask := uint64(1)<<width - 1
			for shift := uint(0); shift < width; shift += 8 {
				lane := imm << shift
				repeat(lane, width, true)
				repeat(mask^lane, width, true)
			}
		}
		for _, shift := range []uint{8, 16} {
			lane := imm<<shift | (1<<shift - 1)
			repeat(lane, 32, true)
			repeat(0xffffffff^lane, 32, true)
		}
		var byteMask uint64
		for bit := uint(0); bit < 8; bit++ {
			if imm&(1<<bit) != 0 {
				byteMask |= uint64(255) << (bit * 8)
			}
		}
		repeat(byteMask, 64, true)
		// Independent mathematical FP oracle: (1 + fraction/16) * 2^exponent.
		power := int(imm>>4&3) + 1
		if imm&64 != 0 {
			power -= 4
		}
		fp := math.Ldexp(float64(16+(imm&15)), power-4)
		if imm&128 != 0 {
			fp = -fp
		}
		repeat(uint64(math.Float32bits(float32(fp))), 32, true)
		repeat(math.Float64bits(fp), 64, true)
	}
	for _, bits := range [][2]uint64{
		{0, 1}, {0xffffffffffffffff, 0}, {0x7ff8000000000001, 0x7ff8000000000001},
		{0x8000000000000000, 0x8000000000000000}, {1, 1},
	} {
		cases = append(cases, sample{bits, false})
	}
	seed := uint64(0x9e3779b97f4a7c15)
	for range 128 {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		cases = append(cases, sample{[2]uint64{seed, ^seed}, false})
	}

	var code []byte
	var entries []uint32
	for i, test := range cases {
		if _, single := armVectorImmediate(test.bits[0], test.bits[1]); single != test.single {
			t.Fatalf("unrepresentable or missed immediate: %#x, single=%v", test.bits, single)
		}
		if !native.Available {
			continue
		}
		// Batch independent native-ABI probes into one publication. The ordinary
		// compiler tests separately cover pool ownership, aliases and cold entries.
		a := assembler{arch: runtime.GOARCH, labels: map[int]int{}}
		e := functionEmitter{a: &a, exit: a.label()}
		dst := []byte{0, 2, 12, 15}[i%4]
		if runtime.GOARCH == "arm64" {
			dst = []byte{0, 2, 19, 31}[i%4]
		}
		e.vectorConstant(dst, test.bits[0], test.bits[1])
		if runtime.GOARCH == "arm64" {
			if test.single && len(a.code) != 4 {
				t.Fatal("modified immediate expanded to multiple instructions")
			}
			a.word(0x3d800000 | uint32(nativeabi.PayloadOffset/16)<<10 | uint32(dst))
		} else {
			a.bytes(0xf3)
			if dst >= 8 {
				a.bytes(0x44)
			}
			a.bytes(0x0f, 0x7f, 0x87|(dst&7)<<3)
			a.word(uint32(nativeabi.PayloadOffset))
		}
		e.rawStatus(native.Return)
		a.mark(e.exit)
		body, err := a.finish()
		if err != nil {
			t.Fatal(err)
		}
		entry, err := nativeabi.Wrap(runtime.GOARCH, body, DefaultLimits().MaxCodeBytes)
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, uint32(len(code)))
		code = append(code, entry.Code...)
	}
	if !native.Available {
		return
	}
	publication, err := native.Publish(code, entries, uint64(len(code)))
	if err != nil {
		t.Fatal(err)
	}
	defer publication.Close()
	lease, err := publication.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	ctx, err := native.NewContext(0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	for i, test := range cases {
		if exit, err := ctx.Enter(lease, uint32(i)); err != nil || exit != native.Return {
			t.Fatal(i, exit, err)
		}
		state, err := ctx.State()
		if err != nil || state.Payload[0] != test.bits[0] || state.Payload[1] != test.bits[1] {
			t.Fatal(i, test.bits, state.Payload, err)
		}
	}
}
