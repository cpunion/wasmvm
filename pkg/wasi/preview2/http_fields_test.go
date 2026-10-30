package preview2

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func httpFieldTestLimits() httpFieldLimits {
	return httpFieldLimits{count: 8, bytes: 256, forbidden: map[string]bool{"x-denied": true}}
}
func TestHTTPFieldsBytesCaseOrderAndCopies(t *testing.T) {
	h, err := newHTTPFields(httpFieldTestLimits(), []httpField{{"X-First", "\t\x80"}, {"x-first", ""}, {"Second", "two"}}, false)
	if err != nil {
		t.Fatal(err)
	}
	values := h.values("X-FIRST")
	if len(values) != 2 || values[0] != "\t\x80" || values[1] != "" {
		t.Fatal(values)
	}
	if !h.has("second") || h.has("missing") || h.has("bad name") || len(h.values("bad name")) != 0 {
		t.Fatal("invalid lookup semantics")
	}
	entries := h.entries()
	if len(entries) != 3 || entries[0].name != "X-First" || entries[1].name != "x-first" {
		t.Fatal(entries)
	}
	entries[0].name = "mutated"
	values[0] = "mutated"
	if h.entries()[0].name != "X-First" || h.values("x-first")[0] != "\t\x80" {
		t.Fatal("aliased snapshot")
	}
	if err := h.mutate("x-FIRST", []string{"replacement", ""}, false); err != nil {
		t.Fatal(err)
	}
	if got := h.entries(); len(got) != 3 || got[0].name != "Second" || got[1].name != "x-FIRST" || got[2].value != "" {
		t.Fatal(got)
	}
	if err := h.mutate("x-First", []string{"append"}, true); err != nil {
		t.Fatal(err)
	}
	clone := h.clone()
	if err := clone.mutate("X-FIRST", nil, false); err != nil {
		t.Fatal(err)
	}
	if clone.has("x-first") || !h.has("x-first") {
		t.Fatal("clone shares mutation")
	}
	if err := h.mutate("absent", nil, false); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", "bad name", "a:b", "\x00", "é"} {
		if h.has(name) || len(h.values(name)) != 0 {
			t.Fatal(name)
		}
		if err := h.mutate(name, nil, false); !errors.Is(err, httpHeaderSyntax) {
			t.Fatal(name, err)
		}
	}
}
func TestHTTPFieldsImmutableForbiddenLimitsAndAtomicFailure(t *testing.T) {
	limits := httpFieldTestLimits()
	immutable, err := newHTTPFields(limits, []httpField{{"x-denied", "secret"}, {"allowed", "yes"}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if immutable.has("x-denied") || !immutable.has("ALLOWED") {
		t.Fatal(immutable.entries())
	}
	if err := immutable.mutate("bad name", nil, false); !errors.Is(err, httpHeaderImmutable) {
		t.Fatal(err)
	}
	clone := immutable.clone()
	if err := clone.mutate("new", []string{"value"}, true); err != nil {
		t.Fatal(err)
	}
	for _, fields := range [][]httpField{{{"bad name", "ok"}}, {{"ok", "bad\r\nvalue"}}, {{"ok", "\x7f"}}} {
		if _, err := newHTTPFields(limits, fields, false); !errors.Is(err, httpHeaderSyntax) {
			t.Fatal(fields, err)
		}
	}
	if _, err := newHTTPFields(limits, []httpField{{"X-Denied", "no"}}, false); !errors.Is(err, httpHeaderForbidden) {
		t.Fatal(err)
	}
	if _, err := newHTTPFields(httpFieldLimits{}, nil, false); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := newHTTPFields(limits, make([]httpField, 9), false); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := newHTTPFields(limits, []httpField{{"a", strings.Repeat("x", 256)}}, false); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	h, err := newHTTPFields(httpFieldLimits{count: 2, bytes: 8, forbidden: limits.forbidden}, []httpField{{"a", "1"}}, false)
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
		if err := h.mutate(c.name, c.values, c.append); !errors.Is(err, c.want) {
			t.Fatal(c, err)
		}
		if h.entries()[0] != (httpField{"a", "1"}) {
			t.Fatal("failed mutation changed fields")
		}
	}
	if h.has(strings.Repeat("a", 9)) || len(h.values(strings.Repeat("a", 9))) != 0 {
		t.Fatal("unbounded lookup")
	}
	if err := clone.mutate("x-Denied", []string{"no"}, false); !errors.Is(err, httpHeaderForbidden) {
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
				if err := h.mutate("x", []string{"value"}, false); err != nil {
					t.Error(err)
				}
				_ = h.values("X")
				_ = h.entries()
				_ = h.clone()
			}
		}()
	}
	wg.Wait()
}
