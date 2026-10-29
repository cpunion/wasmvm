//go:build linux || darwin

package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"golang.org/x/sys/unix"
)

func TestWASICLI(t *testing.T) {
	root := "../../artifacts/preview1/fixtures"
	for _, backend := range []string{"interpreter", "native"} {
		code, out, diag := invoke(t, nil, "run", "--wasi", "--backend", backend, filepath.Join(root, "assemblyscript/fd_write-to-stdout.wasm"))
		if code != 0 || out != "hello" || diag != "" {
			t.Fatal(code, out, diag)
		}
		code, _, diag = invoke(t, nil, "run", "--wasi", "--backend", backend, filepath.Join(root, "assemblyscript/proc_exit-failure.wasm"))
		if code != 33 || diag != "" {
			t.Fatal("guest exit", code, diag)
		}
		code, _, diag = invoke(t, nil, "run", "--wasi", "--backend", backend, "--dir", "/="+t.TempDir(), "--read-only-dir", "/read="+t.TempDir(), filepath.Join(root, "rust/file_seek_tell.wasm"))
		if code != 0 || diag != "" {
			t.Fatal("preopens", code, diag)
		}
		input := filepath.Join(t.TempDir(), "stdin")
		if err := os.WriteFile(input, []byte("input"), 0600); err != nil {
			t.Fatal(err)
		}
		code, out, diag = invoke(t, nil, "run", "--wasi", "--backend", backend, "--fuel", "10000", "--stdin", input, "testdata/wasi-stdio.wasm")
		if code != 0 || out != "input" || diag != "" {
			t.Fatal("stdin", code, out, diag)
		}
	}
	for _, args := range [][]string{
		{"run", "--env", "K=V", "--export", "_start", "testdata/wasi-stdio.wasm"},
		{"run", "--wasi", "--env", "bad", "testdata/wasi-stdio.wasm"},
		{"run", "--wasi", "--dir", "bad", "testdata/wasi-stdio.wasm"},
		{"run", "--wasi", "--read-only-dir", "=missing", "testdata/wasi-stdio.wasm"},
	} {
		code, _, diag := invoke(t, nil, args...)
		if code != 2 || diag == "" {
			t.Fatal(args, code, diag)
		}
	}
	for _, args := range [][]string{
		{"run", "--wasi", "--dir", "/=/definitely/missing", "testdata/wasi-stdio.wasm"},
		{"run", "--wasi", "--stdin", "/definitely/missing", "testdata/wasi-stdio.wasm"},
		{"run", "--wasi", "--export", "missing", "testdata/wasi-stdio.wasm"},
		{"run", "--wasi", "--fuel", "0", "testdata/wasi-stdio.wasm"},
		{"run", "--wasi", "--backend", "native", "--native-code-bytes", "0", "testdata/wasi-stdio.wasm"},
	} {
		code, _, diag := invoke(t, nil, args...)
		if code != 1 || diag == "" {
			t.Fatal(args, code, diag)
		}
	}
}

func TestWASIPipeCancellationAndFlagRestore(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	before, err := unix.FcntlInt(uintptr(host.Descriptor(w)), unix.F_GETFL, 0)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	var diagnostic bytes.Buffer
	code := Main(ctx, []string{"run", "--wasi", "--export", "spam", "testdata/wasi-stdio.wasm"}, w, &diagnostic)
	if code != 1 || !strings.Contains(diagnostic.String(), "deadline") {
		t.Fatal(code, diagnostic.String())
	}
	after, err := unix.FcntlInt(uintptr(host.Descriptor(w)), unix.F_GETFL, 0)
	if err != nil || (before^after)&(unix.O_NONBLOCK|unix.O_APPEND|unix.O_ASYNC) != 0 {
		t.Fatal("stdio flags changed", before, after, err)
	}
	stream, err := cliStreams("", w, &diagnostic)
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); err != nil {
		t.Fatal("double close", err)
	}
	closed, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	closed.Close()
	if _, err := cliStreams("", closed, &diagnostic); err == nil {
		t.Fatal("closed stdout accepted")
	}
}

func TestWASIStdioAliasesRetainNonblockingUntilCleanup(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	if err := unix.SetNonblock(host.Descriptor(w), false); err != nil {
		t.Fatal(err)
	}
	streams, err := cliStreams("", w, w)
	if err != nil {
		t.Fatal(err)
	}
	defer streams.Close()
	if err := streams.streams[1].Close(); err != nil {
		t.Fatal(err)
	}
	flags, err := unix.FcntlInt(uintptr(host.Descriptor(w)), unix.F_GETFL, 0)
	if err != nil || flags&unix.O_NONBLOCK == 0 {
		t.Fatal("closing stdout made live stderr blocking", flags, err)
	}
	if err := streams.Close(); err != nil {
		t.Fatal(err)
	}
	flags, err = unix.FcntlInt(uintptr(host.Descriptor(w)), unix.F_GETFL, 0)
	if err != nil || flags&unix.O_NONBLOCK != 0 {
		t.Fatal("stdio flags not restored after all streams closed", flags, err)
	}
}
