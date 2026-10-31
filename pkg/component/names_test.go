package component

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCompatibleHostAndExportNames(t *testing.T) {
	items := map[string]int{
		"test:pkg/api@1.2.9": 1, "test:pkg/api@1.10.0+z": 2,
		"test:pkg/api@1.10.0+a": 3, "test:pkg/api@2.0.0": 4,
		"test:pkg/api@0.2.1": 5, "test:pkg/api@0.3.0": 6,
		"test:pkg/api@1.99.0-rc.1": 7, "test:pkg/api@1.0.0+bad..metadata": 8,
		"test:pkg/api@0.0.1": 9, "plain": 10,
	}
	for _, row := range []struct {
		name  string
		want  int
		found bool
	}{
		{"plain", 10, true}, {"absent", 0, false},
		{"test:pkg/api@1.2.9", 1, true},
		{"test:pkg/api@1.0.0", 3, true},
		{"test:pkg/api@1.999.0", 3, true},
		{"test:pkg/api@0.2.0", 5, true},
		{"test:pkg/api@0.2.99", 5, true},
		{"test:pkg/api@0.4.0", 0, false},
		{"test:pkg/api@0.0.2", 0, false},
		{"test:pkg/api@1.0.0-rc.1", 0, false},
		{"test:pkg/other@1.0.0", 0, false},
		{"test:pkg/api@1.0", 0, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			for range 20 {
				got, found, err := lookupName(items, row.name, nil)
				if err != nil || found != row.found || got != row.want {
					t.Fatal(got, found, err)
				}
			}
		})
	}
	charged := uint64(0)
	charge := func(n uint64) error { charged += n; return nil }
	if _, found, err := lookupName(items, "test:pkg/api@1.0.0", charge); err != nil || !found || charged == 0 {
		t.Fatal(found, charged, err)
	}
	if _, found, err := lookupName(items, "test:pkg/api@1.0.0", func(uint64) error { return wasm.ErrLimit }); !errors.Is(err, wasm.ErrLimit) || found {
		t.Fatal(found, err)
	}
	if _, found, err := lookupName(items, "plain", func(uint64) error { t.Fatal("exact lookup scanned aliases"); return nil }); err != nil || !found {
		t.Fatal(found, err)
	}
}
