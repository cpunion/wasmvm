//go:build linux || darwin

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasi/preview1"
	"github.com/cpunion/wasmvm/pkg/wasi/threads"
	"golang.org/x/sys/unix"
)

type borrowedStream struct {
	preview1.Stream
	fd, flags int
	once      sync.Once
	err       error
}

func (s *borrowedStream) SetFlags(flags uint16) error {
	return s.Stream.(interface{ SetFlags(uint16) error }).SetFlags(flags)
}
func (s *borrowedStream) Close() error {
	s.once.Do(func() {
		_, s.err = unix.FcntlInt(uintptr(s.fd), unix.F_SETFL, s.flags)
		closeErr := s.Stream.Close()
		if s.err == nil {
			s.err = closeErr
		}
	})
	return s.err
}

func borrowFile(file *os.File, flags int) (preview1.Stream, error) {
	fd, err := unix.Dup(host.Descriptor(file))
	if err != nil {
		return nil, err
	}
	unix.CloseOnExec(fd)
	stream, err := preview1.FileStream(os.NewFile(uintptr(fd), "WASI CLI stdio"))
	if err != nil {
		return nil, err
	}
	return &borrowedStream{Stream: stream, fd: fd, flags: flags}, nil
}
func cliStreams(stdinPath string, stdout, stderr io.Writer) ([]preview1.Stream, error) {
	files := []*os.File{os.Stdin, nil, nil}
	if stdinPath != "" {
		files[0] = nil
	}
	files[1], _ = stdout.(*os.File)
	files[2], _ = stderr.(*os.File)
	var flags [3]int
	for i, file := range files {
		if file != nil {
			v, err := unix.FcntlInt(uintptr(host.Descriptor(file)), unix.F_GETFL, 0)
			if err != nil {
				return nil, err
			}
			flags[i] = v
		}
	}
	streams := make([]preview1.Stream, 3)
	cleanup := func() {
		for _, stream := range streams {
			if stream != nil {
				stream.Close()
			}
		}
	}
	for i, file := range files {
		var err error
		if i == 0 && stdinPath != "" {
			var input *os.File
			input, err = os.OpenFile(stdinPath, os.O_RDONLY|unix.O_NONBLOCK, 0)
			if err == nil {
				streams[0], err = preview1.FileStream(input)
			}
		} else if file != nil {
			streams[i], err = borrowFile(file, flags[i])
		} else {
			writer := stdout
			if i == 2 {
				writer = stderr
			}
			streams[i] = preview1.OutputFunc(func(ctx context.Context, b []byte) (int, error) {
				if err := ctx.Err(); err != nil {
					return 0, err
				}
				return writer.Write(b)
			})
		}
		if err != nil {
			cleanup()
			return nil, err
		}
	}
	return streams, nil
}
func runWASI(ctx context.Context, engine *vm.Engine, store *vm.Store, module *vm.CompiledModule, options vm.CallOptions, args wasiOptions, stdout, stderr io.Writer) int {
	config := preview1.DefaultConfig()
	config.Args = append([]string{filepath.Base(args.argv0)}, args.args...)
	for _, env := range args.env {
		if strings.IndexByte(env, '=') < 1 {
			fmt.Fprintln(stderr, "--env requires key=value")
			return 2
		}
		config.Env = append(config.Env, env)
	}
	for kind, dirs := range []stringList{args.dirs, args.readOnlyDirs} {
		for _, dir := range dirs {
			guest, hostPath, ok := strings.Cut(dir, "=")
			if !ok || guest == "" || hostPath == "" {
				fmt.Fprintln(stderr, "directory capability requires guest=host")
				return 2
			}
			config.Preopens = append(config.Preopens, preview1.Preopen{GuestPath: guest, HostPath: hostPath, ReadOnly: kind == 1})
		}
	}
	streams, err := cliStreams(args.stdinFile, stdout, stderr)
	if err != nil {
		return failure(stderr, err)
	}
	config.Stdin, config.Stdout, config.Stderr = streams[0], streams[1], streams[2]
	p, err := preview1.NewProcess(config)
	if err != nil {
		for _, stream := range streams {
			stream.Close()
		}
		return failure(stderr, err)
	}
	defer func() { p.Close(); p.WaitClosed(nil) }()
	guestCtx, cancel := context.WithCancel(p.Context())
	stop := context.AfterFunc(ctx, cancel)
	defer func() { stop(); cancel() }()
	if ctx.Err() != nil {
		cancel()
	}
	if args.threads {
		config := threads.DefaultConfig()
		config.MaxThreads, config.Execution = uint32(args.maxThreads), options
		group, err := threads.New(guestCtx, engine, module, p, config)
		if err != nil {
			return failure(stderr, err)
		}
		defer func() { group.Close(); group.WaitClosed(nil) }()
		err = group.Run()
		var exit *preview1.ExitError
		if errors.As(err, &exit) {
			return int(exit.Code)
		}
		if err != nil {
			if ctx.Err() != nil {
				err = ctx.Err()
			}
			return failure(stderr, err)
		}
		return 0
	}
	l := vm.NewLinker(store)
	if err = p.Install(store, l); err != nil {
		return failure(stderr, err)
	}
	in, err := l.Instantiate(guestCtx, module, options)
	if err == nil {
		if options.Metered {
			options.Fuel = in.StartFuelRemaining()
		}
		var fn vm.Extern
		fn, err = in.Export(args.exported)
		if err == nil {
			_, err = fn.Call(guestCtx, nil, options)
		}
	}
	var exit *preview1.ExitError
	if errors.As(err, &exit) {
		return int(exit.Code)
	}
	if err != nil {
		if ctx.Err() != nil {
			return failure(stderr, ctx.Err())
		}
		return failure(stderr, err)
	}
	return 0
}
