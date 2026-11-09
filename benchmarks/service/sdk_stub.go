//go:build llgo || (!wasmtime && !wasmer && !wasmedge)

package main

import (
	"errors"

	"github.com/cpunion/wasmvm/internal/bench/service"
)

func sdkKernels(options, *report) ([]service.Call, func() error, error) {
	return nil, nil, errors.New("SDK reference requires a Go build with exactly one wasmtime/wasmer/wasmedge tag and its pinned SDK")
}
