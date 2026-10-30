// Package component compiles and executes synchronous component composition
// through the project-owned Core interpreter or native compiler. Capabilities
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
	Core         vm.EngineConfig
	Store        vm.StoreConfig
	Decode       DecodeLimits
	Validate     ValidationLimits
	MaxValues    uint64
	MaxBytes     uint64
	MaxNesting   uint32
	MaxResources uint32
}

func DefaultConfig() Config {
	core := vm.DefaultEngineConfig()
	core.Profile, core.Features = "core3", wasm.SupportedFeatures
	return Config{Core: core, Store: vm.DefaultStoreConfig(), Decode: decode.DefaultLimits(), Validate: validate.DefaultLimits(), MaxValues: 1 << 20, MaxBytes: 64 << 20, MaxNesting: 128, MaxResources: 1 << 16}
}

type Engine struct {
	mu     sync.Mutex
	closed bool
	core   *vm.Engine
	config Config
}

func NewEngine(config Config) (*Engine, error) {
	if config.MaxValues == 0 || config.MaxBytes == 0 || config.MaxNesting == 0 || config.MaxResources == 0 {
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

type Compiled struct {
	mu     sync.Mutex
	engine *Engine
	plan   *validate.Component
	cores  map[*module.Module]*vm.CompiledModule
}

func (c *Compiled) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.plan = nil
	for _, core := range c.cores {
		core.Close()
	}
	c.cores = nil
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
	if err := preflight(plan, true); err != nil {
		return nil, err
	}
	compiled := &Compiled{engine: e, plan: plan, cores: map[*module.Module]*vm.CompiledModule{}}
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

func preflight(p *validate.Component, root bool) error {
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
		switch t.Kind {
		case model.Future, model.Stream, model.ErrorContext:
			return unsupported("value conversion")
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
		if start.Type.ContainsResource() {
			return unsupported("resource-bearing component start")
		}
	}
	for _, f := range p.Canonical {
		if f.Definition.Op > 4 || f.Async {
			return unsupported("resource or async canonical operation")
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
			if e.Type.Kind == model.AsyncFunction {
				return unsupported("async function")
			}
			for _, p := range e.Type.Parameters {
				if err := value(p.Type); err != nil {
					return err
				}
			}
			return value(e.Type.Result)
		}
		if e.Sort == model.ValueSort {
			if e.Type.ContainsResource() {
				return unsupported("standalone resource value binding")
			}
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
	if root {
		for _, im := range p.Type.Imports {
			if im.Sort == model.TypeSort && im.Type.Value == nil && im.Type.Kind != model.Resource || im.Sort == model.ComponentSort {
				return unsupported("host type/component binding")
			}
		}
	}
	for _, child := range p.Children {
		if err := preflight(child, false); err != nil {
			return err
		}
	}
	return nil
}
