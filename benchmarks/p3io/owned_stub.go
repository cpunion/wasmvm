//go:build p3reference && !llgo

package main

import (
	"context"
	"errors"
	"github.com/cpunion/wasmvm/internal/bench/service"
)

func ownedKernels(context.Context, options, *report) ([]service.Call, func() error, error) {
	return nil, nil, errors.New("owned VM excluded from independent P3 reference build")
}
