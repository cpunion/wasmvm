//go:build p3reference && !llgo && !p3host

package main

/*
#cgo CFLAGS: -O3 -std=gnu11
#include "p3_reference.h"
*/
import "C"

import (
	"errors"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/bench/service"
)

func referenceKernels(args options, out *report) ([]service.Call, func() error, error) {
	start := time.Now()
	data, err := os.ReadFile(args.path)
	out.Phases["load"] = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, nil, err
	}
	if len(data) == 0 {
		return nil, nil, errors.New("empty component")
	}
	owner := C.wasmvm_p3_open((*C.uint8_t)(unsafe.Pointer(&data[0])), C.size_t(len(data)), (*C.uint8_t)(unsafe.Pointer(&args.payload[0])), C.size_t(len(args.payload)), C.uint32_t(args.Workers))
	runtime.KeepAlive(data)
	runtime.KeepAlive(args.payload)
	if owner == nil {
		return nil, nil, errors.New("P3 reference owner allocation failed")
	}
	release := func() error { C.wasmvm_p3_close(owner); return nil }
	for i, name := range []string{"engineSetup", "compileValidate", "providerSetup", "instantiate", "exportLookup", "executorSetup"} {
		out.Phases[name] = int64(C.wasmvm_p3_phase(owner, C.uint32_t(i)))
	}
	if message := C.GoString(C.wasmvm_p3_error(owner, C.uint32_t(args.Workers))); message != "" {
		return nil, release, errors.New(message)
	}
	calls := make([]service.Call, args.Workers)
	for i := range calls {
		calls[i] = func() (int64, uint32, error) {
			var reply C.wasmvm_p3_reply
			start := time.Now()
			status := C.wasmvm_p3_call(owner, C.uint32_t(i), &reply)
			ns := time.Since(start).Nanoseconds()
			if status != 0 {
				return ns, 0, errors.New(C.GoString(C.wasmvm_p3_error(owner, C.uint32_t(i))))
			}
			if uint64(reply.length) != uint64(len(args.payload)) || reply.bytes == nil {
				return ns, 0, errors.New("invalid reference body length")
			}
			// The private worker cannot call again before this immutable lease is
			// checked; joined driver shutdown precedes owner/Store destruction.
			data := unsafe.Slice((*byte)(unsafe.Pointer(reply.bytes)), len(args.payload))
			guest, trailer := "", ""
			if reply.guest != 0 {
				guest = "rust-p3"
			}
			if reply.trailer != 0 {
				trailer = "done"
			}
			if err := checkReply(args, data, int(reply.status), guest, trailer); err != nil {
				return ns, 0, err
			}
			runtime.KeepAlive(owner)
			return ns, args.Expected, nil
		}
	}
	out.Policy = "Wasmtime48.0.1 Cranelift speed via Go C/Rust bridge; P3 Service::handle and complete body/trailer collection inside Store::run_concurrent; independent Store/current-thread Tokio executor per worker sharing compiled component; default unmetered policy; generated export lookup cached; no network socket"
	return calls, release, nil
}
