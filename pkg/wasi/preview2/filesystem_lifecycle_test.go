//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

func TestFileSystemContractRejection(t *testing.T) {
	p, _, _ := fsOwner(t, FileSystemConfig{})
	engine, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	for _, name := range []string{"read-width", "read-result", "stat-field", "error-label", "timestamp-label", "advice-label", "directory-entry", "stream-resource", "preopen-order"} {
		module, err := engine.Compile(nil, fixture(t, "filesystem-contract-"+name))
		if err != nil {
			t.Fatal(name, err)
		}
		in, err := module.Instantiate(nil, fsImports(t, p), component.CallOptions{})
		if in != nil {
			in.Close()
		}
		module.Close()
		if !errors.Is(err, vm.ErrType) {
			t.Fatal(name, err)
		}
	}
}
func TestFileSystemPublicationRollbackAndExpiredCaller(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	p, _, _ := fsOwner(t, FileSystemConfig{MaxDescriptors: 2, Preopens: []FileSystemPreopen{{GuestPath: "/first", HostPath: first}, {GuestPath: "/second", HostPath: second}}})
	engine, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	module, err := engine.Compile(nil, fixture(t, "filesystem-api"))
	if err != nil {
		t.Fatal(err)
	}
	defer module.Close()
	imports := fsImports(t, p)
	original := imports.Instances[FileSystemPreopensInterface].Functions["get-directories"]
	var saved *component.Caller
	imports.Instances[FileSystemPreopensInterface].Functions["get-directories"] = func(c *component.Caller, args []component.Value) (*component.Value, error) {
		saved = c
		return original(c, args)
	}
	in, err := module.Instantiate(nil, imports, component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	a := fsAPI{t, in}
	list := a.call("preopens", "get-directories")
	a.drop(list.Elements[1].Elements[0])
	get, err := in.Function("preopens", "get-directories")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := get.Call(nil, nil, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	p.mu.Lock()
	count := len(p.descriptors)
	p.mu.Unlock()
	if count != 1 {
		t.Fatal("partial preopen list leaked", count)
	}
	a.drop(list.Elements[0].Elements[0])
	if _, err := p.publish(saved, p.preopens[0].descriptor); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	p.mu.Lock()
	count = len(p.descriptors)
	p.mu.Unlock()
	if count != 0 {
		t.Fatal("expired Caller leaked descriptor", count)
	}
	if _, err := p.descriptor(saved, list.Elements[0].Elements[0].Resource); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	p.next = 1 << 32
	if _, err := p.publish(nil, p.preopens[0].descriptor); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}
func TestFileSystemLimitsAndInvalidHostNames(t *testing.T) {
	p, _, rootPath := fsOwner(t, FileSystemConfig{MaxDirectoryEntries: 1})
	a := fsAPI{t, fsInstance(t, p, "interpreter", "filesystem-api")}
	root := a.root()
	if err := os.WriteFile(filepath.Join(rootPath, "second"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	a.fail("read-directory", FSInsufficientMemory, root)
	if err := os.Remove(filepath.Join(rootPath, "second")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(rootPath, "bench.bin")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rootPath, "\xff"), nil, 0600); err != nil {
		if !errors.Is(err, unix.EILSEQ) {
			t.Fatal(err)
		}
		t.Log("host filesystem rejects non-UTF-8 names before enumeration")
	} else {
		a.fail("read-directory", FSIllegalByteSequence, root)
	}
	if err := os.Symlink("\xff", filepath.Join(rootPath, "link")); err != nil {
		t.Fatal(err)
	}
	a.fail("readlink-at", FSIllegalByteSequence, root, component.Value{Text: "link"})
	a.fail("stat-at", FSInvalid, root, component.Value{}, component.Value{Text: "a\x00"})
	a.fail("create-directory-at", FSInvalid, root, component.Value{Text: "a\x00"})
	a.fail("readlink-at", FSNoEntry, root, component.Value{Text: "missing"})
	a.fail("set-times-at", FSInvalid, root, component.Value{}, component.Value{Text: "a\x00"}, component.Value{}, component.Value{})
	a.drop(root)
}
func TestFileSystemErrorDowncastAndGuestRevocation(t *testing.T) {
	p, process, _ := fsOwner(t, FileSystemConfig{})
	engine, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	module, err := engine.Compile(nil, fixture(t, "filesystem-api"))
	if err != nil {
		t.Fatal(err)
	}
	defer module.Close()
	imports := fsImports(t, p)
	original := imports.Instances[FileSystemPreopensInterface].Functions["get-directories"]
	imports.Instances[FileSystemPreopensInterface].Functions["get-directories"] = func(c *component.Caller, args []component.Value) (*component.Value, error) {
		for _, cause := range []error{errors.New("unrelated host failure"), &FileSystemError{Code: FSAccess}} {
			value, err := process.errors.NewError(c, cause)
			if err != nil {
				return nil, err
			}
			code, err := p.errorCode(c, []component.Value{value})
			if err != nil {
				return nil, err
			}
			_, typed := cause.(*FileSystemError)
			if typed && (code.Case != 1 || code.Elements[0].Bits != uint64(FSAccess)) || !typed && code.Case != 0 {
				return nil, errors.New("incorrect filesystem error provenance")
			}
			if err = c.DropResource(value.Resource); err != nil {
				return nil, err
			}
		}
		return original(c, args)
	}
	read := imports.Instances[StreamsInterface].Functions["[method]input-stream.read"]
	imports.Instances[StreamsInterface].Functions["[method]input-stream.read"] = func(c *component.Caller, args []component.Value) (*component.Value, error) {
		p.Close()
		if err := p.WaitClosed(c.Context()); err != nil {
			return nil, err
		}
		return read(c, args)
	}
	in, err := module.Instantiate(nil, imports, component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	f, err := in.Function("error-probe")
	if err != nil {
		t.Fatal(err)
	}
	result, err := f.Call(nil, nil, component.CallOptions{})
	if err != nil || result.Value.Bits != 1 {
		t.Fatal(result, err)
	}
}
func TestFileSystemCloseWaitsForOperationsAndCancellation(t *testing.T) {
	p, _, _ := fsOwner(t, FileSystemConfig{})
	if err := p.begin(); err != nil {
		t.Fatal(err)
	}
	p.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.WaitClosed(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := p.begin(); !errors.Is(err, os.ErrClosed) {
		t.Fatal(err)
	}
	if err := p.check(context.Background()); !errors.Is(err, os.ErrClosed) {
		t.Fatal(err)
	}
	p.active.Done()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := p.WaitClosed(ctx); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}
func TestFileSystemValueAndStreamFailures(t *testing.T) {
	if _, err := fsStatValue(host.Stat{TimeOverflow: true}); fsError(err) != FSOverflow {
		t.Fatal(err)
	}
	for want, mode := range []uint32{0, unix.S_IFBLK, unix.S_IFCHR, unix.S_IFDIR, unix.S_IFIFO, unix.S_IFLNK, unix.S_IFREG, unix.S_IFSOCK} {
		if kind := fsType(mode); kind != uint32(want) {
			t.Fatal(mode, kind, want)
		}
	}
	p, _, root := fsOwner(t, FileSystemConfig{})
	source := fsSource(t, p, filepath.Join(root, "bench.bin"), true)
	if err := source.backing.file.Close(); err != nil {
		t.Fatal(err)
	}
	if n, err := source.Read(context.Background(), make([]byte, 1)); n != 0 || fsError(err) != FSBadDescriptor {
		t.Fatal(n, err)
	}
	source.Close()
	source = fsSource(t, p, filepath.Join(root, "bench.bin"), false)
	source.backing.file.Close()
	if n, err := source.Write(context.Background(), []byte("x")); n != 0 || fsError(err) != FSBadDescriptor {
		t.Fatal(n, err)
	}
	source.appendMode = true
	if n, err := source.Write(context.Background(), []byte("x")); n != 0 || fsError(err) != FSBadDescriptor {
		t.Fatal(n, err)
	}
}

func TestFileSystemFailedAppendRestorationStopsOtherWriters(t *testing.T) {
	p, _, root := fsOwner(t, FileSystemConfig{})
	a := fsAPI{t, fsInstance(t, p, "interpreter", "filesystem-api")}
	directory := a.root()
	file := a.open(directory, "bench.bin", 0, fsRead|fsWrite)
	p.mu.Lock()
	var backing *fsBacking
	for _, d := range p.descriptors {
		if d.backing.directory == nil {
			backing = d.backing
		}
	}
	p.mu.Unlock()
	// Emulate the retained failure state of an unsuccessful O_APPEND restore.
	// Positional writes must fail before touching a potentially appended fd.
	backing.ioMu.Lock()
	backing.writeFailure = fsCode(FSIo)
	backing.ioMu.Unlock()
	a.fail("write", FSIo, file, component.ByteList([]byte("BAD")), component.Value{})
	if data, err := os.ReadFile(filepath.Join(root, "bench.bin")); err != nil || string(data) != "0123456789abcdef" {
		t.Fatal(string(data), err)
	}
	source := fsSource(t, p, filepath.Join(root, "bench.bin"), false)
	source.backing.writeFailure = fsCode(FSIo)
	if n, err := source.Write(context.Background(), []byte("BAD")); n != 0 || fsError(err) != FSIo {
		t.Fatal(n, err)
	}
	a.drop(file)
	a.drop(directory)
}
