package codegen

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func integerOps() []uint32 {
	var ops []uint32
	for op := uint32(0x45); op <= 0x5a; op++ {
		ops = append(ops, op)
	}
	for op := uint32(0x67); op <= 0x8a; op++ {
		ops = append(ops, op)
	}
	ops = append(ops, 0xa7, 0xac, 0xad, 0xc0, 0xc1, 0xc2, 0xc3, 0xc4)
	return ops
}

func TestIntegerTargetsAndBudget(t *testing.T) {
	if len(integerOps()) != 66 {
		t.Fatal("integer instruction inventory")
	}
	for _, arch := range []string{"amd64", "arm64"} {
		for _, op := range integerOps() {
			entry, err := Integer(arch, op, 4096)
			if err != nil {
				t.Fatal(arch, op, err)
			}
			if _, err := Integer(arch, op, uint64(len(entry.Code))); err != nil {
				t.Fatal(err)
			}
			if _, err := Integer(arch, op, uint64(len(entry.Code)-1)); err != wasm.ErrLimit {
				t.Fatal(err)
			}
		}
	}
	if _, err := Integer("unsupported", 0x6a, 4096); err != wasm.ErrUnsupported {
		t.Fatal(err)
	}
	for _, op := range []uint32{0, 0x44, 0x5b, 0x66, 0x8b, 0xa6, 0xa8, 0xae, 0xc5, 0xfd0000} {
		if _, err := Integer("amd64", op, 4096); err != wasm.ErrUnsupported {
			t.Fatal(op, err)
		}
	}
}

func TestIntegerNativeSemantics(t *testing.T) {
	if !native.Available {
		return
	}
	ctx, err := native.NewContext(0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	edges := []uint64{0, 1, 2, 31, 32, 63, 64, 255, 256, 32767, 32768, 65535, 65536, 0x7fffffff,
		0x80000000, 0xffffffff, 0x100000000, 0x7fffffffffffffff, 0x8000000000000000, ^uint64(0), 0xffff000000000080}
	var pairs [][2]uint64
	for _, a := range edges {
		for _, b := range edges {
			pairs = append(pairs, [2]uint64{a, b})
		}
	}
	seed := uint64(0x123456789abcdef0)
	next := func() uint64 { seed ^= seed << 13; seed ^= seed >> 7; seed ^= seed << 17; return seed }
	for i := 0; i < 64; i++ {
		pairs = append(pairs, [2]uint64{next(), next()})
	}
	for _, op := range integerOps() {
		t.Run(fmt.Sprintf("%x", op), func(t *testing.T) {
			entry, err := Integer(runtime.GOARCH, op, 4096)
			if err != nil {
				t.Fatal(err)
			}
			code, err := native.Publish(entry.Code, []uint32{0}, 4096)
			if err != nil {
				t.Fatal(err)
			}
			defer code.Close()
			lease, err := code.Acquire()
			if err != nil {
				t.Fatal(err)
			}
			defer lease.Close()
			for _, p := range pairs {
				want, trap := numeric.Eval(op, p[0], p[1])
				if err := ctx.SetState(native.State{ResultLo: p[0], ResultHi: p[1], Source: 99, RootMap: 17}); err != nil {
					t.Fatal(err)
				}
				status, err := ctx.Enter(lease, 0)
				if err != nil {
					t.Fatal(err)
				}
				state, err := ctx.State()
				if err != nil {
					t.Fatal(err)
				}
				if state.Source != 99 || state.RootMap != 17 {
					t.Fatal("unrelated record fields overwritten", state)
				}
				if trap != wasm.NoTrap {
					if status != native.Trap || state.Payload[0] != uint64(trap) {
						t.Fatalf("%x(%x,%x): exit %v trap %d, want %v", op, p[0], p[1], status, state.Payload[0], trap)
					}
				} else if status != native.Return || state.ResultLo != want {
					t.Fatalf("%x(%x,%x): exit %v value %x, want %x", op, p[0], p[1], status, state.ResultLo, want)
				}
			}
		})
	}
	t.Logf("executed %d integer instructions x %d input pairs on %s/%s (%s)", len(integerOps()), len(pairs), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
}

func TestBranchRelocations(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, conditional := range []bool{false, true} {
			a := assembler{arch: arch, labels: map[int]int{}}
			start, end := a.label(), a.label()
			a.mark(start)
			a.branch(end, conditional, 1)
			a.branch(start, conditional, 1)
			a.mark(end)
			if _, err := a.finish(); err != nil {
				t.Fatal(err)
			}
			for _, r := range a.fixups {
				word := binary.LittleEndian.Uint32(a.code[r.position:])
				var target int
				if arch == "amd64" {
					target = r.position + 4 + int(int32(word))
				} else if conditional {
					target = r.position + int(int32(word<<8)>>13)*4
				} else {
					target = r.position + int(int32(word<<6)>>6)*4
				}
				if target != a.labels[r.label] {
					t.Fatalf("%s branch target %d, want %d", arch, target, a.labels[r.label])
				}
			}
			delete(a.labels, start)
			if _, err := a.finish(); err != wasm.ErrInvalid {
				t.Fatal(err)
			}
		}
	}
	for _, tt := range []struct {
		arch        string
		conditional bool
		target      int
		err         error
	}{
		{"amd64", false, 1 << 32, wasm.ErrLimit}, {"arm64", false, 1 << 28, wasm.ErrLimit},
		{"arm64", true, 1 << 21, wasm.ErrLimit}, {"arm64", false, 3, wasm.ErrInvalid},
	} {
		a := assembler{arch: tt.arch, labels: map[int]int{1: tt.target}}
		a.branch(1, tt.conditional, 0)
		if _, err := a.finish(); err != tt.err {
			t.Fatal(tt, err)
		}
	}
}
