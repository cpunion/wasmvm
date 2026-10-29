package cli

import (
	"github.com/cpunion/wasmvm/pkg/wasm"
	"os"
	"strings"
	"testing"
)

func TestThreadsCLI(t *testing.T) {
	data, err := os.ReadFile("../../pkg/vm/testdata/atomic-fence.wasm")
	if err != nil {
		t.Fatal(err)
	}
	path := writeModule(t, data)
	if code, _, _ := invoke(t, nil, "validate", path); code != 1 {
		t.Fatal("Threads enabled by default")
	}
	for _, command := range []string{"validate", "inspect", "run"} {
		args := []string{command, "--threads"}
		if command == "run" {
			args = append(args, "--export", "fence")
		}
		args = append(args, path)
		code, out, diag := invoke(t, nil, args...)
		if code != 0 || diag != "" || jsonObject(t, out)["features"] != float64(wasm.Threads) {
			t.Fatal(code, out, diag)
		}
	}
	data, err = os.ReadFile("../../pkg/vm/testdata/threads64.wasm")
	if err != nil {
		t.Fatal(err)
	}
	code, out, diag := invoke(t, nil, "inspect", "--profile", "core3", "--threads", writeModule(t, data))
	if code != 0 || diag != "" || !strings.Contains(out, `"shared": true`) {
		t.Fatal(code, out, diag)
	}
}
