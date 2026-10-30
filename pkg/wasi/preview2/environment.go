package preview2

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const EnvironmentInterface = "wasi:cli/environment@0.2.12"

type EnvironmentConfig struct {
	Args, Env  []string
	InitialCWD *string
	MaxBytes   uint64
	MaxEntries uint32
}

func DefaultEnvironmentConfig() EnvironmentConfig {
	return EnvironmentConfig{MaxBytes: 1 << 20, MaxEntries: 4096}
}

// Environment owns an immutable, explicit configuration. It never inherits the
// host process's arguments, environment or current directory implicitly.
type Environment struct {
	args []string
	env  [][2]string
	cwd  *string
}

func NewEnvironment(config EnvironmentConfig) (*Environment, error) {
	defaults := DefaultEnvironmentConfig()
	if config.MaxBytes == 0 {
		config.MaxBytes = defaults.MaxBytes
	}
	if config.MaxEntries == 0 {
		config.MaxEntries = defaults.MaxEntries
	}
	if uint64(len(config.Args))+uint64(len(config.Env)) > uint64(config.MaxEntries) {
		return nil, wasm.ErrLimit
	}
	var bytes uint64
	valid := func(s string) error {
		if !utf8.ValidString(s) || strings.IndexByte(s, 0) >= 0 {
			return wasm.ErrInvalid
		}
		if uint64(len(s)) > config.MaxBytes-bytes {
			return wasm.ErrLimit
		}
		bytes += uint64(len(s))
		return nil
	}
	e := &Environment{args: append([]string(nil), config.Args...)}
	for _, s := range e.args {
		if err := valid(s); err != nil {
			return nil, err
		}
	}
	for _, s := range config.Env {
		if err := valid(s); err != nil {
			return nil, err
		}
		name, value, found := strings.Cut(s, "=")
		if !found || name == "" {
			return nil, wasm.ErrInvalid
		}
		e.env = append(e.env, [2]string{name, value})
	}
	if config.InitialCWD != nil {
		s := *config.InitialCWD
		if err := valid(s); err != nil {
			return nil, err
		}
		e.cwd = &s
	}
	return e, nil
}

func environmentType(name string) component.FunctionCheck {
	return func(f component.FunctionInfo) error {
		if f.ParameterCount() != 0 {
			return vm.ErrType
		}
		t := f.Result()
		switch name {
		case "get-arguments":
			if t.Kind() == component.KindList && t.Element().Kind() == component.KindString {
				return nil
			}
		case "get-environment":
			if t.Kind() == component.KindList && t.Element().Kind() == component.KindTuple && t.Element().FieldCount() == 2 {
				_, key := t.Element().Field(0)
				_, value := t.Element().Field(1)
				if key.Kind() == component.KindString && value.Kind() == component.KindString {
					return nil
				}
			}
		case "initial-cwd":
			if t.Kind() == component.KindOption && t.Element().Kind() == component.KindString {
				return nil
			}
		}
		return vm.ErrType
	}
}

func (e *Environment) arguments(ctx context.Context) (*component.Value, error) {
	out := component.Value{Elements: make([]component.Value, len(e.args))}
	for i, s := range e.args {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		out.Elements[i] = component.Value{Text: s}
	}
	return &out, ctx.Err()
}

func (e *Environment) variables(ctx context.Context) (*component.Value, error) {
	out := component.Value{Elements: make([]component.Value, len(e.env))}
	for i, pair := range e.env {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		out.Elements[i] = component.Value{Elements: []component.Value{{Text: pair[0]}, {Text: pair[1]}}}
	}
	return &out, ctx.Err()
}

// Imports returns the complete versioned environment interface, including
// signature contracts checked before instantiation can execute guest effects.
func (e *Environment) Imports() component.Imports {
	functions := map[string]component.HostFunction{
		"get-arguments": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			return e.arguments(c.Context())
		},
		"get-environment": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			return e.variables(c.Context())
		},
		"initial-cwd": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			if err := c.Context().Err(); err != nil {
				return nil, err
			}
			v := &component.Value{}
			if e.cwd != nil {
				v.Case = 1
				v.Elements = []component.Value{{Text: *e.cwd}}
			}
			return v, nil
		},
	}
	checks := map[string]component.FunctionCheck{}
	for name := range functions {
		checks[name] = environmentType(name)
	}
	return component.Imports{Instances: map[string]component.Imports{EnvironmentInterface: {Functions: functions, FunctionChecks: checks}}}
}
