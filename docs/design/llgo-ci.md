# Go / LLGo Compatibility and CI Design

Verification date: 2026-10-26. This document distinguishes source constraints, observed tool behavior, and future VM acceptance requirements. Passing the probes does not mean Wasm has been implemented.

## Two separate compatibility targets

| Target | Build and execution path | Acceptance requirement |
| --- | --- | --- |
| Compile the host VM with LLGo | `llgo test ./...`, followed later by `llgo build ./cmd/wasmvm`, producing and running native programs | Use the same VM source, semantic tests, and fixtures as the standard Go compiler |
| Execute LLGo-generated guests in the VM | `GOOS=wasip1 GOARCH=wasm llgo build -o fixture.wasm .`, then load and execute it in the VM | Implement the Core instructions and WASI ABI actually required by that guest |

The first target is a CI requirement from project initialization. The second is an integration milestone in M4, after both core backends, WASI P1, threads, EH, and SIMD are implemented. Compiling the host with LLGo does not automatically provide the second capability. Likewise, executing the standard Go compiler's `wasip1/wasm` output does not prove compatibility with LLGo guests.

## Toolchains and source language baseline

- The source design uses Go 1.24 language and standard-library features as a conservative baseline, avoiding new Go 1.27 syntax. The language version declared in `go.mod` is distinct from the CI bootstrap toolchain version and does not establish acceptance on every older toolchain.
- Both initial CI paths use the Go **1.27.3** GOROOT so their standard-library inputs match. LLGo **v1.0.8** itself declares `go 1.27.0` in its `go.mod`, so the selected source installation cannot be bootstrapped with Go 1.24.
- The initial scaffold used LLGo v1.0.8 at `c171abfa130108bc6451e01fcf15a8a3474eb0a7`; M0 used the first memory-hook fix at `bc7c991d8fe91459c8afc1eb7d0e374124d95374`. M1 pins **`8b1389c627c77a02e56af03f08d9cdc19edffe77`**, including non-tearing unordered block operations in [LLGo PR #2919](https://github.com/xgo-dev/llgo/pull/2919). This remains an unreleased contribution, needed for unchanged wazero compatibility. Historical measurements retain their original pins.
- Pinned setup action: `4b2f2c9aed3fd28f139615fe636dee5e2ce5bc91`. LLVM major version: **22**. LLVM and system dependency patch versions still come from the installation sources; the setup does not claim a fully reproducible toolchain.
- Candidate evaluations use separate modules. Wazero v1.12.0 requires Go 1.25; the root scaffold's `go 1.24.0` declaration is not a minimum-version claim for that dependency.
- The workflow runs both compilers on Ubuntu 24.04 amd64 and macOS 15 arm64. Linux arm64 and Windows acceptance remain planned. A configured job or installation-action support is not evidence of complete VM support; final-head job results and actual implemented scope remain authoritative.
- Before formally declaring a minimum host version, add an actual Go 1.24 compatibility job and pin the LLVM patch version and toolchain image. The pinned CI Go version has been exercised; the source baseline still requires verification with the older toolchain.

Evidence: [LLGo v1.0.8 go.mod](https://github.com/xgo-dev/llgo/blob/c171abfa130108bc6451e01fcf15a8a3474eb0a7/go.mod), [v1.0.8 release](https://github.com/xgo-dev/llgo/releases/tag/v1.0.8), [setup-llgo action definition](https://github.com/xgo-dev/setup-llgo/blob/4b2f2c9aed3fd28f139615fe636dee5e2ce5bc91/action.yml), and [setup-llgo inputs and platforms](https://github.com/xgo-dev/setup-llgo/blob/4b2f2c9aed3fd28f139615fe636dee5e2ce5bc91/README.md). The checkout/setup-go v7 commits were also checked through the GitHub API and pinned by SHA in the workflow.

## Portable implementation constraints

1. Use ordinary Go for the interpreter, validator, compiler and object management. Avoid the gc compiler's private ABI, private `runtime` symbols and `go:linkname`. The required native backend uses the small [C ABI adapter](native-abi.md); standard-Go native execution initially requires cgo, while the interpreter and compiler logic remain buildable without it. Isolate genuine platform differences in `internal/platform` using `//go:build llgo` / `!llgo`, while keeping the higher-level API and tests identical.
2. Do not retain Go object addresses encoded as `uintptr`, or hide Go pointers in `[]byte`, numeric value slots, or guest linear memory. Objects that must remain reachable to the host GC need real typed references or an explicit root registry. If guest GC objects use handles, the VM owns handle-table release, generation validation, and root scanning.
3. Do not derive guest semantics from finalizer timing, exact host collection timing, or one compiler's heap statistics. Native LLGo uses conservative BDWGC by default, while standard Go has a different runtime. Guest WasmGC object lifetimes, resource closure, and quotas need explicit contracts.
4. Use `encoding/binary` and explicit bit operations to interpret little-endian guest memory, and `math.Float32bits` / `Float64bits` for floating-point representations. The default path must not cast byte slices through `unsafe.Pointer` for atomic access, avoiding alignment, host-endianness, and GC-root issues. Add lock-free fast paths only after a reference memory implementation and boundary tests exist.
5. The VM memory layer implements Wasm atomic semantics on both backends. Go's `sync/atomic` is a candidate host mechanism, not a substitute for the complete Wasm memory model. Design mixed atomic/ordinary accesses, overlapping widths, `memory.grow`, and wait/notify registration and wakeup together so that the VM itself has no Go data race.
6. Native LLGo maps goroutines 1:1 to OS threads with fixed native stacks. Limit VM workers, guest threads, and recursion depth. Use explicit interpreter frame stacks so unbounded guest recursion cannot exhaust the host stack. Do not assume host goroutines can be created indefinitely at negligible cost.

Sources: [runtime and source compatibility notes at the local snapshot](https://github.com/xgo-dev/llgo/blob/38bfd05ee102e36710a2df5bca47834a622b667c/README.md) and [classification of gc-specific implementation differences](https://github.com/xgo-dev/llgo/blob/38bfd05ee102e36710a2df5bca47834a622b667c/test/goroot/notapplicable.yaml). These constraints are project design decisions; they do not imply that ordinary `unsafe` or standard-library APIs are universally unavailable in LLGo.

## Capabilities required by the current LLGo guest

The local LLGo checkout is at `38bfd05ee102e36710a2df5bca47834a622b667c` (2026-10-16). Its W32 documentation establishes the following contract. Re-extract module imports, features, and ABI requirements whenever the guest compiler is upgraded.

| Capability | Requirement for this VM |
| --- | --- |
| Memory32 + WASI Preview 1 | Implement `wasi_snapshot_preview1`; P2/Component Model support cannot substitute for it |
| Shared `env.memory` + `wasi.thread-spawn` | Provide shared linear memory, atomic operations, and the thread-creation ABI; define instance/stack/TLS ownership for every thread |
| Standard Wasm EH | Support non-legacy EH; validate panic/recover, unwinding across functions, and thread isolation |
| SIMD | Support the SIMD instructions actually used by the module; the default runner explicitly enables SIMD |
| Guest GC within linear memory | LLGo supplies its own STW conservative collector; the VM does not scan its Go objects. This is separate from WasmGC struct/array/type/reference support |
| Thread exit | A returning worker follows wasi-libc pthread cleanup/TLS destructor paths; WASIX `thread_exit` is not required |

The existing documentation lists WasmGC, W64, JSPI, and WASI P2/components as separate work. In this version, the linear-memory GC abandons a collection attempt after a timeout when a C call cannot reach a safepoint. That is a guest-runtime progress limitation; the new VM must still execute its memory and scheduling semantics correctly.

Evidence: [WASI threads execution contract](https://github.com/xgo-dev/llgo/blob/38bfd05ee102e36710a2df5bca47834a622b667c/doc/wasm-wasi-threads.md), [threaded GC contract](https://github.com/xgo-dev/llgo/blob/38bfd05ee102e36710a2df5bca47834a622b667c/doc/wasm-threaded-gc.md), and [Wasm profile/ABI proposal](https://github.com/xgo-dev/llgo/blob/38bfd05ee102e36710a2df5bca47834a622b667c/doc/wasm-proposal.md). M0 now includes an [actual generated guest and successful Wasmer reference run](../../experiments/llgo-guest/README.md), but not a complete wazero threads adapter. Homebrew v1.0.8's named `-target wasi` path lacks its target JSON; the GOOS/GOARCH path above was verified separately.

## CI requirements

The executable configuration is [ci.yml](../../.github/workflows/ci.yml). All four compiler/platform jobs execute `test -count=1 -timeout=10m ./...`. Failure of any job fails the workflow; none uses `continue-on-error`.

| Check | Standard Go | LLGo |
| --- | --- | --- |
| Shared package tests and future VM fixtures | Required | Required |
| Unchanged pinned wazero evaluator | Required, with race checking and owned-code coverage >= 90% | Required with the pinned runtime fix; expected link failures do not satisfy this gate |
| Benchmark harness/fixture integrity and alternate candidate Go probes | Required | Covered by the shared fixture provenance; not an LLGo benchmark claim |
| Pure-Go core constraint | `CGO_ENABLED=0 go test` | Run in the default native LLGo environment with its C runtime dependencies installed |
| Formatting and static checks | `gofmt`, `go vet` | Share the checks on the same source |
| Host data-race checks | `go test -race` | Acceptance of `-race` is not treated as verified race-detection capability |
| Owned executable core package coverage | At least 90% per package under `internal` and `pkg`; empty/missing profiles fail | Same semantic tests; no LLGo coverage instrumentation claim |
| Bounded binary fuzzing | Ten-second fuzz run per platform | Runs the same fuzz seed cases through ordinary package tests |
| Wasm spec / WASI / LLGo guest corpus | Added by capability milestone | Same fixtures and assertions; no whole-package skips to hide differences |

`CGO_ENABLED=0` applies only to the basic Go test step, not the entire workflow environment. LLGo's LLVM and BDWGC dependencies are consistent with the goal that the VM source should not depend on a cgo-based engine. The root scaffold has no external dependency, while the evaluation modules have their own `go.sum` files. Action caching remains disabled; future caches must include all relevant module sums, compiler revision, LLVM, OS, and architecture.

The current repository probes cover typed atomics and goroutine/lock synchronization, numeric bit patterns and little-endian encoding, reference cycles remaining reachable across explicit GC, condition-variable wakeup, and context cancellation. Future VM tests must cover shared-memory growth, lost wakeups, cancellation propagation, GC roots across host calls/stack growth, EH unwinding, and resource closure. Retain minimal reproductions and issues for gc/LLGo differences; skipping an entire feature does not establish compatibility.

## Required own-core matrix (planned)

The current workflow tests binary primitives, candidate engines and host probes on Linux amd64 and macOS arm64. It does **not** yet implement the complete-engine acceptance matrix:

- M1: the complete Core 2.0 interpreter corpus and lifecycle suite under Go/LLGo on Linux amd64 and macOS arm64; Go tests with `CGO_ENABLED=0`.
- M2: the same full corpus with forced interpreter and native backends on each host/compiler pair, plus the [native ABI gates](native-abi.md). Add standard-Go `CGO_ENABLED=1` native execution and pointer-rule checks while retaining the cgo-free interpreter/compiler job. A cgo-free native-unavailable assertion is not counted as native execution.
- M3 onward: extend the same matrix to all accepted feature combinations and guest integrations; no whole-feature compiler skips or silent backend fallback.
- Add controlled performance acceptance separately from advisory hosted PR timing, following the [performance contract](performance-contract.md). Record actual native backend, target CPU profile, bridge mode, emitted-code identity and compiler revision.

Native cancellation/control records use the bridge's defined cross-language atomic handshake. Race-detector success in Go does not prove generated-code memory ordering, root publication or safe executable-code retirement. Architecture-specific tests and forced host/guest collection are mandatory.

## Local verification evidence and commands

An independent temporary probe passed both commands on 2026-10-26:

```sh
go test -v -count=1 ./...
llgo test -v -count=1 ./...
```

That probe checked four goroutines incrementing `atomic.Uint64` to 4000, `sync.WaitGroup` completion, little-endian uint64 round trips, IEEE-754 negative-zero bits, and a cyclic object graph remaining reachable after `runtime.GC()`. Its directory was `/var/folders/wb/8zzrvxd56294wxlxxbpws9440000gn/T/gwasm-llgo-probe-6zzrcjr4`. It was a temporary investigation check; the repository files define the project's ongoing tests.

The temporary investigation environment used Go `go1.27.3 darwin/arm64`. `/Users/lijie/go/bin/llgo` reported `llgo (devel) darwin/arm64`; `go version -m` identified revision `8ac217053d0337b7261cebc923bb6b17955bf49a` with a dirty marker, while `llgo env LLGO_ROOT` pointed to the 2026-10-16 checkout above. This describes only that investigation environment and is not acceptance evidence for the pinned v1.0.8 CI or a complete VM.

**The new repository used a different environment for its actual tests.** In the `wasmvm` working directory, the shell resolved `/opt/homebrew/bin/llgo`, a Homebrew wrapper executing `/opt/homebrew/Cellar/llgo/1.0.8/libexec/bin/llgo`. It reported `llgo v1.0.8 darwin/arm64`, with LLGO_ROOT pointing to the same Cellar installation. The project's five probes passed with this installation and Go 1.27.3; Go race/vet and actionlint also passed. Toolchain verification must record the working directory, PATH resolution, and actual executable instead of reusing assumptions from another checkout. See [validation.md](../validation.md) for the detailed record. CI installs the compiler/runtime from the specified revision and checks the action's reported revision.

Suggested verification commands from the project root:

```sh
go version
llgo version
llgo env LLGO_ROOT GOROOT
# If llgo is a wrapper, locate the actual binary before reading Go build info.
CGO_ENABLED=0 go test -count=1 -timeout=10m ./...
llgo test -count=1 -timeout=10m ./...
go test -race -count=1 -timeout=10m ./...
```

Remote CI has run successfully for both Go and LLGo; see the [verification record and run link](../validation.md). These results validate the current compatibility probes. Future VM functionality must be added to the same dual-compiler matrix phase by phase.
