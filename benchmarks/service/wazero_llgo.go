//go:build llgo

package main

import (
	"context"
	"errors"

	"github.com/cpunion/wasmvm/internal/bench/service"
)

// The independent wazero reference is built by Go. This LLGo binary measures
// only ordinary native C under LLGo's scheduler; it never substitutes an engine.
func wazeroKernels(context.Context, options, *report) ([]service.Call, func() error, error) {
	return nil, nil, errors.New("wazero service reference requires the Go-built adapter")
}
