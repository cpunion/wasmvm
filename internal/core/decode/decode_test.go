package decode

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

var magic = []byte{0, 97, 115, 109, 1, 0, 0, 0}

func uleb(n uint32) []byte {
	var b [5]byte
	size := binary.PutUvarint(b[:], uint64(n))
	return append([]byte(nil), b[:size]...)
}
func section(id byte, content ...byte) []byte {
	out := []byte{id}
	out = append(out, uleb(uint32(len(content)))...)
	return append(out, content...)
}
func fixture(sections ...[]byte) []byte {
	out := append([]byte(nil), magic...)
	for _, s := range sections {
		out = append(out, s...)
	}
	return out
}
func function(body ...byte) []byte {
	body = append([]byte{0}, body...)
	code := append([]byte{1}, uleb(uint32(len(body)))...)
	code = append(code, body...)
	return fixture(section(1, 1, 0x60, 0, 0), section(3, 1, 0), section(10, code...))
}

func TestOwnedMetadata(t *testing.T) {
	input := fixture(section(0, 1, 'x', 0xaa, 0xbb), section(1, 1, 0x60, 1, 0x7f, 1, 0x7e), section(11, 1, 1, 2, 0xcc, 0xdd))
	m, err := Decode(input, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	for i := range input {
		input[i] = 0
	}
	if m.Custom[0].Name != "x" || !bytes.Equal(m.Custom[0].Data, []byte{0xaa, 0xbb}) || !bytes.Equal(m.Data[0].Bytes, []byte{0xcc, 0xdd}) {
		t.Fatal("module aliases input")
	}
	if m.Types[0].Params[0] != wasm.I32 || m.Types[0].Results[0] != wasm.I64 {
		t.Fatal("function type")
	}
	if cap(m.Data[0].Bytes) != len(m.Data[0].Bytes) {
		t.Fatal("data view crosses segment boundary")
	}
	// Mutation of one decode must not affect another decode of the same input.
	other, err := Decode(fixture(section(11, 1, 1, 2, 0xcc, 0xdd)), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	m.Data[0].Bytes[0] = 0
	if other.Data[0].Bytes[0] != 0xcc {
		t.Fatal("shared decode state")
	}
}

func TestLimits(t *testing.T) {
	for _, tc := range []struct {
		name   string
		input  []byte
		change func(*Limits)
	}{
		{"bytes", magic, func(l *Limits) { l.MaxBytes = 7 }},
		{"nodes", function(0x41, 0, 0x1a, 0x0b), func(l *Limits) { l.MaxNodes = 3 }},
		{"names", fixture(section(0, 1, 'a')), func(l *Limits) { l.MaxNameBytes = 0 }},
		{"nesting", function(2, 0x40, 0x0b, 0x0b), func(l *Limits) { l.MaxNesting = 0 }},
		{"locals", fixture(section(3, 1, 0), section(10, 1, 4, 1, 2, 0x7f, 0x0b)), func(l *Limits) { l.MaxLocals = 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := DefaultLimits()
			tc.change(&l)
			m, err := Decode(tc.input, l)
			if m != nil || !errors.Is(err, wasm.ErrLimit) {
				t.Fatalf("module=%v error=%v", m, err)
			}
		})
	}
	if _, err := Decode(magic, Limits{MaxBytes: 8}); err != nil {
		t.Fatal("empty module with zero structural quotas:", err)
	}
}

func TestMalformedBoundaries(t *testing.T) {
	for name, input := range map[string][]byte{
		"magic":                 []byte("bad!\x01\x00\x00\x00"),
		"version":               []byte("\x00asm\x02\x00\x00\x00"),
		"unknown section":       fixture(section(13)),
		"duplicate":             fixture(section(1, 0), section(1, 0)),
		"order":                 fixture(section(3, 0), section(1, 0)),
		"count without code":    fixture(section(3, 1, 0)),
		"code without count":    fixture(section(10, 1, 2, 0, 0x0b)),
		"bad import":            fixture(section(2, 1, 0, 0, 4)),
		"bad export":            fixture(section(7, 1, 0, 4, 0)),
		"bad value":             fixture(section(1, 1, 0x60, 1, 0x40, 0)),
		"bad table":             fixture(section(4, 1, 0x7f, 0, 0)),
		"bad limits":            fixture(section(5, 1, 2, 0)),
		"bad mutable":           fixture(section(6, 1, 0x7f, 2, 0x41, 0, 0x0b)),
		"bad data":              fixture(section(11, 1, 3)),
		"bad element":           fixture(section(9, 1, 8)),
		"data count":            fixture(section(12, 1)),
		"missing data count":    function(0xfc, 9, 0, 0x0b),
		"unknown opcode":        function(0xff, 0x0b),
		"unknown fc":            function(0xfc, 18, 0x0b),
		"unknown simd":          function(0xfd, 0x9a, 1, 0x0b),
		"large simd":            function(0xfd, 0x80, 2, 0x0b),
		"else":                  function(5, 0x0b),
		"duplicate else":        function(4, 0x40, 5, 5, 0x0b, 0x0b),
		"padded negative block": function(2, 0xff, 0x7f, 0x0b, 0x0b),
		"reserved memory":       function(0x3f, 1, 0x0b),
		"extra body":            function(0x0b, 1),
		"vector claim":          fixture(section(1, 0xff, 0xff, 0xff, 0xff, 0x0f)),
		"locals overflow":       fixture(section(3, 1, 0), section(10, 1, 11, 2, 0xff, 0xff, 0xff, 0xff, 0x0f, 0x7f, 1, 0x7f, 0x0b)),
	} {
		t.Run(name, func(t *testing.T) {
			l := DefaultLimits()
			l.MaxLocals = ^uint32(0)
			m, err := Decode(input, l)
			if m != nil || !errors.Is(err, wasm.ErrMalformed) {
				t.Fatalf("module=%v error=%v", m, err)
			}
		})
	}
	input := function(0x42, 0x80, 1, 0x1a, 0x0b)
	for cut := 0; cut < len(input); cut++ {
		_, err := Decode(input[:cut], DefaultLimits())
		// Prefixes ending at optional section boundaries remain valid modules.
		if err != nil && !errors.Is(err, wasm.ErrMalformed) {
			t.Fatalf("cut %d: %v", cut, err)
		}
	}
}

func TestInstructionBitsAndOffsets(t *testing.T) {
	m, err := Decode(function(0x41, 0x7f, 0x42, 0x7f, 0x43, 1, 0, 0xc0, 0x7f, 0x44, 0, 0, 0, 0, 0, 0, 0, 0x80, 0x0b), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	code := m.Functions[0].Body
	if code[0].Bits != 0xffffffff || code[1].Bits != ^uint64(0) || code[2].Bits != 0x7fc00001 || code[3].Bits != 1<<63 {
		t.Fatalf("lost bits: %v", code)
	}
	for i := 1; i < len(code); i++ {
		if code[i].Offset <= code[i-1].Offset {
			t.Fatal("non-increasing byte offsets")
		}
	}
	m, err = Decode(function(2, 0x80, 0, 0x0b, 0x0b), DefaultLimits())
	if err != nil || m.Functions[0].Body[0].Block != 0 {
		t.Fatalf("padded positive block type: %v", err)
	}
}

func TestCompleteCore2BinaryInputs(t *testing.T) {
	root := "../../../artifacts/core2/fixtures"
	if _, err := os.Stat(root); err != nil {
		t.Fatal("prepare required pinned inputs with python3 scripts/core2_corpus.py:", err)
	}
	files, malformed, scripts := 0, 0, 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
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
			Commands []struct {
				Type, Filename string
				ModuleType     string `json:"module_type"`
			}
		}
		if err = json.Unmarshal(data, &script); err != nil {
			return err
		}
		for _, command := range script.Commands {
			if command.Filename == "" || command.ModuleType == "text" {
				continue
			}
			input, err := os.ReadFile(filepath.Join(filepath.Dir(path), command.Filename))
			if err != nil {
				return err
			}
			files++
			m, err := Decode(input, DefaultLimits())
			if command.Type == "assert_malformed" {
				malformed++
				if m != nil || !errors.Is(err, wasm.ErrMalformed) {
					t.Errorf("%s/%s: expected malformed, got %v", filepath.Base(filepath.Dir(path)), command.Filename, err)
				}
			} else if err != nil {
				t.Errorf("%s/%s (%s): %v", filepath.Base(filepath.Dir(path)), command.Filename, command.Type, err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if scripts != 148 || files != 4581 || malformed != 719 {
		t.Fatalf("inventory changed: scripts=%d binaries=%d malformed=%d", scripts, files, malformed)
	}
	t.Logf("Decoded %d binary inputs from %d scripts; rejected %d malformed binaries. No validation or execution assertions run.", files, scripts, malformed)
}

func FuzzDecode(f *testing.F) {
	for _, input := range [][]byte{magic, function(0x41, 0, 0x1a, 0x0b), fixture(section(9, 0)), fixture(section(11, 1, 1, 1, 0))} {
		f.Add(input)
	}
	input, err := os.ReadFile("../../../pkg/vm/testdata/multi-memory.wasm")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(input)
	f.Fuzz(func(t *testing.T, input []byte) {
		l := DefaultLimits()
		l.MaxBytes = 1 << 16
		l.MaxNodes = 4096
		l.MaxLocals = 4096
		l.MaxNesting = 64
		for _, features := range []wasm.Features{0, wasm.SupportedFeatures} {
			m, err := DecodeWithFeatures(input, l, features)
			if err != nil {
				if m != nil || (!errors.Is(err, wasm.ErrMalformed) && !errors.Is(err, wasm.ErrLimit)) {
					t.Fatalf("bad failure: %v", err)
				}
			}
		}
	})
}

func BenchmarkDecode(b *testing.B) {
	for _, tc := range []struct {
		name  string
		input []byte
	}{
		{"empty", magic},
		{"scalar", function(0x41, 1, 0x41, 2, 0x6a, 0x1a, 0x0b)},
		{"vector", function(0xfd, 12, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 0x1a, 0x0b)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			limits := DefaultLimits()
			b.SetBytes(int64(len(tc.input)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := Decode(tc.input, limits); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
