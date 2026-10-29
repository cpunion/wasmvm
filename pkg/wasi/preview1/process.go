//go:build linux || darwin

package preview1

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Stream is an injected, process-owned stdio capability. Methods may run
// concurrently and must honor context cancellation. Poll waits for readiness;
// no guest memory or Caller is passed to asynchronous host work. Close must
// unblock pending operations and release the capability's own resources.
type Stream interface {
	Read(context.Context, []byte) (int, error)
	Write(context.Context, []byte) (int, error)
	Ready(context.Context, bool) (bytes uint64, hangup, ready bool, err error)
	Poll(context.Context, bool) (bytes uint64, hangup bool, err error)
	Close() error
}

// Clock supports realtime (0), monotonic (1), and explicitly selected optional
// CPU clocks (2/3). Wait must observe cancellation. A zero Clock uses real time.
type Clock struct {
	Now        func(uint32) (uint64, error)
	Resolution func(uint32) (uint64, error)
	Wait       func(context.Context, uint32, uint64) error
}
type Limits struct{ MaxDescriptors, MaxIOVectors, MaxIOBytes, MaxSubscriptions, MaxDirectoryEntries, MaxPathBytes, MaxConfigBytes uint32 }

func DefaultLimits() Limits { return Limits{1024, 1024, 8 << 20, 1024, 65536, 4096, 1 << 20} }

type Preopen struct {
	GuestPath, HostPath string
	ReadOnly            bool
}
type Config struct {
	Args, Env             []string
	Stdin, Stdout, Stderr Stream
	Preopens              []Preopen
	Clock                 Clock
	Random                func(context.Context, []byte) error
	Limits                Limits
}

func DefaultConfig() Config { return Config{Limits: DefaultLimits()} }

// Process owns shared P1 resources independently of any guest Store. Install
// can bind it into multiple isolated Stores for legacy wasi-threads. Close is
// a cancellation request; WaitClosed waits for callbacks and resource leases.
type Process struct {
	mu                                      sync.Mutex
	ctx                                     context.Context
	cancel                                  context.CancelFunc
	done                                    chan struct{}
	closing, exited                         bool
	exitCode                                uint32
	active, cleaning, reserved, pollWorkers uint64
	finished                                bool
	config                                  Config
	clock                                   Clock
	fds                                     map[uint32]*descriptor
}
type descriptor struct {
	resource      *resource
	base, inherit Rights
	preopen       string
}
type resource struct {
	mu, readMu, writeMu        sync.Mutex
	file                       *os.File
	dir                        *host.Directory
	stream                     Stream
	socket, listener, datagram bool
	flags                      uint16
	refs                       uint64
	owners                     uint32
	ctx                        context.Context
	cancel                     context.CancelFunc
	entries                    []host.Dirent
}
type ExitError struct{ Code uint32 }

func (e *ExitError) Error() string { return fmt.Sprintf("WASI process exited with code %d", e.Code) }

func NewProcess(config Config) (*Process, error) {
	if config.Limits == (Limits{}) {
		config.Limits = DefaultLimits()
	}
	if config.Limits.MaxDescriptors < 3 {
		return nil, ErrMfile
	}
	config.Args = append([]string(nil), config.Args...)
	config.Env = append([]string(nil), config.Env...)
	config.Preopens = append([]Preopen(nil), config.Preopens...)
	n := uint64(0)
	for _, list := range [][]string{config.Args, config.Env} {
		for _, s := range list {
			if strings.IndexByte(s, 0) >= 0 {
				return nil, ErrInval
			}
			n += uint64(len(s)) + 5
		}
	}
	if n > uint64(config.Limits.MaxConfigBytes) {
		return nil, Err2Big
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &Process{ctx: ctx, cancel: cancel, done: make(chan struct{}), config: config, fds: make(map[uint32]*descriptor)}
	p.clock = config.Clock
	if p.clock.Now == nil && p.clock.Resolution == nil && p.clock.Wait == nil {
		p.clock = systemClock()
	}
	if p.clock.Now == nil || p.clock.Resolution == nil || p.clock.Wait == nil {
		cancel()
		return nil, ErrInval
	}
	if p.config.Random == nil {
		p.config.Random = func(ctx context.Context, b []byte) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			_, err := rand.Read(b)
			return err
		}
	}
	for i, s := range []Stream{config.Stdin, config.Stdout, config.Stderr} {
		if s == nil {
			if i == 0 {
				s = BytesInput(nil)
			} else {
				s = OutputFunc(func(_ context.Context, b []byte) (int, error) { return len(b), nil })
			}
		}
		var r *resource
		for _, existing := range p.fds {
			if reflect.TypeOf(s) == reflect.TypeOf(existing.resource.stream) && reflect.TypeOf(s).Comparable() && s == existing.resource.stream {
				r = existing.resource
				r.refs++
				r.owners++
				break
			}
		}
		if r == nil {
			r = p.newResource(nil, nil, s)
		}
		rights := RightFdFilestatGet | RightFdFdstatSetFlags | RightPollFdReadwrite
		if i == 0 {
			rights |= RightFdRead
		} else {
			rights |= RightFdWrite
		}
		p.fds[uint32(i)] = &descriptor{resource: r, base: rights}
	}
	for _, preopen := range config.Preopens {
		if preopen.GuestPath == "" || strings.IndexByte(preopen.GuestPath, 0) >= 0 {
			p.Close()
			return nil, ErrInval
		}
		d, err := host.OpenDirectory(preopen.HostPath)
		if err != nil {
			p.Close()
			return nil, err
		}
		r := p.newResource(d.File, d, nil)
		rights := allRights
		if preopen.ReadOnly {
			rights &^= writeRights
		}
		_, err = p.addDescriptor(&descriptor{resource: r, base: rights & directoryRights, inherit: rights, preopen: preopen.GuestPath})
		if err != nil {
			r.cancel()
			d.Close()
			p.Close()
			return nil, err
		}
	}
	return p, nil
}
func systemClock() Clock {
	start := time.Now()
	now := func(id uint32) (uint64, error) {
		switch id {
		case 0:
			return uint64(time.Now().UnixNano()), nil
		case 1:
			return uint64(time.Since(start)), nil
		}
		return 0, ErrInval
	}
	return Clock{Now: now, Resolution: func(id uint32) (uint64, error) {
		if id > 1 {
			return 0, ErrInval
		}
		return 1, nil
	}, Wait: func(ctx context.Context, id uint32, end uint64) error {
		for {
			at, err := now(id)
			if err != nil {
				return err
			}
			if end <= at {
				return ctx.Err()
			}
			delta := end - at
			if delta > uint64(1<<63-1) {
				delta = 1<<63 - 1
			}
			t := time.NewTimer(time.Duration(delta))
			select {
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			case <-t.C:
			}
		}
	}}
}
func (p *Process) newResource(f *os.File, d *host.Directory, s Stream) *resource {
	ctx, cancel := context.WithCancel(p.ctx)
	return &resource{file: f, dir: d, stream: s, refs: 1, owners: 1, ctx: ctx, cancel: cancel}
}
func (r *resource) close() {
	r.cancel()
	if r.dir != nil {
		r.dir.Close()
	} else if r.file != nil {
		r.file.Close()
	}
	if r.stream != nil {
		r.stream.Close()
	}
}
func (p *Process) addDescriptor(d *descriptor) (uint32, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closing {
		return 0, wasm.ErrClosed
	}
	if uint64(len(p.fds))+p.reserved >= uint64(p.config.Limits.MaxDescriptors) {
		return 0, ErrMfile
	}
	for fd := uint32(3); ; fd++ {
		if p.fds[fd] == nil {
			p.fds[fd] = d
			return fd, nil
		}
	}
}
func (p *Process) Close() { p.shutdown(false, 0) }

// Context scopes guest execution to this process. Derive guest-call deadlines
// from it so Close/proc_exit also interrupt pure Wasm code between host calls.
// WaitClosed drains P1 callbacks/resources; Store lifetimes remain caller-owned.
func (p *Process) Context() context.Context { return p.ctx }
func (p *Process) shutdown(exit bool, code uint32) {
	p.mu.Lock()
	if exit && !p.exited && !p.closing {
		p.exited = true
		p.exitCode = code
	}
	if p.closing {
		p.mu.Unlock()
		return
	}
	p.closing = true
	p.cancel()
	var closing []*resource
	for fd, d := range p.fds {
		delete(p.fds, fd)
		if p.dropOwnerLocked(d.resource) {
			closing = append(closing, d.resource)
		}
	}
	p.finishLocked()
	p.mu.Unlock()
	for _, r := range closing {
		p.closeResource(r)
	}
}
func (p *Process) finishLocked() {
	if p.closing && p.active == 0 && p.cleaning == 0 && !p.finished {
		p.finished = true
		close(p.done)
	}
}
func (p *Process) releaseLocked(r *resource) bool {
	r.refs--
	if r.refs == 0 {
		p.cleaning++
		return true
	}
	return false
}
func (p *Process) dropOwnerLocked(r *resource) bool {
	r.owners--
	if r.owners == 0 {
		r.cancel()
	}
	return p.releaseLocked(r)
}
func (p *Process) closeResource(r *resource) {
	r.close()
	p.mu.Lock()
	p.cleaning--
	p.finishLocked()
	p.mu.Unlock()
}
func (p *Process) WaitClosed(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-p.done:
		return nil
	default:
	}
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (p *Process) Exited() (uint32, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.exitCode, p.exited
}
func (p *Process) begin(parent context.Context) (context.Context, func(), error) {
	p.mu.Lock()
	if p.closing {
		var err error = wasm.ErrClosed
		if p.exited {
			err = &ExitError{p.exitCode}
		}
		p.mu.Unlock()
		return nil, nil, err
	}
	p.active++
	p.mu.Unlock()
	ctx, cancel := context.WithCancel(parent)
	stop := context.AfterFunc(p.ctx, cancel)
	return ctx, func() {
		stop()
		cancel()
		p.mu.Lock()
		p.active--
		p.finishLocked()
		p.mu.Unlock()
	}, nil
}
func (p *Process) hold(fd uint32, rights Rights) (descriptor, func(), error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	d := p.fds[fd]
	if d == nil {
		return descriptor{}, nil, ErrBadf
	}
	if rights&^d.base != 0 {
		return descriptor{}, nil, ErrNotcapable
	}
	d.resource.refs++
	copy := *d
	return copy, func() {
		p.mu.Lock()
		closeNow := p.releaseLocked(copy.resource)
		p.mu.Unlock()
		if closeNow {
			p.closeResource(copy.resource)
		}
	}, nil
}
func (p *Process) closeFD(fd uint32) error {
	p.mu.Lock()
	d := p.fds[fd]
	if d == nil {
		p.mu.Unlock()
		return ErrBadf
	}
	delete(p.fds, fd)
	closeNow := p.dropOwnerLocked(d.resource)
	p.mu.Unlock()
	if closeNow {
		p.closeResource(d.resource)
	}
	return nil
}
func (p *Process) reserveDescriptor() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closing {
		return wasm.ErrClosed
	}
	if uint64(len(p.fds))+p.reserved >= uint64(p.config.Limits.MaxDescriptors) {
		return ErrMfile
	}
	p.reserved++
	return nil
}
func (p *Process) unreserveDescriptor() { p.mu.Lock(); p.reserved--; p.mu.Unlock() }
func (p *Process) publishDescriptor(d *descriptor) (uint32, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reserved--
	if p.closing {
		return 0, wasm.ErrClosed
	}
	for fd := uint32(3); ; fd++ {
		if p.fds[fd] == nil {
			p.fds[fd] = d
			return fd, nil
		}
	}
}

const allRights Rights = (1 << 30) - 1
const writeRights = RightFdWrite | RightFdAllocate | RightFdDatasync | RightFdSync | RightFdFilestatSetSize | RightFdFilestatSetTimes | RightPathCreateFile | RightPathCreateDirectory | RightPathLinkSource | RightPathLinkTarget | RightPathRenameSource | RightPathRenameTarget | RightPathFilestatSetSize | RightPathFilestatSetTimes | RightPathSymlink | RightPathRemoveDirectory | RightPathUnlinkFile

// Install publishes all 46 typed imports as one module binding. The Process
// retains no Store or guest reference; closing one Store leaves siblings alive.
func (p *Process) Install(store *vm.Store, linker *vm.Linker) error {
	_, finish, err := p.begin(context.Background())
	if err != nil {
		return err
	}
	defer finish()
	exports := make(map[string]vm.Extern, len(functions))
	for _, def := range functions {
		params := make([]wasm.ValueType, len(def.signature))
		for i, c := range def.signature {
			params[i] = wasm.I32
			if c == 'I' {
				params[i] = wasm.I64
			}
		}
		results := []wasm.ValueType{wasm.I32}
		if def.name == "proc_exit" {
			results = nil
		}
		ex, err := store.NewFunction(vm.FunctionType{Params: params, Results: results}, func(c *vm.Caller, args []vm.Value) ([]vm.Value, error) {
			ctx, finish, err := p.begin(c.Context())
			if err != nil {
				return nil, err
			}
			defer finish()
			v := make([]uint64, len(args))
			for i, arg := range args {
				v[i], _, _ = arg.Bits()
			}
			call := abiCall{caller: c, ctx: ctx, p: p}
			err = p.invoke(&call, def.name, v)
			var exit *ExitError
			if errors.As(err, &exit) {
				return nil, exit
			}
			if processErr := p.ctx.Err(); processErr != nil {
				return nil, processErr
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return []vm.Value{vm.I32(uint32(errno(err)))}, nil
		})
		if err != nil {
			return err
		}
		exports[def.name] = ex
	}
	return linker.DefineModule("wasi_snapshot_preview1", exports)
}
