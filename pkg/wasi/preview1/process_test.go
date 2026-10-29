//go:build linux || darwin

package preview1

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

func TestProcessConfigurationAndCleanup(t *testing.T) {
	for _, change := range []func(*Config){
		func(c *Config) { c.Args = []string{"bad\x00arg"} }, func(c *Config) { c.Env = []string{"bad\x00env"} },
		func(c *Config) { c.Args = []string{strings.Repeat("x", 100)}; c.Limits.MaxConfigBytes = 20 },
		func(c *Config) { c.Limits.MaxDescriptors = 2 }, func(c *Config) { c.Clock.Now = func(uint32) (uint64, error) { return 0, nil } },
		func(c *Config) { c.Preopens = []Preopen{{"", t.TempDir(), false}} }, func(c *Config) { c.Preopens = []Preopen{{"/", filepath.Join(t.TempDir(), "missing"), false}} },
		func(c *Config) { c.Limits.MaxDescriptors = 3; c.Preopens = []Preopen{{"/", t.TempDir(), false}} },
	} {
		c := DefaultConfig()
		change(&c)
		if p, err := NewProcess(c); err == nil {
			p.Close()
			t.Fatal("invalid configuration accepted")
		}
	}
	p := process(t, Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.WaitClosed(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, ok := p.Exited(); ok {
		t.Fatal("initial exit")
	}
	p.Close()
	if err := p.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := p.begin(ctx); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := p.ImportSocket(nil, false); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	f, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.ImportSocket(f, false); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err = f.Stat(); !errors.Is(err, fs.ErrClosed) {
		t.Fatal("ownership after import failure", err)
	}
	if err = p.reserveDescriptor(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	for raw, want := range errnoMap {
		if got := errno(raw); got != want {
			t.Fatal(raw, got, want)
		}
	}
	for _, test := range []struct {
		err  error
		want Errno
	}{{nil, ErrSuccess}, {context.Canceled, ErrCanceled}, {context.DeadlineExceeded, ErrTimedout}, {fs.ErrClosed, ErrBadf}, {fs.ErrInvalid, ErrInval}, {io.ErrUnexpectedEOF, ErrIo}, {unix.Errno(99999), ErrIo}, {ErrNomem, ErrNomem}} {
		if got := errno(test.err); got != test.want {
			t.Fatal(test.err, got)
		}
	}
}

func TestStreamsAndSharedOutput(t *testing.T) {
	ctx := context.Background()
	buf := NewBuffer(5)
	if n, err := buf.Write(ctx, []byte("abc")); n != 3 || err != nil {
		t.Fatal(n, err)
	}
	if n, err := buf.Write(ctx, []byte("def")); n != 2 || !errors.Is(err, ErrNospc) {
		t.Fatal(n, err)
	}
	var b [8]byte
	if n, err := buf.Read(ctx, b[:]); n != 5 || err != nil {
		t.Fatal(n, err)
	}
	if n, err := buf.Read(ctx, b[:]); n != 0 || err != io.EOF {
		t.Fatal(n, err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := buf.Read(canceled, b[:]); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := buf.Write(canceled, b[:]); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, _, _, err := buf.Ready(canceled, true); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	buf.Close()
	if _, err := buf.Read(ctx, b[:]); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	if _, err := buf.Write(ctx, b[:]); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	if _, _, err := buf.Poll(ctx, false); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	output := NewBuffer(64)
	cfg := DefaultConfig()
	cfg.Stdout, cfg.Stderr = output, output
	p := process(t, cfg)
	if err := p.closeFD(1); err != nil {
		t.Fatal(err)
	}
	d, release, err := p.hold(2, RightFdWrite)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := d.resource.stream.Write(ctx, []byte("still open")); n != 10 || err != nil {
		t.Fatal(n, err)
	}
	release()
	p.Close()
	if string(output.Bytes()) != "still open" {
		t.Fatal("shared output")
	}
	f := OutputFunc(func(ctx context.Context, b []byte) (int, error) { return len(b), ctx.Err() })
	if _, err := f.Read(ctx, b[:]); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	if _, _, _, err := f.Ready(ctx, true); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	if _, _, _, err := f.Ready(canceled, false); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := f.Write(ctx, b[:]); err != nil {
		t.Fatal(err)
	}
	if _, err := FileStream(nil); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	stream, err := FileStream(r)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if _, _, ready, err := stream.Ready(ctx, true); err != nil || ready {
		t.Fatal(ready, err)
	}
	deadline, stop := context.WithTimeout(ctx, time.Millisecond)
	defer stop()
	if _, err := stream.Read(deadline, b[:]); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	if _, err = w.Write([]byte("pipe")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := stream.Poll(ctx, true); err != nil {
		t.Fatal(err)
	}
	if n, err := stream.Read(ctx, b[:]); n != 4 || err != nil {
		t.Fatal(n, err)
	}
	if _, _, _, err := stream.Ready(canceled, true); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	closed, err := os.CreateTemp(t.TempDir(), "invalid")
	if err != nil {
		t.Fatal(err)
	}
	closed.Close()
	if _, err = FileStream(closed); err == nil {
		t.Fatal("closed stream accepted")
	}
}

func TestCapabilityRightsAndPointerValidation(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "readonly"), []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultConfig()
	cfg.Preopens = []Preopen{{"/", root, true}}
	p := process(t, cfg)
	f := newFixture(t, p, "native")
	f.write(400, []byte("created"))
	f.want(ErrNotcapable, "path_open", 3, 0, 400, 7, 1, 0, 0, 0, 500)
	f.want(ErrNotcapable, "path_link", 3, 0, 400, 7, 3, 400, 7)
	f.want(ErrNotcapable, "path_create_directory", 3, 400, 7)
	f.want(ErrInval, "path_open", 3, 2, 400, 7, 0, 0, 0, 0, 500)
	f.want(ErrInval, "path_open", 3, 0, 400, 7, 16, 0, 0, 0, 500)
	f.want(ErrFault, "path_open", 3, 0, 400, 7, 0, 0, 0, 0, 131071)
	f.want(ErrNotcapable, "path_open", 3, 0, 400, 7, 0, uint64(RightFdWrite), 0, 0, 500)
	f.want(ErrNametoolong, "path_open", 3, 0, 400, 4097, 0, 0, 0, 0, 500)
	f.write(400, []byte("bad\x00path"))
	f.want(ErrInval, "path_open", 3, 0, 400, 8, 0, 0, 0, 0, 500)
	f.want(ErrFault, "path_open", 3, 0, 131071, 8, 0, 0, 0, 0, 500)
	f.want(ErrNotcapable, "path_filestat_get", 0, 0, 400, 7, 500)
	f.want(ErrInval, "path_filestat_get", 3, 2, 400, 7, 500)
	f.want(ErrFault, "path_filestat_get", 3, 0, 400, 7, 131071)
	f.want(ErrInval, "path_filestat_set_times", 3, 0, 400, 7, 0, 0, 3)
	f.want(ErrNotcapable, "path_filestat_set_times", 3, 0, 400, 7, 1<<63, 0, 1)
	f.want(ErrInval, "path_link", 3, 2, 400, 7, 3, 400, 7)
	f.want(ErrFault, "path_readlink", 3, 400, 7, 131071, 2, 300)
	f.want(ErrFault, "path_readlink", 3, 400, 7, 0, 1, 131071)
	f.want(ErrFault, "fd_readdir", 3, 131071, 2, 0, 300)
	f.want(ErrInval, "fd_readdir", 3, 0, 32, 999, 300)
	if _, err := os.Stat(filepath.Join(root, "created")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	e := vm.NewEngine()
	defer e.Close()
	store, err := e.NewStore(vm.DefaultStoreConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	linker := vm.NewLinker(store)
	if err := p.Install(store, linker); err != nil {
		t.Fatal(err)
	}
	if err := p.Install(store, linker); !errors.Is(err, vm.ErrLink) {
		t.Fatal("duplicate install", err)
	}
	small := vm.DefaultStoreConfig()
	small.Limits.MaxFunctions = 0
	limited, err := e.NewStore(small)
	if err != nil {
		t.Fatal(err)
	}
	defer limited.Close()
	if err := p.Install(limited, vm.NewLinker(limited)); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("install limit", err)
	}
	p.Close()
	if err := p.Install(store, linker); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal("closed install", err)
	}
}
