package vm

import (
	"context"
	"sync"

	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type DecodeLimits = decode.Limits
type ValidationLimits = validate.Limits
type LoweringLimits = lower.Limits
type SlotLimits = slots.Limits
type FunctionType = module.FuncType
type MemoryType = module.Limits
type TableType = module.TableType
type GlobalType = module.GlobalType

// EngineConfig selects the complete Core 2.0 interpreter and bounded compiler
// phases. Native execution and other profiles are not implemented in M1.
type EngineConfig struct {
	Backend  string
	Profile  string
	Decode   DecodeLimits
	Validate ValidationLimits
	Lower    LoweringLimits
	Slots    SlotLimits
}

func DefaultEngineConfig() EngineConfig {
	return EngineConfig{"interpreter", "core2", decode.DefaultLimits(), validate.DefaultLimits(), lower.DefaultLimits(), slots.DefaultLimits()}
}

type Engine struct {
	mu     sync.Mutex
	closed bool
	config EngineConfig
}

func NewEngine() *Engine { e, _ := NewEngineWithConfig(DefaultEngineConfig()); return e }
func NewEngineWithConfig(config EngineConfig) (*Engine, error) {
	if config.Backend != "interpreter" || config.Profile != "core2" {
		return nil, wasm.ErrUnsupported
	}
	return &Engine{config: config}, nil
}

// Close prevents new compilation and Store creation. Existing compiled modules
// and Stores retain their code and remain usable until separately closed.
func (e *Engine) Close() { e.mu.Lock(); e.closed = true; e.mu.Unlock() }
func (e *Engine) available() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return wasm.ErrClosed
	}
	return nil
}
func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// Compile owns a snapshot of input and checks cancellation between bounded
// compiler phases. Concurrent compilations never share mutable compiler state.
func (e *Engine) Compile(ctx context.Context, input []byte) (*CompiledModule, error) {
	ctx = contextOrBackground(ctx)
	v, err := e.frontend(ctx, input)
	if err != nil {
		return nil, err
	}
	ir, err := lower.Compile(v, e.config.Lower)
	if err != nil {
		return nil, err
	}
	if err = e.check(ctx); err != nil {
		return nil, err
	}
	code, err := slots.Compile(ir, e.config.Slots)
	if err != nil {
		return nil, err
	}
	if err = e.check(ctx); err != nil {
		return nil, err
	}
	return &CompiledModule{engine: e, code: code}, nil
}

func (e *Engine) check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return e.available()
}
func (e *Engine) frontend(ctx context.Context, input []byte) (*validate.Module, error) {
	if err := e.check(ctx); err != nil {
		return nil, err
	}
	m, err := decode.Decode(input, e.config.Decode)
	if err != nil {
		return nil, err
	}
	if err = e.check(ctx); err != nil {
		return nil, err
	}
	v, err := validate.Check(m, e.config.Validate)
	if err != nil {
		return nil, err
	}
	if err = e.check(ctx); err != nil {
		return nil, err
	}
	return v, nil
}

// Validate performs binary decoding and static Core 2.0 validation. It does not
// lower code, resolve imports, instantiate, or execute a start function.
func (e *Engine) Validate(ctx context.Context, input []byte) error {
	_, err := e.frontend(contextOrBackground(ctx), input)
	return err
}

// Inspect returns copied metadata after static validation, without lowering or
// executing any code. Missing host imports do not prevent inspection.
func (e *Engine) Inspect(ctx context.Context, input []byte) (ModuleInfo, error) {
	v, err := e.frontend(contextOrBackground(ctx), input)
	if err != nil {
		return ModuleInfo{}, err
	}
	return describeModule(v.Definition()), nil
}

type CompiledModule struct {
	mu     sync.Mutex
	engine *Engine
	code   *slots.Module
}

func (m *CompiledModule) Close() { m.mu.Lock(); m.code = nil; m.mu.Unlock() }
func (m *CompiledModule) acquire(e *Engine) (*slots.Module, error) {
	if m == nil {
		return nil, ErrReference
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.engine != e {
		return nil, ErrReference
	}
	if m.code == nil {
		return nil, wasm.ErrClosed
	}
	return m.code, nil
}

type ExternType struct {
	Kind     ExternKind
	Function *FunctionType
	Table    *TableType
	Memory   *MemoryType
	Global   *GlobalType
}
type ImportInfo struct {
	Module, Name string
	ExternType
}
type ExportInfo struct {
	Name string
	ExternType
}
type ModuleInfo struct {
	Backend, Profile string
	Imports          []ImportInfo
	Exports          []ExportInfo
	Functions        uint32
	HasStart         bool
}

func (m *CompiledModule) Info() (ModuleInfo, error) {
	code, err := m.acquire(m.engine)
	if err != nil {
		return ModuleInfo{}, err
	}
	return describeModule(code.IR.Validated.Definition()), nil
}
func describeModule(d *module.Module) ModuleInfo {
	info := ModuleInfo{Backend: "interpreter", Profile: "core2", Functions: uint32(len(d.Functions)), HasStart: d.HasStart}
	objects := [4][]ExternType{}
	functionType := func(t module.FuncType) ExternType {
		return ExternType{Kind: FunctionKind, Function: &FunctionType{Params: append([]wasm.ValueType(nil), t.Params...), Results: append([]wasm.ValueType(nil), t.Results...)}}
	}
	for _, im := range d.Imports {
		desc := ExternType{Kind: ExternKind(im.Kind)}
		switch im.Kind {
		case 0:
			desc = functionType(d.Types[im.Type])
		case 1:
			t := im.Table
			desc.Table = &t
		case 2:
			m := im.Memory
			desc.Memory = &m
		case 3:
			g := im.Global
			desc.Global = &g
		}
		objects[im.Kind] = append(objects[im.Kind], desc)
		info.Imports = append(info.Imports, ImportInfo{im.Module, im.Name, desc})
	}
	for _, f := range d.Functions {
		objects[0] = append(objects[0], functionType(d.Types[f.Type]))
	}
	for _, t := range d.Tables {
		t := t
		objects[1] = append(objects[1], ExternType{Kind: TableKind, Table: &t})
	}
	for _, m := range d.Memories {
		m := m
		objects[2] = append(objects[2], ExternType{Kind: MemoryKind, Memory: &m})
	}
	for _, g := range d.Globals {
		t := g.Type
		objects[3] = append(objects[3], ExternType{Kind: GlobalKind, Global: &t})
	}
	for _, ex := range d.Exports {
		info.Exports = append(info.Exports, ExportInfo{ex.Name, objects[ex.Kind][ex.Index]})
	}
	return info
}
