package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCoreExtensionsCLI(t *testing.T) {
	for _, row := range []struct {
		path, flag, export string
		feature            wasm.Features
		args               []string
	}{
		{"wide-arithmetic/wide-arithmetic/script.0.wasm", "--wide-arithmetic", "i64.mul_wide_u", wasm.WideArithmetic, []string{"i64:2", "i64:3"}},
		{"compact-import-section/imports-compact/script.15.wasm", "--compact-imports", "", wasm.CompactImports, nil},
		{"custom-page-sizes/custom-page-sizes/script.0.wasm", "--custom-page-sizes", "", wasm.CustomPageSizes, nil},
	} {
		data, err := os.ReadFile(filepath.Join("../../testdata", row.path))
		if err != nil {
			t.Fatal(err)
		}
		path := writeModule(t, data)
		if code, _, _ := invoke(t, nil, "validate", path); code != 1 {
			t.Fatal("extension enabled by default", row.flag)
		}
		for _, command := range []string{"validate", "inspect", "run"} {
			if command == "run" && row.export == "" {
				continue
			}
			args := []string{command, row.flag}
			if command == "run" {
				args = append(args, "--export", row.export)
			}
			args = append(args, path)
			if command == "run" {
				args = append(args, row.args...)
			}
			code, out, diag := invoke(t, nil, args...)
			if code != 0 || diag != "" || jsonObject(t, out)["features"] != float64(row.feature) {
				t.Fatal(row.flag, command, code, out, diag)
			}
		}
	}
}
