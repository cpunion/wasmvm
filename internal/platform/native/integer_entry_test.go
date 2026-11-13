package native

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestIntegerEntryPreservesAmbientFloatingStateAndAdmission(t *testing.T) {
	if !Available {
		return
	}
	// This read-only diagnostic canary snapshots the ambient environment;
	// it is deliberately not an input to the compiler's integer proof.
	var body []byte
	if runtime.GOARCH == "amd64" {
		body = []byte{0x0f, 0xae, 0x5f, 0x60, 0xd9, 0x7f, 0x68, 0x8b, 0x47, 0x70, 0xc3}
	} else {
		body = armWords(0xd53b4401, 0xf9003001, 0xd53b4421, 0xf9003401, 0xb9407000, 0xd65f03c0)
	}
	code, err := PublishIntegerCode(body, []uint32{0}, 1024)
	if err != nil {
		t.Fatal(err)
	}
	defer code.Close()
	lease, err := code.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	ctx, err := NewContext(0, 16, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	ctx.record.Payload[0] = uint64(Return)
	if status, err := ctx.Enter(lease, 0); status != Return || err != nil {
		t.Fatal(status, err)
	}
	entry, err := lease.begin(0)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.end()
	for status := Exit(0); status <= Move; status++ {
		ctx.record.Payload[0] = uint64(status)
		if !probeIntegerFP(entry, unsafe.Pointer(ctx.record), status) {
			t.Fatal("ambient FP environment changed", status)
		}
		if ctx.record.ResultLo == 0 || runtime.GOARCH == "amd64" && ctx.record.ResultLo&0xe000 == 0 {
			t.Fatal("integer entry normalized the diagnostic host state")
		}
	}
	ctx.record.Version = 0
	if !probeIntegerFP(entry, unsafe.Pointer(ctx.record), 0) {
		t.Fatal("invalid version admitted")
	}
	ctx.record.Version = ABIVersion
	ctx.record.Used = ctx.record.Capacity + 1
	if !probeIntegerFP(entry, unsafe.Pointer(ctx.record), 0) {
		t.Fatal("invalid arena admitted")
	}
	ctx.record.Used = 0
	ctx.record.Depth = ctx.record.MaxDepth + 1
	if !probeIntegerFP(entry, unsafe.Pointer(ctx.record), 0) {
		t.Fatal("invalid depth admitted")
	}
	ctx.record.Depth = 0
	ctx.Cancel()
	if !probeIntegerFP(entry, unsafe.Pointer(ctx.record), Poll) {
		t.Fatal("cancelled entry admitted")
	}
}
