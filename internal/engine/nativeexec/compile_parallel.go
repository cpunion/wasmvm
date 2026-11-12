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
	result  chan functionCompileResult
}

// At most four bodies can be running or awaiting ordered publication. The ring
// grants a new job only after the preceding result in that slot was consumed.
// Persistent workers avoid per-function goroutines and all close paths join them.
type functionCompiler struct {
	ctx           context.Context
	cancel        context.CancelFunc
	functions     []*slots.Function
	limits        codegen.Limits
	config        codegen.Config
	indices       []int
	queued, taken int
	jobs          chan compilationJob
	results       []chan functionCompileResult
	joined        sync.WaitGroup
	once          sync.Once
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
	p.jobs = make(chan compilationJob, workers)
	p.results = make([]chan functionCompileResult, workers)
	for i := range p.results {
		p.results[i] = make(chan functionCompileResult, 1)
	}
	p.joined.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer p.joined.Done()
			for job := range p.jobs {
				bounded := p.limits
				bounded.MaxCodeBytes = job.maximum
				job.result <- compileFunction(p.ctx, p.functions[job.index], bounded, p.config)
			}
		}()
	}
	return p
}
func (p *functionCompiler) next(maximum uint64) functionCompileResult {
	if p.jobs == nil {
		i := p.indices[p.taken]
		p.taken++
		bounded := p.limits
		bounded.MaxCodeBytes = maximum
		return compileFunction(p.ctx, p.functions[i], bounded, p.config)
	}
	for p.queued < len(p.indices) && p.queued-p.taken < len(p.results) {
		slot := p.queued % len(p.results)
		p.jobs <- compilationJob{p.indices[p.queued], maximum, p.results[slot]}
		p.queued++
	}
	slot := p.taken % len(p.results)
	result := <-p.results[slot]
	p.taken++
	return result
}
func (p *functionCompiler) close() {
	p.once.Do(func() {
		if p.jobs != nil {
			p.cancel()
			close(p.jobs)
			p.joined.Wait()
		}
	})
}
