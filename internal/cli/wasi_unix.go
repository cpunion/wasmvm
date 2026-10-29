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

// cliStdio retains restoration descriptors independently of guest bindings.
// Duplicated stdout/stderr can share one open-file description: restoring its
// flags on fd_close would make a still-live sibling stream block in the kernel.
type cliStdio struct {
	streams [3]preview1.Stream
	keepers [3]*os.File
	flags   [3]int
	once    sync.Once
	err     error
}

func (s *cliStdio) Close() error {
	s.once.Do(func() {
		remember := func(err error) {
			if s.err == nil {
				s.err = err
			}
		}
		for _, stream := range s.streams {
			if stream != nil {
				remember(stream.Close())
			}
		}
		for i, file := range s.keepers {
			if file != nil {
				_, err := unix.FcntlInt(uintptr(host.Descriptor(file)), unix.F_SETFL, s.flags[i])
				remember(err)
				remember(file.Close())
			}
		}
	})
	return s.err
}

func cliStreams(stdinPath string, stdout, stderr io.Writer) (*cliStdio, error) {
	files := []*os.File{os.Stdin, nil, nil}
	if stdinPath != "" {
		files[0] = nil
	}
	files[1], _ = stdout.(*os.File)
	files[2], _ = stderr.(*os.File)
	scope := &cliStdio{}
	// Snapshot every original before FileStream changes any shared status flags.
	for i, file := range files {
		if file != nil {
			flags, err := unix.FcntlInt(uintptr(host.Descriptor(file)), unix.F_GETFL, 0)
			if err != nil {
				return nil, err
			}
			scope.flags[i] = flags
		}
	}
	for i, file := range files {
		if file != nil {
			fd, err := unix.Dup(host.Descriptor(file))
			if err != nil {
				scope.Close()
				return nil, err
			}
			unix.CloseOnExec(fd)
			scope.keepers[i] = os.NewFile(uintptr(fd), "WASI stdio restoration")
		}
	}
	for i, file := range files {
		var err error
		if i == 0 && stdinPath != "" {
			var input *os.File
			input, err = os.OpenFile(stdinPath, os.O_RDONLY|unix.O_NONBLOCK, 0)
			if err == nil {
				scope.streams[i], err = preview1.FileStream(input)
			}
		} else if file != nil {
			var fd int
			fd, err = unix.Dup(host.Descriptor(scope.keepers[i]))
			if err == nil {
				unix.CloseOnExec(fd)
				scope.streams[i], err = preview1.FileStream(os.NewFile(uintptr(fd), "WASI CLI stdio"))
			}
		} else {
			writer := stdout
			if i == 2 {
				writer = stderr
			}
			scope.streams[i] = preview1.OutputFunc(func(ctx context.Context, b []byte) (int, error) {
				if err := ctx.Err(); err != nil {
					return 0, err
				}
				return writer.Write(b)
			})
		}
		if err != nil {
			scope.Close()
			return nil, err
		}
	}
	return scope, nil
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
	defer streams.Close()
	config.Stdin, config.Stdout, config.Stderr = streams.streams[0], streams.streams[1], streams.streams[2]
	p, err := preview1.NewProcess(config)
	if err != nil {
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
