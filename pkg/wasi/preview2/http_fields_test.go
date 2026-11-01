package preview2

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func httpFieldTestLimits() httpFieldLimits {
	return httpFieldLimits{Count: 8, Bytes: 256, Forbidden: map[string]bool{"x-denied": true}}
}
func TestHTTPFieldsBytesCaseOrderAndCopies(t *testing.T) {
	h, err := newHTTPFields(httpFieldTestLimits(), []httpField{{Name: "X-First", Value: "\t\x80"}, {Name: "x-first", Value: ""}, {Name: "Second", Value: "two"}}, false)
	if err != nil {
		t.Fatal(err)
	}
	values := h.Values("X-FIRST")
	if len(values) != 2 || values[0] != "\t\x80" || values[1] != "" {
		t.Fatal(values)
	}
	if !h.Has("second") || h.Has("missing") || h.Has("bad name") || len(h.Values("bad name")) != 0 {
		t.Fatal("invalid lookup semantics")
	}
	entries := h.Entries()
	if len(entries) != 3 || entries[0].Name != "X-First" || entries[1].Name != "x-first" {
		t.Fatal(entries)
	}
	entries[0].Name = "mutated"
	values[0] = "mutated"
	if h.Entries()[0].Name != "X-First" || h.Values("x-first")[0] != "\t\x80" {
		t.Fatal("aliased snapshot")
	}
	if err := h.Mutate("x-FIRST", []string{"replacement", ""}, false); err != nil {
		t.Fatal(err)
	}
	if got := h.Entries(); len(got) != 3 || got[0].Name != "Second" || got[1].Name != "x-FIRST" || got[2].Value != "" {
		t.Fatal(got)
	}
	if err := h.Mutate("x-First", []string{"append"}, true); err != nil {
		t.Fatal(err)
	}
	clone := h.Clone()
	if err := clone.Mutate("X-FIRST", nil, false); err != nil {
		t.Fatal(err)
	}
	if clone.Has("x-first") || !h.Has("x-first") {
		t.Fatal("clone shares mutation")
	}
	if err := h.Mutate("absent", nil, false); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", "bad name", "a:b", "\x00", "é"} {
		if h.Has(name) || len(h.Values(name)) != 0 {
			t.Fatal(name)
		}
		if err := h.Mutate(name, nil, false); !errors.Is(err, httpHeaderSyntax) {
			t.Fatal(name, err)
		}
	}
}
func TestHTTPFieldsImmutableForbiddenLimitsAndAtomicFailure(t *testing.T) {
	limits := httpFieldTestLimits()
	immutable, err := newHTTPFields(limits, []httpField{{Name: "x-denied", Value: "secret"}, {Name: "allowed", Value: "yes"}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if immutable.Has("x-denied") || !immutable.Has("ALLOWED") {
		t.Fatal(immutable.Entries())
	}
	if err := immutable.Mutate("bad name", nil, false); !errors.Is(err, httpHeaderImmutable) {
		t.Fatal(err)
	}
	clone := immutable.Clone()
	if err := clone.Mutate("new", []string{"value"}, true); err != nil {
		t.Fatal(err)
	}
	for _, fields := range [][]httpField{{{Name: "bad name", Value: "ok"}}, {{Name: "ok", Value: "bad\r\nvalue"}}, {{Name: "ok", Value: "\x7f"}}} {
		if _, err := newHTTPFields(limits, fields, false); !errors.Is(err, httpHeaderSyntax) {
			t.Fatal(fields, err)
		}
	}
	if _, err := newHTTPFields(limits, []httpField{{Name: "X-Denied", Value: "no"}}, false); !errors.Is(err, httpHeaderForbidden) {
		t.Fatal(err)
	}
	if _, err := newHTTPFields(httpFieldLimits{}, nil, false); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := newHTTPFields(limits, make([]httpField, 9), false); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := newHTTPFields(limits, []httpField{{Name: "a", Value: strings.Repeat("x", 256)}}, false); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	h, err := newHTTPFields(httpFieldLimits{Count: 2, Bytes: 8, Forbidden: limits.Forbidden}, []httpField{{Name: "a", Value: "1"}}, false)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		values []string
		append bool
		want   error
	}{
		{"a", []string{"ok", "bad\n"}, false, httpHeaderSyntax},
		{"a", []string{"12345678"}, false, wasm.ErrLimit},
		{"a", []string{"1", "2"}, true, wasm.ErrLimit},
		{"a", []string{"1", "2", "3"}, false, wasm.ErrLimit},
		{strings.Repeat("a", 9), nil, false, wasm.ErrLimit},
	}
	for _, c := range cases {
		if err := h.Mutate(c.name, c.values, c.append); !errors.Is(err, c.want) {
			t.Fatal(c, err)
		}
		if h.Entries()[0] != (httpField{Name: "a", Value: "1"}) {
			t.Fatal("failed mutation changed fields")
		}
	}
	if h.Has(strings.Repeat("a", 9)) || len(h.Values(strings.Repeat("a", 9))) != 0 {
		t.Fatal("unbounded lookup")
	}
	if err := clone.Mutate("x-Denied", []string{"no"}, false); !errors.Is(err, httpHeaderForbidden) {
		t.Fatal(err)
	}
	for _, e := range []httpHeaderError{httpHeaderSyntax, httpHeaderForbidden, httpHeaderImmutable} {
		if e.Error() == "" {
			t.Fatal(e)
		}
	}
}
func TestHTTPFieldsConcurrentSnapshotsAndMutations(t *testing.T) {
	h, err := newHTTPFields(httpFieldTestLimits(), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if err := h.Mutate("x", []string{"value"}, false); err != nil {
					t.Error(err)
				}
				_ = h.Values("X")
				_ = h.Entries()
				_ = h.Clone()
			}
		}()
	}
	wg.Wait()
}
