# LLGo Guest Inventory

This standalone, project-authored Go program probes a retained reference cycle across `runtime.GC`, one goroutine with WaitGroup/channel handoff, panic/recover, and `println`. It is an M0 input inventory and reference smoke test, not a claim that wasmvm or wazero runs LLGo guests.

`guest.wasm` is included because it is below 1 MiB. `observation.json` records its SHA-256, source hashes, producer versions, actual imports/instructions, and reference result. The generated artifact includes its compiler/runtime dependencies; it is not an independently implemented VM or a hand-authored core module.

## Regenerate

From this directory, with LLGo v1.0.8 and the matching runtime/SDK installed:

```sh
llgo version
go version
GOOS=wasip1 GOARCH=wasm llgo build -o guest.wasm .
wasm-tools validate --features all guest.wasm
wasm-tools print guest.wasm
```

The local Homebrew LLGo v1.0.8 installation lacked `targets/wasi.json`, so `llgo build -target wasi` failed before compilation. The raw `GOOS=wasip1 GOARCH=wasm` command above built successfully. This packaging issue is distinct from an engine capability gap. LLGo also reports `-trimpath` as unsupported; byte-for-byte output across source locations or toolchain installations is not guaranteed. The recorded hash identifies the included artifact, and regeneration requires reviewing and updating the observation metadata.

The recorded module imports shared `env.memory` with minimum 1,031 and maximum 8,192 64-KiB pages, Preview 1 functions, and `wasi.thread-spawn`. It exports `_start` and `wasi_thread_start`. Its body contains final EH (`try_table`, `throw_ref`) and atomic wait/notify/load/store instructions. No WasmGC struct/array operations are claimed: the guest's collector manages objects inside linear memory. The runner enables SIMD, but this particular fixture's recorded instruction inventory does not establish execution of SIMD operations.

## Reference execution

This command passed with **Wasmer 7.3.0**:

```sh
wasmer run --offline --disable-cache --enable-exceptions --enable-simd \
  --stack-size 16777216 guest.wasm
```

Observed result: exit status `0`, empty stdout, and stderr containing exactly `m0-llgo-guest` followed by a newline. This small fixture does not establish compatibility with all LLGo workloads.

The installed Wasmtime 48.0.1 rejected `-Sthreads` as no longer supported. No Wasmtime execution result for this legacy-threaded guest is claimed.

Check the included artifact and source digests:

```sh
python3 - <<'PY'
from pathlib import Path
import hashlib
import json
observation = json.loads(Path('observation.json').read_text())
for item in [observation['artifact'], *observation['sources']]:
    actual = hashlib.sha256(Path(item['path']).read_bytes()).hexdigest()
    if actual != item['sha256']:
        raise SystemExit(f"SHA-256 mismatch: {item['path']}")
print('Artifact and source SHA-256 checks passed.')
PY
```

Producer/source references: [LLGo v1.0.8](https://github.com/xgo-dev/llgo/releases/tag/v1.0.8), [LLGo license](https://github.com/xgo-dev/llgo/blob/c171abfa130108bc6451e01fcf15a8a3474eb0a7/LICENSE), and [Go license](https://go.dev/LICENSE). Retain applicable runtime/toolchain notices when redistributing generated artifacts.
