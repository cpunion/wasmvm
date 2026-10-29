# WASI Preview 1 inputs

The ABI comes from WebAssembly/WASI revision
`fae981bae14809d91f9bc2d63852d461f331d161`, under its retained W3C notice.
`typenames.witx` and `wasi_snapshot_preview1.witx` define all 46 imports.

The complete applicable official P1 inventory has **72 programs**: 14 C,
46 Rust and 12 AssemblyScript. Source revision:
`58b6206a46013fb53d23990e5a650584a5361043`; prebuilt revision:
`9dfd4867f12498bd6411a57739c24b4880668683`. All 136 C/Rust/AssemblyScript source
files in the original source archive match the prebuilt source snapshot.
All 72 P1 programs are retained; the separate 55 P3 components are later scope.
The archive preserves selected source, configurations, filesystem fixtures,
producer metadata, original binary hashes and the upstream Apache-2.0 notice.

Development conversion uses wasm-tools 1.258.0 `strip --all`, removing custom
sections only. Wasm instructions, data and module sections are unchanged. This
keeps the offline archive small; `manifest.json` retains the original identity
and producer fields. `pins.json` records source, prebuilt and final archive sizes
and SHA-256 hashes. Every extracted asset is checked against the manifest.

An additional project-authored Go 1.27.3 integration guest combines goroutines,
GC, panic/recover, randomness, filesystem I/O and timed waits. Its source and
build flags are retained under `guest-go` and the extracted source directory,
with the Go runtime's BSD notice. It is separate from the 72 official programs.
Compile it with `GOOS=wasip1 GOARCH=wasm go build -trimpath -ldflags='-s -w'`.

Prepare and verify without downloading producers or fixtures:

```sh
python3 scripts/preview1_corpus.py
```

Regeneration uses the three exact archives in pins.json, preserves the source
and configuration trees, and runs `wasm-tools strip --all` on each selected
prebuilt binary. Verify original hashes before conversion. Produce a sorted
POSIX tar of the manifest's files, uid/gid/mtime zero, mode 0644; gzip with
mtime zero and an empty filename. No patched or excluded program is substituted
for a failure. The integration guest has its own producer/source record.

`TestP1Conformance` executes every official program in three tiers, checks its
configured exit/output and deterministic cleanup, and writes per-tier reports
under `artifacts/conformance/wasi` when WASMVM_CONFORMANCE_REPORT is set.
A successful compilation alone is not counted as guest execution.
