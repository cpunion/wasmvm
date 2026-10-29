package vm

import (
	"context"
	"sync"

	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type ExternKind byte

const (
	FunctionKind ExternKind = iota
	TableKind
	MemoryKind
	GlobalKind
)

// Extern is a Store-owned function, table, memory or global. Its fields cannot
// be forged. Published instances and host objects live until Store close.
type Extern struct {
	store    *Store
	external state.Extern
}

func (ex Extern) Kind() ExternKind { return ExternKind(ex.external.Kind) }
func (ex Extern) Call(ctx context.Context, args []Value, options CallOptions) (CallResult, error) {
	if ex.store == nil {
		return CallResult{}, ErrReference
	}
	return ex.store.Call(ctx, ex, args, options)
}
func (s *Store) NewFunction(signature FunctionType, callback HostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	if callback == nil {
		return Extern{}, ErrType
	}
	for _, ts := range [][]wasm.ValueType{signature.Params, signature.Results} {
		for _, typ := range ts {
			if err := s.referenceFeature(typ); err != nil {
				return Extern{}, err
			}
		}
	}
	for _, typ := range referenceType {
		if err := s.referenceFeature(typ); err != nil {
			return Extern{}, err
		}
	}
	ex, err := s.state.NewHost(signature, callback, referenceType...)
	if err != nil {
		return Extern{}, err
	}
	return Extern{store: s, external: ex}, nil
}
func (ex Extern) FunctionType() (FunctionType, error) {
	if ex.store == nil {
		return FunctionType{}, ErrReference
	}
	op, err := ex.store.begin(nil, CallOptions{})
	if err != nil {
		return FunctionType{}, err
	}
	defer op.finish()
	if ex.Kind() != FunctionKind {
		return FunctionType{}, ErrType
	}
	f, ok := ex.store.state.Function(ex.external.Function)
	if !ok {
		return FunctionType{}, ErrReference
	}
	return FunctionType{Params: append([]wasm.ValueType(nil), f.Signature.Params...), Results: append([]wasm.ValueType(nil), f.Signature.Results...)}, nil
}
func (ex Extern) Reference() (Value, error) {
	if ex.store == nil {
		return Value{}, ErrReference
	}
	op, err := ex.store.begin(nil, CallOptions{})
	if err != nil {
		return Value{}, err
	}
	defer op.finish()
	if ex.Kind() != FunctionKind {
		return Value{}, ErrType
	}
	s := ex.store
	s.mu.Lock()
	defer s.mu.Unlock()
	typ := wasm.FuncRef
	if s.engine.config.Features&(wasm.FunctionReferences|wasm.GC) != 0 {
		f, ok := s.state.Function(ex.external.Function)
		if !ok {
			return Value{}, ErrReference
		}
		typ = wasm.CanonicalFunctionReference(f.TypeID, false)
	}
	return s.rootLocked(typ, rawReference(ex.external.Function))
}

type Linker struct {
	mu      sync.Mutex
	store   *Store
	entries map[[2]string]Extern
}

func NewLinker(store *Store) *Linker {
	return &Linker{store: store, entries: make(map[[2]string]Extern)}
}
func (l *Linker) Define(module, name string, ex Extern) error {
	if ex.store != l.store || ex.store == nil {
		return ErrReference
	}
	l.store.mu.Lock()
	closed := l.store.closing
	l.store.mu.Unlock()
	if closed {
		return wasm.ErrClosed
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	key := [2]string{module, name}
	if _, ok := l.entries[key]; ok {
		return ErrLink
	}
	l.entries[key] = ex
	return nil
}

type Instance struct {
	store              *Store
	instance           *state.Instance
	startFuelRemaining uint64
}

// StartFuelRemaining is the immutable budget receipt from instantiation's
// start call. Later calls do not change it; it is meaningful when metered.
func (in *Instance) StartFuelRemaining() uint64 { return in.startFuelRemaining }

func (l *Linker) Instantiate(ctx context.Context, compiled *CompiledModule, options CallOptions) (*Instance, error) {
	op, err := l.store.begin(ctx, options)
	if err != nil {
		return nil, err
	}
	defer op.finish()
	op.collect = true
	code, err := compiled.acquireForStore(l.store)
	if err != nil {
		return nil, err
	}
	l.mu.Lock()
	imports := make(map[[2]string]state.Extern, len(l.entries))
	for key, ex := range l.entries {
		imports[key] = ex.external
	}
	l.mu.Unlock()
	in, err := l.store.state.AllocateContext(op.ctx, code, func(module, name string) (state.Extern, bool) {
		ex, ok := imports[[2]string{module, name}]
		return ex, ok
	})
	if err != nil {
		return nil, err
	}
	op.collect = true
	op.pending = append(op.pending, in)
	if err = op.ctx.Err(); err != nil {
		return nil, err
	}
	if err = in.Initialize(); err != nil {
		return nil, err
	}
	if err = op.ctx.Err(); err != nil {
		return nil, err
	}
	d := code.IR.Validated.Definition()
	if d.HasStart {
		if _, err = op.call(in.Functions[d.Start], nil); err != nil {
			return nil, err
		}
	}
	if err = op.ctx.Err(); err != nil {
		return nil, err
	}
	if err = l.store.state.Publish(in); err != nil {
		return nil, err
	}
	// Successful publication owns its code lease. Without externref payloads
	// or unpublished instances there is nothing a Store scan can reclaim.
	// Failure paths retain collection, including orphaned native leases.
	l.store.mu.Lock()
	op.collect = len(l.store.externs) != 0 || l.store.state.HasUnpublished()
	l.store.mu.Unlock()
	return &Instance{store: l.store, instance: in, startFuelRemaining: op.fuel}, nil
}
func (in *Instance) Export(name string) (Extern, error) {
	if in == nil || in.store == nil {
		return Extern{}, ErrReference
	}
	op, err := in.store.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	ex, ok := in.instance.Exports[name]
	if !ok {
		return Extern{}, ErrLink
	}
	return Extern{store: in.store, external: ex}, nil
}
