package validate

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/decode"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPinnedComponentTypeAndABIValidation(t *testing.T) {
	root := "../../../artifacts/components/binary-corpus"
	for _, name := range []string{"abi", "core-modules", "defined-types"} {
		data, err := os.ReadFile(filepath.Join(root, name+".json"))
		if err != nil {
			t.Fatal(err)
		}
		var script struct {
			Commands []struct {
				Type, Filename string
				Line           uint32
			}
		}
		if err := json.Unmarshal(data, &script); err != nil {
			t.Fatal(err)
		}
		for _, command := range script.Commands {
			input, err := os.ReadFile(filepath.Join(root, command.Filename))
			if err != nil {
				t.Fatal(err)
			}
			definition, err := decode.Decode(input, decode.DefaultLimits())
			if err != nil {
				t.Fatalf("%s line %d decode: %v", name, command.Line, err)
			}
			checked, err := Check(definition, DefaultLimits())
			if command.Type == "assert_invalid" {
				if checked != nil || !errors.Is(err, wasm.ErrInvalid) {
					t.Errorf("%s line %d expected invalid: %v", name, command.Line, err)
				}
			} else if checked == nil || err != nil {
				t.Errorf("%s line %d expected valid: %v", name, command.Line, err)
			}
		}
	}
}
