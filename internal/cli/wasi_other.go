//go:build !linux && !darwin

package cli

import (
	"context"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"io"
)

func runWASI(_ context.Context, _ *vm.Engine, _ *vm.Store, _ *vm.CompiledModule, _ vm.CallOptions, _ wasiOptions, _ io.Writer, stderr io.Writer) int {
	return failure(stderr, wasm.ErrUnsupported)
}
