package preview3

import (
	"context"
	"errors"
	"io"
	"sync"
	"syscall"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasi/preview2"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const (
	CLITypesInterface       = "wasi:cli/types@0.3.2"
	StdinInterface          = "wasi:cli/stdin@0.3.2"
	StdoutInterface         = "wasi:cli/stdout@0.3.2"
	StderrInterface         = "wasi:cli/stderr@0.3.2"
	TerminalInputInterface  = "wasi:cli/terminal-input@0.3.2"
	TerminalOutputInterface = "wasi:cli/terminal-output@0.3.2"
	TerminalStdinInterface  = "wasi:cli/terminal-stdin@0.3.2"
	TerminalStdoutInterface = "wasi:cli/terminal-stdout@0.3.2"
	TerminalStderrInterface = "wasi:cli/terminal-stderr@0.3.2"
)

// Sources retain Preview 2's prompt, cancellable readiness contract. The P3
// binding uses native streams/futures and never exports IO pollable resources.
type InputSource = preview2.InputSource
type OutputSource = preview2.OutputSource

var ErrWouldBlock = preview2.ErrWouldBlock

func NewBytesInput(data []byte) *preview2.BytesInput        { return preview2.NewBytesInput(data) }
func NewOutputBuffer(maximum uint32) *preview2.OutputBuffer { return preview2.NewOutputBuffer(maximum) }

type IOConfig struct {
	Stdin                                         InputSource
	Stdout, Stderr                                OutputSource
	MaxStreams, MaxReadBytes, MaxWriteBytes       uint32
	StdinTerminal, StdoutTerminal, StderrTerminal bool
}

// IO owns explicitly granted stdio sources and bounded operations. Successful
// construction transfers their ownership. Close cancels operations; WaitClosed
// joins source cleanup and external I/O workers. Defaults inherit no host stdio.
type IO struct {
	mu                    sync.Mutex
	config                IOConfig
	ctx                   context.Context
	cancel                context.CancelFunc
	closed                bool
	done                  chan struct{}
	workers               sync.WaitGroup
	activeWorkers         uint64
	operations            map[*ioOperation]struct{}
	cleanupErr            error
	inputType, outputType *component.HostResourceType
}

func NewIO(config IOConfig) (*IO, error) {
	if config.MaxStreams == 0 {
		config.MaxStreams = 1024
	}
	if config.MaxReadBytes == 0 {
		config.MaxReadBytes = 64 << 10
	}
	if config.MaxWriteBytes == 0 {
		config.MaxWriteBytes = 64 << 10
	}
	if config.MaxReadBytes >= 1<<28 || config.MaxWriteBytes >= 1<<28 {
		return nil, wasm.ErrLimit
	}
	if config.Stdin == nil {
		config.Stdin = NewBytesInput(nil)
	}
	if config.Stdout == nil {
		config.Stdout = &discardOutput{}
	}
	if config.Stderr == nil {
		config.Stderr = &discardOutput{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &IO{config: config, ctx: ctx, cancel: cancel, done: make(chan struct{}), operations: map[*ioOperation]struct{}{}, inputType: component.NewHostResourceType(nil), outputType: component.NewHostResourceType(nil)}, nil
}
func (p *IO) Close() {
	if p == nil {
		return
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.cancel()
	ops := make([]*ioOperation, 0, len(p.operations))
	for op := range p.operations {
		ops = append(ops, op)
	}
	p.mu.Unlock()
	for _, op := range ops {
		op.stop()
		op.finish(op.closeCode)
	}
	go func() {
		for _, close := range []func() error{p.config.Stdin.Close, p.config.Stdout.Close, p.config.Stderr.Close} {
			err := closeSource(close)
			if err != nil {
				p.mu.Lock()
				p.cleanupErr = errors.Join(p.cleanupErr, err)
				p.mu.Unlock()
			}
		}
		p.workers.Wait()
		close(p.done)
	}()
}
func closeSource(fn func() error) (err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("WASI source close panicked")
		}
	}()
	return fn()
}
func (p *IO) WaitClosed(ctx context.Context) error {
	if p == nil {
		return vm.ErrReference
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		p.mu.Lock()
		defer p.mu.Unlock()
		return p.cleanupErr
	}
}
func (p *IO) worker(fn func()) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return wasm.ErrClosed
	}
	if p.activeWorkers >= uint64(p.config.MaxStreams)*2 {
		p.mu.Unlock()
		return wasm.ErrLimit
	}
	p.activeWorkers++
	p.workers.Add(1)
	p.mu.Unlock()
	go func() { defer p.workers.Done(); defer func() { p.mu.Lock(); p.activeWorkers--; p.mu.Unlock() }(); fn() }()
	return nil
}
func (p *IO) operation() (*ioOperation, error) {
	return p.operationFor(cliError, cliCompletionValue, 3)
}
func (p *IO) operationFor(code func(error) uint32, encode func(uint32) component.Value, closedCode uint32) (*ioOperation, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, wasm.ErrClosed
	}
	if uint64(len(p.operations)) >= uint64(p.config.MaxStreams) {
		return nil, wasm.ErrLimit
	}
	ctx, cancel := context.WithCancel(p.ctx)
	op := &ioOperation{owner: p, ctx: ctx, cancel: cancel, code: code, closeCode: closedCode, cell: resultCell{done: make(chan struct{}), encode: encode}}
	p.operations[op] = struct{}{}
	return op, nil
}
func cliCompletionValue(code uint32) component.Value {
	value := component.Value{}
	if code != 0 {
		value.Case = 1
		value.Elements = []component.Value{{Bits: uint64(code - 1)}}
	}
	return value
}

type resultCell struct {
	mu     sync.Mutex
	done   chan struct{}
	value  uint32
	ready  bool
	encode func(uint32) component.Value
}

func (s *resultCell) set(code uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.ready {
		s.ready = true
		s.value = code
		close(s.done)
	}
}
func (s *resultCell) source(p *IO, close func()) component.AsyncSource {
	return component.AsyncSource{Close: close, Read: func(c *component.Caller, _ uint32) (*component.ReadCompletion, error) {
		completion := component.NewReadCompletion()
		ctx := c.Context()
		if err := p.worker(func() {
			select {
			case <-ctx.Done():
				completion.Cancelled()
				return
			case <-s.done:
			}
			s.mu.Lock()
			code := s.value
			s.mu.Unlock()
			completion.Resolve(func(*component.Caller) (component.ReadResult, error) {
				value := s.encode(code)
				return component.ReadResult{Values: []component.Value{value}, Closed: true}, nil
			})
		}); err != nil {
			return nil, err
		}
		return completion, nil
	}}
}

type ioOperation struct {
	owner             *IO
	mu                sync.Mutex
	ctx               context.Context
	cancel            context.CancelFunc
	cell              resultCell
	stopped, finished bool
	control           *component.Cancellation
	code              func(error) uint32
	closeCode         uint32
	release           func()
}

func (op *ioOperation) stop() {
	op.mu.Lock()
	op.stopped = true
	control := op.control
	op.cancel()
	op.mu.Unlock()
	if control != nil {
		control.Cancel()
	}
}
func (op *ioOperation) finish(code uint32) {
	op.mu.Lock()
	if op.finished {
		op.mu.Unlock()
		return
	}
	op.finished = true
	release := op.release
	op.release = nil
	op.mu.Unlock()
	op.cell.set(code)
	op.cancel()
	op.owner.mu.Lock()
	delete(op.owner.operations, op)
	op.owner.mu.Unlock()
	if release != nil {
		release()
	}
}
func (op *ioOperation) setRelease(fn func()) {
	op.mu.Lock()
	if op.finished {
		op.mu.Unlock()
		fn()
		return
	}
	op.release = fn
	op.mu.Unlock()
}
func (op *ioOperation) nextControl() *component.Cancellation {
	c := component.NewCancellation()
	op.mu.Lock()
	op.control = c
	stopped := op.stopped
	op.mu.Unlock()
	if stopped {
		c.Cancel()
	}
	return c
}
func waitReady(ctx context.Context, source preview2.PollSource) error {
	ready, changed := source.Readiness()
	if ready {
		return ctx.Err()
	}
	if changed == nil {
		return io.ErrNoProgress
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-changed:
		return ctx.Err()
	}
}
func cliError(err error) uint32 {
	if err == nil || errors.Is(err, io.EOF) {
		return 0
	}
	if errors.Is(err, syscall.EILSEQ) {
		return 2
	}
	if errors.Is(err, io.ErrClosedPipe) {
		return 3
	}
	return 1
}
func (p *IO) input(c *component.Caller) (*component.Value, error) {
	op, err := p.operation()
	if err != nil {
		return nil, err
	}
	_, streamType := c.FunctionInfo().Result().Field(0)
	_, futureType := c.FunctionInfo().Result().Field(1)
	return p.inputWithOperation(c, op, p.config.Stdin, streamType, futureType)
}
func (p *IO) inputWithOperation(c *component.Caller, op *ioOperation, source InputSource, streamType, futureType component.ValueType) (*component.Value, error) {
	accepted := false
	defer func() {
		if !accepted {
			op.stop()
			op.finish(1)
		}
	}()
	stream, err := c.NewSource(streamType, component.AsyncSource{Close: func() { op.stop(); op.finish(0) }, Read: func(c *component.Caller, limit uint32) (*component.ReadCompletion, error) {
		completion := component.NewReadCompletion()
		callCtx := c.Context()
		ctx, cancel := context.WithCancel(callCtx)
		stop := context.AfterFunc(op.ctx, cancel)
		if err := p.worker(func() {
			defer cancel()
			defer stop()
			var data []byte
			var readErr error
			defer func() {
				if recover() != nil {
					readErr = io.ErrUnexpectedEOF
					data = nil
				}
				if callCtx.Err() != nil && len(data) == 0 && (readErr == nil || errors.Is(readErr, context.Canceled) || errors.Is(readErr, context.DeadlineExceeded)) {
					completion.Cancelled()
					return
				}
				closed := readErr != nil
				if closed {
					op.finish(op.code(readErr))
				}
				result := component.ReadResult{Bytes: string(data), Closed: closed}
				completion.Resolve(func(*component.Caller) (component.ReadResult, error) { return result, nil })
			}()
			if limit == 0 {
				return
			}
			buf := make([]byte, min(limit, p.config.MaxReadBytes))
			for {
				if readErr = ctx.Err(); readErr != nil {
					return
				}
				n, failed := source.Read(ctx, buf)
				if n < 0 || n > len(buf) {
					readErr = io.ErrNoProgress
					return
				}
				if n > 0 {
					data = buf[:n]
					readErr = failed
					if errors.Is(failed, ErrWouldBlock) {
						readErr = nil
					}
					return
				}
				if errors.Is(failed, ErrWouldBlock) {
					if readErr = waitReady(ctx, source); readErr == nil {
						continue
					}
					return
				}
				if failed == nil {
					failed = io.ErrNoProgress
				}
				readErr = failed
				return
			}
		}); err != nil {
			stop()
			cancel()
			return nil, err
		}
		return completion, nil
	}})
	if err != nil {
		return nil, err
	}
	future, err := c.NewSource(futureType, op.cell.source(p, nil))
	if err != nil {
		return nil, err
	}
	accepted = true
	return &component.Value{Elements: []component.Value{stream, future}}, nil
}
func (p *IO) write(ctx context.Context, source OutputSource, data []byte, flush bool) (err error) {
	defer func() {
		if recover() != nil {
			err = io.ErrUnexpectedEOF
		}
	}()
	for len(data) > 0 {
		if err = ctx.Err(); err != nil {
			return err
		}
		n, failed := source.Write(ctx, data)
		if n < 0 || n > len(data) {
			return io.ErrNoProgress
		}
		data = data[n:]
		if failed != nil && !errors.Is(failed, ErrWouldBlock) {
			return failed
		}
		if len(data) == 0 {
			break
		}
		if n == 0 && !errors.Is(failed, ErrWouldBlock) {
			return io.ErrNoProgress
		}
		if errors.Is(failed, ErrWouldBlock) {
			if err = waitReady(ctx, source); err != nil {
				return err
			}
		}
	}
	if flush {
		for {
			if err = ctx.Err(); err != nil {
				return err
			}
			err = source.Flush(ctx)
			if !errors.Is(err, ErrWouldBlock) {
				return err
			}
			if err = waitReady(ctx, source); err != nil {
				return err
			}
		}
	}
	return nil
}
func (op *ioOperation) schedule(c *component.Caller, ref *component.Resource, source OutputSource) error {
	var eof bool
	var owned *component.Resource
	var failure error
	return c.Schedule(component.AsyncWork{Owned: []*component.Resource{ref}, Cancel: op.nextControl(), Start: func(c *component.Caller, refs []*component.Resource) (*component.HostCompletion, error) {
		owned = refs[0]
		batch, err := c.ReadAsync(owned, op.owner.config.MaxWriteBytes)
		if err != nil {
			op.finish(op.code(err))
			if c.Context().Err() != nil {
				p := component.NewHostCompletion()
				p.Cancelled()
				return p, nil
			}
			return nil, err
		}
		eof = batch.Closed
		pending := component.NewHostCompletion()
		ctx := c.Context()
		if err := op.owner.worker(func() {
			failure = op.owner.write(ctx, source, []byte(batch.Bytes), eof)
			if ctx.Err() != nil {
				pending.Cancelled()
			} else {
				pending.Resolve(func(*component.Caller) (*component.Value, error) { return nil, nil })
			}
		}); err != nil {
			op.finish(op.code(err))
			return nil, err
		}
		return pending, nil
	}, Done: func(c *component.Caller, cancelled bool) error {
		if cancelled {
			op.finish(0)
			return nil
		}
		if eof || failure != nil {
			op.finish(op.code(failure))
			return nil
		}
		if err := op.schedule(c, owned, source); err != nil {
			op.finish(op.code(err))
			return err
		}
		return nil
	}})
}
func (p *IO) output(c *component.Caller, args []component.Value, source OutputSource) (*component.Value, error) {
	op, err := p.operation()
	if err != nil {
		return nil, err
	}
	return p.outputWithOperation(c, args, source, op)
}
func (p *IO) outputWithOperation(c *component.Caller, args []component.Value, source OutputSource, op *ioOperation) (*component.Value, error) {
	accepted := false
	defer func() {
		if !accepted {
			op.stop()
			op.finish(1)
		}
	}()
	future, err := c.NewSource(c.FunctionInfo().Result(), op.cell.source(p, func() { op.stop(); op.finish(0) }))
	if err != nil {
		return nil, err
	}
	if err := op.schedule(c, args[0].Resource, source); err != nil {
		return nil, err
	}
	accepted = true
	return &future, nil
}
func resultType(t component.ValueType) bool {
	if t.Kind() != component.KindFuture {
		return false
	}
	r := t.Element()
	_, ok := r.Case(0)
	_, bad := r.Case(1)
	if r.Kind() != component.KindResult || r.CaseCount() != 2 || ok.Valid() || bad.Kind() != component.KindEnum || bad.LabelCount() != 3 {
		return false
	}
	return bad.Label(0) == "io" && bad.Label(1) == "illegal-byte-sequence" && bad.Label(2) == "pipe"
}
func streamType(t component.ValueType) bool {
	return t.Kind() == component.KindStream && t.Element().Kind() == component.KindU8
}
func (p *IO) Imports() component.Imports {
	if p == nil || p.inputType == nil {
		return component.Imports{}
	}
	out := component.Imports{Instances: map[string]component.Imports{CLITypesInterface: {}, TerminalInputInterface: {Resources: map[string]*component.HostResourceType{"terminal-input": p.inputType}}, TerminalOutputInterface: {Resources: map[string]*component.HostResourceType{"terminal-output": p.outputType}}}}
	for _, name := range []string{StdinInterface, StdoutInterface, StderrInterface} {
		input := name == StdinInterface
		operation := "write-via-stream"
		var fn component.HostFunction
		if input {
			operation = "read-via-stream"
			fn = func(c *component.Caller, _ []component.Value) (*component.Value, error) { return p.input(c) }
		} else {
			source := p.config.Stdout
			if name == StderrInterface {
				source = p.config.Stderr
			}
			fn = func(c *component.Caller, args []component.Value) (*component.Value, error) {
				return p.output(c, args, source)
			}
		}
		check := func(f component.FunctionInfo) error {
			if f.Async() {
				return vm.ErrType
			}
			if input {
				_, a := f.Result().Field(0)
				_, b := f.Result().Field(1)
				if f.ParameterCount() != 0 || f.Result().Kind() != component.KindTuple || f.Result().FieldCount() != 2 || !streamType(a) || !resultType(b) {
					return vm.ErrType
				}
			} else {
				_, a := f.Parameter(0)
				if f.ParameterCount() != 1 || !streamType(a) || !resultType(f.Result()) {
					return vm.ErrType
				}
			}
			return nil
		}
		out.Instances[name] = component.Imports{Functions: map[string]component.HostFunction{operation: fn}, FunctionChecks: map[string]component.FunctionCheck{operation: check}}
	}
	for i, name := range []string{TerminalStdinInterface, TerminalStdoutInterface, TerminalStderrInterface} {
		operation := []string{"get-terminal-stdin", "get-terminal-stdout", "get-terminal-stderr"}[i]
		present := []bool{p.config.StdinTerminal, p.config.StdoutTerminal, p.config.StderrTerminal}[i]
		typ := p.outputType
		resource := "terminal-output"
		if i == 0 {
			typ = p.inputType
			resource = "terminal-input"
		}
		out.Instances[name] = component.Imports{Resources: map[string]*component.HostResourceType{resource: typ}, Functions: map[string]component.HostFunction{operation: func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			if err := p.ctx.Err(); err != nil {
				return nil, wasm.ErrClosed
			}
			result := &component.Value{}
			if present {
				v, err := c.NewResource(typ, 0)
				if err != nil {
					return nil, err
				}
				result.Case = 1
				result.Elements = []component.Value{v}
			}
			return result, nil
		}}, FunctionChecks: map[string]component.FunctionCheck{operation: func(f component.FunctionInfo) error {
			t := f.Result()
			if f.Async() || f.ParameterCount() != 0 || t.Kind() != component.KindOption || t.Element().Kind() != component.KindOwn || !t.Element().IsResource(typ) {
				return vm.ErrType
			}
			return nil
		}}}
	}
	return out
}

type discardOutput struct{}

func (*discardOutput) Write(ctx context.Context, b []byte) (int, error) { return len(b), ctx.Err() }
func (*discardOutput) Flush(ctx context.Context) error                  { return ctx.Err() }
func (*discardOutput) Close() error                                     { return nil }
func (*discardOutput) Readiness() (bool, <-chan struct{})               { return true, nil }
