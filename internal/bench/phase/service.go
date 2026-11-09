package phase

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"
)

// This is an application admission policy, independent of GOMAXPROCS. Native
// guest entry and LLGo's scheduler do not acquire a Go scheduler CPU permit.
func serviceOptions(args options) error {
	if args.mode != "service" {
		if args.workers != 0 || args.clients != 0 || args.requests != 0 || args.resources {
			return errors.New("service counts require mode service")
		}
		return nil
	}
	if args.workers < 1 || args.workers > 32 || args.clients < 1 || args.clients > 256 || args.requests < 1 || args.requests > 1<<20 {
		return errors.New("service requires workers 1..32, clients 1..256 and requests 1..1048576")
	}
	if int64(args.clients)*int64(args.requests)*int64(1+args.samples+args.warmups) > 1<<20 {
		return errors.New("service exceeds 1048576 recorded requests across all batches")
	}
	return nil
}

type serviceReport struct {
	Workers, Clients, RequestsPerClient int
	GOMAXPROCS                          int
	WorkerSetupNs                       int64
	AllocationScope                     string
	Batches                             []serviceBatch
	ResourceSnapshots                   []serviceResources `json:",omitempty"`
	ResourceInstrumented                bool
}

type serviceBatch struct {
	WallNs, TotalAllocBytes uint64
	Mallocs                 *uint64 `json:",omitempty"`
	HeapBefore, HeapAfter   uint64
	GCBefore, GCAfter       uint32
	GoroutinesBefore        int
	GoroutinesAfter         int
	RequestNs, PublicAPINs  []int64
}

func serviceKernel(ctx context.Context, input []byte, args options, out *report) (kernel, func(), error) {
	if err := serviceOptions(args); err != nil {
		return nil, func() {}, err
	}
	calls, release, err := publicKernels(ctx, input, args, out, args.workers)
	if err != nil {
		return nil, release, err
	}
	start := time.Now()
	call, stop := boundedService(calls, args, out)
	out.Service.WorkerSetupNs = time.Since(start).Nanoseconds()
	return call, func() {
		start := time.Now()
		stop()
		release()
		out.Phases.Cleanup = time.Since(start).Nanoseconds()
	}, nil
}

// Workers and clients live across batches. Setup/teardown, sample allocation,
// MemStats queries and result reduction are outside the batch wall timer.
// The timer includes admission/backpressure, public calls and reply delivery.
func boundedService(calls []kernel, args options, out *report) (kernel, func()) {
	type response struct {
		ns  int64
		err error
	}
	jobs := make(chan int, len(calls))
	replies := make([]chan response, args.clients)
	gates := make([]chan struct{}, args.clients)
	finished := make(chan int, args.clients)
	var workers, clients sync.WaitGroup
	for _, call := range calls {
		workers.Go(func() {
			for client := range jobs {
				ns, _, err := call()
				replies[client] <- response{ns, err}
			}
		})
	}
	var batch serviceBatch
	requestErrors := make([]error, args.clients*args.requests)
	for client := range args.clients {
		replies[client] = make(chan response, 1)
		gates[client] = make(chan struct{})
		clients.Go(func() {
			for range gates[client] {
				for request := range args.requests {
					index := client*args.requests + request
					start := time.Now()
					jobs <- client
					r := <-replies[client]
					batch.RequestNs[index] = time.Since(start).Nanoseconds()
					batch.PublicAPINs[index], requestErrors[index] = r.ns, r.err
				}
				finished <- client
			}
		})
	}

	out.Service = &serviceReport{
		Workers: len(calls), Clients: args.clients, RequestsPerClient: args.requests,
		GOMAXPROCS:           runtime.GOMAXPROCS(0),
		AllocationScope:      "runtime TotalAlloc difference; Go counts managed allocations, LLGo reports BDWGC collector bytes. Mallocs omitted under LLGo. Neither includes all native/JIT/linear memory; heap snapshots are not peak RSS.",
		ResourceInstrumented: args.resources,
	}
	return func() (int64, uint32, error) {
			if args.resources && len(out.Service.Batches) == 0 {
				snapshot, err := serviceResourceSnapshot()
				if err != nil {
					return 0, 0, err
				}
				out.Service.ResourceSnapshots = append(out.Service.ResourceSnapshots, snapshot)
			}
			batch = serviceBatch{RequestNs: make([]int64, len(requestErrors)), PublicAPINs: make([]int64, len(requestErrors))}
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			batch.GoroutinesBefore = runtime.NumGoroutine()
			start := time.Now()
			for _, gate := range gates {
				gate <- struct{}{}
			}
			for range gates {
				<-finished
			}
			ns := time.Since(start).Nanoseconds()
			batch.GoroutinesAfter = runtime.NumGoroutine()
			runtime.ReadMemStats(&after)
			batch.WallNs = uint64(ns)
			batch.TotalAllocBytes = after.TotalAlloc - before.TotalAlloc
			batch.HeapBefore, batch.HeapAfter = before.HeapAlloc, after.HeapAlloc
			batch.GCBefore, batch.GCAfter = before.NumGC, after.NumGC
			if runtime.Compiler != "llgo" {
				mallocs := after.Mallocs - before.Mallocs
				batch.Mallocs = &mallocs
			}
			out.Service.Batches = append(out.Service.Batches, batch)
			if args.resources {
				snapshot, err := serviceResourceSnapshot()
				if err != nil {
					return ns, 0, err
				}
				out.Service.ResourceSnapshots = append(out.Service.ResourceSnapshots, snapshot)
			}
			for _, err := range requestErrors {
				if err != nil {
					return ns, 0, err
				}
			}
			return ns, args.expected, nil
		}, func() {
			for _, gate := range gates {
				close(gate)
			}
			clients.Wait()
			close(jobs)
			workers.Wait()
		}
}
