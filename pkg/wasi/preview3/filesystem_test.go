//go:build linux || darwin

package preview3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

type fileAPI struct {
	t  *testing.T
	in *component.Instance
}

func (a fileAPI) call(name string, args ...component.Value) *component.Value {
	a.t.Helper()
	f, err := a.in.Function("fs", "[method]descriptor."+name)
	if err != nil {
		a.t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	v, err := f.Call(ctx, args, component.CallOptions{})
	if err != nil {
		a.t.Fatal(name, err)
	}
	return v.Value
}
func (a fileAPI) ok(name string, args ...component.Value) component.Value {
	a.t.Helper()
	v := a.call(name, args...)
	if v.Case != 0 {
		a.t.Fatal(name, "filesystem error", v)
	}
	if len(v.Elements) == 0 {
		return component.Value{}
	}
	return v.Elements[0]
}
func (a fileAPI) root() component.Value {
	a.t.Helper()
	f, err := a.in.Function("preopens", "get-directories")
	if err != nil {
		a.t.Fatal(err)
	}
	v, err := f.Call(nil, nil, component.CallOptions{})
	if err != nil || len(v.Value.Elements) != 1 {
		a.t.Fatal(v, err)
	}
	return v.Value.Elements[0].Elements[0]
}
func (a fileAPI) drop(v component.Value) {
	a.t.Helper()
	if _, err := a.in.DropResource(nil, v.Resource, component.CallOptions{}); err != nil {
		a.t.Fatal(err)
	}
}
func (a fileAPI) open(root component.Value, path string, openFlags, flags uint64) component.Value {
	a.t.Helper()
	return a.ok("open-at", root, component.Value{}, component.Value{Text: path}, component.Value{Bits: openFlags}, component.Value{Bits: flags})
}
func fileOwner(t *testing.T, config FileSystemConfig) (*FileSystem, *IO, string) {
	t.Helper()
	root := t.TempDir()
	if len(config.Preopens) == 0 {
		config.Preopens = []FileSystemPreopen{{GuestPath: "/data", HostPath: root}}
	}
	process, err := NewIO(IOConfig{MaxReadBytes: 2, MaxWriteBytes: 2})
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewFileSystem(process, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		p.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := p.WaitClosed(ctx); err != nil {
			t.Error(err)
		}
		process.Close()
		if err := process.WaitClosed(ctx); err != nil {
			t.Error(err)
		}
	})
	return p, process, root
}
func (a fileAPI) drain(stream, future component.Value) ([]component.Value, string) {
	a.t.Helper()
	var values []component.Value
	var text string
	for {
		r, err := a.in.ReadAsync(nil, stream.Resource, 2, component.CallOptions{})
		if err != nil {
			a.t.Fatal(err)
		}
		values = append(values, r.Values...)
		text += r.Bytes
		if r.Closed {
			break
		}
	}
	r, err := a.in.ReadAsync(nil, future.Resource, 1, component.CallOptions{})
	if err != nil || len(r.Values) != 1 || r.Values[0].Case != 0 {
		a.t.Fatal("stream future", r, err)
	}
	for _, ref := range []*component.Resource{stream.Resource, future.Resource} {
		if _, err := a.in.DropAsync(nil, ref, component.CallOptions{}); err != nil {
			a.t.Fatal(err)
		}
	}
	return values, text
}
func TestFileSystemGuestMissingFile(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, _, _ := fileOwner(t, FileSystemConfig{})
		in := instantiate(t, compileFixture(t, backend, "filesystem-guest"), p.Imports())
		if v := call(t, in, "run"); v.Bits != 1<<32|uint64(fileErrorCases["no-entry"]) {
			t.Fatal(v)
		}
		p.mu.Lock()
		remaining := len(p.descriptors)
		p.mu.Unlock()
		if remaining != 0 {
			t.Fatal("failed guest open retained descriptor", remaining)
		}
	})
}
func TestFileSystemVersionedGuestAndNativeStreams(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, _, root := fileOwner(t, FileSystemConfig{})
		if err := os.WriteFile(filepath.Join(root, "data.txt"), []byte("abc"), 0600); err != nil {
			t.Fatal(err)
		}
		in := instantiate(t, compileFixture(t, backend, "filesystem-guest"), p.Imports())
		if v := call(t, in, "run"); v.Bits != 3 {
			t.Fatal("guest async open/stat", v)
		}
		in = instantiate(t, compileFixture(t, backend, "filesystem-api"), p.Imports())
		api := fileAPI{t, in}
		dir := api.root()
		if api.ok("get-flags", dir).Bits != fsRead|fsMutate || api.ok("get-type", dir).Case != 2 {
			t.Fatal("directory flags/type")
		}
		file := api.open(dir, "data.txt", 0, fsRead|fsWrite)
		stat := api.ok("stat", file)
		if stat.Elements[0].Case != 5 || stat.Elements[2].Bits != 3 {
			t.Fatal(stat)
		}
		if api.ok("stat-at", dir, component.Value{}, component.Value{Text: "data.txt"}).Elements[2].Bits != 3 {
			t.Fatal("stat-at")
		}
		for _, name := range []string{"advise", "sync-data", "sync"} {
			args := []component.Value{file}
			if name == "advise" {
				args = append(args, component.Value{}, component.Value{Bits: 3}, component.Value{})
			}
			api.ok(name, args...)
		}
		stamp := component.Value{Case: 2, Elements: []component.Value{{Elements: []component.Value{{Bits: ^uint64(0)}, {Bits: 123}}}}}
		api.ok("set-times", file, stamp, stamp)
		stat = api.ok("stat", file)
		if int64(stat.Elements[3].Elements[0].Elements[0].Bits) != -1 {
			t.Fatal("negative filesystem time narrowed", stat)
		}
		api.ok("set-times-at", dir, component.Value{}, component.Value{Text: "data.txt"}, component.Value{}, component.Value{Case: 1})
		if api.call("is-same-object", file, file).Bits != 1 {
			t.Fatal("object identity")
		}
		api.ok("metadata-hash", file)
		api.ok("metadata-hash-at", dir, component.Value{}, component.Value{Text: "data.txt"})
		api.ok("create-directory-at", dir, component.Value{Text: "sub"})
		api.ok("link-at", dir, component.Value{}, component.Value{Text: "data.txt"}, dir, component.Value{Text: "hard"})
		api.ok("rename-at", dir, component.Value{Text: "hard"}, dir, component.Value{Text: "renamed"})
		api.ok("symlink-at", dir, component.Value{Text: "data.txt"}, component.Value{Text: "link"})
		if api.ok("readlink-at", dir, component.Value{Text: "link"}).Text != "data.txt" {
			t.Fatal("readlink")
		}
		api.ok("unlink-file-at", dir, component.Value{Text: "renamed"})
		api.ok("unlink-file-at", dir, component.Value{Text: "link"})
		api.ok("remove-directory-at", dir, component.Value{Text: "sub"})
		api.ok("set-size", file, component.Value{Bits: 0})
		for _, method := range []string{"write-via-stream", "append-via-stream"} {
			f, _ := in.Function("fs", "[method]descriptor."+method)
			_, typ := f.Info().Parameter(1)
			pending := "hello"
			source, err := in.NewSource(nil, typ, component.AsyncSource{Read: func(_ *component.Caller, limit uint32) (*component.ReadCompletion, error) {
				n := min(int(limit), len(pending))
				chunk := pending[:n]
				pending = pending[n:]
				done := pending == ""
				p := component.NewReadCompletion()
				p.Resolve(func(*component.Caller) (component.ReadResult, error) {
					return component.ReadResult{Bytes: chunk, Closed: done}, nil
				})
				return p, nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			args := []component.Value{file, source}
			if method == "write-via-stream" {
				args = append(args, component.Value{})
			}
			future := api.call(method, args...)
			r, err := in.ReadAsync(nil, future.Resource, 1, component.CallOptions{})
			if err != nil || r.Values[0].Case != 0 {
				t.Fatal(method, r, err)
			}
			in.DropAsync(nil, future.Resource, component.CallOptions{})
		}
		pair := api.call("read-via-stream", file, component.Value{})
		api.drop(file)
		_, text := api.drain(pair.Elements[0], pair.Elements[1])
		if text != "hellohello" {
			t.Fatal("independent stream lifetime", text)
		}
		pair = api.call("read-directory", dir)
		entries, _ := api.drain(pair.Elements[0], pair.Elements[1])
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Elements[1].Text)
		}
		sort.Strings(names)
		if len(names) != 1 || names[0] != "data.txt" {
			t.Fatal("directory stream", names)
		}
		api.drop(dir)
		p.mu.Lock()
		handles := len(p.descriptors)
		streams := len(p.streams)
		p.mu.Unlock()
		if handles != 0 || streams != 0 {
			t.Fatal("filesystem retained handles/streams", handles, streams)
		}
	})
}
func TestFileSystemBindingContracts(t *testing.T) {
	p, _, _ := fileOwner(t, FileSystemConfig{})
	for _, name := range []string{"filesystem-wrong-flags", "filesystem-wrong-time"} {
		m := compileFixture(t, "interpreter", name)
		if _, err := m.Instantiate(nil, p.Imports(), component.CallOptions{}); err == nil {
			t.Fatal(name, "accepted wrong WIT shape")
		}
	}
}

func (a fileAPI) fail(name, label string, args ...component.Value) {
	a.t.Helper()
	v := a.call(name, args...)
	if v.Case != 1 || len(v.Elements) != 1 || v.Elements[0].Case != fileErrorCases[label] {
		a.t.Fatal(name, label, v)
	}
}
func (a fileAPI) failedStream(pair *component.Value, label string) {
	a.t.Helper()
	r, err := a.in.ReadAsync(nil, pair.Elements[0].Resource, 1, component.CallOptions{})
	if err != nil || !r.Closed {
		a.t.Fatal(r, err)
	}
	r, err = a.in.ReadAsync(nil, pair.Elements[1].Resource, 1, component.CallOptions{})
	if err != nil || len(r.Values) != 1 || r.Values[0].Case != 1 || r.Values[0].Elements[0].Case != fileErrorCases[label] {
		a.t.Fatal(label, r, err)
	}
	a.in.DropAsync(nil, pair.Elements[0].Resource, component.CallOptions{})
	a.in.DropAsync(nil, pair.Elements[1].Resource, component.CallOptions{})
}
func TestFileSystemErrorsLimitsAndCapabilities(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, _, root := fileOwner(t, FileSystemConfig{})
		os.WriteFile(filepath.Join(root, "data.txt"), []byte("abc"), 0600)
		in := instantiate(t, compileFixture(t, backend, "filesystem-api"), p.Imports())
		api := fileAPI{t, in}
		dir := api.root()
		file := api.open(dir, "data.txt", 0, fsRead)
		api.fail("open-at", "no-entry", dir, component.Value{}, component.Value{Text: "missing"}, component.Value{}, component.Value{Bits: fsRead})
		api.fail("open-at", "not-permitted", dir, component.Value{}, component.Value{Text: "../escape"}, component.Value{}, component.Value{Bits: fsRead})
		api.fail("open-at", "invalid", dir, component.Value{}, component.Value{Text: "a\x00b"}, component.Value{}, component.Value{})
		api.fail("open-at", "not-directory", file, component.Value{}, component.Value{Text: "child"}, component.Value{}, component.Value{})
		api.fail("stat-at", "not-directory", file, component.Value{}, component.Value{Text: "child"})
		api.fail("stat-at", "no-entry", dir, component.Value{}, component.Value{Text: "missing"})
		api.fail("readlink-at", "no-entry", dir, component.Value{Text: "missing"})
		api.fail("create-directory-at", "not-directory", file, component.Value{Text: "sub"})
		api.fail("set-size", "not-permitted", file, component.Value{Bits: 1})
		api.fail("set-size", "is-directory", dir, component.Value{Bits: 1})
		writable := api.open(dir, "data.txt", 0, fsRead|fsWrite)
		api.fail("set-size", "overflow", writable, component.Value{Bits: 1 << 63})
		api.fail("advise", "overflow", writable, component.Value{Bits: 1 << 63}, component.Value{}, component.Value{})
		badTime := component.Value{Case: 2, Elements: []component.Value{{Elements: []component.Value{{}, {Bits: 1e9}}}}}
		api.fail("set-times", "invalid", writable, badTime, component.Value{})
		api.fail("set-times-at", "invalid", dir, component.Value{}, component.Value{Text: "data.txt"}, component.Value{}, badTime)
		api.fail("set-times", "not-permitted", file, component.Value{}, component.Value{})
		api.fail("set-times-at", "not-directory", file, component.Value{}, component.Value{Text: "x"}, component.Value{}, component.Value{})
		if err := os.Symlink("/outside", filepath.Join(root, "absolute")); err != nil {
			t.Fatal(err)
		}
		api.fail("readlink-at", "not-permitted", dir, component.Value{Text: "absolute"})
		if api.call("is-same-object", dir, file).Bits != 0 {
			t.Fatal("distinct objects aliased")
		}
		api.failedStream(api.call("read-via-stream", file, component.Value{Bits: 1 << 63}), "overflow")
		api.failedStream(api.call("read-directory", file), "not-directory")
		noRead := api.open(dir, "data.txt", 0, 0)
		api.failedStream(api.call("read-via-stream", noRead, component.Value{}), "not-permitted")
		api.drop(noRead)
		api.drop(writable)
		api.drop(file)
		api.drop(dir)
		closed, _ := in.Function("fs", "[method]descriptor.get-type")
		if _, err := closed.Call(nil, []component.Value{file}, component.CallOptions{}); err == nil {
			t.Fatal("stale resource reached filesystem")
		}
	})
}
func TestFileSystemReadOnlyAndRestoredQuotas(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		root := t.TempDir()
		os.WriteFile(filepath.Join(root, "data.txt"), []byte("abc"), 0600)
		p, _, _ := fileOwner(t, FileSystemConfig{Preopens: []FileSystemPreopen{{GuestPath: "/ro", HostPath: root, ReadOnly: true}}})
		in := instantiate(t, compileFixture(t, backend, "filesystem-api"), p.Imports())
		api := fileAPI{t, in}
		dir := api.root()
		api.fail("create-directory-at", "read-only", dir, component.Value{Text: "new"})
		api.fail("open-at", "read-only", dir, component.Value{}, component.Value{Text: "data.txt"}, component.Value{Bits: 8}, component.Value{Bits: fsWrite})
		api.fail("set-times", "read-only", dir, component.Value{}, component.Value{})
		api.fail("link-at", "read-only", dir, component.Value{}, component.Value{Text: "data.txt"}, dir, component.Value{Text: "hard"})
		api.fail("rename-at", "read-only", dir, component.Value{Text: "data.txt"}, dir, component.Value{Text: "new"})
		api.fail("symlink-at", "read-only", dir, component.Value{Text: "data.txt"}, component.Value{Text: "new"})
		api.drop(dir)
		limited, _, root := fileOwner(t, FileSystemConfig{MaxDescriptors: 2, MaxDirectoryEntries: 1, MaxDirectoryBatch: 1})
		os.WriteFile(filepath.Join(root, "data.txt"), []byte("x"), 0600)
		os.WriteFile(filepath.Join(root, "other"), []byte("y"), 0600)
		in = instantiate(t, compileFixture(t, backend, "filesystem-api"), limited.Imports())
		api = fileAPI{t, in}
		dir = api.root()
		file := api.open(dir, "data.txt", 0, fsRead)
		api.fail("open-at", "insufficient-memory", dir, component.Value{}, component.Value{Text: "other"}, component.Value{}, component.Value{Bits: fsRead})
		api.failedStream(api.call("read-directory", dir), "insufficient-memory")
		api.drop(file)
		pair := api.call("read-directory", dir)
		r, err := in.ReadAsync(nil, pair.Elements[0].Resource, 1, component.CallOptions{})
		if err != nil || r.Count != 1 {
			t.Fatal(r, err)
		}
		api.failedStream(pair, "insufficient-memory")
		file = api.open(dir, "data.txt", 0, fsRead)
		api.drop(file)
		api.drop(dir)
	})
}
func TestFileSystemConstructorAndClosedOwners(t *testing.T) {
	if _, err := NewFileSystem(nil, FileSystemConfig{}); err == nil {
		t.Fatal("missing IO owner")
	}
	process, _ := NewIO(IOConfig{})
	defer closeIO(t, process)
	root := t.TempDir()
	for _, config := range []FileSystemConfig{
		{MaxDescriptors: 1, Preopens: []FileSystemPreopen{{HostPath: root}, {HostPath: root}}},
		{MaxDirectoryEntries: 1 << 24}, {MaxDirectoryBatch: 1 << 24},
		{Preopens: []FileSystemPreopen{{GuestPath: string([]byte{255}), HostPath: root}}},
		{MaxPathBytes: 1, Preopens: []FileSystemPreopen{{GuestPath: "/long", HostPath: root}}},
		{Preopens: []FileSystemPreopen{{GuestPath: "/first", HostPath: root}, {GuestPath: "/missing", HostPath: filepath.Join(root, "absent")}}},
	} {
		if _, err := NewFileSystem(process, config); err == nil {
			t.Fatal("invalid grant/config", config)
		}
	}
	p, err := NewFileSystem(process, FileSystemConfig{})
	if err != nil {
		t.Fatal(err)
	}
	in := instantiate(t, compileFixture(t, "interpreter", "filesystem-api"), p.Imports())
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if err := p.WaitClosed(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	p.Close()
	p.Close()
	if err := p.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	f, _ := in.Function("preopens", "get-directories")
	if _, err := f.Call(nil, nil, component.CallOptions{}); !errors.Is(err, os.ErrClosed) {
		t.Fatal(err)
	}
	process.Close()
	if _, err := NewFileSystem(process, FileSystemConfig{}); err == nil {
		t.Fatal("closed IO grant accepted")
	}
}

func fileStreamSnapshot(t *testing.T, p *FileSystem) *fileStream {
	t.Helper()
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.streams) != 1 {
		t.Fatal("expected one stream", len(p.streams))
	}
	for s := range p.streams {
		return s
	}
	return nil
}
func awaitFilePoll(t *testing.T, s *fileStream) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		s.mu.Lock()
		watching := s.watching
		s.mu.Unlock()
		if watching {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("FIFO never suspended")
		}
		time.Sleep(time.Millisecond)
	}
}
func TestFileSystemFIFOBackpressureAndShutdown(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, mode := range []string{"read", "write", "provider-close", "instance-close"} {
			p, _, root := fileOwner(t, FileSystemConfig{MaxWorkers: 1})
			if err := unix.Mkfifo(filepath.Join(root, "pipe"), 0600); err != nil {
				t.Fatal(err)
			}
			in := instantiate(t, compileFixture(t, backend, "filesystem-api"), p.Imports())
			api := fileAPI{t, in}
			dir := api.root()
			file := api.open(dir, "pipe", 0, fsRead|fsWrite)
			pair := api.call("read-via-stream", file, component.Value{})
			s := fileStreamSnapshot(t, p)
			type outcome struct {
				r   component.ReadResult
				err error
			}
			done := make(chan outcome, 1)
			go func() {
				r, err := in.ReadAsync(nil, pair.Elements[0].Resource, 2, component.CallOptions{})
				done <- outcome{r, err}
			}()
			awaitFilePoll(t, s)
			if ready, changed := s.Readiness(); ready || changed == nil {
				t.Fatal("incoherent pending readiness")
			}
			if mode == "provider-close" {
				p.Close()
			} else if mode == "instance-close" {
				in.Close()
			} else {
				writer, err := os.OpenFile(filepath.Join(root, "pipe"), os.O_WRONLY|unix.O_NONBLOCK, 0)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := writer.Write([]byte("ab")); err != nil {
					t.Fatal(err)
				}
				writer.Close()
			}
			select {
			case result := <-done:
				switch mode {
				case "provider-close":
					if result.err != nil || !result.r.Closed {
						t.Fatal(result)
					}
				case "instance-close":
					if !errors.Is(result.err, context.Canceled) && !errors.Is(result.err, wasm.ErrClosed) {
						t.Fatal(result)
					}
				default:
					if result.err != nil || result.r.Bytes != "ab" || result.r.Closed {
						t.Fatal(result)
					}
				}
			case <-time.After(time.Second):
				t.Fatal("FIFO remained parked", mode)
			}
			if mode == "instance-close" {
				continue
			}
			in.DropAsync(nil, pair.Elements[0].Resource, component.CallOptions{})
			r, err := in.ReadAsync(nil, pair.Elements[1].Resource, 1, component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if mode == "provider-close" {
				if r.Values[0].Case != 1 || r.Values[0].Elements[0].Case != fileErrorCases["bad-descriptor"] {
					t.Fatal(r)
				}
				continue
			}
			if r.Values[0].Case != 0 {
				t.Fatal(r)
			}
			in.DropAsync(nil, pair.Elements[1].Resource, component.CallOptions{})
			if ready, _ := s.Readiness(); !ready {
				t.Fatal("closed stream did not wake readiness")
			}
			api.failedStream(api.call("read-via-stream", file, component.Value{Bits: 1}), "invalid-seek")
			if mode == "write" {
				reader, err := os.OpenFile(filepath.Join(root, "pipe"), os.O_RDONLY, 0)
				if err != nil {
					t.Fatal(err)
				}
				defer reader.Close()
				fn, _ := in.Function("fs", "[method]descriptor.write-via-stream")
				_, typ := fn.Info().Parameter(1)
				source, err := in.NewSource(nil, typ, component.AsyncSource{Read: func(*component.Caller, uint32) (*component.ReadCompletion, error) {
					p := component.NewReadCompletion()
					p.Resolve(func(*component.Caller) (component.ReadResult, error) {
						return component.ReadResult{Bytes: "xy", Closed: true}, nil
					})
					return p, nil
				}})
				if err != nil {
					t.Fatal(err)
				}
				future := api.call("write-via-stream", file, source, component.Value{})
				r, err := in.ReadAsync(nil, future.Resource, 1, component.CallOptions{})
				if err != nil || r.Values[0].Case != 0 {
					t.Fatal(r, err)
				}
				in.DropAsync(nil, future.Resource, component.CallOptions{})
				var data [2]byte
				if _, err := io.ReadFull(reader, data[:]); err != nil || string(data[:]) != "xy" {
					t.Fatal(data, err)
				}
			}
			api.drop(file)
			api.drop(dir)
		}
	})
}
func TestFileSystemDirectoryModesAndInvalidNames(t *testing.T) {
	p, _, root := fileOwner(t, FileSystemConfig{})
	os.Mkdir(filepath.Join(root, "sub"), 0700)
	os.WriteFile(filepath.Join(root, "file"), nil, 0600)
	os.Symlink("file", filepath.Join(root, "link"))
	unix.Mkfifo(filepath.Join(root, "pipe"), 0600)
	in := instantiate(t, compileFixture(t, "interpreter", "filesystem-api"), p.Imports())
	api := fileAPI{t, in}
	dir := api.root()
	first := api.call("read-directory", dir)
	second := api.call("read-directory", dir)
	a, _ := api.drain(first.Elements[0], first.Elements[1])
	b, _ := api.drain(second.Elements[0], second.Elements[1])
	if len(a) != 4 || len(b) != 4 {
		t.Fatal("directory cursors interfered", len(a), len(b))
	}
	modes := map[string]uint32{}
	for _, entry := range a {
		modes[entry.Elements[1].Text] = entry.Elements[0].Case
	}
	if modes["sub"] != 2 || modes["file"] != 5 || modes["link"] != 4 || modes["pipe"] != 3 {
		t.Fatal(modes)
	}
	name := string([]byte{255})
	if err := os.WriteFile(filepath.Join(root, name), nil, 0600); err != nil {
		if runtime.GOOS == "darwin" && errors.Is(err, unix.EILSEQ) {
			api.drop(dir)
			return
		}
		t.Fatal(err)
	}
	pair := api.call("read-directory", dir)
	api.failedStream(pair, "illegal-byte-sequence")
	api.drop(dir)
}

func TestFileSystemDeferredCloseWaitsAndCancellation(t *testing.T) {
	p, _, root := fileOwner(t, FileSystemConfig{MaxDescriptors: 2, MaxWorkers: 1})
	os.WriteFile(filepath.Join(root, "data.txt"), []byte("x"), 0600)
	if err := unix.Mkfifo(filepath.Join(root, "pipe"), 0600); err != nil {
		t.Fatal(err)
	}
	in := instantiate(t, compileFixture(t, "interpreter", "filesystem-api"), p.Imports())
	api := fileAPI{t, in}
	dir := api.root()
	pipe := api.open(dir, "pipe", 0, fsRead|fsWrite)
	var object *fileObject
	p.mu.Lock()
	for _, d := range p.descriptors {
		if d.object.directory == nil {
			object = d.object
		}
	}
	p.mu.Unlock()
	raw, err := object.file.SyscallConn()
	if err != nil {
		t.Fatal(err)
	}
	held, release, finished := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() { finished <- raw.Control(func(uintptr) { close(held); <-release }) }()
	<-held
	released := false
	defer func() {
		if !released {
			close(release)
		}
	}()
	// A close must return without waiting for an in-flight host FD lease.
	api.drop(pipe)
	fn, _ := in.Function("fs", "[method]descriptor.open-at")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, err = fn.Call(ctx, []component.Value{dir, {}, {Text: "data.txt"}, {}, {Bits: fsRead}}, component.CallOptions{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("reclamation wait ignored cancellation", err)
	}
	close(release)
	released = true
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		p.mu.Lock()
		active := p.active
		closing := p.closing
		p.mu.Unlock()
		if active == 0 && closing == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("close/reservation did not drain")
		}
		time.Sleep(time.Millisecond)
	}
	file := api.open(dir, "data.txt", 0, fsRead)
	api.drop(file)
	api.drop(dir)
}
func TestFileSystemWITVariantsAndSpecialModes(t *testing.T) {
	for _, item := range []struct {
		unix   uint32
		goMode os.FileMode
		want   uint32
	}{
		{unix.S_IFBLK, os.ModeDevice, 0}, {unix.S_IFCHR, os.ModeDevice | os.ModeCharDevice, 1},
		{unix.S_IFDIR, os.ModeDir, 2}, {unix.S_IFIFO, os.ModeNamedPipe, 3}, {unix.S_IFLNK, os.ModeSymlink, 4},
		{unix.S_IFREG, 0, 5}, {unix.S_IFSOCK, os.ModeSocket, 6}, {0, os.ModeIrregular, 7},
	} {
		a, b := fileType(item.unix), directoryType(item.goMode)
		if a.Case != item.want || b.Case != item.want {
			t.Fatal(item, a, b)
		}
		if item.want == 7 && (len(a.Elements) != 1 || len(b.Elements) != 1 || a.Elements[0].Case != 0) {
			t.Fatal("other type lost its optional payload")
		}
	}
	for _, item := range []struct {
		err  error
		name string
	}{
		{errors.New("opaque host detail"), "other"}, {unix.EILSEQ, "illegal-byte-sequence"}, {io.ErrClosedPipe, "bad-descriptor"},
		{context.Canceled, "interrupted"}, {context.DeadlineExceeded, "interrupted"}, {fileError("io"), "io"},
	} {
		v := fileCompletionValue(fileCompletionCode(item.err))
		if v.Case != 1 || v.Elements[0].Case != fileErrorCases[item.name] {
			t.Fatal(item, v)
		}
	}
	if string(fileError("read-only")) != fileError("read-only").Error() {
		t.Fatal("structured error lost its name")
	}
}

func (a fileAPI) writeResult(name string, file component.Value, offset uint64, text string) component.Value {
	a.t.Helper()
	f, _ := a.in.Function("fs", "[method]descriptor."+name)
	_, typ := f.Info().Parameter(1)
	pending := text
	source, err := a.in.NewSource(nil, typ, component.AsyncSource{Read: func(_ *component.Caller, limit uint32) (*component.ReadCompletion, error) {
		n := min(len(pending), int(limit))
		chunk := pending[:n]
		pending = pending[n:]
		closed := pending == ""
		p := component.NewReadCompletion()
		p.Resolve(func(*component.Caller) (component.ReadResult, error) {
			return component.ReadResult{Bytes: chunk, Closed: closed}, nil
		})
		return p, nil
	}})
	if err != nil {
		a.t.Fatal(err)
	}
	args := []component.Value{file, source}
	if name == "write-via-stream" {
		args = append(args, component.Value{Bits: offset})
	}
	future := a.call(name, args...)
	r, err := a.in.ReadAsync(nil, future.Resource, 1, component.CallOptions{})
	if err != nil || len(r.Values) != 1 {
		a.t.Fatal(r, err)
	}
	a.in.DropAsync(nil, future.Resource, component.CallOptions{})
	return r.Values[0]
}
func TestFileSystemOutputPermissionsOffsetsAndOpenModes(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, _, root := fileOwner(t, FileSystemConfig{})
		os.WriteFile(filepath.Join(root, "data.txt"), []byte("abc"), 0600)
		in := instantiate(t, compileFixture(t, backend, "filesystem-api"), p.Imports())
		api := fileAPI{t, in}
		dir := api.root()
		readOnly := api.open(dir, "data.txt", 0, fsRead)
		for _, name := range []string{"write-via-stream", "append-via-stream"} {
			r := api.writeResult(name, readOnly, 0, "x")
			if r.Case != 1 || r.Elements[0].Case != fileErrorCases["not-permitted"] {
				t.Fatal(name, r)
			}
		}
		api.drop(readOnly)
		writable := api.open(dir, "data.txt", 8, fsWrite|fsFileSync)
		if r := api.writeResult("write-via-stream", writable, 1<<63, "x"); r.Case != 1 || r.Elements[0].Case != fileErrorCases["overflow"] {
			t.Fatal(r)
		}
		api.failedStream(api.call("read-via-stream", writable, component.Value{}), "not-permitted")
		if r := api.writeResult("write-via-stream", writable, 0, "q"); r.Case != 0 {
			t.Fatal(r)
		}
		api.drop(writable)
		file := api.open(dir, "new", 1|4, fsRead|fsWrite)
		api.drop(file)
		api.fail("open-at", "exist", dir, component.Value{}, component.Value{Text: "new"}, component.Value{Bits: 1 | 4}, component.Value{Bits: fsRead | fsWrite})
		directory := api.open(dir, ".", 2, fsRead)
		api.drop(directory)
		if err := os.Mkdir(filepath.Join(root, "empty"), 0700); err != nil {
			t.Fatal(err)
		}
		directory = api.open(dir, "empty", 2, fsRead)
		pair := api.call("read-directory", directory)
		r, err := in.ReadAsync(nil, pair.Elements[0].Resource, 0, component.CallOptions{})
		if err != nil || r.Count != 0 || r.Closed {
			t.Fatal("zero directory readiness", r, err)
		}
		api.drop(directory)
		entries, _ := api.drain(pair.Elements[0], pair.Elements[1])
		if len(entries) != 0 {
			t.Fatal(entries)
		}
		file = api.open(dir, "data.txt", 0, fsRead)
		pair = api.call("read-via-stream", file, component.Value{})
		s := fileStreamSnapshot(t, p)
		if ready, _ := s.Readiness(); !ready {
			t.Fatal("regular file was not ready")
		}
		api.drain(pair.Elements[0], pair.Elements[1])
		api.drop(file)
		api.drop(dir)
	})
}
func TestFileSystemBlockedPipeOutputProgress(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, process, root := fileOwner(t, FileSystemConfig{MaxWorkers: 1})
		process.config.MaxWriteBytes = 64 << 10
		if err := unix.Mkfifo(filepath.Join(root, "pipe"), 0600); err != nil {
			t.Fatal(err)
		}
		in := instantiate(t, compileFixture(t, backend, "filesystem-api"), p.Imports())
		api := fileAPI{t, in}
		dir := api.root()
		file := api.open(dir, "pipe", 0, fsRead|fsWrite)
		f, _ := in.Function("fs", "[method]descriptor.write-via-stream")
		_, typ := f.Info().Parameter(1)
		payload := string(make([]byte, 128<<10))
		pending := payload
		source, err := in.NewSource(nil, typ, component.AsyncSource{Read: func(_ *component.Caller, limit uint32) (*component.ReadCompletion, error) {
			n := min(len(pending), int(limit))
			data := pending[:n]
			pending = pending[n:]
			closed := pending == ""
			p := component.NewReadCompletion()
			p.Resolve(func(*component.Caller) (component.ReadResult, error) {
				return component.ReadResult{Bytes: data, Closed: closed}, nil
			})
			return p, nil
		}})
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		var result component.CallResult
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		go func() {
			var err error
			result, err = f.Call(ctx, []component.Value{file, source, {}}, component.CallOptions{})
			done <- err
		}()
		deadline := time.Now().Add(time.Second)
		var stream *fileStream
		for {
			p.mu.Lock()
			for s := range p.streams {
				stream = s
			}
			p.mu.Unlock()
			if stream != nil {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("pipe output did not start")
			}
			time.Sleep(time.Millisecond)
		}
		awaitFilePoll(t, stream)
		reader, err := os.OpenFile(filepath.Join(root, "pipe"), os.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		read := make(chan error, 1)
		data := make([]byte, len(payload))
		go func() { _, err := io.ReadFull(reader, data); read <- err }()
		readDone, writeDone := false, false
		for !readDone || !writeDone {
			select {
			case err := <-done:
				if err != nil {
					t.Fatal("write", err)
				}
				writeDone = true
			case err := <-read:
				if err != nil || string(data) != payload {
					t.Fatal("read", err)
				}
				readDone = true
			case <-ctx.Done():
				stream.mu.Lock()
				failure := stream.failure
				watching := stream.watching
				stream.mu.Unlock()
				t.Fatal("pipe timeout", readDone, writeDone, failure, watching, ctx.Err())
			}
		}

		r, err := in.ReadAsync(nil, result.Value.Resource, 1, component.CallOptions{})
		if err != nil || r.Values[0].Case != 0 {
			t.Fatal(r, err)
		}
		in.DropAsync(nil, result.Value.Resource, component.CallOptions{})
		api.drop(file)
		api.drop(dir)
	})
}
func TestFileSystemFixtureProvenance(t *testing.T) {
	check := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("changed input", path)
		}
	}
	data, err := os.ReadFile("testdata/filesystem.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Version   string                       `json:"wasiVersion"`
		WIT       string                       `json:"witArchiveSHA256"`
		Producer  string                       `json:"producer"`
		Generator string                       `json:"generatorSHA256"`
		Sources   map[string]string            `json:"sourceSHA256"`
		Files     map[string]map[string]string `json:"sha256"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Version != "0.3.2" || pins.WIT != "2e918e047c6ad251a44b4d1141359495a27219058fd7090f8b8142f6bc0ec311" || pins.Producer != "wasm-tools 1.258.0" || len(pins.Files) != 5 {
		t.Fatal("changed provenance")
	}
	check("../../../scripts/build_preview3_filesystem_fixtures.py", pins.Generator)
	for path, hash := range pins.Sources {
		check("../../../"+path, hash)
	}
	for name, files := range pins.Files {
		for ext, hash := range files {
			check("testdata/"+name+"."+ext, hash)
		}
	}
	var schema struct {
		Version   string `json:"wasiVersion"`
		WIT       string `json:"witArchiveSHA256"`
		Generator string `json:"generatorSHA256"`
		Schema    string `json:"schemaSHA256"`
	}
	data, err = os.ReadFile("testdata/filesystem-schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	if schema.Version != pins.Version || schema.WIT != pins.WIT {
		t.Fatal("changed schema pin")
	}
	check("../../../scripts/build_preview3_filesystem_schema.py", schema.Generator)
	check("filesystem_schema.go", schema.Schema)
}

func TestFileSystemLateOpenResultCleanup(t *testing.T) {
	p, _, root := fileOwner(t, FileSystemConfig{MaxDescriptors: 2})
	for _, late := range []bool{false, true} {
		if err := p.reserveObject(context.Background()); err != nil {
			t.Fatal(err)
		}
		fd, err := os.Create(filepath.Join(root, "unpublished"))
		if err != nil {
			t.Fatal(err)
		}
		object, err := p.registerObject(fd, nil)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		job := &fileCall{cancel: cancel, stop: func() bool { return true }}
		if late {
			job.close()
		}
		job.publish(&fileDescriptor{object: object})
		job.close()
		job.close()
		if ctx.Err() == nil {
			t.Fatal("unpublished call retained context")
		}
		deadline := time.Now().Add(time.Second)
		for {
			p.mu.Lock()
			_, retained := p.objects[object]
			p.mu.Unlock()
			if !retained {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("late result retained descriptor")
			}
			time.Sleep(time.Millisecond)
		}
		if _, err := fd.Stat(); !errors.Is(err, os.ErrClosed) {
			t.Fatal("unpublished FD", err)
		}
	}
}
func TestFileSystemWorkerAdmissionAndPublicationRollback(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, _, root := fileOwner(t, FileSystemConfig{MaxWorkers: 1})
		if err := os.WriteFile(filepath.Join(root, "data.txt"), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
		cfg := component.DefaultConfig()
		cfg.Core.Backend = backend
		cfg.MaxResources = 1
		e, err := component.NewEngine(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		data, err := os.ReadFile("testdata/filesystem-api.wasm")
		if err != nil {
			t.Fatal(err)
		}
		m, err := e.Compile(nil, data)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		in := instantiate(t, m, p.Imports())
		api := fileAPI{t, in}
		dir := api.root()
		api.fail("open-at", "insufficient-memory", dir, component.Value{}, component.Value{Text: "data.txt"}, component.Value{}, component.Value{Bits: fsRead})
		p.mu.Lock()
		descriptors := len(p.descriptors)
		p.mu.Unlock()
		if descriptors != 1 {
			t.Fatal("failed publication retained handle", descriptors)
		}
		api.drop(dir)
		in = instantiate(t, m, p.Imports())
		api = fileAPI{t, in}
		dir = api.root()
		if err := p.begin(); err != nil {
			t.Fatal(err)
		}
		f, _ := in.Function("fs", "[method]descriptor.stat")
		_, err = f.Call(nil, []component.Value{dir}, component.CallOptions{})
		p.end()
		if !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("worker quota", err)
		}
		api.drop(dir)
	})
}
