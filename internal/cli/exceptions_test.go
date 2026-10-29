package cli

import (
	"os"
	"strings"
	"testing"
)

func TestExceptionCLI(t *testing.T) {
	data, err := os.ReadFile("../../testdata/exception-handling/cast/script.0.wasm")
	if err != nil {
		t.Fatal(err)
	}
	path := writeModule(t, data)
	if code, _, _ := invoke(t, nil, "validate", path); code != 1 {
		t.Fatal(code)
	}
	for _, command := range []string{"validate", "inspect", "run"} {
		args := []string{command, "--exception-handling", "--gc"}
		if command == "run" {
			args = append(args, "--export", "test")
		}
		args = append(args, path)
		code, out, diag := invoke(t, nil, args...)
		if code != 0 || diag != "" {
			t.Fatal(code, out, diag)
		}
	}
	tagData, err := os.ReadFile("../../testdata/exception-handling/tag/script.0.wasm")
	if err != nil {
		t.Fatal(err)
	}
	code, out, diag := invoke(t, nil, "inspect", "--exception-handling", writeModule(t, tagData))
	if code != 0 || diag != "" || !strings.Contains(out, `"kind": "tag"`) {
		t.Fatal(code, out, diag)
	}
}
