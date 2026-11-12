package vm

import (
	"context"
	"errors"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// WithCaller runs a sequence of Core calls, memory transfers and instantiations
// under one Store lease, cancellation context and fuel budget. The capability
// expires when fn returns. This supports compound protocols such as canonical
// realloc/call/post-return without resetting fuel between their Core calls.
// Store operations that require a separate execution lease return ErrBusy;
// use the supplied Caller for operations within the session.
func (s *Store) WithCaller(ctx context.Context, options CallOptions, fn func(*Caller) error) (uint64, error) {
	return s.withCaller(ctx, options, fn, nil)
}

// WithCallerCleanup runs cleanup before the Caller expires on failure, panic or
// cancellation at the final publication check. Cleanup shares the existing
// lease, fuel and context, and cannot publish a successful result after failure.
func (s *Store) WithCallerCleanup(ctx context.Context, options CallOptions, fn func(*Caller) error, cleanup func(*Caller) error) (uint64, error) {
	return s.withCaller(ctx, options, fn, cleanup)
}

func (s *Store) withCaller(ctx context.Context, options CallOptions, fn func(*Caller) error, cleanup func(*Caller) error) (fuel uint64, err error) {
	if fn == nil {
		return 0, ErrType
	}
	op, err := s.begin(ctx, options)
	if err != nil {
		return 0, err
	}
	defer op.finish()
	caller := op.newCaller(nil)
	guard := caller.callerState
	defer func() {
		// A capability may still be executing on another goroutine. Cancel
		// before waiting for its lock, including on normal return and panic.
		op.stopWatcher()
		op.cancel()
		guard.mu.Lock()
		guard.active = false
		guard.mu.Unlock()
		for _, v := range guard.temporary {
			v.Release()
		}
		fuel = op.fuel
	}()
	complete := false
	defer func() {
		if !complete && cleanup != nil {
			if failed := cleanup(caller); failed != nil {
				err = errors.Join(err, failed)
			}
		}
	}()
	err = fn(caller)
	if err == nil {
		err = op.ctx.Err()
	}
	complete = err == nil
	return 0, err // The capability-draining defer publishes final fuel.
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

// ReadExternalMemoryString snapshots a same-Store memory as immutable bytes.
// The result remains valid after guest writes, growth, reentry and Store close.
// Its contents need not be UTF-8. The full range is checked before allocation.
func (c *Caller) ReadExternalMemoryString(memory Extern, offset, size uint64) (string, error) {
	if err := c.enter(); err != nil {
		return "", err
	}
	defer c.mu.Unlock()
	if memory.store != c.op.store || memory.store == nil {
		return "", ErrReference
	}
	if memory.Kind() != MemoryKind {
		return "", ErrType
	}
	data, ok := memory.external.Memory.ReadString(offset, size)
	if !ok {
		return "", &wasm.Trap{Code: wasm.TrapMemoryBounds}
	}
	return data, nil
}

// WriteExternalMemoryString copies immutable host bytes into a same-Store
// memory. The callback-scoped capability and full-range checks match the slice
// variant. No host or guest backing storage escapes.
func (c *Caller) WriteExternalMemoryString(memory Extern, offset uint64, src string) error {
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
	if !memory.external.Memory.WriteString(offset, src) {
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
