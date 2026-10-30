package component

import (
	"context"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type HostResourceDestructor func(*Caller, uint32) error

// HostResourceType is one nominal embedder-defined resource declaration.
// Only its constructor grants an identity; a zero value cannot bind or mint.
type HostResourceType struct{ definition *hostResourceDefinition }

type hostResourceDefinition struct {
	identity   *canonical.ResourceType
	destructor HostResourceDestructor
}

func NewHostResourceType(destructor HostResourceDestructor) *HostResourceType {
	return &HostResourceType{definition: &hostResourceDefinition{identity: &canonical.ResourceType{}, destructor: destructor}}
}

func (in *Instance) newResource(t *HostResourceType, rep uint32) (Value, error) {
	if t == nil {
		return Value{}, vm.ErrReference
	}
	rt := in.hostTypes[t.definition]
	if rt == nil {
		return Value{}, vm.ErrReference
	}
	id, err := in.hostResources.New(rt.identity, uint64(rep))
	if err != nil {
		return Value{}, err
	}
	return Value{Resource: canonical.NewReference(in.hostResources, rt.identity, id, rt, nil)}, nil
}

// NewResource grants host ownership for a type bound to this instance. Numeric
// representations alone never grant guest or host resource authority.
func (in *Instance) NewResource(ctx context.Context, t *HostResourceType, rep uint32) (Value, error) {
	var v Value
	_, err := in.store.WithCaller(ctx, CallOptions{}, func(c *vm.Caller) error { var err error; v, err = in.newResource(t, rep); return err })
	if err != nil {
		in.discardResource(v.Resource)
		return Value{}, err
	}
	return v, nil
}

func (c *Caller) NewResource(t *HostResourceType, rep uint32) (Value, error) {
	if err := c.enter(); err != nil {
		return Value{}, err
	}
	defer c.mu.Unlock()
	v, err := c.instance.newResource(t, rep)
	if err == nil {
		c.created = append(c.created, v.Resource)
	}
	return v, err
}

func (rt *resourceType) destroyHost(core *vm.Caller, rep uint64) error {
	if rt.instance.callDepth > rt.instance.config.Store.MaxReentry {
		return wasm.ErrLimit
	}
	rt.instance.callDepth++
	defer func() { rt.instance.callDepth-- }()
	if rt.host.destructor == nil {
		return nil
	}
	c := &Caller{callerState: &callerState{core: core, instance: rt.instance, active: true}}
	complete := false
	defer func() {
		c.mu.Lock()
		c.active = false
		c.mu.Unlock()
		if !complete {
			for _, ref := range c.created {
				rt.instance.cleanupResource(core, ref)
			}
		}
		c.created = nil
	}()
	err := rt.host.destructor(c, uint32(rep))
	complete = err == nil
	return err
}

func (in *Instance) resourceRep(t *HostResourceType, ref *Resource) (uint32, error) {
	table, identity, id, owner, _, err := canonical.ReferenceInfo(ref)
	if err != nil {
		return 0, err
	}
	rt, ok := owner.(*resourceType)
	if !ok || t == nil || rt.instance != in || rt.host == nil || rt.host != t.definition {
		return 0, vm.ErrReference
	}
	rep, err := table.Rep(identity, id)
	return uint32(rep), err
}

func (in *Instance) ResourceRep(ctx context.Context, t *HostResourceType, ref *Resource) (uint32, error) {
	var rep uint32
	_, err := in.store.WithCaller(ctx, CallOptions{}, func(c *vm.Caller) error { var err error; rep, err = in.resourceRep(t, ref); return err })
	return rep, err
}

func (c *Caller) ResourceRep(t *HostResourceType, ref *Resource) (uint32, error) {
	if err := c.enter(); err != nil {
		return 0, err
	}
	defer c.mu.Unlock()
	return c.instance.resourceRep(t, ref)
}
