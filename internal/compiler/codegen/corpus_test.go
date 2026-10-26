package codegen

import (
	"encoding/json"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCompleteCore2NativeCompilation(t *testing.T) {
	scripts, modules, functions := 0, 0, 0
	var codeBytes, fastBytes, helpers uint64
	err := filepath.WalkDir("../../../artifacts/core2/fixtures", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "script.json" {
			return nil
		}
		scripts++
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var script struct {
			Commands []struct{ Type, Filename string }
		}
		if err := json.Unmarshal(data, &script); err != nil {
			return err
		}
		for _, cmd := range script.Commands {
			if cmd.Filename == "" || cmd.Type == "assert_malformed" || cmd.Type == "assert_invalid" {
				continue
			}
			name := filepath.Join(filepath.Dir(path), cmd.Filename)
			data, err := os.ReadFile(name)
			if err != nil {
				return err
			}
			m, err := decode.Decode(data, decode.DefaultLimits())
			if err != nil {
				t.Fatal(name, err)
			}
			checked, err := validate.Check(m, validate.DefaultLimits())
			if err != nil {
				t.Fatal(name, err)
			}
			lowered, err := lower.Compile(checked, lower.DefaultLimits())
			if err != nil {
				t.Fatal(name, err)
			}
			compiled, err := slots.Compile(lowered, slots.DefaultLimits())
			if err != nil {
				t.Fatal(name, err)
			}
			if compiled.IR != lowered || len(compiled.Functions) != len(lowered.Functions) {
				t.Fatal("module identity")
			}
			config := Config{Unmetered: true}
			if len(m.Memories) != 0 {
				config.MemoryMinimumPages = m.Memories[0].Min
			}
			for _, imported := range m.Imports {
				if imported.Kind == 2 {
					config.MemoryMinimumPages = imported.Memory.Min
				}
			}
			for _, function := range compiled.Functions {
				native, err := CompileFunction(runtime.GOARCH, function, DefaultLimits())
				if err != nil {
					t.Fatal(name, err)
				}
				fast, err := CompileFunctionWithConfig(runtime.GOARCH, function, DefaultLimits(), config)
				if err != nil {
					t.Fatal(name, err)
				}
				fastBytes += uint64(len(fast.Entry.Code))
				functions++
				codeBytes += uint64(len(native.Entry.Code))
				helpers += uint64(native.Helpers)
			}
			modules++
		}
		return nil
	})
	if err != nil {
		t.Fatal("prepare pinned corpus:", err)
	}
	if scripts != 148 || modules != 1716 || functions != 6236 {
		t.Fatal(scripts, modules, functions)
	}
	t.Logf("Unmetered tier: %d native bytes; both tiers compiled all functions.", fastBytes)
	t.Logf("Compiled %d Core 2.0 functions in %d modules: %d native bytes, %d explicit helper sites; compilation coverage is separate from execution conformance.", functions, modules, codeBytes, helpers)
}
