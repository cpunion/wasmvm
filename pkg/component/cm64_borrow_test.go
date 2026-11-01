package component

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCM64LocalBorrowDirect(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, err := NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "cm64-borrow-direct"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			in, err := m.Instantiate(nil, Imports{}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			makeResource, _ := in.Function("make-owned")
			borrow, _ := in.Function("borrow-rep")
			consume, _ := in.Function("consume-owned")
			for _, rep := range []uint64{0, 1, 1 << 32, 1 << 63, 0xfedcba9876543210, ^uint64(0)} {
				v, err := makeResource.Call(nil, []Value{{Bits: rep}}, CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				for range 2 {
					result, err := borrow.Call(nil, []Value{*v.Value}, CallOptions{})
					if err != nil || result.Value == nil || result.Value.Bits != rep {
						t.Fatal(result, err)
					}
				}
				result, err := consume.Call(nil, []Value{*v.Value}, CallOptions{})
				if err != nil || result.Value.Bits != rep {
					t.Fatal(result, err)
				}
			}
		})
	}
}

func TestCM64BorrowAggregatesAndComposition(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, address := range []string{"32", "64"} {
			for _, shape := range []string{"local", "composed"} {
				t.Run(backend+"/"+shape+address, func(t *testing.T) {
					config := DefaultConfig()
					config.Core.Backend = backend
					config.MaxResources = 128
					e, err := NewEngine(config)
					if err != nil {
						t.Fatal(err)
					}
					defer e.Close()
					m, err := e.Compile(nil, fixture(t, "cm64-borrow-"+shape+address))
					if err != nil {
						t.Fatal(err)
					}
					defer m.Close()
					in, err := m.Instantiate(nil, Imports{}, CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					defer in.Close()
					invoke := func(name string, args ...Value) Value {
						t.Helper()
						f, err := in.Function(name)
						if err != nil {
							t.Fatal(err)
						}
						result, err := f.Call(nil, args, CallOptions{Metered: true, Fuel: 1000000})
						if err != nil || result.Value == nil {
							t.Fatal(name, result, err)
						}
						return *result.Value
					}
					for _, rep := range []uint64{0, 1, 1 << 32, 1 << 63, 0xfedcba9876543210, ^uint64(0)} {
						if shape == "composed" {
							if got := invoke("run", Value{Bits: rep}); got.Bits != rep+1 {
								t.Fatal("cross-component rep", got, rep)
							}
							if got := invoke("cancel", Value{Bits: rep}); got.Bits != 42 {
								t.Fatal("cancelled borrow", got)
							}
							continue
						}
						a := invoke("make", Value{Bits: rep})
						other := rep ^ 0x1234567800000001
						b := invoke("make", Value{Bits: other})
						if got := invoke("direct", a); got.Bits != rep {
							t.Fatal("direct", got)
						}
						if got := invoke("record", Value{Elements: []Value{{Bits: 4}, a, {Bits: 13}}}); got.Bits != rep+17 {
							t.Fatal("record", got)
						}
						for _, v := range []struct {
							value Value
							want  uint64
						}{
							{Value{Case: 0}, 0}, {Value{Case: 1, Elements: []Value{a}}, rep},
							{Value{Case: 2, Elements: []Value{{Bits: 0xfedcba98}}}, 0xfedcba98},
						} {
							if got := invoke("variant", v.value); got.Bits != v.want {
								t.Fatal("variant", got, v.want)
							}
						}
						if got := invoke("list", Value{Elements: []Value{a, b}}); got.Bits != rep^other {
							t.Fatal("list", got)
						}
						values := make([]Value, 17)
						for i := range values {
							values[i] = a
						}
						values[16] = b
						if got := invoke("spill", Value{Elements: values}); got.Bits != rep^other {
							t.Fatal("spilled borrows", got)
						}
						if got := invoke("async", a); got.Bits != rep {
							t.Fatal("suspended borrow", got)
						}
						for _, v := range []Value{a, b} {
							if _, err := in.DropResource(nil, v.Resource, CallOptions{}); err != nil {
								t.Fatal("borrow retained lender", err)
							}
						}
					}
				})
			}
		}
	}
}

func TestCM64BorrowFailureReleasesLender(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, address := range []string{"32", "64"} {
			for _, kind := range []string{"align", "range", "trap"} {
				t.Run(backend+"/"+kind+address, func(t *testing.T) {
					config := DefaultConfig()
					config.Core.Backend = backend
					e, err := NewEngine(config)
					if err != nil {
						t.Fatal(err)
					}
					defer e.Close()
					name := "cm64-borrow-bad-" + kind + address
					if kind == "trap" {
						name = "cm64-borrow-local" + address
					}
					m, err := e.Compile(nil, fixture(t, name))
					if err != nil {
						t.Fatal(err)
					}
					defer m.Close()
					in, err := m.Instantiate(nil, Imports{}, CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					defer in.Close()
					makeResource, _ := in.Function("make")
					entry := "list"
					if kind == "trap" {
						entry = "trap"
					}
					read, _ := in.Function(entry)
					v, err := makeResource.Call(nil, []Value{{Bits: 0xfedcba9876543210}}, CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					args := []Value{{Elements: []Value{*v.Value, *v.Value}}}
					if kind == "trap" {
						args = []Value{*v.Value}
					}
					_, err = read.Call(nil, args, CallOptions{})
					if !errors.Is(err, wasm.ErrTrap) {
						t.Fatal("wide-borrow failure did not trap", err)
					}
					table, typ, id, _, _, err := canonical.ReferenceInfo(v.Value.Resource)
					if err != nil {
						t.Fatal(err)
					}
					if err := table.CheckOwn(typ, id); err != nil {
						t.Fatal("failed lower retained a lend", err)
					}
				})
			}
		}
	}
}

func TestCM64BorrowFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/cm64-borrows.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Revision  string                       `json:"componentModelCorrection"`
		Producer  string                       `json:"producer"`
		Generator string                       `json:"generatorSHA256"`
		Spec      string                       `json:"specManifestSHA256"`
		Sources   map[string]string            `json:"sourceSHA256"`
		Fixtures  map[string]map[string]string `json:"fixtures"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Revision != "fed10b0b823b82553dd8363f16a640816f5eac65" || pins.Producer != "wasm-tools 1.258.0 parse" || len(pins.Fixtures) != 9 {
		t.Fatal("changed cm64 producer", pins)
	}
	check := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		sum := sha256.Sum256(data)
		if err != nil || hex.EncodeToString(sum[:]) != want {
			t.Fatal(path, err)
		}
	}
	check("../../scripts/build_cm64_borrow_fixtures.py", pins.Generator)
	check("../../testdata/component-cm64/manifest.json", pins.Spec)
	for name, want := range pins.Sources {
		check("testdata/"+name, want)
	}
	for name, files := range pins.Fixtures {
		for ext, want := range files {
			check("testdata/"+name+"."+ext, want)
		}
	}
	data, err = os.ReadFile("../../testdata/component-cm64/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Revision string
		Archive  string            `json:"archiveSHA256"`
		Files    map[string]string `json:"sha256"`
	}
	if err := json.Unmarshal(data, &spec); err != nil || spec.Revision != pins.Revision || len(spec.Files) != 4 {
		t.Fatal(spec, err)
	}
	path := "../../testdata/component-cm64/inputs.tar.gz"
	check(path, spec.Archive)
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	compressed, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	archive := tar.NewReader(compressed)
	seen := map[string]bool{}
	for {
		member, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(io.LimitReader(archive, 1<<20))
		sum := sha256.Sum256(data)
		if err != nil || seen[member.Name] || hex.EncodeToString(sum[:]) != spec.Files[member.Name] {
			t.Fatal("changed specification asset", member.Name, err)
		}
		seen[member.Name] = true
	}
	if len(seen) != len(spec.Files) {
		t.Fatal("missing pinned specification assets")
	}
}
