package component

import (
	"context"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
)

// CallWithResult transfers the result to consume as soon as an async task
// resolves, inside the current Store lease. consume may read a returned stream
// while its producer continues. The call then drains background tasks. The
// callback receives ownership and a fresh Caller which expires on return.
func (f *Function) CallWithResult(ctx context.Context, args []Value, options CallOptions, consume func(*Caller, *Value) error) (uint64, error) {
	if f == nil || f.instance == nil {
		return 0, vm.ErrReference
	}
	if consume == nil {
		return 0, vm.ErrType
	}
	var result *Value
	in := f.instance
	return in.store.WithCallerCleanup(ctx, options, func(c *vm.Caller) (err error) {
		complete := false
		defer in.finishTasks(c, &err, &complete)
		if f.start == nil {
			result, err = f.call(c, args)
		} else {
			leave, failed := f.enterCall(c, args)
			if failed != nil {
				return failed
			}
			defer leave()
			task, failed := in.beginAsyncCall(c, f, f.plan, args)
			if failed != nil {
				return failed
			}
			if err = in.awaitTasks(c, func() bool { return task.task.resolved }); err != nil {
				return err
			}
			result = task.task.result
		}
		if err != nil {
			return err
		}
		parameters := []model.Field(nil)
		values := []Value(nil)
		if result != nil {
			parameters = []model.Field{{Type: f.typ.Result}}
			values = []Value{*result}
		}
		receiver := &Function{instance: in, scope: f.scope, typ: &validate.Type{Kind: model.Function, Parameters: parameters}}
		handoff, err := receiver.hostCall(c, values)
		if err != nil {
			return err
		}
		defer handoff.finishDeferred(c, &complete, &err)
		_, err = handoff.invoke(c, func(fresh *Caller) (*Value, error) {
			var value *Value
			if len(handoff.args) != 0 {
				value = &handoff.args[0]
			}
			return nil, consume(fresh, value)
		})
		complete = err == nil
		return err
	}, func(c *vm.Caller) error {
		if result != nil {
			canonical.VisitReferences(f.typ.Result, result, func(ref *Resource) { in.cleanupResource(c, ref) })
		}
		return nil
	})
}
