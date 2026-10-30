package preview2

import (
	"strings"
	"sync"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const IOErrorInterface = "wasi:io/error@0.2.12"

type IOErrorConfig struct{ MaxErrors, MaxDebugBytes uint32 }

// IOErrorInfo retains the host cause for typed downstream downcasts. Guest code
// sees only Debug, an owned human-readable snapshot; its contents are not an ABI.
type IOErrorInfo struct {
	Debug string
	Cause error
}

// IOErrors owns its nominal error resource domain. Close releases abandoned
// entries independently of the component Store; explicit drops release promptly.
// Causes supplied by the embedder must remain safe for concurrent host access.
type IOErrors struct {
	mu       sync.Mutex
	config   IOErrorConfig
	resource *component.HostResourceType
	entries  map[uint32]IOErrorInfo
	next     uint64
	closed   bool
}

func NewIOErrors(config IOErrorConfig) (*IOErrors, error) {
	if config.MaxErrors == 0 {
		config.MaxErrors = 1024
	}
	if config.MaxDebugBytes == 0 {
		config.MaxDebugBytes = 4096
	}
	if config.MaxDebugBytes >= 1<<28 {
		return nil, wasm.ErrLimit
	}
	e := &IOErrors{config: config, entries: map[uint32]IOErrorInfo{}, next: 1}
	e.resource = component.NewHostResourceType(func(_ *component.Caller, id uint32) error {
		e.mu.Lock()
		delete(e.entries, id)
		e.mu.Unlock()
		return nil
	})
	return e, nil
}
func (e *IOErrors) Close() {
	if e == nil {
		return
	}
	e.mu.Lock()
	e.closed = true
	e.entries = nil
	e.mu.Unlock()
}

// NewError snapshots bounded valid UTF-8 before granting own<error> in the
// Caller's bound Store. It never parses a debug string into a host error code.
func (e *IOErrors) NewError(c *component.Caller, cause error) (component.Value, error) {
	if e == nil || e.resource == nil {
		return component.Value{}, vm.ErrReference
	}
	if cause == nil {
		return component.Value{}, wasm.ErrInvalid
	}
	debug := cause.Error()
	if uint64(len(debug)) > uint64(e.config.MaxDebugBytes) {
		debug = debug[:e.config.MaxDebugBytes]
	}
	// A truncated or invalid UTF-8 subsequence becomes one ASCII marker, so the
	// sanitized snapshot cannot exceed the original bounded byte length.
	debug = strings.ToValidUTF8(debug, "?")
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return component.Value{}, wasm.ErrClosed
	}
	if uint64(len(e.entries)) >= uint64(e.config.MaxErrors) || e.next > 1<<32-1 {
		return component.Value{}, wasm.ErrLimit
	}
	id := uint32(e.next)
	e.next++
	e.entries[id] = IOErrorInfo{Debug: debug, Cause: cause}
	value, err := c.NewResource(e.resource, id)
	if err != nil {
		delete(e.entries, id)
	}
	return value, err
}

// Describe requires an active Caller and the exact nominal Store capability.
// Returning the original host cause supports typed filesystem/socket downcasts.
func (e *IOErrors) Describe(c *component.Caller, ref *component.Resource) (IOErrorInfo, error) {
	if e == nil || e.resource == nil {
		return IOErrorInfo{}, vm.ErrReference
	}
	id, err := c.ResourceRep(e.resource, ref)
	if err != nil {
		return IOErrorInfo{}, err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return IOErrorInfo{}, wasm.ErrClosed
	}
	value, ok := e.entries[id]
	if !ok {
		return IOErrorInfo{}, vm.ErrReference
	}
	return value, nil
}
func (e *IOErrors) Imports() component.Imports {
	if e == nil || e.resource == nil {
		return component.Imports{}
	}
	const method = "[method]error.to-debug-string"
	check := func(f component.FunctionInfo) error {
		_, self := f.Parameter(0)
		if f.ParameterCount() != 1 || self.Kind() != component.KindBorrow || !self.IsResource(e.resource) || f.Result().Kind() != component.KindString {
			return vm.ErrType
		}
		return nil
	}
	read := func(c *component.Caller, args []component.Value) (*component.Value, error) {
		info, err := e.Describe(c, args[0].Resource)
		if err != nil {
			return nil, err
		}
		return &component.Value{Text: info.Debug}, c.Context().Err()
	}
	return component.Imports{Instances: map[string]component.Imports{IOErrorInterface: {Resources: map[string]*component.HostResourceType{"error": e.resource}, Functions: map[string]component.HostFunction{method: read}, FunctionChecks: map[string]component.FunctionCheck{method: check}}}}
}
