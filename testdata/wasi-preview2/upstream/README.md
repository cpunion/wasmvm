# Upstream P2 world consumers

The complete `p2_*.rs` inventory from Wasmtime 48.0.1, revision
`7bac2c2775808aaec5d4aa5627a5e447b51102cf`, contains 95 programs. This corpus
executes all 92 consumers of the declared stable worlds with unchanged sources
and guest assertions. It retains compiled config/keyvalue consumers but excludes
their `0.2.0-rc.1`/`0.2.0-draft` interfaces from runtime claims. The draft TLS
source is also retained and explicitly excluded. No failed stable program is
silently skipped or counted as a normal success.

Rust 1.98.0, the original Cargo.lock, `wasm32-wasip2`, wit-bindgen 0.60.0 and
wasm-tools 1.262.0 reproduce and independently validate 94 binaries. The manifest
pins all 206 archive assets, producer executable, generator, original source
archive and each consumer source/binary. The [shared P3 archive](../../wasi-preview3/upstream)
already contains the original test sources/support, Cargo.lock, upstream
licenses, and Rust/dependency notices; their hashes are pinned instead of
duplicating those files. This archive adds P2 WIT and reference harnesses.

Both owned backends execute every stable consumer under Go and LLGo. Expected
outcomes include exit codes, an intentional guest panic, three protocol traps,
two quota failures, two observed-wait cancellations and two incomplete HTTP
responses. The other cases must succeed. Additional scenarios cover all eight
P2/P3 header-append modes, streaming echo through an outbound request, hash
fanout and cloned/new field maps. Reactor tests call its actual string, imported
stream and imported-record exports, rather than only its empty CLI entry.

A separately named `p2_api_proxy_streaming-executor-lifetimes` variant
covers both the additional small duplex interaction and the 160 KiB stress
interaction through 4 KiB body buffers. The small case is explicitly named
`double-echo-executor-lifetimes`; the unchanged inventory still uses the
original guest. The sample executor in the
unchanged consumer retains nonready subscriptions while repolling futures; a
future can then consume EOF and drop its stream with an old pollable still live.
The strict runtime traps that invalid resource drop. The original large-body
diagnostic is retained as a failure, not claimed as a successful original test.
The original small duplex interaction also failed with a live pollable in
[PR #95 race CI](https://github.com/cpunion/wasmvm/actions/runs/41383051587/job/125077801323).
Neither original duplex failure is counted as a passing original interaction.
Wasmtime 48.0.1 passed 40 ordinary and 100 paced diagnostic requests, so a
reference-runtime failure is not claimed. The explicit patch retires all old
subscriptions and wakes their futures for fresh registration. Its patch,
repaired-source and binary hashes are separate from the 92 unchanged inventory
entries. Guest HTTP behavior and assertions are unchanged; resource lifetime
checks remain enforced, including the original negative lifetime consumer.

The host harness is adapted, while guest sources and assertions remain original:

- Fresh filesystem grants reproduce upstream files, read-only mounts and
  preopen names. Loopback sockets and deterministic DNS avoid public services.
- Mock clocks reproduce the reference API-time inputs. Directory stdin and
  broken stdout preserve OS error provenance. Socket allocation denial is
  distinct from denial of a later bind/connect/send operation.
- A local HTTP server supplies the reference method/URI headers and echo body.
  The duplex case publishes response headers before consuming the request body,
  matching the Hyper server's streaming behavior. The invalid-version case uses
  a malformed response line, since Go otherwise accepts its CONNECT request.
- Services without a response report `HTTP-response-incomplete`, including the
  intentional trap case, as in the reference CLI harness. Indefinite sleepers
  are canceled only after the guest invokes its poll/wait interface.
- Header quotas use wasmvm's explicit bounded provider limits. P2 must fail with
  `wasm.ErrLimit`; P3 returns the guest-visible recoverable error. The retained
  hostcall-fuel program runs its bounded read scenario, including a `u64::MAX`
  request. Wasmtime's separate `-Shostcall-fuel=5000` accounting and its other
  command-line modes are reference-specific policy, not a wasmvm fuel claim.

Prepare and verify offline, without Rust or wasm-tools:

```sh
python3 scripts/wasi_upstream_corpus.py
go test ./pkg/wasi/preview2 -run TestUpstreamP2World
llgo test -O3 -lto=full ./pkg/wasi/preview2 -run TestUpstreamP2World
```

The existing component CI preparation includes both archives. Override the
extracted directory with `WASMVM_WASI_UPSTREAM_CORPUS`. Rebuild with the exact
source archive and tools pinned by the manifest:

```sh
python3 scripts/build_wasi_p2_world_fixtures.py \
  --source-archive /path/to/wasmtime-source.tar.gz \
  --work artifacts/components/wasi-upstream-producer \
  --wasm-tools /path/to/wasm-tools
```

Passing this consumer inventory is evidence for the declared worlds, not
exhaustive WASI conformance or performance acceptance. Draft interfaces remain
outside the selected scope. Existing performance gates remain open.
