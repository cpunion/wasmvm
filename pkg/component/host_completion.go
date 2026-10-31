package component

import (
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"sync"
)

// AsyncHostFunction starts work without retaining its Caller. The completion's
// resolver runs in the Store session with a fresh Caller when work is ready.
type AsyncHostFunction func(*Caller, []Value) (*HostCompletion, error)

// HostCompletion is a one-shot handoff from external work to one invocation.
// Resolve and Cancelled may be called from another goroutine. Resource creation
// and guest reentry belong in Resolve's function, where Caller is valid.
type HostCompletion struct {
	mu      sync.Mutex
	state   uint8
	bound   bool
	resolve func(*Caller) (*Value, error)
	wake    chan<- struct{}
	cleanup func()
}

func NewHostCompletion() *HostCompletion { return &HostCompletion{} }

// NewHostCompletionWithCleanup releases external, unpublished results after the
// invocation finishes or is abandoned. Cleanup runs once after an accepted
// completion, outside its mutex, and must be nonblocking and must not enter the
// VM. The starter retains cleanup responsibility if it fails before acceptance.
func NewHostCompletionWithCleanup(cleanup func()) *HostCompletion {
	return &HostCompletion{cleanup: cleanup}
}
func (p *HostCompletion) Resolve(fn func(*Caller) (*Value, error)) error {
	if p == nil || fn == nil {
		return vm.ErrType
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state != 0 {
		return wasm.ErrClosed
	}
	p.resolve, p.state = fn, 1
	p.notify()
	return nil
}

// Cancelled acknowledges a cancellation request without publishing a result.
func (p *HostCompletion) Cancelled() error {
	if p == nil {
		return vm.ErrType
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state != 0 {
		return wasm.ErrClosed
	}
	p.state = 2
	p.notify()
	return nil
}
func (p *HostCompletion) notify() {
	if p.wake != nil {
		select {
		case p.wake <- struct{}{}:
		default:
		}
	}
}
func (p *HostCompletion) bind(wake chan<- struct{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bound || p.state == 3 {
		return vm.ErrType
	}
	p.bound = true
	p.wake = wake
	if p.state != 0 {
		p.notify()
	}
	return nil
}
func (p *HostCompletion) ready() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state == 1 || p.state == 2
}
func (p *HostCompletion) take() (func(*Caller) (*Value, error), bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.bound || (p.state != 1 && p.state != 2) {
		return nil, false, vm.ErrType
	}
	fn, cancelled := p.resolve, p.state == 2
	p.resolve, p.state, p.wake = nil, 3, nil
	return fn, cancelled, nil
}
func (p *HostCompletion) close() {
	p.mu.Lock()
	cleanup := p.cleanup
	p.cleanup = nil
	p.resolve, p.state, p.wake = nil, 3, nil
	p.mu.Unlock()
	if cleanup != nil {
		cleanup()
	}
}
