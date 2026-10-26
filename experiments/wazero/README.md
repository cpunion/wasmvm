# wazero Interpreter Evaluation

This standalone module evaluates pinned wazero **v1.12.0** using its interpreter, CoreFeaturesV2, and every experimental feature flag exposed by that release. It is an M0 experiment; it does not implement the wasmvm API or certify full WebAssembly conformance.

## Run

```sh
cd experiments/wazero
go test -v -count=1 -timeout=30s ./...
go test -race -count=1 -timeout=30s ./...
go test -count=1 -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Run the isolated comparison from the repository root with `experiments/wazero/scripts/run.sh go upstream` or `experiments/wazero/scripts/run.sh llgo upstream`. The LLGo path requires the memory-hook fix from [PR #2919](https://github.com/xgo-dev/llgo/pull/2919), pinned in CI. It runs unchanged wazero and must actually pass the tests.

The `interpreter-only` mode and `record-upstream-llgo.sh` preserve the earlier diagnostic with released LLGo v1.0.8. They are not the required CI path or a permanent wazero fork. The old failure assertion intentionally rejects newly successful builds so that the historical expectation cannot masquerade as compatibility verification.

`TestEvaluation` emits one machine-readable `EVALUATION_REPORT` JSON array. The 13 fixtures produce **7 executed cases and 6 precisely matched rejections**. An expected rejection records an unsupported capability; it is not a feature-support pass. Separate tests execute concurrent atomic increments from four instances sharing memory, memory growth, and bounded wait/notify. They emit `CONCURRENCY_REPORT`.

`TestAtomicWaitCancellationObservation` uses a finite one-second guest wait with a 20 ms context deadline and emits `CANCELLATION_REPORT`. The observed delayed return is a known lifecycle limitation, not unsupported syntax. Timing is recorded rather than compared against a brittle scheduler threshold, and the finite wait prevents an indefinite test hang.

The local Go run covered **100% of statements in this module's `evaluate.go`**. This measures the evaluation/reporting code only; it is neither wazero engine coverage nor Wasm opcode/specification coverage. Hash/schema rejection and wrong-result/trap/linkage tests prevent malformed inputs and arbitrary failures from being counted as capability evidence.

## Fixture provenance and regeneration

All WAT files in `testdata` were authored for this project; no third-party fixture source was copied. `manifest.json` records the candidate revision, expected outcomes, execution signatures, and SHA-256 digests of both each WAT source and its generated Wasm.

Regenerate with **wasm-tools 1.258.0** from this directory:

```sh
wasm-tools --version
for source in testdata/*.wat; do
  wasm-tools parse "$source" -o "${source%.wat}.wasm"
  wasm-tools validate --features all "${source%.wat}.wasm"
done
python3 - <<'PY'
from pathlib import Path
import hashlib
import json
manifest_path = Path('testdata/manifest.json')
manifest = json.loads(manifest_path.read_text())
for fixture in manifest['fixtures']:
    for key, digest_key in [('wat', 'watSha256'), ('wasm', 'sha256')]:
        data = (manifest_path.parent / fixture[key]).read_bytes()
        fixture[digest_key] = hashlib.sha256(data).hexdigest()
manifest_path.write_text(json.dumps(manifest, indent=2) + '\n')
PY
go test -v -count=1 ./...
```

Review source, binary, hashes, and expected outcomes together. Regeneration is an explicit development operation; tests do not rewrite hashes or download a floating tool version. Validating a fixture establishes its format independently of wazero's rejection, not broad conformance. The empty component probes only container acceptance; it exercises no WASI 0.2/0.3 interface.

The larger LLGo-produced guest and its exact producer/reference observations are in `../llgo-guest/`.
