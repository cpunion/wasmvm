package component

import (
	"context"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// DropAsync consumes a host-owned readable future/stream endpoint. It notifies
// its writable peer; copying endpoints must first finish or cancel their copy.
func (in *Instance) DropAsync(ctx context.Context, ref *Resource, options CallOptions) (uint64, error) {
	return in.store.WithCaller(ctx, options, func(*vm.Caller) error { return in.dropAsync(ref) })
}
func (c *Caller) DropAsync(ref *Resource) error {
	if err := c.enter(); err != nil {
		return err
	}
	defer c.mu.Unlock()
	return c.instance.dropAsync(ref)
}
func (in *Instance) dropAsync(ref *Resource) error {
	table, _, id, owner, borrow, err := canonical.ReferenceInfo(ref)
	if err != nil {
		return err
	}
	e, ok := owner.(*asyncEnd)
	if !ok || e.in != in || borrow || table != in.hostResources || e.n != nil || e.id != id || !e.read {
		return vm.ErrReference
	}
	if e.state == endCopying || e.state == endCancelling {
		return wasm.ErrTrap
	}
	if got, err := table.Opaque(id); err != nil || got != e {
		return wasm.ErrTrap
	}
	defer canonical.InvalidateReference(ref)
	e.abort()
	return nil
}
