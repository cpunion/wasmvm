//go:build linux || darwin

package threads

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasi/preview1"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const maxTID = 1 << 29

// HostPanicError reports a panic in a synchronous host callback. The group
// cancels its other threads and completes reclamation before Run returns it.
type HostPanicError struct{ Value any }

func (e *HostPanicError) Error() string {
	return fmt.Sprintf("WASI host callback panicked: %v", e.Value)
}

type Config struct {
	// MaxThreads includes the main thread and reservations under construction.
	MaxThreads uint32
	Store      vm.StoreConfig
	// Execution is a separate instruction budget for each thread, including
	// that instance's start function. A process deadline bounds the whole group.
	Execution vm.CallOptions
}

func DefaultConfig() Config {
	return Config{MaxThreads: DefaultMaxThreads, Store: vm.DefaultStoreConfig()}
}

// Group owns process after New succeeds. Engine and module remain borrowed
// and must stay open while the group can spawn. Close requests cancellation;
// WaitClosed joins all Stores, host callbacks and final resource reclamation.
type Group struct {
	mu                         sync.Mutex
	engine                     *vm.Engine
	module                     *vm.CompiledModule
	process                    *preview1.Process
	config                     Config
	ctx                        context.Context
	cancel                     context.CancelFunc
	done                       chan struct{}
	memories                   map[[2]string]*vm.SharedMemory
	stores                     map[uint32]*vm.Store
	active                     uint32
	nextTID                    uint32
	running, closing, cleaning bool
	ended                      bool
	cause                      error
}

func New(ctx context.Context, engine *vm.Engine, module *vm.CompiledModule, process *preview1.Process, config Config) (*Group, error) {
	if engine == nil || module == nil || process == nil {
		return nil, vm.ErrReference
	}
	if config == (Config{}) {
		config = DefaultConfig()
	}
	if config.MaxThreads == 0 || config.MaxThreads >= maxTID {
		return nil, wasm.ErrLimit
	}
	info, err := module.Info()
	if err != nil {
		return nil, err
	}
	if info.Features&wasm.Threads == 0 {
		return nil, wasm.ErrUnsupported
	}
	imports := map[[2]string]vm.MemoryType{}
	spawns := false
	for _, im := range info.Imports {
		switch {
		case im.Kind == vm.MemoryKind && im.Memory.Shared && !im.Memory.Index64:
			key := [2]string{im.Module, im.Name}
			typ := *im.Memory
			if old, found := imports[key]; found {
				typ.Min, typ.Max = max(old.Min, typ.Min), min(old.Max, typ.Max)
				if typ.Min > typ.Max {
					return nil, fmt.Errorf("%w: inconsistent shared memory %s.%s", vm.ErrLink, im.Module, im.Name)
				}
			}
			imports[key] = typ
		case im.Kind == vm.FunctionKind && im.Module == "wasi_snapshot_preview1":
		case im.Kind == vm.FunctionKind && im.Module == "wasi" && im.Name == "thread-spawn" && signature(im.Function, []wasm.ValueType{wasm.I32}, []wasm.ValueType{wasm.I32}):
			spawns = true
		default:
			return nil, fmt.Errorf("%w: unsupported thread import %s.%s", vm.ErrLink, im.Module, im.Name)
		}
	}
	if len(imports) == 0 {
		return nil, fmt.Errorf("%w: thread module must import shared memory", vm.ErrLink)
	}
	want := map[string][]wasm.ValueType{"_start": nil}
	if spawns {
		want["wasi_thread_start"] = []wasm.ValueType{wasm.I32, wasm.I32}
	}
	for _, ex := range info.Exports {
		if params, found := want[ex.Name]; found && ex.Kind == vm.FunctionKind && signature(ex.Function, params, nil) {
			delete(want, ex.Name)
		}
	}
	if len(want) != 0 {
		return nil, fmt.Errorf("%w: require _start() and wasi_thread_start(i32, i32)", vm.ErrType)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := process.Context().Err(); err != nil {
		return nil, err
	}
	creator, err := engine.NewStore(config.Store)
	if err != nil {
		return nil, err
	}
	defer creator.Close()
	g := &Group{engine: engine, module: module, process: process, config: config, done: make(chan struct{}), memories: map[[2]string]*vm.SharedMemory{}, stores: map[uint32]*vm.Store{}, nextTID: 1}
	for key, typ := range imports {
		memory, err := creator.NewMemory(typ)
		if err == nil {
			g.memories[key], err = memory.ShareMemory()
		}
		if err != nil {
			for _, memory := range g.memories {
				memory.Close()
			}
			return nil, err
		}
	}
	g.ctx, g.cancel = context.WithCancel(ctx)
	// These registrations are drained by cleanup. No callback retains a Caller.
	stopProcess := context.AfterFunc(process.Context(), g.Close)
	stopContext := context.AfterFunc(g.ctx, g.Close)
	go func() {
		<-g.done
		stopProcess()
		stopContext()
	}()
	return g, nil
}

func signature(t *vm.FunctionType, params, results []wasm.ValueType) bool {
	return t != nil && slices.Equal(t.Params, params) && slices.Equal(t.Results, results)
}

// Run invokes the command entry once and joins the entire group. A main return,
// a trap, proc_exit or cancellation terminates the process. Worker returns only
// release that worker. TLS, stack allocation and pthread cleanup belong to libc.
func (g *Group) Run() (err error) {
	g.mu.Lock()
	if g.closing {
		g.mu.Unlock()
		return wasm.ErrClosed
	}
	if g.running {
		g.mu.Unlock()
		return vm.ErrBusy
	}
	g.running = true
	g.active++
	g.mu.Unlock()
	var store *vm.Store
	completed := false
	defer func() {
		if !completed {
			err = &HostPanicError{Value: recover()}
		}
		g.fail(err)
		g.release(0, store)
		<-g.done
		g.mu.Lock()
		defer g.mu.Unlock()
		if code, exited := g.process.Exited(); exited {
			err = &preview1.ExitError{Code: code}
		} else {
			err = g.cause
		}
	}()
	var instance *vm.Instance
	store, instance, err = g.instantiate(0)
	if err == nil {
		err = g.call(instance, "_start", nil)
	}
	completed = true
	return err
}

func (g *Group) call(instance *vm.Instance, name string, args []vm.Value) error {
	fn, err := instance.Export(name)
	if err != nil {
		return err
	}
	options := g.config.Execution
	if options.Metered {
		options.Fuel = instance.StartFuelRemaining()
	}
	_, err = fn.Call(g.ctx, args, options)
	return err
}

func (g *Group) instantiate(tid uint32) (*vm.Store, *vm.Instance, error) {
	store, err := g.engine.NewStore(g.config.Store)
	if err != nil {
		return nil, nil, err
	}
	g.mu.Lock()
	if g.closing {
		g.mu.Unlock()
		store.Close()
		return nil, nil, context.Canceled
	}
	g.stores[tid] = store
	g.mu.Unlock()
	linker := vm.NewLinker(store)
	for key, shared := range g.memories {
		memory, err := store.ImportMemory(shared)
		if err == nil {
			err = linker.Define(key[0], key[1], memory)
		}
		if err != nil {
			return store, nil, err
		}
	}
	if err := g.process.Install(store, linker); err != nil {
		return store, nil, err
	}
	spawn, err := store.NewFunction(vm.FunctionType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, func(c *vm.Caller, args []vm.Value) ([]vm.Value, error) {
		arg, _, _ := args[0].Bits()
		tid, err := g.spawn(c.Context(), uint32(arg))
		return []vm.Value{vm.I32(tid)}, err
	})
	if err == nil {
		err = linker.Define("wasi", "thread-spawn", spawn)
	}
	if err != nil {
		return store, nil, err
	}
	instance, err := linker.Instantiate(g.ctx, g.module, g.config.Execution)
	return store, instance, err
}

func (g *Group) spawn(ctx context.Context, arg uint32) (id uint32, err error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	g.mu.Lock()
	if g.closing {
		g.mu.Unlock()
		return 0, context.Canceled
	}
	if g.active >= g.config.MaxThreads || g.nextTID >= maxTID {
		g.mu.Unlock()
		return ^uint32(preview1.ErrAgain - 1), nil
	}
	tid := g.nextTID
	g.nextTID++
	g.active++
	g.mu.Unlock()
	var store *vm.Store
	owned, completed := true, false
	defer func() {
		if !completed {
			err = &HostPanicError{Value: recover()}
			g.fail(err)
		}
		if owned {
			g.release(tid, store)
		}
	}()
	var instance *vm.Instance
	store, instance, err = g.instantiate(tid)
	if err != nil {
		completed = true
		if errors.Is(err, wasm.ErrTrap) {
			g.fail(err)
			return 0, err
		}
		if g.ctx.Err() != nil {
			return 0, g.ctx.Err()
		}
		return ^uint32(preview1.ErrAgain - 1), nil
	}
	owned, completed = false, true
	go g.worker(tid, arg, store, instance)
	return tid, nil
}

func (g *Group) worker(tid, arg uint32, store *vm.Store, instance *vm.Instance) {
	defer g.release(tid, store)
	completed := false
	defer func() {
		if !completed {
			g.fail(&HostPanicError{Value: recover()})
		}
	}()
	if err := g.call(instance, "wasi_thread_start", []vm.Value{vm.I32(tid), vm.I32(arg)}); err != nil {
		g.fail(err)
	}
	completed = true
}

func (g *Group) fail(err error) {
	g.mu.Lock()
	if !g.closing {
		g.ended = true
	}
	if err != nil && g.cause == nil && !g.closing {
		g.cause = err
	}
	g.mu.Unlock()
	g.Close()
}

func (g *Group) release(tid uint32, store *vm.Store) {
	if store == nil {
		// Initialization can panic before returning its already registered Store.
		g.mu.Lock()
		store = g.stores[tid]
		g.mu.Unlock()
	}
	if store != nil {
		store.Close()
		store.WaitClosed(nil)
	}
	g.mu.Lock()
	delete(g.stores, tid)
	g.active--
	g.finishLocked()
	g.mu.Unlock()
}

func (g *Group) Close() {
	g.mu.Lock()
	if g.closing {
		g.mu.Unlock()
		return
	}
	g.closing = true
	if g.cause == nil && !g.ended {
		g.cause = g.ctx.Err()
		if g.cause == nil {
			g.cause = context.Canceled
		}
	}
	stores := make([]*vm.Store, 0, len(g.stores))
	for _, store := range g.stores {
		stores = append(stores, store)
	}
	g.mu.Unlock()
	g.cancel()
	g.process.Close()
	for _, store := range stores {
		store.Close()
	}
	g.mu.Lock()
	g.finishLocked()
	g.mu.Unlock()
}

func (g *Group) finishLocked() {
	if !g.closing || g.active != 0 || g.cleaning {
		return
	}
	g.cleaning = true
	go func() {
		g.process.WaitClosed(nil)
		for _, memory := range g.memories {
			memory.Close()
		}
		close(g.done)
	}()
}

func (g *Group) WaitClosed(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-g.done:
		return nil
	default:
	}
	select {
	case <-g.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
