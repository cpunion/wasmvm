//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type fsAPI struct {
	t  *testing.T
	in *component.Instance
}

func (a fsAPI) call(instance, name string, args ...component.Value) component.Value {
	a.t.Helper()
	f, err := a.in.Function(instance, name)
	if err != nil {
		a.t.Fatal(err)
	}
	result, err := f.Call(nil, args, component.CallOptions{})
	if err != nil {
		a.t.Fatal(name, err)
	}
	if result.Value == nil {
		return component.Value{}
	}
	return *result.Value
}
func (a fsAPI) ok(name string, args ...component.Value) component.Value {
	a.t.Helper()
	value := a.call("types", "[method]descriptor."+name, args...)
	if value.Case != 0 {
		a.t.Fatalf("%s failed: %+v", name, value)
	}
	if len(value.Elements) == 0 {
		return component.Value{}
	}
	return value.Elements[0]
}
func (a fsAPI) fail(name string, code FileSystemErrorCode, args ...component.Value) {
	a.t.Helper()
	value := a.call("types", "[method]descriptor."+name, args...)
	if value.Case != 1 || len(value.Elements) != 1 || value.Elements[0].Bits != uint64(code) {
		a.t.Fatalf("%s: %+v; expected %d", name, value, code)
	}
}
func (a fsAPI) root() component.Value {
	a.t.Helper()
	value := a.call("preopens", "get-directories")
	if len(value.Elements) != 1 || value.Elements[0].Elements[1].Text != "/sandbox" {
		a.t.Fatal(value)
	}
	return value.Elements[0].Elements[0]
}
func (a fsAPI) open(root component.Value, path string, openFlags, flags uint64) component.Value {
	a.t.Helper()
	return a.ok("open-at", root, component.Value{}, component.Value{Text: path}, component.Value{Bits: openFlags}, component.Value{Bits: flags})
}
func (a fsAPI) drop(value component.Value) {
	a.t.Helper()
	if _, err := a.in.DropResource(nil, value.Resource, component.CallOptions{}); err != nil {
		a.t.Fatal(err)
	}
}
func TestFileSystemAPIAndBoundaries(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			p, process, rootPath := fsOwner(t, FileSystemConfig{MaxReadBytes: 3, MaxWriteBytes: 3, MaxPathBytes: 64})
			a := fsAPI{t, fsInstance(t, p, backend, "filesystem-api")}
			root := a.root()
			file := a.open(root, "data", 1, fsRead|fsWrite)
			// A mutation capability is not the open-flags directory constraint.
			// Language adapters may request it before knowing the object kind.
			writable := a.open(root, "data", 0, fsWrite|fsMutate)
			if flags := a.ok("get-flags", writable); flags.Bits&fsMutate != 0 {
				t.Fatal("regular file advertised directory mutation", flags)
			}
			a.fail("open-at", FSNotDirectory, root, component.Value{}, component.Value{Text: "data"}, component.Value{Bits: 2}, component.Value{Bits: fsRead})
			a.drop(writable)
			n := a.ok("write", file, component.ByteList([]byte("abcdef")), component.Value{})
			if n.Bits != 3 {
				t.Fatal(n)
			}
			bytes := a.ok("read", file, component.Value{Bits: 99}, component.Value{})
			text, err := bytes.Elements[0].ByteString()
			if err != nil || text != "abc" {
				t.Fatal(bytes, err)
			}
			bytes = a.ok("read", file, component.Value{Bits: 3}, component.Value{Bits: 3})
			if bytes.Elements[1].Bits != 1 {
				t.Fatal(bytes)
			}
			a.fail("read", FSOverflow, file, component.Value{Bits: 1}, component.Value{Bits: 1 << 63})
			a.fail("read", FSOverflow, file, component.Value{Bits: 3}, component.Value{Bits: 1<<63 - 2})
			a.fail("write", FSOverflow, file, component.ByteList([]byte("abc")), component.Value{Bits: 1<<63 - 2})
			a.fail("set-size", FSOverflow, file, component.Value{Bits: 1 << 63})
			a.fail("set-size", FSIsDirectory, root, component.Value{Bits: 1})
			a.fail("advise", FSOverflow, file, component.Value{Bits: 1 << 63}, component.Value{}, component.Value{})
			a.fail("advise", FSOverflow, file, component.Value{Bits: 1<<63 - 2}, component.Value{Bits: 3}, component.Value{})
			a.fail("read-directory", FSNotDirectory, file)
			a.fail("create-directory-at", FSNotDirectory, file, component.Value{Text: "sub"})
			a.fail("read-via-stream", FSIsDirectory, root, component.Value{})
			a.fail("write-via-stream", FSOverflow, file, component.Value{Bits: 1 << 63})
			readOnly := a.open(root, "data", 0, fsRead)
			for _, method := range []string{"sync", "sync-data"} {
				a.ok(method, readOnly)
			}
			a.fail("write", FSNotPermitted, readOnly, component.ByteList([]byte("x")), component.Value{})
			a.fail("write-via-stream", FSNotPermitted, readOnly, component.Value{})
			a.fail("set-size", FSNotPermitted, readOnly, component.Value{Bits: 1})
			writeOnly := a.open(root, "data", 0, fsWrite)
			a.fail("read", FSNotPermitted, writeOnly, component.Value{Bits: 1}, component.Value{})
			a.fail("read-via-stream", FSNotPermitted, writeOnly, component.Value{})
			for _, path := range []string{"../outside", "/etc/passwd", "sub/../../outside"} {
				a.fail("open-at", FSNotPermitted, root, component.Value{Bits: 1}, component.Value{Text: path}, component.Value{}, component.Value{Bits: fsRead})
			}
			for _, path := range []string{"abs", "escape"} {
				target := "/etc/passwd"
				if path == "escape" {
					target = "../outside"
				}
				if err := os.Symlink(target, filepath.Join(rootPath, path)); err != nil {
					t.Fatal(err)
				}
				a.fail("stat-at", FSNotPermitted, root, component.Value{Bits: 1}, component.Value{Text: path})
			}
			a.fail("readlink-at", FSNotPermitted, root, component.Value{Text: "abs"})
			a.fail("symlink-at", FSNotPermitted, root, component.Value{Text: "/absolute"}, component.Value{Text: "new"})
			a.fail("open-at", FSNoEntry, root, component.Value{}, component.Value{Text: "missing"}, component.Value{}, component.Value{Bits: fsRead})
			a.fail("open-at", FSNameTooLong, root, component.Value{}, component.Value{Text: strings.Repeat("x", 65)}, component.Value{}, component.Value{Bits: fsRead})
			a.fail("open-at", FSInvalid, root, component.Value{}, component.Value{Text: "x\x00y"}, component.Value{}, component.Value{Bits: fsRead})
			stamp := component.Value{Case: 2, Elements: []component.Value{{Elements: []component.Value{{Bits: 123}, {Bits: 456789012}}}}}
			a.ok("set-times", file, stamp, stamp)
			stat := a.ok("stat", file)
			if stat.Elements[3].Elements[0].Elements[0].Bits != 123 || stat.Elements[3].Elements[0].Elements[1].Bits != 456789012 {
				t.Fatal(stat)
			}
			invalid := component.Value{Case: 2, Elements: []component.Value{{Elements: []component.Value{{Bits: 123}, {Bits: 1000000000}}}}}
			a.fail("set-times", FSInvalid, file, invalid, stamp)
			overflow := component.Value{Case: 2, Elements: []component.Value{{Elements: []component.Value{{Bits: ^uint64(0)}, {}}}}}
			a.fail("set-times-at", FSOverflow, root, component.Value{}, component.Value{Text: "data"}, overflow, stamp)
			a.ok("set-times", file, component.Value{}, component.Value{})
			// Streams retain their own lease after the descriptor is dropped.
			stream := a.ok("read-via-stream", file, component.Value{})
			a.drop(file)
			got := a.call("streams", "[method]input-stream.blocking-read", stream, component.Value{Bits: 3})
			text, err = got.Elements[0].ByteString()
			if err != nil || text != "abc" {
				t.Fatal(got, err)
			}
			a.drop(stream)
			process.Close()
			if err := process.WaitClosed(nil); err != nil {
				t.Fatal(err)
			}
			// Filesystem ownership is independent of the IO owner, including failures
			// to publish a new stream into an already closed stream domain.
			f, err := a.in.Function("types", "[method]descriptor.read-via-stream")
			if err != nil {
				t.Fatal(err)
			}
			result, err := f.Call(nil, []component.Value{readOnly, {}}, component.CallOptions{})
			if err != nil || result.Value.Case != 1 {
				t.Fatal(result, err)
			}
			a.ok("get-type", readOnly)
			a.drop(readOnly)
			a.drop(writeOnly)
			a.drop(root)
			p.Close()
			if err := p.WaitClosed(nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestFileSystemReadOnlyDirectory(t *testing.T) {
	rootPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootPath, "data"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	p, _, _ := fsOwner(t, FileSystemConfig{Preopens: []FileSystemPreopen{{GuestPath: "/sandbox", HostPath: rootPath, ReadOnly: true}}})
	a := fsAPI{t, fsInstance(t, p, "interpreter", "filesystem-api")}
	root := a.root()
	for _, flags := range []uint64{fsWrite, fsMutate} {
		a.fail("open-at", FSReadOnly, root, component.Value{}, component.Value{Text: "data"}, component.Value{}, component.Value{Bits: flags})
	}
	for _, flags := range []uint64{1, 8} {
		a.fail("open-at", FSReadOnly, root, component.Value{}, component.Value{Text: "data"}, component.Value{Bits: flags}, component.Value{Bits: fsRead})
	}
	for _, method := range []string{"create-directory-at", "remove-directory-at", "unlink-file-at"} {
		a.fail(method, FSReadOnly, root, component.Value{Text: "sub"})
	}
	a.fail("set-times", FSReadOnly, root, component.Value{}, component.Value{})
	a.fail("set-times-at", FSReadOnly, root, component.Value{}, component.Value{Text: "data"}, component.Value{}, component.Value{})
	a.fail("symlink-at", FSReadOnly, root, component.Value{Text: "data"}, component.Value{Text: "new"})
	a.fail("rename-at", FSReadOnly, root, component.Value{Text: "data"}, root, component.Value{Text: "new"})
	a.fail("link-at", FSReadOnly, root, component.Value{}, component.Value{Text: "data"}, root, component.Value{Text: "new"})
	dir := a.open(root, ".", 2, 0)
	a.fail("read-directory", FSNotPermitted, dir)
	a.ok("sync", root)
	a.ok("sync-data", root)
	other := a.open(root, "data", 0, fsRead)
	if got := a.ok("get-type", other); got.Bits != 6 {
		t.Fatal(got)
	}
	if same := a.call("types", "[method]descriptor.is-same-object", root, other); same.Bits != 0 {
		t.Fatal(same)
	}
	a.drop(other)
	a.drop(dir)
	a.drop(root)
}
func TestFileSystemOwnerAndResourceLimits(t *testing.T) {
	p, _, _ := fsOwner(t, FileSystemConfig{MaxDescriptors: 1, MaxDirectoryStreams: 1, MaxDirectoryEntries: 1})
	a := fsAPI{t, fsInstance(t, p, "interpreter", "filesystem-api")}
	root := a.root()
	a.fail("open-at", FSInsufficientMemory, root, component.Value{}, component.Value{Text: "bench.bin"}, component.Value{}, component.Value{Bits: fsRead})
	entries := a.ok("read-directory", root)
	a.fail("read-directory", FSInsufficientMemory, root)
	got := a.call("types", "[method]directory-entry-stream.read-directory-entry", entries)
	if got.Case != 0 || got.Elements[0].Case != 1 {
		t.Fatal(got)
	}
	got = a.call("types", "[method]directory-entry-stream.read-directory-entry", entries)
	if got.Case != 0 || got.Elements[0].Case != 0 {
		t.Fatal(got)
	}
	a.drop(entries)
	p.Close()
	if err := p.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	a.fail("get-flags", FSBadDescriptor, root)
	f, err := a.in.Function("types", "[method]descriptor.is-same-object")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Call(nil, []component.Value{root, root}, component.CallOptions{}); !errors.Is(err, os.ErrClosed) {
		t.Fatal(err)
	}
	a.drop(root)
}
func TestFileSystemConstructorAndErrorMapping(t *testing.T) {
	if _, err := NewFileSystem(nil, FileSystemConfig{}); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	if got := (*FileSystem)(nil).Imports(); len(got.Instances) != 0 {
		t.Fatal(got)
	}
	var zero FileSystem
	zero.Close()
	if err := zero.WaitClosed(nil); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	process, err := NewIO(IOConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer process.Close()
	for _, config := range []FileSystemConfig{{MaxReadBytes: 1 << 28}, {MaxWriteBytes: 1 << 28}, {MaxDirectoryEntries: 1 << 25}, {Preopens: []FileSystemPreopen{{}, {}}, MaxDescriptors: 1}} {
		if _, err := NewFileSystem(process, config); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(config, err)
		}
	}
	for _, config := range []FileSystemConfig{{Preopens: []FileSystemPreopen{{GuestPath: "\xff"}}}, {Preopens: []FileSystemPreopen{{GuestPath: "x\x00"}}}, {MaxPathBytes: 1, Preopens: []FileSystemPreopen{{GuestPath: "long"}}}} {
		if _, err := NewFileSystem(process, config); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(config, err)
		}
	}
	if _, err := NewFileSystem(process, FileSystemConfig{Preopens: []FileSystemPreopen{{HostPath: filepath.Join(t.TempDir(), "missing")}}}); err == nil {
		t.Fatal("accepted missing directory")
	}
	process.Close()
	if _, err := NewFileSystem(process, FileSystemConfig{}); err == nil {
		t.Fatal("accepted closed IO owner")
	}
	for _, entry := range fileSystemErrnos {
		if code := fsError(entry.err); code != entry.code {
			t.Fatal(entry, code)
		}
	}
	cause := errors.New("unknown")
	if code := fsError(cause); code != FSIo {
		t.Fatal(code)
	}
	typed := &FileSystemError{Code: FSAccess, Cause: cause}
	if typed.Error() != cause.Error() || !errors.Is(typed, cause) || fsError(typed) != FSAccess {
		t.Fatal(typed)
	}
	if (&FileSystemError{Code: FSAccess}).Error() == "" {
		t.Fatal("empty error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := zero.WaitClosed(ctx); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
}
