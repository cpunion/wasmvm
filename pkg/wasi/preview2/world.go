//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"sync"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
)

const CLICommandInterface = "wasi:cli/run@0.2.12"

// WorldConfig supplies explicit capabilities for stable CLI and HTTP worlds.
// It never inherits process stdio, environment, filesystem or network grants.
type WorldConfig struct {
	Environment EnvironmentConfig
	IO          IOConfig
	FileSystem  FileSystemConfig
	Network     NetworkConfig
	TCP         TCPConfig
	UDP         UDPConfig
	HTTP        HTTPConfig
	Random      RandomConfig
	WallClock   WallClockConfig
}

type worldOwner interface {
	Close()
	WaitClosed(context.Context) error
}

// World composes existing providers with one shared IO/resource domain. Close
// and WaitClosed reclaim its resource owners; close bound instances separately
// to revoke their immutable configuration imports and guest execution as well.
type World struct {
	io        *IO
	http      *HTTP
	owners    []worldOwner
	providers []interface{ Imports() component.Imports }
	once      sync.Once
}

// NewWorld transfers IO sources once the inner NewIO succeeds. If a later
// constructor fails, rollback closes/joins all accepted owners and sources.
func NewWorld(config WorldConfig) (world *World, err error) {
	environment, err := NewEnvironment(config.Environment)
	if err != nil {
		return nil, err
	}
	random, err := NewRandom(config.Random)
	if err != nil {
		return nil, err
	}
	wall, err := NewWallClock(config.WallClock)
	if err != nil {
		return nil, err
	}
	w := &World{}
	defer func() {
		if err != nil {
			w.Close()
			if len(w.owners) != 0 {
				err = errors.Join(err, w.WaitClosed(context.Background()))
			}
		}
	}()
	w.io, err = NewIO(config.IO)
	if err != nil {
		return nil, err
	}
	w.owners = append(w.owners, w.io)
	filesystem, err := NewFileSystem(w.io, config.FileSystem)
	if err != nil {
		return nil, err
	}
	w.owners = append(w.owners, filesystem)
	network, err := NewNetwork(w.io, config.Network)
	if err != nil {
		return nil, err
	}
	w.owners = append(w.owners, network)
	tcp, err := NewTCP(network, config.TCP)
	if err != nil {
		return nil, err
	}
	w.owners = append(w.owners, tcp)
	udp, err := NewUDP(network, config.UDP)
	if err != nil {
		return nil, err
	}
	w.owners = append(w.owners, udp)
	w.http, err = NewHTTP(network, config.HTTP)
	if err != nil {
		return nil, err
	}
	w.owners = append(w.owners, w.http)
	// Later providers use the same nominal IO/network resources. The HTTP
	// provider supplies the single exported monotonic clock for this world.
	w.providers = []interface{ Imports() component.Imports }{environment, random, wall, filesystem, network, tcp, udp, w.http}
	return w, nil
}

func (w *World) Imports() component.Imports {
	if w == nil || w.io == nil {
		return component.Imports{}
	}
	out := ExitImports()
	for _, provider := range w.providers {
		for name, instance := range provider.Imports().Instances {
			out.Instances[name] = instance
		}
	}
	return out
}

func (w *World) IO() *IO {
	if w == nil {
		return nil
	}
	return w.io
}
func (w *World) HTTP() *HTTP {
	if w == nil {
		return nil
	}
	return w.http
}

func (w *World) Close() {
	if w != nil {
		w.once.Do(func() {
			for i := len(w.owners) - 1; i >= 0; i-- {
				w.owners[i].Close()
			}
		})
	}
}
func (w *World) WaitClosed(ctx context.Context) error {
	if w == nil || w.io == nil {
		return vm.ErrReference
	}
	var failure error
	for i := len(w.owners) - 1; i >= 0; i-- {
		if err := w.owners[i].WaitClosed(ctx); failure == nil {
			failure = err
		}
	}
	return failure
}

// RunCommand checks the standard export before invoking it. Ordinary
// result failure maps to code 1; nonreturning WASI exit preserves its code.
// Traps and cancellation remain errors. Instance/provider ownership is explicit.
func RunCommand(ctx context.Context, in *component.Instance, options component.CallOptions) (uint32, error) {
	if in == nil {
		return 0, vm.ErrReference
	}
	function, err := in.Function(CLICommandInterface, "run")
	if err != nil {
		return 0, err
	}
	info := function.Info()
	result := info.Result()
	if info.Async() || info.ParameterCount() != 0 || result.Kind() != component.KindResult || result.CaseCount() != 2 {
		return 0, vm.ErrType
	}
	for i := 0; i < 2; i++ {
		if _, payload := result.Case(i); payload.Valid() {
			return 0, vm.ErrType
		}
	}
	value, err := function.Call(ctx, nil, options)
	var exit *ExitError
	if errors.As(err, &exit) {
		return exit.Code, nil
	}
	if err != nil {
		return 0, err
	}
	return value.Value.Case, nil
}
