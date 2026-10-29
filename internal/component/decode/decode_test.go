package decode

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

var preamble = []byte("\x00asm\x0d\x00\x01\x00")

func leb(n uint32) []byte {
	var b []byte
	for n >= 128 {
		b = append(b, byte(n)|128)
		n >>= 7
	}
	return append(b, byte(n))
}
func section(id byte, payload []byte) []byte {
	b := append([]byte{id}, leb(uint32(len(payload)))...)
	return append(b, payload...)
}
func component(parts ...[]byte) []byte {
	b := append([]byte(nil), preamble...)
	for _, p := range parts {
		b = append(b, p...)
	}
	return b
}

func TestOrderedCompositionAndSnapshot(t *testing.T) {
	input, err := os.ReadFile("testdata/composition.wasm")
	if err != nil {
		t.Fatal(err)
	}
	m, err := Decode(input, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	counts := map[byte]int{}
	for _, d := range m.Definitions {
		counts[d.Kind]++
	}
	if counts[1] != 1 || counts[2] != 1 || counts[4] != 1 || counts[8] != 1 || counts[11] != 1 {
		t.Fatal(counts)
	}
	var child *model.Binary
	for _, d := range m.Definitions {
		if d.Component != nil {
			child = d.Component
		}
	}
	if child == nil || len(child.Definitions) != 3 {
		t.Fatal(child)
	}
	core := m.Definitions[0].Bytes
	before := append([]byte(nil), core...)
	clear(input)
	if !bytes.Equal(core, before) || m.Definitions[0].CoreModule.Exports[0].Name != "add" {
		t.Fatal("decoder retained caller bytes")
	}
	// Component sections repeat and do not follow the Core section ordering rule.
	m, err = Decode(component(section(7, []byte{1, 0x7f}), section(7, []byte{1, 0x79})), DefaultLimits())
	if err != nil || len(m.Definitions) != 2 {
		t.Fatal(m, err)
	}
}

func TestPinnedBinaryGrammar(t *testing.T) {
	root := "../../../artifacts/components/binary-corpus"
	b, err := os.ReadFile(filepath.Join(root, "binary.json"))
	if err != nil {
		t.Fatal("prepare pinned component binary corpus", err)
	}
	var script struct {
		Commands []struct {
			Type, Filename, ModuleType string
			Line                       int
		}
	}
	if err := json.Unmarshal(b, &script); err != nil || len(script.Commands) != 132 {
		t.Fatal("complete binary script", len(script.Commands), err)
	}
	passed := 0
	for _, cmd := range script.Commands {
		input, err := os.ReadFile(filepath.Join(root, cmd.Filename))
		if err != nil {
			t.Fatal(err)
		}
		_, err = Decode(input, DefaultLimits())
		if cmd.Type == "assert_malformed" {
			if !errors.Is(err, wasm.ErrMalformed) {
				t.Errorf("line %d %s: malformed assertion returned %v", cmd.Line, cmd.Filename, err)
				continue
			}
		} else if err != nil {
			t.Errorf("line %d %s: %s failed to decode: %v", cmd.Line, cmd.Filename, cmd.Type, err)
			continue
		}
		passed++
	}
	if passed != 132 {
		t.Fatalf("executed %d/132 binary grammar commands", passed)
	}
}

func TestLimitsAndMalformedBoundaries(t *testing.T) {
	for _, input := range [][]byte{nil, preamble[:7], component(section(13, nil)), component(section(7, []byte{0xff, 0xff, 0xff, 0xff, 0x0f})), component(section(7, []byte{1, 0x70, 0xff, 0x7e}))} {
		m, err := Decode(input, DefaultLimits())
		if !errors.Is(err, wasm.ErrMalformed) || m != nil {
			t.Fatal(input, m, err)
		}
	}
	for _, limits := range []Limits{{}, {MaxBytes: 100, MaxNodes: 100}, {MaxBytes: 100, MaxNodes: 0, MaxNesting: 4}} {
		input := component(section(7, []byte{1, 0x7f}))
		m, err := Decode(input, limits)
		if !errors.Is(err, wasm.ErrLimit) || m != nil {
			t.Fatal(limits, m, err)
		}
	}
	nested := component(section(4, component(section(4, component()))))
	l := DefaultLimits()
	l.MaxNesting = 2
	if _, err := Decode(nested, l); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	l = DefaultLimits()
	l.MaxNameBytes = 1
	if _, err := Decode(component(section(0, []byte{2, 'a', 'b'})), l); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}

func FuzzDecode(f *testing.F) {
	f.Add(append([]byte(nil), preamble...))
	f.Add(component(section(7, []byte{1, 0x7f})))
	f.Fuzz(func(t *testing.T, b []byte) {
		l := Limits{MaxBytes: 4096, MaxNodes: 512, MaxNameBytes: 128, MaxNesting: 16}
		m, err := Decode(b, l)
		if err == nil && m == nil {
			t.Fatal("successful nil decode")
		}
	})
}
