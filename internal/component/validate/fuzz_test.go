package validate

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/component/decode"
)

// Exercise declaration prefixes through the real binary decoder. Encountering
// a runtime definition ends the prefix; this is not a whole-component checker.
func FuzzTypeDeclarations(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte("\x00asm\x0d\x00\x01\x00"),
		[]byte("\x00asm\x0d\x00\x01\x00\x07\x05\x01\x40\x00\x01\x00"),
		[]byte("\x00asm\x0d\x00\x01\x00\x03\x04\x01\x60\x00\x00"),
		[]byte("\x00asm\x0d\x00\x01\x00\x07\x03\x01\x42\x00"),
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		definition, err := decode.Decode(input, decode.Limits{MaxBytes: 1 << 16, MaxNodes: 1 << 14, MaxNameBytes: 1024, MaxNesting: 16})
		if err != nil {
			return
		}
		s, err := NewTypePool(Limits{MaxNodes: 1 << 14, MaxDepth: 32}).Scope(nil, true)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range definition.Definitions {
			switch d.Kind {
			case 3:
				err = s.DefineCore(d.CoreType, definition.CoreTypes)
			case 6:
				err = s.aliasDeclaration(d.Alias)
			case 7:
				_, err = s.Define(d.Type, nil)
			case 10:
				_, err = s.Declare(d.Import, false)
			default:
				return
			}
			if err != nil {
				return
			}
		}
	})
}
