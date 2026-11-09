//go:build !llgo && (wasmtime || wasmer || wasmedge)

package main

/*
#cgo CFLAGS: -O3 -std=gnu11 -D_POSIX_C_SOURCE=200809L
#include "sdk.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/bench/service"
)

func sdkIdentity() (string, string) { return C.GoString(C.sdk_name()), C.GoString(C.sdk_version()) }

func sdkKernels(args options, out *report) ([]service.Call, func() error, error) {
	name, version := sdkIdentity()
	if name != args.engine {
		return nil, nil, fmt.Errorf("adapter contains %s, requested %s", name, args.engine)
	}
	pins := map[string]string{"wasmtime": "48.0.1", "wasmer": "7.3.0", "wasmedge": "0.17.1"}
	if version != pins[name] {
		return nil, nil, fmt.Errorf("%s SDK version %s differs from pin %s", name, version, pins[name])
	}
	start := time.Now()
	data, err := os.ReadFile(args.path)
	out.Phases["load"] = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, nil, err
	}
	if err := pureModule(data); err != nil {
		return nil, nil, err
	}
	owner := C.sdk_open((*C.uint8_t)(unsafe.Pointer(&data[0])), C.size_t(len(data)), C.uint(args.Workers))
	runtime.KeepAlive(data)
	if owner == nil {
		return nil, nil, errors.New("cannot allocate SDK owner")
	}
	release := func() error { C.sdk_close(owner); return nil }
	for index, phase := range []string{"engineSetup", "compileValidate", "instantiate", "exportLookup"} {
		out.Phases[phase] = int64(C.sdk_phase(owner, C.uint(index)))
	}
	if message := C.GoString(C.sdk_error(owner, C.uint(args.Workers))); message != "" {
		return nil, release, errors.New(message)
	}
	out.Policy = fmt.Sprintf("%s %s %s through Go C bridge; one immutable compiled module/AST, independent store/executor per worker; default unmetered policy; no disk compilation cache", name, version, args.backend)
	calls := make([]service.Call, args.Workers)
	for worker := range calls {
		calls[worker] = func() (int64, uint32, error) {
			start := time.Now()
			packed := uint64(C.sdk_call(owner, C.uint(worker), C.uint32_t(args.seed), C.uint32_t(args.iterations)))
			ns := time.Since(start).Nanoseconds()
			value := uint32(packed)
			if packed>>32 != 0 {
				return ns, value, errors.New(C.GoString(C.sdk_error(owner, C.uint(worker))))
			}
			if value != args.Expected {
				return ns, value, fmt.Errorf("%s checksum: got %d, expected %d", name, value, args.Expected)
			}
			return ns, value, nil
		}
	}
	return calls, release, nil
}
