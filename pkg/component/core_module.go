package component

import (
	"context"
	"sync"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// CoreModule is a compiled module in this component Engine's Core domain.
// Close prevents new bindings; existing component instances retain publication.
type CoreModule struct {
	mu       sync.Mutex
	engine   *Engine
	compiled *vm.CompiledModule
	schema   vm.ModuleInterface
	closed   bool
	refs     uint64
}

// CompileCore uses the same bounded frontend, interpreter/native compiler and
// engine domain as embedded Core modules. No module is instantiated here.
func (e *Engine) CompileCore(ctx context.Context, input []byte) (*CoreModule, error) {
	ctx = contextOrBackground(ctx)
	if err := e.check(ctx); err != nil {
		return nil, err
	}
	m, err := e.core.Compile(ctx, input)
	if err != nil {
		return nil, err
	}
	schema, err := m.Interface()
	if err == nil {
		err = e.check(ctx)
	}
	if err != nil {
		m.Close()
		return nil, err
	}
	return &CoreModule{engine: e, compiled: m, schema: schema, refs: 1}, nil
}

func (m *CoreModule) Close() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed || m.compiled == nil {
		return
	}
	m.closed = true
	m.releaseLocked()
}
func (m *CoreModule) releaseLocked() {
	m.refs--
	if m.refs == 0 {
		m.compiled.Close()
		m.compiled = nil
		m.schema = vm.ModuleInterface{}
	}
}
func (m *CoreModule) acquire(e *Engine) (*coreModuleLease, error) {
	if m == nil {
		return nil, vm.ErrReference
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.engine != e || m.engine == nil {
		return nil, vm.ErrReference
	}
	if m.closed || m.compiled == nil {
		return nil, wasm.ErrClosed
	}
	m.refs++
	return &coreModuleLease{owner: m, compiled: m.compiled, schema: m.schema}, nil
}

// Info returns the Core VM's owned module metadata snapshot.
func (m *CoreModule) Info() (vm.ModuleInfo, error) {
	if m == nil {
		return vm.ModuleInfo{}, vm.ErrReference
	}
	l, err := m.acquire(m.engine)
	if err != nil {
		return vm.ModuleInfo{}, err
	}
	defer l.Close()
	return l.compiled.Info()
}

type coreModuleLease struct {
	once     sync.Once
	owner    *CoreModule
	compiled *vm.CompiledModule
	schema   vm.ModuleInterface
}

func (l *coreModuleLease) Close() {
	l.once.Do(func() {
		m := l.owner
		m.mu.Lock()
		defer m.mu.Unlock()
		m.releaseLocked()
		l.owner, l.compiled, l.schema = nil, nil, vm.ModuleInterface{}
	})
}
