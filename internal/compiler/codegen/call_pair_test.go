package codegen

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestCallPairsPreserveBothWordsAndAdjacentBytes(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, offset := range []uint32{0, 496, 504, 512, 65536} {
			a := assembler{arch: arch, longBranches: true}
			e := functionEmitter{a: &a, exit: a.label()}
			e.callLoad(4, e.callContext(), uint32(nativeabi.ArenaOffset))
			e.callPair(0, 1, e.callReg(4), offset, false)
			e.callPair(0, 1, e.callReg(4), offset+32, true)
			e.rawStatus(nativeabi.Return)
			a.mark(e.exit)
			body, err := a.finish()
			if err != nil {
				t.Fatal(arch, offset, err)
			}
			entry, err := nativeabi.Wrap(arch, body, 16384)
			if err != nil {
				t.Fatal(arch, offset, err)
			}
			if arch != runtime.GOARCH || !native.Available {
				continue
			}
			code, err := native.Publish(entry.Code, []uint32{0}, uint64(len(entry.Code)))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(code.Close)
			lease, err := code.Acquire()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(lease.Close)
			capacity := uint64(offset) + 64
			ctx, err := native.NewContext(capacity, capacity, 1)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(ctx.Close)
			data := bytes.Repeat([]byte{0xa5}, int(capacity))
			binary.LittleEndian.PutUint64(data[offset:], 0xfedcba9876543210)
			binary.LittleEndian.PutUint64(data[offset+8:], 0x8123456789abcdef)
			if err := ctx.Write(0, data); err != nil {
				t.Fatal(err)
			}
			state := native.State{Depth: 1}
			if status, err := ctx.RunState(lease, 0, &state); err != nil || status != native.Return {
				t.Fatal(arch, offset, status, err)
			}
			copy(data[offset+32:offset+48], data[offset:offset+16])
			got := make([]byte, len(data))
			if err := ctx.Read(0, got); err != nil || !bytes.Equal(got, data) {
				t.Fatal("pair copy lost opaque words or adjacent memory", arch, offset, err)
			}
		}
	}
}
