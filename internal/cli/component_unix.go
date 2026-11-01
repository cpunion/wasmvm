//go:build linux || darwin

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasi/preview1"
	"github.com/cpunion/wasmvm/pkg/wasi/preview2"
	"github.com/cpunion/wasmvm/pkg/wasi/preview3"
	"golang.org/x/sys/unix"
)

type commandWorld interface {
	Imports() component.Imports
	Close()
	WaitClosed(context.Context) error
}

func runComponentWASI(ctx context.Context, core vm.EngineConfig, options vm.CallOptions, args wasiOptions, stdout, stderr io.Writer) (code int) {
	env, grants, err := cliCapabilities(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	config := component.DefaultConfig()
	config.Core = core
	engine, err := component.NewEngine(config)
	if err != nil {
		return failure(stderr, err)
	}
	defer engine.Close()
	input, err := readModule(args.argv0, config.Decode.MaxBytes)
	if err != nil {
		return failure(stderr, err)
	}
	compiled, err := engine.Compile(ctx, input)
	if err != nil {
		return failure(stderr, err)
	}
	defer compiled.Close()
	stdio, err := cliStreams(args.stdinFile, stdout, stderr)
	if err != nil {
		return failure(stderr, err)
	}
	defer func() {
		if err := stdio.Close(); err != nil && code == 0 {
			code = failure(stderr, err)
		}
	}()
	var sources [3]*commandSource
	defer func() {
		for _, s := range sources {
			if s != nil {
				s.Close()
				s.waiters.Wait()
			}
		}
	}()
	for i, stream := range stdio.streams {
		if flags, ok := stream.(interface{ SetFlags(uint16) error }); ok {
			if err := flags.SetFlags(4); err != nil {
				return failure(stderr, err)
			}
		}
		sources[i] = newCommandSource(stream, i == 0)
	}
	environment := preview2.EnvironmentConfig{Args: append([]string{filepath.Base(args.argv0)}, args.args...), Env: env}
	var world commandWorld
	var run func(context.Context, *component.Instance, component.CallOptions) (uint32, error)
	if args.preview == 2 {
		cfg := preview2.WorldConfig{Environment: environment, IO: preview2.IOConfig{Stdin: sources[0], Stdout: sources[1], Stderr: sources[2]}}
		for _, grant := range grants {
			cfg.FileSystem.Preopens = append(cfg.FileSystem.Preopens, preview2.FileSystemPreopen{GuestPath: grant.GuestPath, HostPath: grant.HostPath, ReadOnly: grant.ReadOnly})
		}
		world, err = preview2.NewWorld(cfg)
		run = preview2.RunCommand
	} else {
		cfg := preview3.WorldConfig{Environment: environment, IO: preview3.IOConfig{Stdin: sources[0], Stdout: sources[1], Stderr: sources[2]}}
		for _, grant := range grants {
			cfg.FileSystem.Preopens = append(cfg.FileSystem.Preopens, preview3.FileSystemPreopen{GuestPath: grant.GuestPath, HostPath: grant.HostPath, ReadOnly: grant.ReadOnly})
		}
		world, err = preview3.NewWorld(cfg)
		run = preview3.RunCommand
	}
	if err != nil {
		return failure(stderr, err)
	}
	defer func() {
		world.Close()
		if err := world.WaitClosed(nil); err != nil && code == 0 {
			code = failure(stderr, err)
		}
	}()
	in, err := compiled.Instantiate(ctx, world.Imports(), options)
	if err != nil {
		return failure(stderr, err)
	}
	defer func() { in.Close(); in.WaitClosed(nil) }()
	if options.Metered {
		options.Fuel = in.StartFuelRemaining()
	}
	result, err := run(ctx, in, options)
	if err != nil {
		return failure(stderr, err)
	}
	return int(result)
}

// commandSource adapts existing nonblocking CLI descriptors to the shared P2/P3
// prompt-source contract. At most one external readiness waiter runs per stream.
type commandSource struct {
	mu      sync.Mutex
	stream  preview1.Stream
	input   bool
	ctx     context.Context
	cancel  context.CancelFunc
	changed chan struct{}
	waiters sync.WaitGroup
	waiting bool
	closed  bool
	failure error
}

func newCommandSource(stream preview1.Stream, input bool) *commandSource {
	ctx, cancel := context.WithCancel(context.Background())
	return &commandSource{stream: stream, input: input, ctx: ctx, cancel: cancel, changed: make(chan struct{})}
}

func (s *commandSource) state(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return io.ErrClosedPipe
	}
	return s.failure
}

func (s *commandSource) Read(ctx context.Context, data []byte) (int, error) {
	if err := s.state(ctx); err != nil {
		return 0, err
	}
	n, err := s.stream.Read(ctx, data)
	if n == 0 && err == nil && len(data) != 0 {
		err = io.EOF
	}
	return n, commandIOError(err)
}

func (s *commandSource) Write(ctx context.Context, data []byte) (int, error) {
	if err := s.state(ctx); err != nil {
		return 0, err
	}
	n, err := s.stream.Write(ctx, data)
	return n, commandIOError(err)
}

func commandIOError(err error) error {
	if errors.Is(err, unix.EAGAIN) {
		return preview2.ErrWouldBlock
	}
	return err
}

func (s *commandSource) Flush(ctx context.Context) error { return s.state(ctx) }

func (s *commandSource) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.failure != nil {
		return true, s.changed
	}
	_, _, ready, err := s.stream.Ready(s.ctx, s.input)
	if err != nil && !errors.Is(err, unix.EINTR) {
		s.failure = err
		return true, s.changed
	}
	if ready {
		return true, s.changed
	}
	if !s.waiting {
		s.waiting = true
		s.waiters.Add(1)
		go func() {
			defer s.waiters.Done()
			_, _, err := s.stream.Poll(s.ctx, s.input)
			s.mu.Lock()
			defer s.mu.Unlock()
			s.waiting = false
			if !s.closed {
				s.failure = err
				close(s.changed)
				s.changed = make(chan struct{})
			}
		}()
	}
	return false, s.changed
}

func (s *commandSource) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.cancel()
	close(s.changed)
	s.mu.Unlock()
	return s.stream.Close()
}
