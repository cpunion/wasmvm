package service

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
)

func TestServiceCancellationDrainsAllClientsAndJoinsWorkers(t *testing.T) {
	args := Options{Workers: 4, Clients: 8, Requests: 16, Samples: 1, Expected: 7}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var count, active atomic.Int32
	calls := make([]Call, args.Workers)
	for i := range calls {
		var entered atomic.Bool
		calls[i] = func() (int64, uint32, error) {
			if !entered.CompareAndSwap(false, true) {
				t.Error("one Store invoked concurrently")
			}
			if active.Add(1) > int32(args.Workers) {
				t.Error("admission limit exceeded")
			}
			if count.Add(1) == 7 {
				cancel()
			}
			runtime.Gosched()
			active.Add(-1)
			entered.Store(false)
			return 17, 7, ctx.Err()
		}
	}
	call, stop, out, setupErr := Start(calls, args)
	if setupErr != nil {
		t.Fatal(setupErr)
	}
	_, _, err := call()
	stop()
	if !errors.Is(err, context.Canceled) || count.Load() != 128 || active.Load() != 0 || len(out.Batches[0].RequestNs) != 128 {
		t.Fatal(err, count.Load(), active.Load(), out)
	}
}

func TestInvalidAdmissionCannotStartWorkers(t *testing.T) {
	valid := Options{Workers: 1, Clients: 2, Requests: 3, Samples: 1}
	for _, change := range []func(*Options){
		func(v *Options) { v.Workers = 0 }, func(v *Options) { v.Workers = 33 },
		func(v *Options) { v.Clients = 0 }, func(v *Options) { v.Clients = 257 },
		func(v *Options) { v.Requests = 0 }, func(v *Options) { v.Requests = 1<<20 + 1 },
		func(v *Options) { v.Samples = 0 }, func(v *Options) { v.Samples = 1<<20 + 1 },
		func(v *Options) { v.Warmups = -1 }, func(v *Options) { v.Warmups = 1<<20 + 1 },
		func(v *Options) { v.Requests = 1 << 20 },
	} {
		args := valid
		change(&args)
		if call, stop, out, err := Start(nil, args); err == nil || call != nil || stop != nil || out != nil {
			t.Fatal("invalid admission created a live driver", args, err)
		}
	}
	for _, calls := range [][]Call{nil, {nil}} {
		if _, _, _, err := Start(calls, valid); err == nil {
			t.Fatal("invalid callback set accepted")
		}
	}
}
