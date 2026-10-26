package native

import (
	"encoding/binary"
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
)

// These independent C-ABI probes deliberately bypass the future Wasm compiler.
// They test real entry, record access, off-heap stores and resumption on each ISA.
func addProbe() []byte {
	if runtime.GOARCH == "amd64" {
		return []byte{
			0x48, 0x8b, 0x47, 0x60, // mov rax,[rdi+result_lo]
			0x48, 0x03, 0x47, 0x68, // add rax,[rdi+result_hi]
			0x48, 0x89, 0x47, 0x60, // mov [rdi+result_lo],rax
			0x48, 0x8b, 0x57, 0x08, // mov rdx,[rdi+arena]
			0x48, 0x89, 0x02, // mov [rdx],rax
			0x48, 0x83, 0x47, 0x38, 0x01, // add qword [rdi+continuation],1
			0x8b, 0x47, 0x70, 0xc3, // mov eax,[rdi+payload]; ret
		}
	}
	return armWords(0xf9403001, 0xf9403402, 0x8b020021, 0xf9003001,
		0xf9400403, 0xf9000061, 0xf9401c01, 0x91000421, 0xf9001c01, 0xb9407000, 0xd65f03c0)
}

func cancelProbe() []byte {
	if runtime.GOARCH == "amd64" {
		return []byte{0x48, 0x83, 0xbf, 0x90, 0, 0, 0, 0, 0x74, 0xf6, 0xb8, 6, 0, 0, 0, 0xc3}
	}
	// add x1,x0,#control; ldar x2,[x1]; cbz x2,-4; mov w0,#Poll; ret
	return armWords(0x91024001, 0xc8dffc22, 0xb4ffffe2, 0x528000c0, 0xd65f03c0)
}

func armWords(words ...uint32) []byte {
	out := make([]byte, len(words)*4)
	for i, w := range words {
		binary.LittleEndian.PutUint32(out[i*4:], w)
	}
	return out
}

func makeProbe(t *testing.T, bytes []byte) (*Code, *Lease, *Context) {
	t.Helper()
	code, err := Publish(bytes, []uint32{0}, 1024)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(code.Close)
	lease, err := code.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(lease.Close)
	ctx, err := NewContext(64, 64, 16)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ctx.Close)
	return code, lease, ctx
}

func TestAvailability(t *testing.T) {
	if Available {
		return
	}
	if _, err := Publish([]byte{0}, []uint32{0}, 1); err != ErrUnavailable {
		t.Fatal(err)
	}
	if _, err := NewContext(0, 0, 1); err != ErrUnavailable {
		t.Fatal(err)
	}
}

func TestLayout(t *testing.T) {
	if !Available {
		return
	}
	offsets := []uintptr{RecordSize, ArenaOffset, CapacityOffset, MaxDepthOffset, UsedOffset, DepthOffset,
		FunctionOffset, ContinuationOffset, SourceOffset, RootMapOffset, FuelOffset, BudgetOffset,
		ResultLoOffset, ResultHiOffset, PayloadOffset, ControlOffset}
	for i, want := range offsets {
		if got := layout(uint32(i)); got != uint64(want) {
			t.Fatalf("field %d: C %d Go %d", i, got, want)
		}
	}
	if layout(100) != ^uint64(0) {
		t.Fatal("invalid layout field")
	}
}

func TestEntryResumeAndRetirement(t *testing.T) {
	if !Available {
		return
	}
	bytes := addProbe()
	code, lease, ctx := makeProbe(t, bytes)
	clear(bytes) // Publication owns a copy.
	code.Close()
	if _, err := code.Acquire(); err != ErrClosed {
		t.Fatal(err)
	}
	state := State{ResultLo: 19, ResultHi: 23, RootMap: 7, Source: 91, Function: 3, Fuel: 100}
	for status := Return; status <= Move; status++ {
		state.Payload[0] = uint64(status)
		if err := ctx.SetState(state); err != nil {
			t.Fatal(err)
		}
		got, err := ctx.Enter(lease, 0)
		if err != nil || got != status {
			t.Fatalf("exit %v %v", got, err)
		}
		state, err = ctx.State()
		if err != nil {
			t.Fatal(err)
		}
		if state.ResultLo != 19+23*uint64(status) || state.Continuation != uint64(status) ||
			state.RootMap != 7 || state.Source != 91 || state.Function != 3 || state.Fuel != 100 {
			t.Fatal(state)
		}
		var stored [8]byte
		if err := ctx.Read(0, stored[:]); err != nil || binary.LittleEndian.Uint64(stored[:]) != state.ResultLo {
			t.Fatal(stored, err)
		}
	}
	lease.Close()
	lease.Close()
	if code.address != nil {
		t.Fatal("retired storage retained without leases")
	}
	if _, err := ctx.Enter(lease, 0); err != ErrClosed {
		t.Fatal(err)
	}
	code.Close()
}

func TestInvalidAndLimits(t *testing.T) {
	if !Available {
		return
	}
	bytes := addProbe()
	for _, tt := range []struct {
		bytes   []byte
		entries []uint32
		max     uint64
		err     error
	}{
		{nil, []uint32{0}, 100, ErrInvalid}, {bytes, nil, 100, ErrInvalid},
		{bytes, []uint32{uint32(len(bytes))}, 100, ErrInvalid}, {bytes, []uint32{0}, 0, ErrLimit},
	} {
		if _, err := Publish(tt.bytes, tt.entries, tt.max); err != tt.err {
			t.Fatal(err, tt.err)
		}
	}
	if runtime.GOARCH == "arm64" {
		if _, err := Publish(bytes, []uint32{1}, 100); err != ErrInvalid {
			t.Fatal(err)
		}
		if _, err := Publish(bytes[:len(bytes)-1], []uint32{0}, 100); err != ErrInvalid {
			t.Fatal(err)
		}
	}
	for _, args := range [][3]uint64{{1, 0, 1}, {0, 0, 0}, {^uint64(0), ^uint64(0), 1}} {
		if _, err := NewContext(args[0], args[1], args[2]); err != ErrLimit {
			t.Fatal(err)
		}
	}
	empty, err := NewContext(0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	empty.Close()
	_, lease, ctx := makeProbe(t, bytes)
	if _, err := ctx.Enter(nil, 0); err != ErrInvalid {
		t.Fatal(err)
	}
	var emptyCode Code
	if _, err := emptyCode.Acquire(); err != ErrClosed {
		t.Fatal(err)
	}
	if _, err := ctx.Enter(lease, 1); err != ErrInvalid {
		t.Fatal(err)
	}
	if err := ctx.SetState(State{Used: 65}); err != ErrLimit {
		t.Fatal(err)
	}
	if err := ctx.SetState(State{Depth: 17}); err != ErrLimit {
		t.Fatal(err)
	}
	if err := ctx.SetState(State{Used: 64, Depth: 16, Payload: [4]uint64{999}}); err != nil {
		t.Fatal(err)
	}
	if _, err := ctx.Enter(lease, 0); err != ErrInvalid {
		t.Fatal(err)
	}
	ctx.record.Version = 0
	if _, err := ctx.Enter(lease, 0); err != ErrInvalid {
		t.Fatal(err)
	}
	ctx.record.Version = ABIVersion
	ctx.Close()
	ctx.Close()
	ctx.Cancel()
	if ctx.record != nil || ctx.arena != nil {
		t.Fatal("context storage retained")
	}
	if _, err := ctx.State(); err != ErrClosed {
		t.Fatal(err)
	}
	if err := ctx.SetState(State{}); err != ErrClosed {
		t.Fatal(err)
	}
	if err := ctx.Read(0, nil); err != ErrClosed {
		t.Fatal(err)
	}
	if err := ctx.Write(0, nil); err != ErrClosed {
		t.Fatal(err)
	}
	if _, err := ctx.Enter(lease, 0); err != ErrClosed {
		t.Fatal(err)
	}
}

func TestMultipleEntriesAndFloatingEnvironment(t *testing.T) {
	if !Available {
		return
	}
	var fp []byte
	if runtime.GOARCH == "amd64" {
		// stmxcsr [rdi+result_lo]; fnstcw [rdi+result_hi]; mov eax,1; ret
		fp = []byte{0x0f, 0xae, 0x5f, 0x60, 0xd9, 0x7f, 0x68, 0xb8, 1, 0, 0, 0, 0xc3}
	} else {
		// mrs x1,fpcr; str x1,[x0,#result_lo]; mov w0,1; ret
		fp = armWords(0xd53b4401, 0xf9003001, 0x52800020, 0xd65f03c0)
	}
	bytes := append(addProbe(), fp...)
	code, err := Publish(bytes, []uint32{0, uint32(len(addProbe()))}, 1024)
	if err != nil {
		t.Fatal(err)
	}
	defer code.Close()
	lease, err := code.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	ctx, err := NewContext(8, 8, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	if status, err := ctx.Enter(lease, 1); status != Return || err != nil {
		t.Fatal(status, err)
	}
	owner, entry, err := lease.begin(1)
	if err != nil {
		t.Fatal(err)
	}
	good := probeFP(entry, unsafe.Pointer(ctx.record))
	owner.end()
	if !good {
		t.Fatal("host floating-point environment was not restored")
	}
	state, err := ctx.State()
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOARCH == "amd64" {
		if state.ResultLo&0xe040 != 0 || state.ResultHi&0xc00 != 0 {
			t.Fatalf("non-default MXCSR/x87: %+v", state)
		}
	} else if state.ResultLo&0x3c80000 != 0 {
		t.Fatalf("non-default FPCR: %+v", state)
	}
}

func TestGeneratedEntryPreservesRegisters(t *testing.T) {
	if !Available {
		return
	}
	var clobber []byte
	if runtime.GOARCH == "amd64" {
		// Deliberately overwrite every preserved integer register, including rbp.
		clobber = []byte{0x48, 0x31, 0xed, 0x48, 0x31, 0xdb, 0x4d, 0x31, 0xe4, 0x4d, 0x31, 0xed, 0x4d, 0x31, 0xf6, 0x4d, 0x31, 0xff,
			0x48, 0x89, 0xe1, 0x83, 0xe1, 0x0f, 0x48, 0x89, 0x4f, 0x78}
	} else {
		for r := uint32(19); r <= 28; r++ {
			clobber = append(clobber, armWords(0xd2800000|r)...)
		}
		for r := uint32(8); r <= 15; r++ {
			clobber = append(clobber, armWords(0x4f00e400|r)...)
		}
		clobber = append(clobber, armWords(0x910003e1, 0x92400c21, 0xf9003c01)...)
	}
	body := addProbe()
	ret := 1
	if runtime.GOARCH == "arm64" {
		ret = 4
	}
	body = append(clobber, body[:len(body)-ret]...)
	entry, err := nativeabi.Wrap(runtime.GOARCH, body, 4096)
	if err != nil {
		t.Fatal(err)
	}
	_, lease, ctx := makeProbe(t, entry.Code)
	if err := ctx.SetState(State{ResultLo: 20, ResultHi: 22, Payload: [4]uint64{uint64(Return), 99}}); err != nil {
		t.Fatal(err)
	}
	owner, address, err := lease.begin(0)
	if err != nil {
		t.Fatal(err)
	}
	good := probeRegisters(address, unsafe.Pointer(ctx.record))
	owner.end()
	if !good {
		t.Fatal("generated entry corrupted preserved registers")
	}
	state, err := ctx.State()
	if err != nil || state.ResultLo != 42 || state.Payload[1] != 0 {
		t.Fatalf("entry result/stack alignment: %+v %v", state, err)
	}
}

func TestFrameStorage(t *testing.T) {
	if !Available {
		return
	}
	_, _, ctx := makeProbe(t, addProbe())
	var zero [64]byte
	if err := ctx.Read(0, zero[:]); err != nil {
		t.Fatal(err)
	}
	if zero != [64]byte{} {
		t.Fatal("arena not zeroed")
	}
	source := []byte{1, 2, 3}
	if err := ctx.Write(61, source); err != nil {
		t.Fatal(err)
	}
	source[0] = 99
	if err := ctx.Read(61, source); err != nil || source[0] != 1 {
		t.Fatal(source, err)
	}
	for _, off := range []uint64{62, 65, ^uint64(0)} {
		if err := ctx.Read(off, source); err != ErrLimit {
			t.Fatal(err)
		}
		if err := ctx.Write(off, source); err != ErrLimit {
			t.Fatal(err)
		}
	}
	if err := ctx.Read(64, nil); err != nil {
		t.Fatal(err)
	}
	if err := ctx.Write(64, nil); err != nil {
		t.Fatal(err)
	}
}

func TestCancelCloseWhileRunning(t *testing.T) {
	if !Available {
		return
	}
	for _, closeContext := range []bool{false, true} {
		code, lease, ctx := makeProbe(t, cancelProbe())
		done := make(chan error, 1)
		go func() {
			status, err := ctx.Enter(lease, 0)
			if err == nil && status != Poll {
				err = ErrInvalid
			}
			done <- err
		}()
		deadline := time.Now().Add(5 * time.Second)
		for {
			ctx.mu.Lock()
			active := ctx.active
			ctx.mu.Unlock()
			if active {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("entry did not start")
			}
			runtime.Gosched()
		}
		if _, err := ctx.State(); err != ErrBusy {
			t.Fatal(err)
		}
		if err := ctx.SetState(State{}); err != ErrBusy {
			t.Fatal(err)
		}
		if err := ctx.Read(0, nil); err != ErrBusy {
			t.Fatal(err)
		}
		if err := ctx.Write(0, nil); err != ErrBusy {
			t.Fatal(err)
		}
		if _, err := ctx.Enter(lease, 0); err != ErrBusy {
			t.Fatal(err)
		}
		code.Close()
		lease.Close()
		code.mu.Lock()
		retained := code.address != nil
		code.mu.Unlock()
		if !retained {
			t.Fatal("active code unmapped")
		}
		// Host collectors must not need to trace the generated frame or registers.
		runtime.GC()
		if closeContext {
			ctx.Close()
		} else {
			ctx.Cancel()
		}
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("cancellation did not exit")
		}
		if code.address != nil {
			t.Fatal("active code not released after exit")
		}
		if closeContext && (ctx.record != nil || ctx.arena != nil) {
			t.Fatal("active context not released after exit")
		}
	}
}

func TestConcurrentPublicationAndReuse(t *testing.T) {
	if !Available {
		return
	}
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				code, err := Publish(addProbe(), []uint32{0}, 1024)
				if err != nil {
					t.Error(err)
					return
				}
				lease, err := code.Acquire()
				if err != nil {
					t.Error(err)
					return
				}
				ctx, err := NewContext(8, 8, 1)
				if err != nil {
					t.Error(err)
					return
				}
				if err = ctx.SetState(State{Payload: [4]uint64{uint64(Return)}}); err != nil {
					t.Error(err)
				}
				if status, err := ctx.Enter(lease, 0); err != nil || status != Return {
					t.Error(status, err)
				}
				ctx.Cancel()
				if status, err := ctx.Enter(lease, 0); err != nil || status != Poll {
					t.Error(status, err)
				}
				ctx.Close()
				lease.Close()
				code.Close()
			}
		}()
	}
	wg.Wait()
}
