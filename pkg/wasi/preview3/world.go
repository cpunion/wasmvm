//go:build linux || darwin

package preview3

import (
	"context"
	"errors"
	"sync"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
)

// WorldConfig supplies explicit capabilities for the pinned CLI/service/middleware worlds.
// It inherits no process configuration, stdio, filesystem or network authority.
type WorldConfig struct {
	Environment    EnvironmentConfig
	IO             IOConfig
	FileSystem     FileSystemConfig
	Sockets        SocketsConfig
	HTTP           HTTPConfig
	Random         RandomConfig
	SystemClock    SystemClockConfig
	MonotonicClock MonotonicClockConfig
}
type worldOwner interface {
	Close()
	WaitClosed(context.Context) error
}

// World composes the existing versioned providers in one IO ownership domain.
// Bound instances have independent lifetimes and must also be closed/joined.
type World struct {
	io        *IO
	http      *HTTP
	owners    []worldOwner
	providers []interface{ Imports() component.Imports }
	once      sync.Once
}

// NewWorld accepts IO sources when NewIO succeeds. Later failures close and
// join all accepted owners; earlier validation failures leave sources untouched.
func NewWorld(config WorldConfig) (world *World, err error) {
	environment, err := NewEnvironment(config.Environment)
	if err != nil {
		return nil, err
	}
	random, err := NewRandom(config.Random)
	if err != nil {
		return nil, err
	}
	system, err := NewSystemClock(config.SystemClock)
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
	monotonic, err := NewMonotonicClock(config.MonotonicClock)
	if err != nil {
		return nil, err
	}
	w.owners = append(w.owners, monotonic)
	filesystem, err := NewFileSystem(w.io, config.FileSystem)
	if err != nil {
		return nil, err
	}
	w.owners = append(w.owners, filesystem)
	sockets, err := NewSockets(w.io, config.Sockets)
	if err != nil {
		return nil, err
	}
	w.owners = append(w.owners, sockets)
	w.http, err = NewHTTP(sockets, config.HTTP)
	if err != nil {
		return nil, err
	}
	w.owners = append(w.owners, w.http)
	w.providers = []interface{ Imports() component.Imports }{environment, random, system, monotonic, w.io, filesystem, sockets, w.http}
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
