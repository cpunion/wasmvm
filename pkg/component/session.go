package component

import (
	"context"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
)

// WithCaller runs an embedding callback and its scheduled work in one Store
// lease. The callback's Caller expires on return; scheduled callbacks receive
// fresh Callers. Successful callback ownership effects remain committed if a
// later background task fails.
func (in *Instance) WithCaller(ctx context.Context, options CallOptions, fn func(*Caller) error) (uint64, error) {
	if fn == nil {
		return 0, vm.ErrType
	}
	return in.store.WithCaller(ctx, options, func(c *vm.Caller) (err error) {
		complete := false
		defer in.finishTasks(c, &err, &complete)
		receiver := &Function{instance: in, scope: &namespace{handles: in.hostResources}, typ: &validate.Type{Kind: model.Function}}
		handoff, err := receiver.hostCall(c, nil)
		if err != nil {
			return err
		}
		defer handoff.finishDeferred(c, &complete, &err)
		_, err = handoff.invoke(c, func(fresh *Caller) (*Value, error) { return nil, fn(fresh) })
		complete = err == nil
		return err
	})
}
