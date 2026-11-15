package codegen

import (
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestChildInitializationPreservesParametersAndFrameBoundaries(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, test := range []struct{ parameters, locals uint32 }{{0, 0}, {0, 1}, {1, 0}, {1, 5}, {9, 257}, {256, 257}, {4097, 513}} {
			a := assembler{arch: arch, longBranches: true}
			e := functionEmitter{a: &a, exit: a.label()}
			cancelled := a.label()
			e.callLoad(4, e.callContext(), uint32(nativeabi.ArenaOffset))
			e.callAdvance(4, 16)
			e.contextLoad(0, nativeabi.PayloadOffset)
			e.contextStore(0, nativeabi.ControlOffset)
			e.callZero(test.parameters+test.locals, test.parameters, cancelled)
			e.rawStatus(nativeabi.Return)
			a.mark(cancelled)
			e.rawStatus(nativeabi.Poll)
			a.mark(e.exit)
			body, err := a.finish()
			if err != nil {
				t.Fatal(arch, test, err)
			}
			entry, err := nativeabi.Wrap(arch, body, 16384)
			if err != nil {
				t.Fatal(err)
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
			for _, cancel := range []bool{false, true} {
				total := test.parameters + test.locals + 2
				ctx, err := native.NewContext(uint64(total)*16, uint64(total)*16, 1)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(ctx.Close)
				for i := uint32(0); i < total; i++ {
					writeSlot(t, ctx, 0, slots.Slot(i), value.Value{Lo: uint64(i) + 1, Hi: ^uint64(i)})
				}
				state := native.State{Depth: 1}
				if cancel {
					state.Payload[0] = 1
				}
				status, err := ctx.RunState(lease, 0, &state)
				wantStatus := native.Return
				cleared := test.locals
				if cancel && cleared >= 256 {
					wantStatus, cleared = native.Poll, 256
				}
				if err != nil || status != wantStatus {
					t.Fatal(arch, test, cancel, status, err)
				}
				for i := uint32(0); i < total; i++ {
					want := value.Value{Lo: uint64(i) + 1, Hi: ^uint64(i)}
					if i > test.parameters && i <= test.parameters+cleared {
						want = value.Value{}
					}
					if got := readSlot(t, ctx, 0, slots.Slot(i)); got != want {
						t.Fatal("parameter, local or guard corrupted", arch, test, cancel, i, got, want)
					}
				}
			}
		}
	}
}
