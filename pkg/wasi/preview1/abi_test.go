//go:build linux || darwin

package preview1

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

type fixture struct {
	t      *testing.T
	p      *Process
	s      *vm.Store
	in     *vm.Instance
	memory vm.Extern
	ctx    context.Context
}

func newFixture(t *testing.T, p *Process, backend string) *fixture {
	t.Helper()
	cfg := vm.DefaultEngineConfig()
	cfg.Backend = backend
	cfg.Profile = "core3"
	e, err := vm.NewEngineWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	s, err := e.NewStore(vm.DefaultStoreConfig())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close(); s.WaitClosed(nil) })
	l := vm.NewLinker(s)
	if err = p.Install(s, l); err != nil {
		t.Fatal(err)
	}
	bytes, err := os.ReadFile("testdata/abi.wasm")
	if err != nil {
		t.Fatal(err)
	}
	m, err := e.Compile(nil, bytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Close)
	in, err := l.Instantiate(nil, m, vm.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	memory, err := in.Export("memory")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return &fixture{t, p, s, in, memory, ctx}
}
func (f *fixture) raw(ctx context.Context, name string, a ...uint64) (uint32, error) {
	f.t.Helper()
	fn, err := f.in.Export(name)
	if err != nil {
		return 0, err
	}
	typ, err := fn.FunctionType()
	if err != nil {
		return 0, err
	}
	values := make([]vm.Value, len(a))
	for i, v := range a {
		if typ.Params[i] == wasm.I64 {
			values[i] = vm.I64(v)
		} else {
			values[i] = vm.I32(uint32(v))
		}
	}
	r, err := fn.Call(ctx, values, vm.CallOptions{})
	if err != nil {
		return 0, err
	}
	if len(r.Values) == 0 {
		return 0, nil
	}
	n, _, err := r.Values[0].Bits()
	return uint32(n), err
}
func (f *fixture) call(name string, a ...uint64) Errno {
	f.t.Helper()
	n, err := f.raw(f.ctx, name, a...)
	if err != nil {
		f.t.Fatal(name, err)
	}
	return Errno(n)
}
func (f *fixture) want(want Errno, name string, a ...uint64) {
	f.t.Helper()
	if got := f.call(name, a...); got != want {
		f.t.Fatalf("%s = %v want %v", name, got, want)
	}
}
func (f *fixture) write(at uint64, b []byte) {
	f.t.Helper()
	if err := f.memory.WriteMemory(at, b); err != nil {
		f.t.Fatal(err)
	}
}
func (f *fixture) read(at, n uint64) []byte {
	f.t.Helper()
	b := make([]byte, n)
	if err := f.memory.ReadMemory(at, b); err != nil {
		f.t.Fatal(err)
	}
	return b
}
func (f *fixture) u32(at uint64) uint32 { return binary.LittleEndian.Uint32(f.read(at, 4)) }
func (f *fixture) u64(at uint64) uint64 { return binary.LittleEndian.Uint64(f.read(at, 8)) }
func (f *fixture) iovec(at, ptr, n uint32) {
	var b [8]byte
	binary.LittleEndian.PutUint32(b[:], ptr)
	binary.LittleEndian.PutUint32(b[4:], n)
	f.write(uint64(at), b[:])
}
func process(t *testing.T, cfg Config) *Process {
	t.Helper()
	p, err := NewProcess(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		p.Close()
		if err := p.WaitClosed(nil); err != nil {
			t.Error(err)
		}
	})
	return p
}

func TestABIAndResources(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		t.Run(backend, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Args = []string{"program", "x"}
			cfg.Env = []string{"K=V"}
			cfg.Limits.MaxDescriptors = 4
			cfg.Preopens = []Preopen{{"/", t.TempDir(), false}}
			output := NewBuffer(3)
			cfg.Stdout = output
			p := process(t, cfg)
			f := newFixture(t, p, backend)
			f.want(ErrSuccess, "args_sizes_get", 0, 4)
			if f.u32(0) != 2 || f.u32(4) != 10 {
				t.Fatal("args sizes")
			}
			f.want(ErrSuccess, "args_get", 16, 32)
			if string(f.read(32, 10)) != "program\x00x\x00" {
				t.Fatal("args bytes")
			}
			f.want(ErrSuccess, "environ_sizes_get", 0, 4)
			f.want(ErrSuccess, "environ_get", 16, 32)
			f.want(ErrFault, "args_sizes_get", 0, 131071)
			f.want(ErrFault, "args_get", 131069, 32)
			f.want(ErrFault, "environ_get", 16, 131071)
			f.want(ErrFault, "random_get", 0xffffffff, 2)
			f.want(ErrSuccess, "random_get", 131072, 0)
			f.want(ErrInval, "clock_res_get", 99, 0)
			f.want(ErrFault, "clock_time_get", 0, 0, 131071)
			f.want(ErrSuccess, "clock_res_get", 1, 0)
			if f.u64(0) == 0 {
				t.Fatal("clock resolution")
			}
			f.iovec(100, 200, 5)
			f.write(200, []byte("hello"))
			f.want(ErrSuccess, "fd_write", 1, 100, 1, 300)
			if f.u32(300) != 3 || string(output.Bytes()) != "hel" {
				t.Fatal("partial write")
			}
			f.want(ErrNospc, "fd_write", 1, 100, 1, 300)
			f.want(ErrNotcapable, "fd_write", 0, 100, 1, 300)
			f.want(ErrBadf, "fd_read", 999, 100, 1, 300)
			f.iovec(100, 131071, 2)
			f.want(ErrFault, "fd_write", 1, 100, 1, 300)
			f.want(ErrInval, "fd_write", 1, 100, 1025, 300)
			f.want(ErrFault, "fd_fdstat_get", 1, 131071)
			f.want(ErrFault, "fd_filestat_get", 1, 131071)
			f.want(ErrSuccess, "fd_filestat_get", 1, 0)
			f.want(ErrNotcapable, "fd_seek", 1, 0, 0, 0)
			f.want(ErrNotcapable, "fd_filestat_set_size", 1, 0)
			f.want(ErrInval, "fd_fdstat_set_flags", 1, 32)
			f.want(ErrSuccess, "fd_fdstat_set_flags", 1, 4)
			f.want(ErrNotcapable, "fd_fdstat_set_rights", 1, uint64(allRights), 0)
			f.want(ErrBadf, "fd_prestat_get", 1, 0)
			f.want(ErrNametoolong, "fd_prestat_dir_name", 3, 0, 0)
			f.want(ErrFault, "fd_prestat_get", 3, 131071)
			f.want(ErrFault, "fd_prestat_dir_name", 3, 131071, 2)
			f.write(400, []byte("created"))
			f.want(ErrMfile, "path_open", 3, 0, 400, 7, 1, uint64(regularRights), 0, 0, 500)
			if _, err := os.Stat(filepath.Join(cfg.Preopens[0].HostPath, "created")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("quota failure created a file", err)
			}
			f.want(ErrSuccess, "fd_renumber", 1, 2)
			f.want(ErrBadf, "fd_renumber", 2, 1)
			f.want(ErrSuccess, "fd_renumber", 2, 2)
			f.want(ErrBadf, "fd_close", 1)
			f.want(ErrSuccess, "fd_close", 2)
			f.want(ErrBadf, "fd_close", 2)
			f.want(ErrNotsup, "proc_raise", 1)
			f.want(ErrInval, "proc_raise", 31)
			f.want(ErrSuccess, "proc_raise", 0)
			_, err := f.raw(f.ctx, "proc_exit", 17)
			var exit *ExitError
			if !errors.As(err, &exit) || exit.Code != 17 {
				t.Fatal("exit", err)
			}
			if code, ok := p.Exited(); !ok || code != 17 {
				t.Fatal("exit state")
			}
			p.Close()
			if err := p.WaitClosed(f.ctx); err != nil {
				t.Fatal(err)
			}
			_, err = f.raw(f.ctx, "sched_yield")
			if !errors.As(err, &exit) {
				t.Fatal("process entry after exit", err)
			}
		})
	}
}

func TestSocketsAndPoll(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		t.Run(backend, func(t *testing.T) {
			p := process(t, DefaultConfig())
			f := newFixture(t, p, backend)
			pair, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer unix.Close(pair[1])
			fd, err := p.ImportSocket(os.NewFile(uintptr(pair[0]), "test socket"), false)
			if err != nil {
				t.Fatal(err)
			}
			f.iovec(100, 200, 3)
			f.write(200, []byte("one"))
			f.want(ErrSuccess, "sock_send", uint64(fd), 100, 1, 0, 300)
			var bytes [3]byte
			if n, err := unix.Read(pair[1], bytes[:]); err != nil || n != 3 || string(bytes[:]) != "one" {
				t.Fatal("send", n, err)
			}
			if _, err := unix.Write(pair[1], []byte("two")); err != nil {
				t.Fatal(err)
			}
			f.want(ErrSuccess, "sock_recv", uint64(fd), 100, 1, 1, 300, 304)
			if string(f.read(200, 3)) != "two" {
				t.Fatal("peek")
			}
			f.want(ErrSuccess, "sock_recv", uint64(fd), 100, 1, 2, 300, 304)
			f.want(ErrInval, "sock_send", uint64(fd), 100, 1, 1, 300)
			f.want(ErrInval, "sock_recv", uint64(fd), 100, 1, 4, 300, 304)
			f.want(ErrFault, "sock_recv", uint64(fd), 100, 1, 0, 300, 131071)
			f.want(ErrNotsock, "sock_shutdown", 1, 1)
			f.want(ErrInval, "sock_shutdown", uint64(fd), 0)
			f.want(ErrSuccess, "fd_fdstat_set_flags", uint64(fd), 4)
			f.want(ErrAgain, "sock_recv", uint64(fd), 100, 1, 0, 300, 304)
			f.want(ErrSuccess, "fd_fdstat_get", uint64(fd), 0)
			if f.read(0, 1)[0] != 6 {
				t.Fatal("socket type")
			}
			f.want(ErrSuccess, "sock_shutdown", uint64(fd), 2)
			if n, err := unix.Read(pair[1], bytes[:]); n != 0 || err != nil {
				t.Fatal("shutdown", n, err)
			}
			var input [144]byte
			for i, descriptor := range []uint32{0, 1, 999} {
				b := input[i*48:]
				binary.LittleEndian.PutUint64(b, uint64(i+10))
				b[8] = 1
				if i == 1 {
					b[8] = 2
				}
				binary.LittleEndian.PutUint32(b[16:], descriptor)
			}
			f.write(1024, input[:])
			f.want(ErrSuccess, "poll_oneoff", 1024, 2048, 3, 300)
			if f.u32(300) != 3 || binary.LittleEndian.Uint16(f.read(2048+64+8, 2)) != uint16(ErrBadf) {
				t.Fatal("poll events", f.read(2048, 96))
			}
			input[8] = 0
			binary.LittleEndian.PutUint32(input[16:], 1)
			binary.LittleEndian.PutUint64(input[24:], 1_000_000)
			f.write(1024, input[:48])
			f.want(ErrSuccess, "poll_oneoff", 1024, 2048, 1, 300)
			if f.u32(300) != 1 {
				t.Fatal("clock event")
			}
			f.want(ErrInval, "poll_oneoff", 1024, 2048, 0, 300)
			f.want(ErrFault, "poll_oneoff", 131071, 2048, 1, 300)
			input[8] = 3
			f.write(1024, input[:48])
			f.want(ErrInval, "poll_oneoff", 1024, 2048, 1, 300)
		})
	}
}

func TestSocketAcceptAndBlockedClose(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		t.Run(backend, func(t *testing.T) {
			p := process(t, DefaultConfig())
			f := newFixture(t, p, backend)
			listener, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
			if err != nil {
				t.Fatal(err)
			}
			if err = unix.Bind(listener, &unix.SockaddrInet4{Addr: [4]byte{127, 0, 0, 1}}); err != nil {
				unix.Close(listener)
				t.Fatal(err)
			}
			if err = unix.Listen(listener, 2); err != nil {
				unix.Close(listener)
				t.Fatal(err)
			}
			address, err := unix.Getsockname(listener)
			if err != nil {
				t.Fatal(err)
			}
			fd, err := p.ImportSocket(os.NewFile(uintptr(listener), "listener"), true)
			if err != nil {
				t.Fatal(err)
			}
			f.want(ErrSuccess, "fd_fdstat_set_flags", uint64(fd), 4)
			f.want(ErrAgain, "sock_accept", uint64(fd), 0, 300)
			client, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer unix.Close(client)
			if err = unix.Connect(client, address); err != nil {
				t.Fatal(err)
			}
			f.want(ErrSuccess, "sock_accept", uint64(fd), 4, 300)
			child := f.u32(300)
			f.want(ErrSuccess, "fd_fdstat_set_flags", uint64(child), 0)
			f.iovec(100, 200, 3)
			fn, err := f.in.Export("sock_recv")
			if err != nil {
				t.Fatal(err)
			}
			result := make(chan error, 1)
			go func() {
				_, err := fn.Call(f.ctx, []vm.Value{vm.I32(child), vm.I32(100), vm.I32(1), vm.I32(0), vm.I32(300), vm.I32(304)}, vm.CallOptions{})
				result <- err
			}()
			// Both Store close and Process close must drain an outstanding socket read.
			f.s.Close()
			if err := f.s.WaitClosed(f.ctx); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-result:
				if err == nil {
					t.Fatal("blocked read survived Store close")
				}
			case <-f.ctx.Done():
				t.Fatal(f.ctx.Err())
			}
			p.Close()
			if err := p.WaitClosed(f.ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}
