//go:build !p3reference || llgo || p3host

package main

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/bench/service"
)

func referenceKernels(options, *report) ([]service.Call, func() error, error) {
	return nil, nil, errors.New("P3 reference requires a Go p3reference build and the pinned Rust bridge")
}
