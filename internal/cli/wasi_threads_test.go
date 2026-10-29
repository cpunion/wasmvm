//go:build linux || darwin

package cli

import (
	"strings"
	"testing"
)

func TestLegacyWASIThreadsCLI(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		code, out, diag := invoke(t, nil, "run", "--wasi-threads", "--backend", backend, "../../pkg/wasi/threads/testdata/group.wasm")
		if code != 0 || out != "" || diag != "" {
			t.Fatal(code, out, diag)
		}
		code, _, diag = invoke(t, nil, "run", "--wasi-threads", "--backend", backend, "--fuel", "1099511627776", "../../artifacts/wasi-threads/official/wasi_threads_exit_nonmain_busy.wasm")
		if code != 99 || diag != "" {
			t.Fatal("worker exit", code, diag)
		}
		code, _, diag = invoke(t, nil, "run", "--wasi-threads", "--backend", backend, "--max-threads", "1", "../../pkg/wasi/threads/testdata/group.wasm")
		if code != 1 || !strings.Contains(diag, "unreachable") {
			t.Fatal("bounded spawn", code, diag)
		}
		code, _, diag = invoke(t, nil, "run", "--wasi-threads", "--backend", backend, "testdata/wasi-stdio.wasm")
		if code != 1 || !strings.Contains(diag, "shared memory") {
			t.Fatal("invalid thread module", code, diag)
		}
	}
	for _, args := range [][]string{
		{"run", "--wasi-threads", "--max-threads", "0", "testdata/wasi-stdio.wasm"},
		{"run", "--wasi-threads", "--max-threads", "536870912", "testdata/wasi-stdio.wasm"},
		{"run", "--wasi-threads", "--export", "spam", "testdata/wasi-stdio.wasm"},
		{"run", "--max-threads", "1", "--wasi", "testdata/wasi-stdio.wasm"},
	} {
		code, _, diag := invoke(t, nil, args...)
		if code != 2 || diag == "" {
			t.Fatal(args, code, diag)
		}
	}
}
