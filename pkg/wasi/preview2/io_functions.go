package preview2

import (
	"io"
	"strings"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

var inputMethods = []string{"read", "blocking-read", "skip", "blocking-skip", "subscribe"}
var outputMethods = []string{"check-write", "write", "blocking-write-and-flush", "flush", "blocking-flush", "subscribe", "write-zeroes", "blocking-write-zeroes-and-flush", "splice", "blocking-splice"}

func (p *IO) streamFailureType(t component.ValueType) bool {
	if t.Kind() != component.KindVariant || t.CaseCount() != 2 {
		return false
	}
	first, payload := t.Case(0)
	second, empty := t.Case(1)
	return first == "last-operation-failed" && payload.Kind() == component.KindOwn && payload.IsResource(p.errors.resource) && second == "closed" && !empty.Valid()
}
func (p *IO) methodType(input bool, name string) component.FunctionCheck {
	return func(f component.FunctionInfo) error {
		count := 1
		switch name {
		case "read", "blocking-read", "skip", "blocking-skip", "write", "blocking-write-and-flush", "write-zeroes", "blocking-write-zeroes-and-flush":
			count = 2
		case "splice", "blocking-splice":
			count = 3
		}
		_, self := f.Parameter(0)
		resource := p.outputType
		if input {
			resource = p.inputType
		}
		if f.ParameterCount() != count || self.Kind() != component.KindBorrow || !self.IsResource(resource) {
			return vm.ErrType
		}
		if count == 2 {
			_, arg := f.Parameter(1)
			if name == "write" || name == "blocking-write-and-flush" {
				if arg.Kind() != component.KindList || arg.Element().Kind() != component.KindU8 {
					return vm.ErrType
				}
			} else if arg.Kind() != component.KindU64 {
				return vm.ErrType
			}
		}
		if count == 3 {
			_, src := f.Parameter(1)
			_, length := f.Parameter(2)
			if src.Kind() != component.KindBorrow || !src.IsResource(p.inputType) || length.Kind() != component.KindU64 {
				return vm.ErrType
			}
		}
		result := f.Result()
		if name == "subscribe" {
			if result.Kind() == component.KindOwn && result.IsResource(p.poller.resource) {
				return nil
			}
			return vm.ErrType
		}
		if result.Kind() != component.KindResult || result.CaseCount() != 2 {
			return vm.ErrType
		}
		_, ok := result.Case(0)
		_, failure := result.Case(1)
		if !p.streamFailureType(failure) {
			return vm.ErrType
		}
		switch name {
		case "read", "blocking-read":
			if ok.Kind() != component.KindList || ok.Element().Kind() != component.KindU8 {
				return vm.ErrType
			}
		case "skip", "blocking-skip", "check-write", "splice", "blocking-splice":
			if ok.Kind() != component.KindU64 {
				return vm.ErrType
			}
		default:
			if ok.Valid() {
				return vm.ErrType
			}
		}
		return nil
	}
}
func (p *IO) method(input bool, name string) component.HostFunction {
	return func(c *component.Caller, args []component.Value) (*component.Value, error) {
		if err := p.begin(); err != nil {
			return nil, err
		}
		defer p.active.Done()
		if err := p.check(c.Context()); err != nil {
			return nil, err
		}
		r, err := p.record(c, args[0].Resource, input)
		if err != nil {
			return nil, err
		}
		if name == "subscribe" {
			return p.subscribe(c, r)
		}
		r.mu.Lock()
		closed := r.closed
		r.mu.Unlock()
		if closed {
			return p.streamError(c, r, io.ErrClosedPipe)
		}
		var value *component.Value
		switch name {
		case "read", "blocking-read", "skip", "blocking-skip":
			var data []byte
			data, err = r.read(c.Context(), args[1].Bits, strings.HasPrefix(name, "blocking-"))
			if err == nil {
				v := component.Value{Bits: uint64(len(data))}
				if name == "read" || name == "blocking-read" {
					v = component.ByteList(data)
				}
				value = &v
			}
		case "check-write":
			var n uint64
			n, err = r.checkWrite()
			value = &component.Value{Bits: n}
		case "write", "blocking-write-and-flush":
			var data string
			data, err = args[1].ByteString()
			if err == nil {
				if name == "write" {
					err = r.write(data)
				} else if len(data) > 4096 {
					err = wasm.ErrLimit
				} else {
					err = r.blockingWrite(c.Context(), data)
				}
			}
		case "write-zeroes", "blocking-write-zeroes-and-flush":
			n := args[1].Bits
			if name == "write-zeroes" {
				err = r.enqueue("", n)
			} else if n > 4096 {
				err = wasm.ErrLimit
			} else {
				err = r.blockingWrite(c.Context(), strings.Repeat("\x00", int(n)))
			}
		case "flush", "blocking-flush":
			err = r.flush()
			if err == nil && name == "blocking-flush" {
				err = r.waitFlush(c.Context())
			}
		case "splice", "blocking-splice":
			src, failed := p.record(c, args[1].Resource, true)
			if failed != nil {
				return nil, failed
			}
			var n uint64
			var failedStream *ioStream
			n, failedStream, err = r.splice(c.Context(), src, args[2].Bits, name == "blocking-splice")
			value = &component.Value{Bits: n}
			if err != nil {
				return p.streamError(c, failedStream, err)
			}
		}
		if err != nil {
			return p.streamError(c, r, err)
		}
		if err = p.check(c.Context()); err != nil {
			return nil, err
		}
		return streamOK(value), nil
	}
}

func (p *IO) Imports() component.Imports {
	if p == nil || p.inputType == nil {
		return component.Imports{}
	}
	out := p.poller.Imports()
	for name, imports := range p.errors.Imports().Instances {
		out.Instances[name] = imports
	}
	functions := map[string]component.HostFunction{}
	checks := map[string]component.FunctionCheck{}
	for _, input := range []bool{true, false} {
		prefix, methods := "output-stream", outputMethods
		if input {
			prefix, methods = "input-stream", inputMethods
		}
		for _, method := range methods {
			name := "[method]" + prefix + "." + method
			functions[name] = p.method(input, method)
			checks[name] = p.methodType(input, method)
		}
	}
	out.Instances[StreamsInterface] = component.Imports{Resources: map[string]*component.HostResourceType{"input-stream": p.inputType, "output-stream": p.outputType, "error": p.errors.resource, "pollable": p.poller.resource}, Functions: functions, FunctionChecks: checks}
	for _, kind := range []string{"stdin", "stdout", "stderr"} {
		input, source := kind == "stdin", p.stderr
		if kind == "stdin" {
			source = p.stdin
		} else if kind == "stdout" {
			source = p.stdout
		}
		resource, export := p.outputType, "output-stream"
		if input {
			resource, export = p.inputType, "input-stream"
		}
		name := "get-" + kind
		fn := func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			if err := p.begin(); err != nil {
				return nil, err
			}
			defer p.active.Done()
			v, err := p.publish(c, source)
			return &v, err
		}
		check := func(f component.FunctionInfo) error {
			if f.ParameterCount() != 0 || f.Result().Kind() != component.KindOwn || !f.Result().IsResource(resource) {
				return vm.ErrType
			}
			return nil
		}
		out.Instances["wasi:cli/"+kind+"@0.2.12"] = component.Imports{Resources: map[string]*component.HostResourceType{export: resource}, Functions: map[string]component.HostFunction{name: fn}, FunctionChecks: map[string]component.FunctionCheck{name: check}}
	}
	p.terminals(&out)

	return out
}

func (p *IO) terminals(out *component.Imports) {
	for name, typ := range map[string]*component.HostResourceType{"terminal-input": p.terminalInputType, "terminal-output": p.terminalOutputType} {
		out.Instances["wasi:cli/"+name+"@0.2.12"] = component.Imports{Resources: map[string]*component.HostResourceType{name: typ}}
	}
	for _, kind := range []string{"stdin", "stdout", "stderr"} {
		typ, name, present := p.terminalOutputType, "terminal-output", p.config.StderrTerminal
		if kind == "stdin" {
			typ, name, present = p.terminalInputType, "terminal-input", p.config.StdinTerminal
		} else if kind == "stdout" {
			present = p.config.StdoutTerminal
		}
		fn := func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			if err := p.begin(); err != nil {
				return nil, err
			}
			defer p.active.Done()
			value := &component.Value{}
			if present {
				resource, err := c.NewResource(typ, 0)
				if err != nil {
					return nil, err
				}
				value.Case = 1
				value.Elements = []component.Value{resource}
			}
			return value, p.check(c.Context())
		}
		check := func(f component.FunctionInfo) error {
			t := f.Result()
			if f.ParameterCount() != 0 || t.Kind() != component.KindOption || t.Element().Kind() != component.KindOwn || !t.Element().IsResource(typ) {
				return vm.ErrType
			}
			return nil
		}
		method := "get-terminal-" + kind
		out.Instances["wasi:cli/terminal-"+kind+"@0.2.12"] = component.Imports{Resources: map[string]*component.HostResourceType{name: typ}, Functions: map[string]component.HostFunction{method: fn}, FunctionChecks: map[string]component.FunctionCheck{method: check}}
	}
}
