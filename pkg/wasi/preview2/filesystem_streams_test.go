//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"golang.org/x/sys/unix"
)

func fsSource(t *testing.T, p *FileSystem, path string, input bool) *fsStream {
	t.Helper()
	mode := unix.O_RDWR | unix.O_NONBLOCK
	fd, err := unix.Open(path, mode, 0600)
	if err != nil {
		t.Fatal(err)
	}
	file := os.NewFile(uintptr(fd), path)
	stat, err := host.Fstat(file)
	if err != nil {
		t.Fatal(err)
	}
	b := &fsBacking{owner: p, file: file, refs: 1}
	p.mu.Lock()
	p.backings[b] = struct{}{}
	p.mu.Unlock()
	ctx, cancel := context.WithCancel(p.ctx)
	source := &fsStream{backing: b, input: input, regular: fsType(stat.Mode) == 6, ctx: ctx, cancel: cancel, changed: make(chan struct{})}
	t.Cleanup(func() { source.Close() })
	return source
}
func TestFileSystemStreamsIndependentCursorsAndAppend(t *testing.T) {
	p, _, root := fsOwner(t, FileSystemConfig{})
	read := fsSource(t, p, filepath.Join(root, "bench.bin"), true)
	// Two streams over the same backing exercise flag serialization against
	// positional writes, with no reopen or shared seek position.
	if err := read.backing.acquire(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(p.ctx)
	write := &fsStream{backing: read.backing, regular: true, offset: 4, ctx: ctx, cancel: cancel, changed: make(chan struct{})}
	defer write.Close()
	if err := read.backing.acquire(); err != nil {
		t.Fatal(err)
	}
	ctx2, cancel2 := context.WithCancel(p.ctx)
	appendSource := &fsStream{backing: read.backing, regular: true, appendMode: true, ctx: ctx2, cancel: cancel2, changed: make(chan struct{})}
	defer appendSource.Close()
	var wg sync.WaitGroup
	for _, source := range []*fsStream{write, appendSource} {
		wg.Add(1)
		go func(s *fsStream) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if n, err := s.Write(context.Background(), []byte("X")); err != nil || n != 1 {
					t.Errorf("write %d %v", n, err)
					return
				}
			}
		}(source)
	}
	wg.Wait()
	data := make([]byte, 4)
	if n, err := read.Read(context.Background(), data); err != nil || n != 4 || string(data) != "0123" {
		t.Fatal(n, err, string(data))
	}
	if ready, _ := read.Readiness(); !ready {
		t.Fatal("regular file not ready")
	}
	if err := write.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Appends extend the current end atomically; the positional writer never
	// changes its own cursor to the append position.
	if write.offset != 104 || read.offset != 4 {
		t.Fatal(write.offset, read.offset)
	}
	st, err := host.Fstat(read.backing.file)
	if err != nil || st.Size < 116 || st.Size > 204 {
		t.Fatal(st, err)
	}
	if n, err := read.Write(context.Background(), nil); n != 0 || fsError(err) != FSNotPermitted {
		t.Fatal(n, err)
	}
	if n, err := write.Read(context.Background(), nil); n != 0 || fsError(err) != FSNotPermitted {
		t.Fatal(n, err)
	}
	read.offset = 1<<63 - 1
	if n, err := read.Read(context.Background(), make([]byte, 1)); n != 0 || fsError(err) != FSOverflow {
		t.Fatal(n, err)
	}
	write.offset = 1<<63 - 1
	if n, err := write.Write(context.Background(), []byte("x")); n != 0 || fsError(err) != FSOverflow {
		t.Fatal(n, err)
	}
	read.offset = st.Size
	if n, err := read.Read(context.Background(), data); n != 0 || !errors.Is(err, io.EOF) {
		t.Fatal(n, err)
	}
	ctx3, cancel3 := context.WithCancel(context.Background())
	cancel3()
	if _, err := read.Read(ctx3, data); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := write.Flush(ctx3); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := read.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := read.Read(context.Background(), data); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if _, err := read.Write(context.Background(), data); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if err := read.Flush(context.Background()); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if ready, _ := read.Readiness(); !ready {
		t.Fatal("closed stream not ready")
	}
}
func TestFileSystemPipeReadinessCancellationAndShutdown(t *testing.T) {
	for _, closeOwner := range []bool{false, true} {
		t.Run(map[bool]string{false: "stream-close", true: "owner-close"}[closeOwner], func(t *testing.T) {
			p, _, root := fsOwner(t, FileSystemConfig{})
			path := filepath.Join(root, "pipe")
			if err := unix.Mkfifo(path, 0600); err != nil {
				t.Fatal(err)
			}
			source := fsSource(t, p, path, true)
			if n, err := source.Read(context.Background(), make([]byte, 8)); n != 0 || !errors.Is(err, ErrWouldBlock) {
				t.Fatal(n, err)
			}
			ready, changed := source.Readiness()
			if ready || changed == nil {
				t.Fatal(ready, changed)
			}
			ready, again := source.Readiness()
			if ready || again != changed {
				t.Fatal("duplicate watcher", ready, changed, again)
			}
			if _, err := unix.Write(host.Descriptor(source.backing.file), []byte("ok")); err != nil {
				t.Fatal(err)
			}
			select {
			case <-changed:
			case <-time.After(time.Second):
				t.Fatal("readiness not notified")
			}
			if ready, _ := source.Readiness(); !ready {
				t.Fatal("pipe not ready")
			}
			data := make([]byte, 8)
			n, err := source.Read(context.Background(), data)
			if err != nil || string(data[:n]) != "ok" {
				t.Fatal(n, err)
			}
			ready, changed = source.Readiness()
			if ready {
				t.Fatal("drained pipe ready")
			}
			if closeOwner {
				p.Close()
				if err := p.WaitClosed(nil); err != nil {
					t.Fatal(err)
				}
			} else {
				source.Close()
			}
			select {
			case <-changed:
			case <-time.After(time.Second):
				t.Fatal("close did not notify")
			}
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() { defer wg.Done(); source.Close() }()
			}
			wg.Wait()
			if ready, _ := source.Readiness(); !ready {
				t.Fatal("closed pipe not ready")
			}
		})
	}
}
func TestFileSystemPipeOutputAndErrors(t *testing.T) {
	p, _, root := fsOwner(t, FileSystemConfig{})
	path := filepath.Join(root, "pipe")
	if err := unix.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	source := fsSource(t, p, path, false)
	if n, err := source.Write(context.Background(), []byte("ok")); err != nil || n != 2 {
		t.Fatal(n, err)
	}
	if ready, _ := source.Readiness(); !ready {
		t.Fatal("pipe output not ready")
	}
	for i := 0; i < 10000; i++ {
		_, err := source.Write(context.Background(), make([]byte, 4096))
		if errors.Is(err, ErrWouldBlock) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if i == 9999 {
			t.Fatal("pipe never filled")
		}
	}
	if ready, changed := source.Readiness(); ready || changed == nil {
		t.Fatal("full pipe ready")
	}
	source.Close()
	for _, err := range []error{io.EOF, context.Canceled, context.DeadlineExceeded} {
		if !errors.Is(fsStreamError(err), err) {
			t.Fatal(err)
		}
	}
	if !errors.Is(fsStreamError(unix.EAGAIN), ErrWouldBlock) {
		t.Fatal("would block")
	}
	var typed *FileSystemError
	if !errors.As(fsStreamError(unix.EPERM), &typed) || typed.Code != FSNotPermitted {
		t.Fatal(typed)
	}
	source = fsSource(t, p, filepath.Join(root, "bench.bin"), false)
	source.failure = fsCode(FSIo)
	if err := source.Flush(context.Background()); fsError(err) != FSIo {
		t.Fatal(err)
	}
	if ready, _ := source.Readiness(); !ready {
		t.Fatal("failed stream not ready")
	}
}
func TestFileSystemStreamDowncastAndRevocation(t *testing.T) {
	p, _, _ := fsOwner(t, FileSystemConfig{})
	a := fsAPI{t, fsInstance(t, p, "interpreter", "filesystem-api")}
	root := a.root()
	file := a.open(root, "bench.bin", 0, fsRead)
	stream := a.ok("read-via-stream", file, component.Value{})
	p.Close()
	if err := p.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	result := a.call("streams", "[method]input-stream.read", stream, component.Value{Bits: 1})
	if result.Case != 1 || result.Elements[0].Case != 0 {
		t.Fatal(result)
	}
	failure := result.Elements[0].Elements[0]
	code := a.call("types", "filesystem-error-code", failure)
	if code.Case != 1 || code.Elements[0].Bits != uint64(FSBadDescriptor) {
		t.Fatal(code)
	}
	debug := a.call("errors", "[method]error.to-debug-string", failure)
	if debug.Text == "" {
		t.Fatal(debug)
	}
	a.drop(failure)
	a.drop(stream)
	a.drop(file)
	a.drop(root)
}
