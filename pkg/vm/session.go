package vm

import (
	"context"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// WithCaller runs a sequence of Core calls, memory transfers and instantiations
// under one Store lease, cancellation context and fuel budget. The capability
// expires when fn returns. This supports compound protocols such as canonical
// realloc/call/post-return without resetting fuel between their Core calls.
// Store operations that require a separate execution lease return ErrBusy;
// use the supplied Caller for operations within the session.
func (s *Store) WithCaller(ctx context.Context, options CallOptions, fn func(*Caller) error) (uint64, error) {
	if fn == nil {
		return 0, ErrType
	}
	op, err := s.begin(ctx, options)
	if err != nil {
		return 0, err
	}
	defer op.finish()
	caller := &Caller{active: true, op: op}
	defer func() {
		caller.mu.Lock()
		caller.active = false
		caller.mu.Unlock()
		for _, v := range caller.temporary {
			v.Release()
		}
	}()
	err = fn(caller)
	if err == nil {
		err = op.ctx.Err()
	}
	return op.fuel, err
}

// ReadExternalMemory and WriteExternalMemory copy through a same-Store memory
// capability, including memories belonging to a different suspended module.
// No guest slice survives realloc, reentry or a grow operation.
func (c *Caller) ReadExternalMemory(memory Extern, offset uint64, dst []byte) error {
	if err := c.enter(); err != nil {
		return err
	}
	defer c.mu.Unlock()
	if memory.store != c.op.store || memory.store == nil {
		return ErrReference
	}
	if memory.Kind() != MemoryKind {
		return ErrType
	}
	if !memory.external.Memory.Read(offset, dst) {
		return &wasm.Trap{Code: wasm.TrapMemoryBounds}
	}
	return nil
}

func (c *Caller) WriteExternalMemory(memory Extern, offset uint64, src []byte) error {
	if err := c.enter(); err != nil {
		return err
	}
	defer c.mu.Unlock()
	if memory.store != c.op.store || memory.store == nil {
		return ErrReference
	}
	if memory.Kind() != MemoryKind {
		return ErrType
	}
	if !memory.external.Memory.Write(offset, src) {
		return &wasm.Trap{Code: wasm.TrapMemoryBounds}
	}
	return nil
}

func (c *Caller) NewFunction(signature FunctionType, callback HostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	if err := c.enter(); err != nil {
		return Extern{}, err
	}
	defer c.mu.Unlock()
	return c.op.store.newFunction(signature, callback, referenceType...)
}

func (c *Caller) Instantiate(linker *Linker, compiled *CompiledModule) (*Instance, error) {
	if err := c.enter(); err != nil {
		return nil, err
	}
	defer c.mu.Unlock()
	if linker == nil || linker.store != c.op.store {
		return nil, ErrReference
	}
	return linker.instantiate(c.op, compiled)
}

func (c *Caller) Export(instance *Instance, name string) (Extern, error) {
	if err := c.enter(); err != nil {
		return Extern{}, err
	}
	defer c.mu.Unlock()
	if instance == nil || instance.store != c.op.store {
		return Extern{}, ErrReference
	}
	ex, ok := instance.instance.Exports[name]
	if !ok {
		return Extern{}, ErrLink
	}
	return Extern{store: c.op.store, external: ex}, nil
}
