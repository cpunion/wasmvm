package preview2

import (
	"fmt"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
)

// ExitInterface is the implemented, pinned WASI exit interface.
const ExitInterface = "wasi:cli/exit@0.2.12"

// ExitError reports a guest-requested exit, including a successful exit (zero).
// Use errors.As to distinguish this outcome from a trap or cancellation.
type ExitError struct{ Code uint32 }

func (e *ExitError) Error() string { return fmt.Sprintf("WASI component exited with code %d", e.Code) }

func exitType(code bool) component.FunctionCheck {
	return func(f component.FunctionInfo) error {
		if f.ParameterCount() != 1 || f.Result().Valid() {
			return vm.ErrType
		}
		_, status := f.Parameter(0)
		if code {
			if status.Kind() == component.KindU8 {
				return nil
			}
			return vm.ErrType
		}
		if status.Kind() != component.KindResult || status.CaseCount() != 2 {
			return vm.ErrType
		}
		_, ok := status.Case(0)
		_, failed := status.Case(1)
		if ok.Valid() || failed.Valid() {
			return vm.ErrType
		}
		return nil
	}
}

// ExitImports implements both nonreturning operations. Closing the shared
// component Store stops every linked instance, not only the calling namespace.
func ExitImports() component.Imports {
	leave := func(c *component.Caller, code uint32) (*component.Value, error) {
		if err := c.CloseInstance(); err != nil {
			return nil, err
		}
		return nil, &ExitError{Code: code}
	}
	return component.Imports{Instances: map[string]component.Imports{ExitInterface: {
		Functions: map[string]component.HostFunction{
			"exit": func(c *component.Caller, args []component.Value) (*component.Value, error) {
				return leave(c, args[0].Case)
			},
			"exit-with-code": func(c *component.Caller, args []component.Value) (*component.Value, error) {
				return leave(c, uint32(args[0].Bits))
			},
		},
		FunctionChecks: map[string]component.FunctionCheck{"exit": exitType(false), "exit-with-code": exitType(true)},
	}}}
}
