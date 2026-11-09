package main

/*
#cgo CFLAGS: -O3 -std=gnu11 -D_POSIX_C_SOURCE=200809L
#include <stdint.h>
uint32_t run(uint32_t seed, uint32_t count);
*/
import "C"

import (
	"fmt"
	"time"

	"github.com/cpunion/wasmvm/internal/bench/service"
)

func nativeKernels(args options) ([]service.Call, func() error, error) {
	calls := make([]service.Call, args.Workers)
	for i := range calls {
		calls[i] = func() (int64, uint32, error) {
			start := time.Now()
			value := uint32(C.run(C.uint32_t(args.seed), C.uint32_t(args.iterations)))
			ns := time.Since(start).Nanoseconds()
			if value != args.Expected {
				return ns, value, fmt.Errorf("native checksum: got %d, expected %d", value, args.Expected)
			}
			return ns, value, nil
		}
	}
	return calls, func() error { return nil }, nil
}
