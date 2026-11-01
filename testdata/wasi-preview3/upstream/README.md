# Upstream native WASI consumers

This offline corpus preserves the complete 61-program `p3_*.rs` inventory from
Wasmtime 48.0.1, revision `7bac2c2775808aaec5d4aa5627a5e447b51102cf`.
The 60 stable consumers retain their original source and assertions. The draft
`p3_tls_sample_application` source is retained and explicitly excluded because
`wasi:tls` is outside the declared stable profile. No failed stable consumer is
classified as a pass or silently skipped.

The original Cargo lockfile, shared test support, WIT, reference harnesses,
upstream licenses, and notices for 101 producer dependencies are archived with
the binaries. Rust 1.98.0 targets `wasm32-wasip2`; wit-bindgen 0.60.0 supplies
native P3 bindings alongside the standard library's P2 imports. Components are
stripped and independently validated with wasm-tools 1.262.0. The chain service
also has a composed binary connecting its original import to `p3_http_echo`.
`manifest.json` pins the complete source archive, producer executable, generator,
all 941 archive files, and both original and composed binaries.

Tests execute all 60 cases with both owned backends and both host compilers:
58 normal successes, one expected resource-limit failure, and one cancellation
after observing the guest's indefinite clock wait. The resource tests keep the
upstream 100-handle limit: 200 sequential streams succeed; 1,000 retained clock
tasks must fail with `wasm.ErrLimit`. Additional tests exercise body/trailer
forwarding, compression, host forwarding, composition and mixed P2/P3 CLI calls.

Host inputs are hermetic adaptations of the upstream harness, not a claim that
the original Wasmtime host harness is executed unchanged:

- Filesystem grants use fresh temporary directories; DNS returns fixed loopback
  addresses or an explicit name-not-found error. Socket tests use loopback.
- HTTP command tests use a local Go echo server. The invalid-version test gets
  an explicitly malformed response line; Go's server otherwise accepts the
  CONNECT request which Hyper's upstream test expects to fail.
- The forbidden-header service uses an explicit embedding client interceptor,
  matching upstream `TestHooks`. It checks that custom and transport-control
  fields are removed while ordinary fields survive. Other outbound tests use
  the actual granted HTTP transport.
- The original chain import is satisfied by composition, without editing guest
  code or replacing it with a dummy host implementation.

Prepare and verify without network access, Rust, or wasm-tools:

```sh
python3 scripts/wasi_upstream_corpus.py
go test ./pkg/wasi/preview3 -run TestUpstreamWASI
llgo test -O3 -lto=full ./pkg/wasi/preview3 -run TestUpstreamWASI
```

`scripts/component_corpus.py` includes this preparation in the existing CI input
step. Override the extracted directory with `WASMVM_WASI_UPSTREAM_CORPUS`.

Rebuilding requires the complete source archive identified by the manifest,
Rust 1.98.0 with the `wasm32-wasip2` target, and the pinned wasm-tools executable:

```sh
python3 scripts/build_wasi_upstream_fixtures.py \
  --source-archive /path/to/wasmtime-source.tar.gz \
  --work artifacts/components/wasi-upstream-producer \
  --wasm-tools /path/to/wasm-tools
```

The rebuild uses `--locked`, preserves all guest sources, remaps build paths and
emits a deterministic archive. Missing tools, sources, licenses, or hashes fail
preparation. Corpus execution does not close exhaustive WASI conformance,
unselected experimental proposals, or performance acceptance.
