//go:build linux || darwin

package preview3

import (
	"context"
	"strings"
	"sync"

	"github.com/cpunion/wasmvm/pkg/component"
)

type fileCall struct {
	mu     sync.Mutex
	closed bool
	result *fileDescriptor
	cancel context.CancelFunc
	stop   func() bool
}

func (j *fileCall) close() {
	j.mu.Lock()
	if j.closed {
		j.mu.Unlock()
		return
	}
	j.closed = true
	result := j.result
	j.result = nil
	j.mu.Unlock()
	j.cancel()
	j.stop()
	if result != nil {
		result.object.release()
	}
}
func (j *fileCall) publish(result *fileDescriptor) {
	if result == nil {
		return
	}
	j.mu.Lock()
	if j.closed {
		j.mu.Unlock()
		result.object.release()
		return
	}
	j.result = result
	j.mu.Unlock()
}
func (p *FileSystem) method(name string) component.AsyncHostFunction {
	return func(c *component.Caller, args []component.Value) (*component.HostCompletion, error) {
		if err := p.check(c.Context()); err != nil {
			return nil, err
		}
		d, err := p.descriptor(c, args[0].Resource)
		if err != nil {
			return nil, err
		}
		var other fileDescriptor
		if name == "link-at" || name == "rename-at" || name == "is-same-object" {
			index := 2
			if name == "link-at" {
				index = 3
			}
			if name == "is-same-object" {
				index = 1
			}
			other, err = p.descriptor(c, args[index].Resource)
			if err != nil {
				d.object.release()
				return nil, err
			}
		}
		if err := p.begin(); err != nil {
			d.object.release()
			if other.object != nil {
				other.object.release()
			}
			return nil, err
		}
		callCtx := c.Context()
		ctx, cancel := context.WithCancel(callCtx)
		job := &fileCall{cancel: cancel}
		job.stop = context.AfterFunc(p.ctx, cancel)
		completion := component.NewHostCompletionWithCleanup(job.close)
		snapshot := append([]component.Value(nil), args...)
		for i := range snapshot {
			snapshot[i].Resource = nil
		}
		go func() {
			defer p.workers.Done()
			var value *component.Value
			var opened *fileDescriptor
			var failed error
			defer func() {
				panicked := recover()
				job.publish(opened)
				d.object.release()
				if other.object != nil {
					other.object.release()
				}
				p.releaseSlot()
				if callCtx.Err() != nil {
					completion.Cancelled()
					return
				}
				completion.Resolve(func(fresh *component.Caller) (*component.Value, error) {
					if panicked != nil {
						panic(panicked)
					}
					if opened != nil && failed == nil {
						var v component.Value
						v, failed = p.publish(fresh, *opened)
						if failed == nil {
							value = &v
						}
					}
					if name == "is-same-object" {
						return value, failed
					}
					return fileResult(value, failed), nil
				})
			}()
			value, opened, failed = p.execute(ctx, d, other, name, snapshot)
		}()
		return completion, nil
	}
}
func (p *FileSystem) Imports() component.Imports {
	if p == nil || p.resource == nil {
		return component.Imports{}
	}
	resources := map[string]*component.HostResourceType{"descriptor": p.resource}
	types := component.Imports{Resources: resources, Functions: map[string]component.HostFunction{}, AsyncCompletions: map[string]component.AsyncHostFunction{}, FunctionChecks: map[string]component.FunctionCheck{}}
	for key, schema := range filesystemWITFunctions["types"] {
		name := strings.TrimPrefix(key, "[method]descriptor.")
		types.FunctionChecks[key] = schema.check(resources)
		if schema.async {
			types.AsyncCompletions[key] = p.method(name)
		} else {
			types.Functions[key] = func(c *component.Caller, args []component.Value) (*component.Value, error) {
				return p.stream(c, name, args)
			}
		}
	}
	preopens := component.Imports{Resources: resources, Functions: map[string]component.HostFunction{"get-directories": p.getDirectories}, FunctionChecks: map[string]component.FunctionCheck{"get-directories": filesystemWITFunctions["preopens"]["get-directories"].check(resources)}}
	return component.Imports{Instances: map[string]component.Imports{FileSystemTypesInterface: types, FileSystemPreopensInterface: preopens}}
}
