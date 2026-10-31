// Package component executes component composition and suspended canonical calls
// through the owned Core interpreter or native compiler. Capabilities
// without complete execution contracts are rejected before guest code runs.
package component

import (
	"context"
	"fmt"
	"sync"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/decode"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Value = canonical.Value
type CallOptions = vm.CallOptions
type DecodeLimits = decode.Limits
type ValidationLimits = validate.Limits

type Config struct {
	Core                        vm.EngineConfig
	Store                       vm.StoreConfig
	Decode                      DecodeLimits
	Validate                    ValidationLimits
	MaxValues                   uint64
	MaxBytes                    uint64
	MaxNesting                  uint32
	MaxResources                uint32
	MaxTasks                    uint32
	DiscardErrorContextMessages bool
}

func DefaultConfig() Config {
	core := vm.DefaultEngineConfig()
	core.Profile, core.Features = "core3", wasm.SupportedFeatures
	return Config{Core: core, Store: vm.DefaultStoreConfig(), Decode: decode.DefaultLimits(), Validate: validate.DefaultLimits(), MaxValues: 1 << 20, MaxBytes: 64 << 20, MaxNesting: 128, MaxResources: 1 << 16, MaxTasks: 1024}
}

type Engine struct {
	mu     sync.Mutex
	closed bool
	core   *vm.Engine
	config Config
}

func NewEngine(config Config) (*Engine, error) {
	if config.MaxValues == 0 || config.MaxBytes == 0 || config.MaxNesting == 0 || config.MaxResources == 0 || config.MaxTasks == 0 {
		return nil, wasm.ErrLimit
	}
	core, err := vm.NewEngineWithConfig(config.Core)
	if err != nil {
		return nil, err
	}
	return &Engine{core: core, config: config}, nil
}

func (e *Engine) Close() {
	e.mu.Lock()
	e.closed = true
	e.mu.Unlock()
	e.core.Close()
}

func (e *Engine) check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return wasm.ErrClosed
	}
	return nil
}

// Compiled owns a validated component and its code. It must not be copied.
// Close rejects new bindings; existing bindings retain their code leases.
type Compiled struct {
	mu     sync.Mutex
	engine *Engine
	closed bool
	refs   uint64
	plan   *validate.Component
	cores  map[*module.Module]*vm.CompiledModule
}

func (c *Compiled) Close() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	if c.refs != 0 {
		c.releaseLocked()
	}
}

func (e *Engine) Compile(ctx context.Context, input []byte) (*Compiled, error) {
	ctx = contextOrBackground(ctx)
	if err := e.check(ctx); err != nil {
		return nil, err
	}
	definition, err := decode.Decode(input, e.config.Decode)
	if err != nil {
		return nil, err
	}
	plan, err := validate.Check(definition, e.config.Validate)
	if err != nil {
		return nil, err
	}
	if err := preflight(plan); err != nil {
		return nil, err
	}
	compiled := &Compiled{engine: e, plan: plan, refs: 1, cores: map[*module.Module]*vm.CompiledModule{}}
	var compile func(*validate.Component) error
	compile = func(p *validate.Component) error {
		for _, d := range p.Definition().Definitions {
			if err := ctx.Err(); err != nil {
				return err
			}
			if d.Kind == 1 {
				core, err := e.core.Compile(ctx, d.Bytes)
				if err != nil {
					return err
				}
				compiled.cores[d.CoreModule] = core
			}
		}
		for _, child := range p.Children {
			if err := compile(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := compile(plan); err != nil {
		compiled.Close()
		return nil, err
	}
	if err := e.check(ctx); err != nil {
		compiled.Close()
		return nil, err
	}
	return compiled, nil
}

func unsupported(detail string) error {
	return fmt.Errorf("%w: component runtime %s", wasm.ErrUnsupported, detail)
}
func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func preflight(p *validate.Component) error {
	seen := map[*model.Type]bool{}
	var value func(*model.Type) error
	value = func(t *model.Type) error {
		if t == nil || seen[t] {
			return nil
		}
		seen[t] = true
		if t.Kind == model.Borrow && p.ResourceRepresentation(t.ResourceID) == wasm.I64 {
			return unsupported("cm64 defining-instance borrow conversion")
		}
		for _, fields := range [][]model.Field{t.Fields, t.Cases} {
			for _, f := range fields {
				if err := value(f.Type); err != nil {
					return err
				}
			}
		}
		if err := value(t.Key); err != nil {
			return err
		}
		return value(t.Element)
	}
	for _, start := range p.Starts {
		for _, param := range start.Type.Parameters {
			if err := value(param.Type); err != nil {
				return err
			}
		}
		if err := value(start.Type.Result); err != nil {
			return err
		}
	}
	for _, f := range p.Canonical {
		op := f.Definition.Op
		if op > 6 && (op < 9 || op > 0x1b) && (op < 0x1c || op > 0x25) && op != 0x2e && op != 0x2f {
			return unsupported("resource or async canonical operation")
		}
		if f.Type == nil {
			continue
		}
		for _, param := range f.Type.Parameters {
			if err := value(param.Type); err != nil {
				return err
			}
		}
		if err := value(f.Type.Result); err != nil {
			return err
		}
	}
	interfaces := map[*validate.Type]bool{}
	var external func(validate.Extern) error
	external = func(e validate.Extern) error {
		if interfaces[e.Type] {
			return nil
		}
		interfaces[e.Type] = true
		if e.Sort == model.FunctionSort {
			for _, p := range e.Type.Parameters {
				if err := value(p.Type); err != nil {
					return err
				}
			}
			return value(e.Type.Result)
		}
		if e.Sort == model.ValueSort {
			return value(e.Type.Value)
		}
		if e.Sort == model.TypeSort && e.Type.Value != nil {
			return value(e.Type.Value)
		}
		if e.Sort == model.InstanceSort || e.Sort == model.ComponentSort {
			for _, group := range [][]validate.Extern{e.Type.Imports, e.Type.Exports} {
				for _, child := range group {
					if err := external(child); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	for _, group := range [][]validate.Extern{p.Type.Imports, p.Type.Exports} {
		for _, e := range group {
			if err := external(e); err != nil {
				return err
			}
		}
	}
	for _, child := range p.Children {
		if err := preflight(child); err != nil {
			return err
		}
	}
	return nil
}
