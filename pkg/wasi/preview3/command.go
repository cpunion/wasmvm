package preview3

import (
	"context"
	"errors"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
)

const CLICommandInterface = "wasi:cli/run@0.3.2"

// RunCommand checks the versioned native async CLI export before executing it.
// A normal error result maps to status 1; explicit exit preserves its code.
// The embedder retains responsibility for Instance and provider shutdown.
func RunCommand(ctx context.Context, in *component.Instance, options component.CallOptions) (uint32, error) {
	if in == nil {
		return 0, vm.ErrReference
	}
	f, err := in.Function(CLICommandInterface, "run")
	if err != nil {
		return 0, err
	}
	info := f.Info()
	result := info.Result()
	if !info.Async() || info.ParameterCount() != 0 || result.Kind() != component.KindResult || result.CaseCount() != 2 {
		return 0, vm.ErrType
	}
	for i := 0; i < 2; i++ {
		if _, payload := result.Case(i); payload.Valid() {
			return 0, vm.ErrType
		}
	}
	v, err := f.Call(ctx, nil, options)
	var exit *ExitError
	if errors.As(err, &exit) {
		return exit.Code, nil
	}
	if err != nil {
		return 0, err
	}
	return v.Value.Case, nil
}
