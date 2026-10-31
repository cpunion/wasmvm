package component

import (
	"testing"

	"github.com/cpunion/wasmvm/pkg/vm"
)

// Backpressured calls must not read argument memory or transfer ownership when
// cancelled before entry. Exercise the same task entry used by async lower.
func TestPendingCancellationBeforeEntry(t *testing.T) {
	in := asyncInstance(t, "interpreter", "pending-callback", Imports{})
	f, _ := in.Function("run")
	_, err := in.store.WithCaller(nil, CallOptions{}, func(c *vm.Caller) (err error) {
		complete := false
		defer in.finishTasks(c, &err, &complete)
		f.scope.entered = true
		defer func() { f.scope.entered = false }()
		read, resolved := false, 0
		task, err := f.start(c, func(*vm.Caller) ([]Value, error) { read = true; return []Value{{Bits: 9}}, nil }, func(_ *vm.Caller, value *Value, cancelled bool) error {
			resolved++
			if value != nil || !cancelled {
				t.Error("incorrect pre-entry resolution")
			}
			return nil
		})
		if err != nil {
			return err
		}
		if read || task.started || task.exited {
			t.Fatal("backpressure consumed arguments")
		}
		task.cancelRequested = true
		if !task.ready() {
			t.Fatal("pre-entry cancellation blocked by exclusive turn")
		}
		if err := task.resume(c); err != nil {
			return err
		}
		if read || task.started || !task.exited || resolved != 1 {
			t.Fatal("pre-entry cancellation transferred arguments", read, resolved)
		}
		complete = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
