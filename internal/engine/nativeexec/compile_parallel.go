package nativeexec

import (
	"context"
	"runtime"
	"sync"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

type functionCompileResult struct {
	body       *codegen.CompiledFunction
	err        error
	panicked   bool
	panicValue any
}

func compileFunction(ctx context.Context, f *slots.Function, limits codegen.Limits, config codegen.Config) (result functionCompileResult) {
	panicking := true
	defer func() {
		if panicking {
			result.panicValue = recover()
			result.panicked = true
		}
	}()
	if result.err = ctx.Err(); result.err == nil {
		result.body, result.err = codegen.CompileFunctionWithConfig(runtime.GOARCH, f, limits, config)
	}
	panicking = false
	return
}

type compilationJob struct {
	index   int
	maximum uint64
}

type compilationWorker struct {
	jobs    chan compilationJob
	results chan functionCompileResult
}

// At most four bodies can be running or awaiting ordered publication. The ring
// grants a new job only after the preceding result in that slot was consumed.
// Each slot has a persistent worker and single-producer, single-consumer queues.
type functionCompiler struct {
	ctx           context.Context
	cancel        context.CancelFunc
	functions     []*slots.Function
	limits        codegen.Limits
	config        codegen.Config
	indices       []int
	queued, taken int
	workers       []compilationWorker
	joined        sync.WaitGroup
	once          sync.Once
}

func (p *functionCompiler) functionConfig(index int) codegen.Config {
	config := p.config
	if config.NativeFrameRegisters {
		config.NativeFunction = uint64(config.ScalarImports) + uint64(index) + 1
	}
	return config
}

func newFunctionCompiler(ctx context.Context, functions []*slots.Function, limits codegen.Limits, config codegen.Config, optional bool, wanted map[*slots.Function]bool) *functionCompiler {
	p := &functionCompiler{ctx: ctx, functions: functions, limits: limits, config: config}
	for i, f := range functions {
		if !optional || wanted[f] {
			p.indices = append(p.indices, i)
		}
	}
	workers := min(4, runtime.GOMAXPROCS(0), len(p.indices))
	if workers <= 1 {
		return p
	}
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.workers = make([]compilationWorker, workers)
	p.joined.Add(workers)
	for i := range p.workers {
		worker := compilationWorker{jobs: make(chan compilationJob, 1), results: make(chan functionCompileResult, 1)}
		p.workers[i] = worker
		go func(worker compilationWorker) {
			defer p.joined.Done()
			for job := range worker.jobs {
				bounded := p.limits
				bounded.MaxCodeBytes = job.maximum
				worker.results <- compileFunction(p.ctx, p.functions[job.index], bounded, p.functionConfig(job.index))
			}
		}(worker)
	}
	return p
}
func (p *functionCompiler) next(maximum uint64) functionCompileResult {
	workers := len(p.workers)
	if workers == 0 {
		i := p.indices[p.taken]
		p.taken++
		bounded := p.limits
		bounded.MaxCodeBytes = maximum
		return compileFunction(p.ctx, p.functions[i], bounded, p.functionConfig(i))
	}
	for p.queued < len(p.indices) && p.queued-p.taken < workers {
		worker := p.workers[p.queued%workers]
		worker.jobs <- compilationJob{p.indices[p.queued], maximum}
		p.queued++
	}
	result := <-p.workers[p.taken%workers].results
	p.taken++
	return result
}
func (p *functionCompiler) close() {
	p.once.Do(func() {
		if len(p.workers) != 0 {
			p.cancel()
			for _, worker := range p.workers {
				close(worker.jobs)
			}
			p.joined.Wait()
		}
	})
}
