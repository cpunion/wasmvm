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
	return f.callWithResult(ctx, args, options, nil, consume)
}

// CallWithResultPrepared creates arguments and consumes the result in the same
// Store lease as guest execution. A rejected Store admission never calls prepare.
// The preparation Caller expires before the guest starts; newly created
// capabilities are reclaimed if preparation or argument acceptance fails.
func (f *Function) CallWithResultPrepared(ctx context.Context, options CallOptions, prepare func(*Caller) ([]Value, error), consume func(*Caller, *Value) error) (uint64, error) {
	if prepare == nil {
		return 0, vm.ErrType
	}
	return f.callWithResult(ctx, nil, options, prepare, consume)
}

func (f *Function) callWithResult(ctx context.Context, args []Value, options CallOptions, prepare func(*Caller) ([]Value, error), consume func(*Caller, *Value) error) (uint64, error) {
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
		if prepare != nil {
			// Preparation creates capabilities through its Caller, tracked by
			// hostCall.created. It performs no canonical ownership conversion.
			factory := in.newHostCall(c, f, nil, nil)
			factory.limit()
			defer factory.finishDeferred(c, &complete, &err)
			defer factory.account()
			_, err = factory.invoke(c, func(fresh *Caller) (*Value, error) {
				args, err = prepare(fresh)
				return nil, err
			})
			if err != nil {
				return err
			}
		}
		if f.start == nil {
			result, err = f.call(c, args)
		} else {
			if failed := f.enterCall(c, args); failed != nil {
				return failed
			}
			defer f.leaveCall()
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
		values := []Value(nil)
		if result != nil {
			values = []Value{*result}
		}
		receiver := f.resultReceiver(result != nil)
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

// resultReceiver is called inside the exclusive Store lease. The two immutable
// shapes distinguish an absent result from a present result, including an empty
// value. Arguments, ownership and Caller lifetimes belong to the new hostCall.
func (f *Function) resultReceiver(present bool) *Function {
	index := 0
	if present {
		index = 1
	}
	if f.resultReceivers[index] == nil {
		var parameters []model.Field
		if present {
			parameters = []model.Field{{Type: f.typ.Result}}
		}
		f.resultReceivers[index] = &Function{instance: f.instance, scope: f.scope,
			typ: &validate.Type{Kind: model.Function, Parameters: parameters}}
	}
	return f.resultReceivers[index]
}
