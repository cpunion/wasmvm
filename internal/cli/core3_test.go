package cli

import (
	"github.com/cpunion/wasmvm/pkg/wasm"
	"os"
	"strings"
	"testing"
)

func TestCore3CLIProfileAndInspection(t *testing.T) {
	data, err := os.ReadFile("../../pkg/vm/testdata/core3.wasm")
	if err != nil {
		t.Fatal(err)
	}
	path := writeModule(t, data)
	for _, command := range []string{"validate", "inspect"} {
		code, out, diag := invoke(t, nil, command, "--profile", "core3", path)
		if code != 0 || diag != "" || jsonObject(t, out)["profile"] != "core3" || jsonObject(t, out)["features"] != float64(wasm.Core3Features) {
			t.Fatal(code, out, diag)
		}
	}
	// The authored combined module needs a host import; the pure EH/cast module
	// verifies actual CLI execution with only the profile selector.
	data, err = os.ReadFile("../../testdata/exception-handling/cast/script.0.wasm")
	if err != nil {
		t.Fatal(err)
	}
	code, out, diag := invoke(t, nil, "run", "--profile", "core3", "--export", "test", writeModule(t, data))
	if code != 0 || diag != "" || !strings.Contains(out, "0x00000001") || jsonObject(t, out)["profile"] != "core3" {
		t.Fatal(code, out, diag)
	}
}
