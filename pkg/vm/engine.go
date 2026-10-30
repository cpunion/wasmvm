package vm

import (
	"context"
	"sync"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/memory"
	"github.com/cpunion/wasmvm/internal/engine/nativeexec"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type DecodeLimits = decode.Limits
type ValidationLimits = validate.Limits
type LoweringLimits = lower.Limits
type SlotLimits = slots.Limits
type NativeLimits = codegen.Limits
type FunctionType = module.FuncType
type MemoryType = module.Limits
type TableType = module.TableType
type GlobalType = module.GlobalType

// EngineConfig selects the Core 2.0 or 3.0 profile and bounded compiler phases.
// Native requires a supported platform adapter; no backend fallback is used.
type EngineConfig struct {
	Backend    string
	Profile    string
	Features   wasm.Features
	MaxWaiters uint32
	Decode     DecodeLimits
	Validate   ValidationLimits
	Lower      LoweringLimits
	Slots      SlotLimits
	Native     NativeLimits
}

func DefaultEngineConfig() EngineConfig {
	return EngineConfig{MaxWaiters: 1 << 16, Backend: "interpreter", Profile: "core2", Decode: decode.DefaultLimits(), Validate: validate.DefaultLimits(), Lower: lower.DefaultLimits(), Slots: slots.DefaultLimits(), Native: codegen.DefaultLimits()}
}

type Engine struct {
	mu     sync.Mutex
	closed bool
	config EngineConfig
	domain *memory.Domain
}

func NewEngine() *Engine { e, _ := NewEngineWithConfig(DefaultEngineConfig()); return e }
func NewEngineWithConfig(config EngineConfig) (*Engine, error) {
	if config.Profile == "core3" {
		config.Features |= wasm.Core3Features
	}
	if config.Backend != "interpreter" && config.Backend != "native" || config.Backend == "native" && !native.Available || config.Profile != "core2" && config.Profile != "core3" || config.Features & ^wasm.SupportedFeatures != 0 {
		return nil, wasm.ErrUnsupported
	}
	return &Engine{config: config, domain: memory.NewDomain(config.MaxWaiters, native.Available)}, nil
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
	out := &CompiledModule{engine: e, code: code}
	if e.config.Backend == "native" {
		out.native, err = nativeexec.Compile(ctx, code, e.config.Native)
		if err != nil {
			return nil, err
		}
		if err = e.check(ctx); err != nil {
			out.Close()
			return nil, err
		}
	}
	return out, nil
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
	m, err := decode.DecodeWithFeatures(input, e.config.Decode, e.config.Features)
	if err != nil {
		return nil, err
	}
	if err = e.check(ctx); err != nil {
		return nil, err
	}
	v, err := validate.CheckWithFeatures(m, e.config.Validate, e.config.Features)
	if err != nil {
		return nil, err
	}
	if err = e.check(ctx); err != nil {
		return nil, err
	}
	return v, nil
}

// Validate decodes and checks the selected binary profile. It does not
// lower code, resolve imports, instantiate, or execute a start function.
func (e *Engine) Validate(ctx context.Context, input []byte) error {
	_, err := e.frontend(contextOrBackground(ctx), input)
	return err
}

// Inspect returns copied metadata after static validation, without lowering or
// executing any code. Missing host imports do not prevent inspection.
func (e *Engine) Inspect(ctx context.Context, input []byte) (ModuleInfo, error) {
	ctx = contextOrBackground(ctx)
	v, err := e.frontend(ctx, input)
	if err != nil {
		return ModuleInfo{}, err
	}
	info := describeModule(v.Definition())
	info.Backend, info.Features = e.config.Backend, e.config.Features
	info.Profile = e.config.Profile
	if err := e.check(ctx); err != nil {
		return ModuleInfo{}, err
	}
	return info, nil
}

type CompiledModule struct {
	mu     sync.Mutex
	engine *Engine
	code   *slots.Module
	native *nativeexec.Module
}

func (m *CompiledModule) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.code = nil
	if m.native != nil {
		m.native.Close()
		m.native = nil
	}
}
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

// ExternType describes an imported or exported object. Descriptor pointers may
// be shared within one metadata snapshot, but never alias engine code or another
// snapshot. Copy before editing if other entries must remain unchanged.
type ExternType struct {
	Kind     ExternKind
	Function *FunctionType
	Table    *TableType
	Memory   *MemoryType
	Global   *GlobalType
	Tag      *FunctionType
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
	Features         wasm.Features
	Imports          []ImportInfo
	Exports          []ExportInfo
	Functions        uint32
	HasStart         bool
}

// ModuleInterface is an immutable type-only view used for checked composition.
type ModuleInterface = module.Interface

// Interface retains no code publication, bodies, data or global initializers.
func (m *CompiledModule) Interface() (ModuleInterface, error) {
	if m == nil {
		return ModuleInterface{}, ErrReference
	}
	code, err := m.acquire(m.engine)
	if err != nil {
		return ModuleInterface{}, err
	}
	return module.SnapshotInterface(code.IR.Validated.Definition()), nil
}

func (m *CompiledModule) Info() (ModuleInfo, error) {
	code, err := m.acquire(m.engine)
	if err != nil {
		return ModuleInfo{}, err
	}
	info := describeModule(code.IR.Validated.Definition())
	info.Backend, info.Features = m.engine.config.Backend, m.engine.config.Features
	info.Profile = m.engine.config.Profile
	return info, nil
}
func describeModule(d *module.Module) ModuleInfo {
	info := ModuleInfo{Backend: "interpreter", Profile: "core2", Functions: uint32(len(d.Functions)), HasStart: d.HasStart}
	objects := [5][]ExternType{}
	// Copy each referenced type once. A short binary can declare many functions
	// with one large signature; per-function copies would multiply their memory
	// cost, bypassing the frontend's aggregate byte/node bounds.
	types := make([]*FunctionType, len(d.Types))
	functionType := func(index uint32) ExternType {
		if types[index] == nil {
			t := d.Types[index]
			types[index] = &FunctionType{Params: append([]wasm.ValueType(nil), t.Params...), Results: append([]wasm.ValueType(nil), t.Results...)}
		}
		return ExternType{Kind: FunctionKind, Function: types[index]}
	}

	tagType := func(index uint32) ExternType {
		t := functionType(index)
		return ExternType{Kind: TagKind, Tag: t.Function}
	}
	for _, im := range d.Imports {
		desc := ExternType{Kind: ExternKind(im.Kind)}
		switch im.Kind {
		case 0:
			desc = functionType(im.Type)
		case 1:
			t := im.Table
			desc.Table = &t
		case 2:
			m := im.Memory
			desc.Memory = &m
		case 4:
			desc = tagType(im.Type)
		case 3:
			g := im.Global
			desc.Global = &g
		}
		objects[im.Kind] = append(objects[im.Kind], desc)
		info.Imports = append(info.Imports, ImportInfo{im.Module, im.Name, desc})
	}
	for _, f := range d.Functions {
		objects[0] = append(objects[0], functionType(f.Type))
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
	for _, t := range d.Tags {
		objects[4] = append(objects[4], tagType(t))
	}
	for _, ex := range d.Exports {
		info.Exports = append(info.Exports, ExportInfo{ex.Name, objects[ex.Kind][ex.Index]})
	}
	return info
}

// acquireForStore retains executable publication before a concurrent Close can
// retire it. Existing instances keep leases; no new code is acquired after close.
func (m *CompiledModule) acquireForStore(s *Store) (*slots.Module, error) {
	if m == nil {
		return nil, ErrReference
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.engine != s.engine {
		return nil, ErrReference
	}
	if m.code == nil {
		return nil, wasm.ErrClosed
	}
	if m.native != nil && s.nativeModules[m.code] == nil {
		lease, err := m.native.Acquire()
		if err != nil {
			return nil, err
		}
		s.nativeModules[m.code] = lease
	}
	return m.code, nil
}
