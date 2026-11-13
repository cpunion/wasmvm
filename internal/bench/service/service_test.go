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

func TestServiceReportsManagedAllocationEvents(t *testing.T) {
	const count = 8
	retained := make([]*[1024]byte, 0, count)
	callback := func() (int64, uint32, error) {
		buffer := new([1024]byte)
		buffer[0] = byte(len(retained) + 1)
		retained = append(retained, buffer)
		return 1, 7, nil
	}
	call, stop, out, err := Start([]Call{callback}, Options{Workers: 1, Clients: 1, Requests: count, Samples: 1, Expected: 7})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	if _, result, err := call(); err != nil || result != 7 {
		t.Fatal(result, err)
	}
	batch := out.Batches[0]
	if batch.Mallocs == nil || *batch.Mallocs < count || batch.TotalAllocBytes < count*1024 {
		t.Fatal("managed allocation events missing", batch.Mallocs, batch.TotalAllocBytes)
	}
	if len(retained) != count {
		t.Fatal(len(retained))
	}
	for i, buffer := range retained {
		if buffer[0] != byte(i+1) {
			t.Fatal("retained storage was reused", i, buffer[0])
		}
	}
}

func TestResourceCohortChecksSnapshotsBeforeAdmittingRequests(t *testing.T) {
	args := Options{Workers: 1, Clients: 1, Requests: 1, Samples: 1, Expected: 7, Resources: true}
	var invoked atomic.Int32
	callback := func() (int64, uint32, error) { invoked.Add(1); return 1, 7, nil }
	call, stop, out, err := Start([]Call{callback}, args)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = call()
	stop()
	if err != nil || len(out.ResourceSnapshots) != 2 || invoked.Load() != 1 {
		t.Fatal(err, out, invoked.Load())
	}
	for _, snapshot := range out.ResourceSnapshots {
		if snapshot.RSSBytes == 0 || snapshot.PeakRSSBytes == 0 || snapshot.OSThreads <= 0 || snapshot.Scope == "" {
			t.Fatal(snapshot)
		}
	}
	t.Setenv("PATH", t.TempDir())
	call, stop, out, err = Start([]Call{callback}, args)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = call()
	stop()
	if err == nil || invoked.Load() != 1 || len(out.Batches) != 0 || len(out.ResourceSnapshots) != 0 {
		t.Fatal("failed snapshot admitted requests", err, invoked.Load(), out)
	}
}
