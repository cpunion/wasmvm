package preview3

import "github.com/cpunion/wasmvm/pkg/wasm"

// Each accepted operation retains the existing active quota. Idle goroutines
// are separately bounded and cannot keep a request closure or body alive.
// A blocking operation never queues behind another operation on a busy worker.
func (p *IO) worker(fn func()) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return wasm.ErrClosed
	}
	if p.activeWorkers >= uint64(p.config.MaxStreams)*2 {
		return wasm.ErrLimit
	}
	p.activeWorkers++
	if n := len(p.idleWorkers); n != 0 {
		jobs := p.idleWorkers[n-1]
		p.idleWorkers[n-1] = nil
		p.idleWorkers = p.idleWorkers[:n-1]
		// The one-slot channel belongs to an idle worker. Receiving its last
		// job precedes its publication here, so this send cannot block.
		jobs <- fn
		return nil
	}
	jobs := make(chan func(), 1)
	p.workers.Add(1)
	go p.runWorker(jobs, fn)
	return nil
}

func (p *IO) runWorker(jobs chan func(), fn func()) {
	defer p.workers.Done()
	for {
		p.runWorkerCall(fn)
		fn = nil
		p.mu.Lock()
		if p.closed || uint64(len(p.idleWorkers)) >= uint64(p.config.MaxIdleWorkers) {
			p.mu.Unlock()
			return
		}
		p.idleWorkers = append(p.idleWorkers, jobs)
		p.mu.Unlock()
		var ok bool
		fn, ok = <-jobs
		if !ok {
			return
		}
	}
}

func (p *IO) runWorkerCall(fn func()) {
	defer p.finishWorkerCall()
	fn()
}

func (p *IO) finishWorkerCall() {
	p.mu.Lock()
	p.activeWorkers--
	p.mu.Unlock()
}
