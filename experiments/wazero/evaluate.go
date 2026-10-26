// Package evaluation records reproducible interpreter capability observations.
// It is a candidate evaluation harness, not a WebAssembly VM implementation.
package evaluation

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

// Corpus contains project-authored WAT and generated Wasm with a pinned manifest.
//
//go:embed testdata/*.wat testdata/*.wasm testdata/manifest.json
var Corpus embed.FS

// Manifest identifies the candidate, fixture producer, and expected observations.
type Manifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	Origin        string `json:"origin"`
	Candidate     struct {
		Module               string `json:"module"`
		Version              string `json:"version"`
		Commit               string `json:"commit"`
		Backend              string `json:"backend"`
		FeatureConfiguration string `json:"featureConfiguration"`
	} `json:"candidate"`
	Producer struct {
		Name       string `json:"name"`
		Version    string `json:"version"`
		Command    string `json:"command"`
		Validation string `json:"validation"`
	} `json:"producer"`
	Fixtures []Fixture `json:"fixtures"`
}

// Fixture describes one valid input; unsupported is an observed capability gap.
type Fixture struct {
	ID        string      `json:"id"`
	Wasm      string      `json:"wasm"`
	WAT       string      `json:"wat"`
	SHA256    string      `json:"sha256"`
	WATSHA256 string      `json:"watSha256"`
	Features  []string    `json:"features"`
	Expected  Expectation `json:"expected"`
	Calls     []Call      `json:"calls"`
	Stdout    string      `json:"stdout,omitempty"`
	Notes     string      `json:"notes"`
}

type Expectation struct {
	Phase   string `json:"phase"`
	Outcome string `json:"outcome"`
}

type Call struct {
	Export  string   `json:"export"`
	Args    []uint64 `json:"args"`
	Results []uint64 `json:"results,omitempty"`
	Trap    bool     `json:"trap,omitempty"`
}

// Observation separates a correctly recorded gap from implemented support.
type Observation struct {
	ID              string      `json:"id"`
	Phase           string      `json:"phase"`
	Outcome         string      `json:"outcome"`
	Error           string      `json:"error,omitempty"`
	Stdout          string      `json:"stdout,omitempty"`
	Expected        Expectation `json:"expected"`
	MatchesExpected bool        `json:"matchesExpected"`
}

// LoadManifest verifies schema, provenance, safe paths, and every binary hash.
func LoadManifest(source fs.FS) (Manifest, error) {
	var manifest Manifest
	data, err := fs.ReadFile(source, "testdata/manifest.json")
	if err != nil {
		return manifest, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&manifest); err != nil {
		return manifest, err
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return manifest, errors.New("manifest must contain exactly one JSON value")
	}
	if manifest.SchemaVersion != 1 || manifest.Origin == "" || len(manifest.Fixtures) == 0 {
		return manifest, errors.New("manifest requires schemaVersion 1, provenance, and fixtures")
	}
	if manifest.Candidate.Module != "github.com/tetratelabs/wazero" || manifest.Candidate.Version != "v1.12.0" || manifest.Candidate.Commit != "2ab480b55fa408d6b35df97fe32a60d08bd6e201" || manifest.Candidate.Backend != "interpreter" {
		return manifest, errors.New("manifest candidate does not match the pinned interpreter")
	}
	if manifest.Candidate.FeatureConfiguration == "" || manifest.Producer.Name != "wasm-tools" || manifest.Producer.Version != "1.258.0" || manifest.Producer.Command == "" || manifest.Producer.Validation == "" {
		return manifest, errors.New("manifest lacks pinned feature or producer metadata")
	}
	seen := map[string]bool{}
	for _, f := range manifest.Fixtures {
		if f.ID == "" || seen[f.ID] {
			return manifest, fmt.Errorf("missing or duplicate fixture id %q", f.ID)
		}
		seen[f.ID] = true
		if len(f.Features) == 0 || f.Notes == "" {
			return manifest, fmt.Errorf("%s lacks feature scope or notes", f.ID)
		}
		if !safeName(f.Wasm, ".wasm") || !safeName(f.WAT, ".wat") {
			return manifest, fmt.Errorf("%s has unsafe fixture paths", f.ID)
		}
		wat, err := fs.ReadFile(source, "testdata/"+f.WAT)
		if err != nil {
			return manifest, err
		}
		watDigest, watErr := hex.DecodeString(f.WATSHA256)
		watSum := sha256.Sum256(wat)
		if watErr != nil || !bytes.Equal(watDigest, watSum[:]) {
			return manifest, fmt.Errorf("%s WAT SHA-256 mismatch", f.ID)
		}
		binary, readErr := fs.ReadFile(source, "testdata/"+f.Wasm)
		if readErr != nil {
			return manifest, readErr
		}
		digest, decodeErr := hex.DecodeString(f.SHA256)
		sum := sha256.Sum256(binary)
		if decodeErr != nil || !bytes.Equal(digest, sum[:]) {
			return manifest, fmt.Errorf("%s binary SHA-256 mismatch", f.ID)
		}
		if !validExpectation(f.Expected) {
			return manifest, fmt.Errorf("%s has invalid expected phase/outcome", f.ID)
		}
		for _, call := range f.Calls {
			if call.Export == "" || (call.Trap && len(call.Results) != 0) {
				return manifest, fmt.Errorf("%s has an invalid call", f.ID)
			}
		}
	}
	return manifest, nil
}

func safeName(name, suffix string) bool {
	return fs.ValidPath(name) && path.Base(name) == name && strings.HasSuffix(name, suffix)
}

func validExpectation(want Expectation) bool {
	return want == (Expectation{"execute", "supported"}) || want == (Expectation{"compile", "unsupported"}) || want == (Expectation{"instantiate", "unsupported"})
}

func runtimeConfig() wazero.RuntimeConfig {
	features := api.CoreFeaturesV2 | experimental.CoreFeaturesThreads | experimental.CoreFeaturesTailCall | experimental.CoreFeaturesExtendedConst | experimental.CoreFeaturesExceptionHandling | experimental.CoreFeaturesTypedFunctionReferences
	return wazero.NewRuntimeConfigInterpreter().WithCoreFeatures(features).WithCloseOnContextDone(true).WithMemoryLimitPages(64)
}

// Evaluate executes every fixture or records its precise known rejection.
// Every supported result comes from a call, not just successful compilation.
func Evaluate(ctx context.Context, source fs.FS) ([]Observation, error) {
	manifest, err := LoadManifest(source)
	if err != nil {
		return nil, err
	}
	results := make([]Observation, 0, len(manifest.Fixtures))
	for _, fixture := range manifest.Fixtures {
		results = append(results, evaluateFixture(ctx, source, fixture))
	}
	return results, nil
}

func evaluateFixture(ctx context.Context, source fs.FS, fixture Fixture) (result Observation) {
	result = Observation{ID: fixture.ID, Phase: "compile", Outcome: "error", Expected: fixture.Expected}
	defer func() {
		result.MatchesExpected = result.Phase == result.Expected.Phase && result.Outcome == result.Expected.Outcome
	}()
	r := wazero.NewRuntimeWithConfig(ctx, runtimeConfig())
	defer r.Close(context.Background())
	binary, err := fs.ReadFile(source, "testdata/"+fixture.Wasm)
	if err != nil {
		result.Error = err.Error()
		return
	}
	compiled, err := r.CompileModule(ctx, binary)
	if err != nil {
		result.Error = err.Error()
		if knownRejection(fixture.ID, result.Phase, err) {
			result.Outcome = "unsupported"
		}
		return
	}
	result.Phase = "instantiate"
	if err = installImports(ctx, source, r, fixture.ID); err != nil {
		result.Error = err.Error()
		return
	}
	var stdout bytes.Buffer
	module, err := r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName("fixture").WithStartFunctions().WithStdout(&stdout))
	if err != nil {
		result.Error = err.Error()
		if knownRejection(fixture.ID, result.Phase, err) {
			result.Outcome = "unsupported"
		}
		return
	}
	result.Phase = "execute"
	for _, call := range fixture.Calls {
		if err = checkCall(ctx, module, call); err != nil {
			result.Error = err.Error()
			return
		}
	}
	if fixture.ID == "p1" {
		err = checkCall(ctx, module, Call{Export: "_start"})
		var exit *sys.ExitError
		if err != nil && (!errors.As(err, &exit) || exit.ExitCode() != 0) {
			result.Error = err.Error()
			return
		}
		result.Stdout = stdout.String()
		if result.Stdout != fixture.Stdout {
			result.Error = fmt.Sprintf("stdout mismatch: got %q want %q", result.Stdout, fixture.Stdout)
			return
		}
	}
	if len(fixture.Calls) == 0 && fixture.ID != "p1" {
		result.Error = "fixture has no executable assertion"
		return
	}
	result.Outcome = "supported"
	return
}

func installImports(ctx context.Context, source fs.FS, r wazero.Runtime, id string) error {
	switch id {
	case "host-callback":
		_, err := r.NewHostModuleBuilder("host").NewFunctionBuilder().WithFunc(func(v uint32) uint32 { return v + 1 }).Export("increment").Instantiate(ctx)
		return err
	case "p1":
		_, err := wasi_snapshot_preview1.Instantiate(ctx, r)
		return err
	case "thread-worker":
		binary, err := fs.ReadFile(source, "testdata/shared-memory.wasm")
		if err != nil {
			return err
		}
		_, err = r.InstantiateWithConfig(ctx, binary, wazero.NewModuleConfig().WithName("shared"))
		return err
	default:
		return nil
	}
}

func checkCall(ctx context.Context, module api.Module, call Call) error {
	function := module.ExportedFunction(call.Export)
	if function == nil {
		return fmt.Errorf("missing export %q", call.Export)
	}
	got, err := function.Call(ctx, call.Args...)
	if call.Trap {
		if err != nil && strings.Contains(err.Error(), "wasm error: unreachable") {
			return nil
		}
		return fmt.Errorf("%s: expected unreachable trap, got results=%v error=%v", call.Export, got, err)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", call.Export, err)
	}
	if len(got) != len(call.Results) {
		return fmt.Errorf("%s result count: got %v want %v", call.Export, got, call.Results)
	}
	for i := range got {
		if got[i] != call.Results[i] {
			return fmt.Errorf("%s result %d: got %d want %d", call.Export, i, got[i], call.Results[i])
		}
	}
	return nil
}

// Rejections are matched to a validated fixture and a pinned engine version.
// An arbitrary decode or runtime error must never be accepted as a known gap.
func knownRejection(id, phase string, err error) bool {
	if err == nil {
		return false
	}
	want := map[string]string{
		"gc":           "section type: read 0-th type: invalid byte: 0x5f != 0x60",
		"memory64":     "section memory: invalid byte for limits: 0x4 not in (0x00, 0x01, 0x02, 0x03)",
		"multi-memory": "section memory: at most one memory allowed in module, but read 2",
		"relaxed-simd": "invalid function[0] export[\"run\"]: read block: type index out of range: 3581",
		"component":    "invalid version header",
	}
	if phase == "compile" {
		return want[id] != "" && err.Error() == want[id]
	}
	return phase == "instantiate" && id == "wasi-threads-import" && err.Error() == "module[wasi] not instantiated"
}
