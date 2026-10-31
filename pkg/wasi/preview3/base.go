package preview3

import (
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasi/preview2"
)

const (
	EnvironmentInterface    = "wasi:cli/environment@0.3.2"
	ExitInterface           = "wasi:cli/exit@0.3.2"
	RandomInterface         = "wasi:random/random@0.3.2"
	InsecureRandomInterface = "wasi:random/insecure@0.3.2"
	InsecureSeedInterface   = "wasi:random/insecure-seed@0.3.2"
)

// EnvironmentConfig grants explicit immutable process configuration.
type EnvironmentConfig = preview2.EnvironmentConfig
type Environment struct{ provider *preview2.Environment }

func DefaultEnvironmentConfig() EnvironmentConfig { return preview2.DefaultEnvironmentConfig() }
func NewEnvironment(config EnvironmentConfig) (*Environment, error) {
	p, err := preview2.NewEnvironment(config)
	if err != nil {
		return nil, err
	}
	return &Environment{p}, nil
}

// bindFunctions selects exact operations and their existing signature checks.
// It never exposes another version's instance or implicitly renames resources.
func bindFunctions(source component.Imports, names map[string]string) component.Imports {
	out := component.Imports{Functions: map[string]component.HostFunction{}, FunctionChecks: map[string]component.FunctionCheck{}}
	for to, from := range names {
		out.Functions[to] = source.Functions[from]
		out.FunctionChecks[to] = source.FunctionChecks[from]
	}
	return out
}
func (e *Environment) Imports() component.Imports {
	if e == nil || e.provider == nil {
		return component.Imports{}
	}
	source := e.provider.Imports().Instances[preview2.EnvironmentInterface]
	return component.Imports{Instances: map[string]component.Imports{EnvironmentInterface: bindFunctions(source, map[string]string{
		"get-environment": "get-environment", "get-arguments": "get-arguments", "get-initial-cwd": "initial-cwd",
	})}}
}

// ExitError distinguishes a guest-requested exit from a trap or cancellation.
type ExitError = preview2.ExitError

func ExitImports() component.Imports {
	source := preview2.ExitImports().Instances[preview2.ExitInterface]
	return component.Imports{Instances: map[string]component.Imports{ExitInterface: bindFunctions(source, map[string]string{"exit": "exit", "exit-with-code": "exit-with-code"})}}
}

// RandomConfig bounds the maximum batch, rather than rejecting larger requests.
// P3 permits a short read and requires a nonempty result for a positive request.
type RandomConfig = preview2.RandomConfig
type Random struct {
	provider *preview2.Random
	maximum  uint64
}

func NewRandom(config RandomConfig) (*Random, error) {
	p, err := preview2.NewRandom(config)
	if err != nil {
		return nil, err
	}
	maximum := config.MaxBytes
	if maximum == 0 {
		maximum = 64 << 10
	}
	return &Random{p, maximum}, nil
}
func (r *Random) Imports() component.Imports {
	if r == nil || r.provider == nil {
		return component.Imports{}
	}
	source := r.provider.Imports()
	out := component.Imports{Instances: map[string]component.Imports{}}
	for i, pair := range [][2]string{{RandomInterface, preview2.RandomInterface}, {InsecureRandomInterface, preview2.InsecureRandomInterface}} {
		prefix := "get-random-"
		if i == 1 {
			prefix = "get-insecure-random-"
		}
		p := bindFunctions(source.Instances[pair[1]], map[string]string{prefix + "bytes": prefix + "bytes", prefix + "u64": prefix + "u64"})
		read := p.Functions[prefix+"bytes"]
		p.Functions[prefix+"bytes"] = func(c *component.Caller, args []component.Value) (*component.Value, error) {
			return read(c, []component.Value{{Bits: min(args[0].Bits, r.maximum)}})
		}
		out.Instances[pair[0]] = p
	}
	out.Instances[InsecureSeedInterface] = bindFunctions(source.Instances[preview2.InsecureSeedInterface], map[string]string{"get-insecure-seed": "insecure-seed"})
	return out
}
