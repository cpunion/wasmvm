package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/pkg/vm"
)

func (in *Instance) initializeValues(core *vm.Caller, scope *namespace) error {
	// Validate supplied capabilities before any factory can have host effects.
	for _, v := range in.initialValues {
		v.scope = scope
		if v.factory == nil && v.resource {
			if err := in.initializing.Validate(v.typ, v.value); err != nil {
				return err
			}
		}
	}
	for _, v := range in.initialValues {
		if v.factory == nil && !v.resource {
			continue
		}
		if v.factory != nil {
			value, err := in.createInitialValue(core, v.factory)
			if err != nil {
				return err
			}
			v.value = value
			v.factory = nil
		}
		copy, err := in.initializing.Clone(v.typ, v.value)
		if err != nil {
			return err
		}
		if err := in.initializing.AcceptResult(v.typ, &copy); err != nil {
			return err
		}
		v.value = copy
		v.validated = true
		// Each import is a separate occurrence. Live leases and moved tokens
		// enforce cross-value conflicts; duplicate checks span one shape.
		in.initializing.Resources.(*resourceCodec).seen = nil
		if literal, ok := v.expected.LiteralValue(); ok {
			literal, err = in.initializing.Clone(v.typ, literal)
			if err != nil {
				return err
			}
			if !equalValue(copy, literal) {
				return vm.ErrType
			}
		}
		key := v.expected.ValueKey()
		if earlier, ok := in.bindings[key]; ok && !equalValue(earlier, copy) {
			return vm.ErrType
		}
		in.bindings[key] = copy
	}
	return nil
}

func (in *Instance) createInitialValue(core *vm.Caller, factory ValueFactory) (Value, error) {
	c := &Caller{callerState: &callerState{core: core, instance: in, active: true}}
	defer func() {
		c.mu.Lock()
		c.active = false
		c.mu.Unlock()
		in.initialCreated = append(in.initialCreated, c.created...)
		c.created = nil
	}()
	return factory(c)
}

func (in *Instance) cleanupInitialValues(core *vm.Caller) error {
	var failed error
	cleanup := func(ref *Resource) {
		if err := in.cleanupResource(core, ref); failed == nil {
			failed = err
		}
	}
	for _, v := range in.initialValues {
		// Failed factories may return unvalidated aggregates. Only successfully
		// accepted shapes are walked; created tokens are tracked separately.
		if v.resource && v.validated {
			canonical.VisitReferences(v.typ, &v.value, cleanup)
		}
	}
	for _, ref := range in.initialCreated {
		cleanup(ref)
	}
	return failed
}
