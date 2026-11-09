// Package service supplies the shared bounded admission and measurement driver
// for owned-engine, independent-engine and native application benchmarks.
package service

import (
	"errors"
	"runtime"
	"sync"
	"time"
)

// Call times one public operation and checks its result before returning.
// Each worker receives an independent callback and its private mutable state.
type Call func() (elapsed int64, result uint32, err error)

type Options struct {
	Workers, Clients, Requests, Samples, Warmups int
	Expected                                     uint32
	Resources                                    bool
}

// Validate bounds all recorded first, warmup and measured requests.
func (args Options) Validate() error {
	if args.Workers < 1 || args.Workers > 32 || args.Clients < 1 || args.Clients > 256 || args.Requests < 1 || args.Requests > 1<<20 {
		return errors.New("service requires workers 1..32, clients 1..256 and requests 1..1048576")
	}
	if args.Samples < 1 || args.Warmups < 0 || args.Samples > 1<<20 || args.Warmups > 1<<20 || int64(args.Clients)*int64(args.Requests)*int64(1+args.Samples+args.Warmups) > 1<<20 {
		return errors.New("service exceeds 1048576 recorded requests across all batches")
	}
	return nil
}

type Report struct {
	Workers, Clients, RequestsPerClient int
	GOMAXPROCS                          int
	WorkerSetupNs                       int64
	AllocationScope                     string
	Batches                             []Batch
	ResourceSnapshots                   []Resources `json:",omitempty"`
	ResourceInstrumented                bool
}

type Batch struct {
	WallNs, TotalAllocBytes uint64
	ReadinessNs             int64
	ReadyClients            int
	Mallocs                 *uint64 `json:",omitempty"`
	HeapBefore, HeapAfter   uint64
	GCBefore, GCAfter       uint32
	GoroutinesBefore        int
	GoroutinesAfter         int
	RequestNs, PublicAPINs  []int64
}

// Workers and clients live across batches. Setup/teardown, sample allocation,
// MemStats queries and result reduction are outside the batch wall timer.
// The timer includes admission/backpressure, public calls and reply delivery.
// The coordinator invokes the returned call serially and stops only after the
// last call returns. Callbacks must return errors rather than panic.
func Start(calls []Call, args Options) (Call, func(), *Report, error) {
	if err := args.Validate(); err != nil {
		return nil, nil, nil, err
	}
	if len(calls) != args.Workers {
		return nil, nil, nil, errors.New("service callback count must equal workers")
	}
	for _, call := range calls {
		if call == nil {
			return nil, nil, nil, errors.New("service callback must not be nil")
		}
	}
	type response struct {
		ns  int64
		err error
	}
	type gate struct {
		start chan struct{}
		epoch time.Time
	}
	jobs := make(chan int, len(calls))
	replies := make([]chan response, args.Clients)
	gates := make([]chan *gate, args.Clients)
	ready := make(chan int, args.Clients)
	workerReady := make(chan struct{}, len(calls))
	finished := make(chan int, args.Clients)
	var workers, clients sync.WaitGroup
	for _, call := range calls {
		workers.Go(func() {
			workerReady <- struct{}{}
			for client := range jobs {
				ns, _, err := call()
				replies[client] <- response{ns, err}
			}
		})
	}
	for range calls {
		<-workerReady
	}
	var batch Batch
	requestErrors := make([]error, args.Clients*args.Requests)
	for client := range args.Clients {
		replies[client] = make(chan response, 1)
		gates[client] = make(chan *gate)
		clients.Go(func() {
			for next := range gates[client] {
				ready <- client
				<-next.start
				for request := range args.Requests {
					index := client*args.Requests + request
					start := time.Now()
					if request == 0 {
						// All initial arrivals share the release epoch. Scheduling
						// delay after broadcast belongs to first-request latency.
						start = next.epoch
					}
					jobs <- client
					r := <-replies[client]
					batch.RequestNs[index] = time.Since(start).Nanoseconds()
					batch.PublicAPINs[index], requestErrors[index] = r.ns, r.err
				}
				finished <- client
			}
		})
	}

	out := &Report{
		Workers: len(calls), Clients: args.Clients, RequestsPerClient: args.Requests,
		GOMAXPROCS:           runtime.GOMAXPROCS(0),
		AllocationScope:      "runtime TotalAlloc difference; Go counts managed allocations, LLGo reports BDWGC collector bytes. Mallocs omitted under LLGo. Neither includes all native/JIT/linear memory; heap snapshots are not peak RSS.",
		ResourceInstrumented: args.Resources,
	}
	return func() (int64, uint32, error) {
			if args.Resources && len(out.Batches) == 0 {
				snapshot, err := resourceSnapshot()
				if err != nil {
					return 0, 0, err
				}
				out.ResourceSnapshots = append(out.ResourceSnapshots, snapshot)
			}
			batch = Batch{RequestNs: make([]int64, len(requestErrors)), PublicAPINs: make([]int64, len(requestErrors))}
			next := &gate{start: make(chan struct{})}
			prepare := time.Now()
			for _, gate := range gates {
				gate <- next
			}
			for range gates {
				<-ready
				batch.ReadyClients++
			}
			batch.ReadinessNs = time.Since(prepare).Nanoseconds()
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			batch.GoroutinesBefore = runtime.NumGoroutine()
			start := time.Now()
			next.epoch = start
			close(next.start)
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
			out.Batches = append(out.Batches, batch)
			if args.Resources {
				snapshot, err := resourceSnapshot()
				if err != nil {
					return ns, 0, err
				}
				out.ResourceSnapshots = append(out.ResourceSnapshots, snapshot)
			}
			for _, err := range requestErrors {
				if err != nil {
					return ns, 0, err
				}
			}
			return ns, args.Expected, nil
		}, func() {
			for _, gate := range gates {
				close(gate)
			}
			clients.Wait()
			close(jobs)
			workers.Wait()
		}, out, nil
}
