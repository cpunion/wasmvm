# wasmvm

[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![CI](https://github.com/cpunion/wasmvm/actions/workflows/ci.yml/badge.svg?event=pull_request)](https://github.com/cpunion/wasmvm/actions/workflows/ci.yml)
[![Go module](https://img.shields.io/badge/Go_module-source_only-blue)](pkg/README.md)
[![API](https://img.shields.io/badge/API-experimental-orange)](docs/embedding.md)
[![Release](https://img.shields.io/badge/release-not_yet_published-lightgrey)](https://github.com/cpunion/wasmvm/releases)
[![Coverage gate](https://img.shields.io/badge/coverage_gate-%E2%89%A590%25_per_package-blue)](scripts/check_coverage.py)

A WebAssembly runtime written in Go, with its own portable interpreter and
amd64/arm64 native compiler. The same source builds and executes with Go and
LLGo. It supports Core 2.0/3.0, WasmGC, Core Threads, the Component Model, and
versioned WASI Preview 1/2/3 providers.

wasmvm owns decoding, validation, compilation, execution, guest GC and host
integration. Competing runtimes are independent test and benchmark references;
execution never silently switches to one of them.

The source APIs and proposal snapshots are experimental. Performance and
production-release acceptance remain open; see [current measurements](#measured-performance)
and the [release contract](docs/release-readiness.md). Supported builds pin
Go1.27.3, LLGo `56fa9ae`, LLVM22.1.8 and the LLGo `-O3 -lto=full` profile.

## Project status

Verified CLI candidate archives are available from successful Go/LLGo CI runs; see [download, checksums and runtime requirements](docs/cli-packages.md).

**Experimental; no tagged release or stable API commitment yet.** The module
path is `github.com/cpunion/wasmvm`; public packages are available from source.
As of 2026-11-01, the Go proxy lists no tagged versions and pkg.go.dev has not
indexed `pkg/vm`. The package badge describes source availability rather than
a published package. Pin a commit when depending on the module.

The declared Core and WASI implementations are present, with the conformance
boundaries below. The stable world inventory executes 92 P2 and 60 P3
consumers; this is not exhaustive WASI certification. Performance and production
release acceptance remain open; see the [release contract](docs/release-readiness.md). Separately gated experimental snapshots cover the settled non-JS
Phase 3–4 extensions; [PR #87](https://github.com/cpunion/wasmvm/pull/87) records
their integration and final CI evidence.

| Execution target | Go | LLGo | Status |
| --- | --- | --- | --- |
| Portable interpreter | Yes; `CGO_ENABLED=0` supported | Yes | Same semantic tests |
| Linux amd64 native | Yes; C compiler/cgo required | Yes; pinned release profile | CI target |
| macOS arm64 native | Yes; C compiler/cgo required | Yes; pinned release profile | CI target |
| Other native OS/architecture combinations | Not validated | Not validated | No native support claim |

CI runs Go and LLGo on both targets; eight LLGo shards execute all 32 test
packages exactly once per target. Shared Go checks include race detection,
native pointer ownership, fuzz smoke checks and a **90% statement-coverage gate
for each executable package**. The coverage badge states this policy, not an
aggregate measurement. The CI badge tracks PR runs; merging requires the final
PR head to pass every semantic shard on all four compiler/platform combinations,
plus the shared race/quality checks. See the [validation record](docs/validation.md).

## Install and run

The module requires Go **1.26 or newer**; CI uses **Go 1.27.3**. Build from source:

```sh
git clone https://github.com/cpunion/wasmvm.git
cd wasmvm
go build -o wasmvm ./cmd/wasmvm
./wasmvm --help
```

The default backend is the interpreter; the default Core profile is Core 2.0.
For a portable build, use `CGO_ENABLED=0 go build -o wasmvm ./cmd/wasmvm`.
Go native execution requires cgo and a system C compiler on a supported target.
For LLGo, use the [pinned toolchain](docs/design/llgo-ci.md) and
`llgo build -O3 -lto=full -o wasmvm ./cmd/wasmvm`.

These examples use a checked-in fixture and run from the repository root.
All options precede the module path:

```sh
./wasmvm validate benchmarks/embedded/fixtures/call-control.wasm
./wasmvm inspect --profile core3 benchmarks/embedded/fixtures/call-control.wasm
./wasmvm run --export run benchmarks/embedded/fixtures/call-control.wasm i32:42 i32:1
./wasmvm run --backend native --export run \
  benchmarks/embedded/fixtures/call-control.wasm i32:42 i32:1
```

Core calls accept typed arguments and produce JSON with exact numeric results.
Use `--profile core3` for Core 3.0, `--fuel` for deterministic execution limits,
and `--timeout` for cancellation. Input is binary `.wasm`; convert WAT using
external tooling such as `wasm-tools parse input.wat -o input.wasm`.

Run WASI commands with explicit environment and filesystem access:

```sh
# Preview 1: _start entry; following values are guest arguments.
./wasmvm run --wasi --env MODE=test --dir /work=/tmp/work program.wasm argument

# Component commands: standard wasi:cli/run entry.
./wasmvm run --wasi-preview 2 command-p2.wasm argument
./wasmvm run --wasi-preview 3 --backend native --timeout 10s command-p3.wasm

# Legacy wasi-threads is a separate P1 thread-spawn adapter.
./wasmvm run --wasi-threads --profile core3 --backend native threaded.wasm
```

Environment, directories and networking are not inherited. Network operations
require explicit embedding capabilities. Large Go/LLGo guests can require a
larger `--native-code-bytes` quota. See the [CLI guide](docs/cli.md) for limits,
stdin, preopens, exits and component command options.

## Go packages and embedding

All packages belong to the same untagged module and have experimental APIs.
Import paths use the prefix `github.com/cpunion/wasmvm/`.

| Package | Available capability | Publication status |
| --- | --- | --- |
| [`pkg/wasm`](pkg/wasm) | Value types, features, errors and traps | Source only; experimental |
| [`pkg/vm`](pkg/vm) | Engine/Store, compilation, linking, typed calls and references | Source only; experimental |
| [`pkg/component`](pkg/component) | Components, Canonical ABI, resources, native async, futures/streams | Source only; experimental |
| [`pkg/wasi/preview1`](pkg/wasi/preview1) | P1 processes, streams and host capabilities | Source only; experimental |
| [`pkg/wasi/threads`](pkg/wasi/threads) | Legacy P1 thread-spawn groups | Source only; experimental |
| [`pkg/wasi/preview2`](pkg/wasi/preview2) | Versioned 0.2 CLI/HTTP providers and worlds | Source only; experimental |
| [`pkg/wasi/preview3`](pkg/wasi/preview3) | Versioned 0.3 async CLI/service/middleware providers | Source only; experimental |

For example, run this Go program from the repository root:

```go
package main

import (
	"fmt"
	"os"

	"github.com/cpunion/wasmvm/pkg/vm"
)

func main() {
	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	binary, err := os.ReadFile("benchmarks/embedded/fixtures/call-control.wasm")
	must(err)
	engine := vm.NewEngine()
	defer engine.Close()
	module, err := engine.Compile(nil, binary)
	must(err)
	defer module.Close()
	store, err := engine.NewStore(vm.DefaultStoreConfig())
	must(err)
	defer store.Close()
	instance, err := vm.NewLinker(store).Instantiate(nil, module, vm.CallOptions{})
	must(err)
	run, err := instance.Export("run")
	must(err)
	result, err := run.Call(nil, []vm.Value{vm.I32(42), vm.I32(1)}, vm.CallOptions{})
	must(err)
	bits, _, err := result.Values[0].Bits()
	must(err)
	fmt.Println(bits) // 42
}
```

To select native/Core 3 execution, start with `vm.DefaultEngineConfig()`, set
`Backend = "native"` / `Profile = "core3"`, and call `vm.NewEngineWithConfig`.
Stores bound guest frames, slots, memory, tables, heap objects and roots.
References require explicit root/release ownership. See [embedding](docs/embedding.md),
[executable examples](pkg/vm/example_test.go) and [package architecture](docs/design/packages.md).

## Standards and proposal support

Snapshot: **2026-11-01**. The tables describe binary and Go embedding contracts.
JS/browser APIs are outside this runtime. Proposal maturity and implementation
status are separate: an implemented Phase 3 snapshot remains experimental.

### Finished Core proposals

All implemented binary features execute through both owned backends. Core 2.0
is the default; Core 3.0 rows require `--profile core3` or individual feature bits.

| Proposal / feature | Profile | Support |
| --- | --- | --- |
| MVP | Core 1 / 2 | Implemented |
| Import/export of mutable globals | Core 1 / 2 | Implemented |
| Non-trapping float-to-int conversions | Core 2 | Implemented |
| Sign-extension operators | Core 2 | Implemented |
| Multi-value | Core 2 | Implemented |
| Reference types | Core 2 | Implemented |
| Bulk memory operations | Core 2 | Implemented |
| Fixed-width SIMD | Core 2 | Implemented; 236 operations |
| Tail calls | Core 3 | Implemented |
| Extended constant expressions | Core 3 | Implemented |
| Typed function references | Core 3 | Implemented |
| Garbage collection | Core 3 | Implemented; typed objects, tracing, roots and quotas |
| Multiple memories | Core 3 | Implemented |
| Relaxed SIMD | Core 3 | Implemented; all 20 operations |
| Final exception handling | Core 3 | Implemented; `try_table` / `exnref` |
| Memory64, including table64 | Core 3 | Implemented |
| Branch hinting | Core 3 | Accepted and ignored; no hint-driven optimization |
| Custom annotations in WAT | Text tooling | External producer; no owned WAT parser |
| JS BigInt / i64 integration | JS API | Outside scope |
| JS string builtins | JS API | Outside scope |

The complete pinned Core 2.0 suite has **51,139 binary assertions / 148 scripts**;
Core 3.0 has **61,395 / 259**. Text-parser assertions are separately classified.
Each inventory runs in the interpreter and both native tiers, under both host
compilers on both CI platforms. Counts identify pinned suites, rather than
certifying arbitrary future specification changes.

### Active Phase 3–5 proposals

These gates are independent of `--profile core3`. Pins, contracts and complete
affected inventories are in [explicit core extensions](docs/design/core-extensions.md).

| Proposal | Phase | Support / CLI gate |
| --- | ---: | --- |
| Threads | 4 | Implemented; `--threads`, 67 atomics, shared memory, wait/notify |
| Compact Import Section | 4 | Implemented; `--compact-imports`, both grouped encodings |
| Wide Arithmetic | 4 | Implemented; `--wide-arithmetic`, four i64/128-bit operations |
| Custom Page Sizes | 3 | Implemented experimental snapshot; `--custom-page-sizes`, 1-byte and 64-KiB pages |
| Custom Descriptors and JS Interop | 3 | Implemented experimental non-JS snapshot; `--custom-descriptors`, non-JS descriptors and exact references |
| Stack Switching | 3 | Implemented experimental snapshot; `--stack-switching`, typed one-shot guest continuations |
| JS Promise Integration | 5 | Outside scope; JS embedding |
| Web Content Security Policy | 5 | Outside scope; browser embedding |
| ESM Integration | 3 | Outside scope; JS embedding |

The three Phase 3 implementations execute their full affected binary inventories
under Go and LLGo in the interpreter and both native tiers. The final-head CI
matrix verifies every shard on both target platforms before merge.
Their pinned snapshots are experimental and disabled by default. Core Threads,
legacy wasi-threads, component async, Core Stack Switching and shared-everything
threads are distinct contracts. Shared-everything threads is not implemented.
The [full proposal audit](docs/research/status-2026-11-01.md) lists every Phase
1–2 and inactive entry at its historical engine snapshot; use this table for
subsequent extension progress. Rounding variants, FP16, acquire-release atomics,
stringref and other early proposals have no implementation claim. Maturity
comes from the [official proposal register](https://github.com/WebAssembly/proposals/tree/df6a2aca1cd9d779eb4f2f3e19dae8277f336015).

### Components and WASI

P1/P2/P3 mean Preview 1/2/3, not WASI semantic versions 1/2/3. Providers pin
WASI **0.2.12** and **0.3.2**; there is no WASI 1.0 release claim.

| Capability | Implemented scope | Evidence and remaining boundary |
| --- | --- | --- |
| Component Model | Composition, values/resources, Canonical ABI, local i64 borrows, host bindings, sync/async | All 75 pinned sources convert; declared runtime profile executes 962 commands / 555 guest invocations. Optional component threading is separate. |
| WASI Preview 1 | All 46 declared imports; explicit process capabilities | Complete 72-program upstream inventory; actual Go/LLGo guests |
| Legacy wasi-threads | P1 thread-spawn ABI, separate Stores, shared memory, joined cleanup | Complete 15-program inventory and LLGo pthread integration |
| WASI Preview 2 | Stable 0.2.12 CLI/HTTP providers, filesystem, clocks, random, I/O, sockets and HTTP | 92 unchanged stable upstream consumers plus separately identified interactions; expanded world conformance remains open |
| WASI Preview 3 | Stable 0.3.2 CLI/service/middleware providers, native async, futures/streams, cancellation and backpressure | 60 unchanged upstream consumers, C/Rust guests and composed HTTP scenarios; expanded world conformance remains open |

Component Model maturity is separate from released WASI interface versions.
See [component inputs](testdata/component-runtime/README.md),
[P2 inventory](testdata/wasi-preview2/upstream/README.md),
[P3 inventory](testdata/wasi-preview3/upstream/README.md) and
[world contracts](docs/design/wasi-worlds.md) for exclusions and provenance.

## Comparison with other runtimes

This is a dated, version-specific summary, not fresh certification of every
backend. The [capability audit](docs/research/status-2026-11-01.md#competitors-accepted-core-features)
records primary sources, feature flags and ambiguous entries.

| Runtime / audited version | Implementation and execution | Capability comparison |
| --- | --- | --- |
| **wasmvm** | Go/LLGo; owned interpreter and amd64/arm64 compiler | Declared Core 2/3, GC/Threads, components and P1/P2/P3; experimental, no release, performance acceptance open |
| [wazero 1.12.0](https://github.com/wazero/wazero/tree/v1.12.0) | Go; interpreter and native compiler | Go embedding and P1; experimental modern features, without declared GC/multi-memory/memory64/P2/P3 parity |
| [Wasmtime 48.0.1](https://github.com/bytecodealliance/wasmtime/tree/v48.0.1) | Rust; Cranelift in this audit | Broad Core/GC/Threads and component/WASI reference; experimental features are individually gated |
| [Wasmer 7.3.0](https://github.com/wasmerio/wasmer/tree/v7.3.0) | Rust; Cranelift in this audit | Core 2 and selected modern features; no declared GC feature; WASIX is a separate ABI |
| [WasmEdge 0.17.1](https://github.com/WasmEdge/WasmEdge/tree/0.17.1) | C++; interpreter and AOT | Broad Core/GC features; only its interpreter was timed; no equivalent P2/P3 coverage claim here |
| [Wago 90f2b57](https://github.com/wago-org/wago/tree/90f2b57dafa7f8c3cc7e3aca38b91148f15e1d1e) | Go; beta runtime | Reports Core 3, optional Threads, compact imports and wide arithmetic; no fresh performance measurement here |

### Measured performance

These are dated advisory events with exact source/toolchain identities. Hosts,
compiler versions, lifecycle phases and sampling cohorts remain separate.
**Performance parity is not achieved.** Full process arrays, unresolved controls,
adverse results and unsupported references remain in the linked records.

The [shared indirect-dispatch study](docs/research/performance-shared-indirect-dispatch-2026-11-13.md)
reduces native code size **90.4%** in a 512-function scaling case and **11.4%**
in the P3 Rust component, with complete Go P3 CPU unchanged. Its fresh Linux
portfolio records the remaining memory, SIMD and SHA-256 gaps; P3 still executes
**3.8–4.0× Wasmtime's instructions**. Competitive parity remains open.

The [canonical buffer study](docs/research/performance-canonical-buffer-state-2026-11-13.md)
reduces Go complete P3 CPU by **0.6%/1.5%** for 16 B/64 KiB and removes
**7/31 allocations per request**. Small-request throughput remains unresolved.
The 64 KiB CPU ratio is still **2.24× Wasmtime**; LLGo controls remain unresolved.

The [integer-image entry study](docs/research/performance-integer-entry-effects-2026-11-13.md)
reduces complete P3 CPU by **2.1–2.3%** and raises throughput **2.2–2.7%** in
the final Go source comparison. Go A/A controls resolve; competitor controls
and the larger remaining gaps are reported separately.

The [nested scalar call study](docs/research/performance-nested-scalar-calls-2026-11-13.md)
records Go at **1.119 ns/iteration**, versus wazero **3.421**, Wasmtime **1.955**
and Wasmer **1.943** on one bounded call-tree kernel. Head/reference controls
resolve; the old source control does not. Complete P3 requests show no material
CPU improvement in the separate screen. Overall parity remains open.

The [2026-11-13 follow-up](docs/research/performance-go-and-llgo-main-2026-11-13.md)
reproduces a **3× Go improvement** for one private-word loop (0.814 → 0.271
ns/iteration), records unresolved competitor controls, and tests the updated
LLGo main baseline separately. P3 and full-profile acceptance remain open.

The [native compiler update](docs/research/performance-compiler-layout-2026-11-12.md)
reduces compilation of the P3 HTTP component from **179.0 to 73.9 ms**, with CPU
**210.3 to 136.5 ms**; Wasmtime 48 observes **57.7 ms** in that separate cohort.
Bounded workers, indexed labels and bitmap relocation preserve both policy
tiers and checks. This preparation improvement does not close execution gaps.

The latest [native runtime and transaction comparison](docs/research/performance-native-runtime-image-2026-11-12.md)
reduces native entries from 122 to 58 per complete P3 request and retains a
smaller runtime metadata image. Snapshot Go observes **51,170 / 38,338 req/s**
at 16 B / 64 KiB, versus Wasmtime **75,733 / 66,238**; throughput gaps remain
**1.48× / 1.73×**. Matched head/base throughput improves about 20% for snapshots
and 22% for Reader input. All 36 timing A/A controls stay within ±2.5%; 34/36
CPU controls do. These local measurements leave performance acceptance open.
The [scalar-carrier event](docs/research/performance-canonical-scalars-2026-11-12.md)
and [ready-byte event](docs/research/performance-ready-bytes-2026-11-12.md)
retain their separate source cohorts and adverse controls. The
[immutable-input API](docs/research/performance-http-snapshot-2026-11-11.md)
keeps preparation separate; changing mutable input still pays its snapshot cost.
Canonical conversion, owned forwarding and VM/host transitions remain priorities.

The preceding [Go-priority comparison](docs/research/performance-go-frame-plans-2026-11-11.md)
retains bounded host-request storage, saving 82 allocations/request. A frame-plan
expansion reduced native entries but failed to establish throughput gains and
was reverted. Go performance takes priority; LLGo semantic compatibility remains
required. [Earlier schema](docs/research/performance-value-comparison-plans-2026-11-11.md)
and [call-veneer](docs/research/performance-call-veneers-2026-11-11.md) events
retain their distinct source/toolchain cohorts. Performance acceptance stays open.

The [complete service comparison](docs/research/performance-service-all-runtimes-2026-11-10.md)
uses one bounded driver for all runtimes. M4 Max, Go1.27.3/LLGo3d92a58,
C SHA,64 blocks/request,4 workers; throughput in req/s:

| wasmvm LLGo | wasmvm Go | wazero1.12 | Wasmtime48 | Wasmer7.3 | Native C/LLGo | Native C/Go |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 135706 | 263847 | 221635 | 304512 | 300650 | 133541 | 318452 |

All72 profiles include both producers, interpreters and policy variants,720
throughput processes and72 separate resource processes.35/72 controls resolve;
all six owned LLGo native controls remain unresolved. These point estimates do
not establish a controlled ranking.

The latest [native indirect-call comparison](docs/research/performance-native-indirect-calls-2026-11-11.md)
keeps ordinary same-instance indirect/reference calls in native frames.
Counter diagnostics on both compilers observe approximately 317 -> 186 native
entries per 16-byte HTTP request, a 41.3% reduction. Core execution boundaries
remain approximately 93 and component tasks 9 per request.
On M4 Max/macOS with four workers, Go native observes 22254 -> 25871 req/s
for 16-byte responses and 16224 -> 18169 req/s for 64-KiB responses.
LLGo native observes 6598 -> 6246 and 4440 -> 4588 req/s respectively.
Same-event Wasmtime 48 observes 74362/62307 req/s. Ordinary Go applications
observe 1798840/40155 and LLGo applications 116366/9346 req/s.
**Only 1/40 A/A controls resolve.** These timings retain their advisory scope,
including the adverse LLGo point estimate. LLGo throughput improvement and
competitive parity remain unestablished. All source/compiler/reference pairs,
resource results and startup/compile/first/warm/cleanup phases are retained.
The [previous component allocation event](docs/research/performance-component-buffer-state-2026-11-10.md)
retains its independent results for capability coallocation and leaf-buffer codecs.

The 2026-11-10 [compiler-source comparison](docs/research/performance-llgo-merged-2026-11-10.md)
evaluates the full merged LLGo #3012 integration, `3d92a58` -> `717fbe442`,
on identical wasmvm sources. That dated event retained `3d92a58`; the current compiler baseline is listed above.
HTTP fixture, 64 KiB, four workers:

| Path | Old LLGo req/s | Candidate req/s | Old / candidate B/request |
| --- | ---: | ---: | ---: |
| wasmvm streaming native | 2696 | 3920 | 846277 / 754768 |
| wasmvm buffered native | 3226 | 5506 | 561018 / 474806 |
| Ordinary native-Go host control | 8854 | 7985 | 732584 / 732057 |

Only **5/50 P3 timing controls resolve**, with no old/new pair resolving both;
these are advisory observations. Buffered peak RSS rises 159.4 -> 175.8 MiB.
Same-event buffered Go native observes 12798 req/s and Wasmtime 48 observes
56515; substantial gaps remain. The original22 core portfolio is mostly flat,
with an adverse short-batch GC allocation interpreter ratio
**1.0643 [1.0434,1.0856]**, both controls resolved, reported as
[LLGo #3045](https://github.com/xgo-dev/llgo/issues/3045).

Independent probes confirm channel blocking-exchange allocation about416 ->0
B/round-trip and map-key/range temporaries ->0. Existing typed VM calls stay
near96 LLGo B/call; reusable scalar calls remain0. The opt-in frontend cache
reduces SSA events92 ->24 but does not improve this warm-build median;
default persistence stays disabled. [#3040](https://github.com/xgo-dev/llgo/issues/3040)
fresh-zero-slice append duplication remains reproducible. Candidate macOS
semantics pass all32 root packages and both measurement modules; candidate
Linux acceptance and the GC investigation were still open for that candidate;
the current main baseline is evaluated separately in the follow-up above.

The earlier [Store-local API](docs/research/performance-p3-store-local-2026-11-10.md)
and [readiness](docs/research/performance-p3-read-ahead-2026-11-10.md) studies
retain their own compiler/source cohorts, counters and unresolved controls.
Earlier [transfer-buffer](docs/research/performance-p3-io-buffers-2026-11-10.md),
[worker-reuse](docs/research/performance-p3-worker-reuse-2026-11-10.md),
[immutable-copy](docs/research/performance-immutable-memory-2026-11-10.md) and
[rejected-pipe](docs/research/performance-p3-body-pipe-evaluation-2026-11-10.md)
events retain all profiles, adverse results and resource tradeoffs independently.

The [Linux routing snapshot](docs/research/performance-linux-routing-2026-11-09.md)
retains32 complete cohort-cases, original partial/timed-out observations and all
controls. Native memory/SIMD, shared-memory, native-C and public-call gaps remain
open. The [compiler baseline](docs/research/performance-llgo-escape-upgrade-2026-11-09.md)
and [invocation cleanup](docs/research/performance-invocation-cleanup-2026-11-09.md)
retain their own source/compiler comparisons.

The latest targeted [GC comparison (PR #140)](https://github.com/cpunion/wasmvm/pull/140#issuecomment-6735033151)
uses source `da8d9f5` (merged as `9852f63`), Go1.27.3 and LLGo38de1e218
`-O3 -lto=full`. Sustained native public-call time in ns/source iteration:

| GC workload | M4 ARM64 LLGo | / Go | / Wasmtime | EPYC7763 amd64 LLGo | / Go | / Wasmtime |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| i31 roundtrip | 1.321 | 0.985× | 0.745× | 1.589 | 0.992× | 0.722× |
| Struct access | 1.272 | 0.979× | 1.251× | 1.601 | 0.996× | 1.261× |
| Allocation, including pauses | 30.161 | 0.947× | 0.613× | 80.195 | 1.107× | 0.706× |
| Indirect struct access | 1.621 | 0.981× | 0.638× | 4.422 | 0.987× | 0.933× |

ARM64 indirect access observes a63.2% reduction against its own baseline;
struct access and Linux LLGo/Go allocation remain gaps. Five independent
process blocks are advisory: only one ARM64 case resolves all diagnostic A/A
controls, versus all four Linux sustained cases. Original short-case adverse
results remain explicit in the [full record](docs/research/performance-arm-heap-operands-2026-11-09.md).
Wazero1.12.0 compiler/interpreter and Wasmer7.3.0 lack this GC profile; their
unsupported probes are retained. Wasmtime48.0.1 Cranelift and WasmEdge0.17.1
interpreter are independently measured supported references. These GC cases
have no equivalent ordinary C control. Startup, preparation, first/warm calls,
cleanup and interpreter rows remain separate; this targeted event does not
replace the full portfolio or establish controlled acceptance.

The [historical complete Core baseline](docs/research/performance-2026-11-01-core.md)
and [original full audit](docs/research/performance-2026-11-01.md) preserve all68
Core cases at their original2026-11-01 source. Later
[memory selection](docs/research/performance-memory-instruction-selection-2026-11-08.md),
[private-memory forwarding](docs/research/performance-private-memory-forwarding-2026-11-08.md),
[GC slot reuse](docs/research/performance-gc-slot-reuse-2026-11-08.md),
[Threads](docs/research/performance-native-threads-2026-11-02.md) and
[interpreter](docs/research/performance-interpreter-dispatch-2026-11-02.md)
records retain their distinct events, rejected experiments and remaining gaps.

Startup, preparation, first/warm calls, cleanup, request wall and complete API
intervals are measured separately; no empty-call or phase median is subtracted.
Benchmark CI is opt-in through the `run-benchmarks` PR label or
[workflow dispatch](https://github.com/cpunion/wasmvm/actions/workflows/benchmarks.yml).
It compares pinned base/head/reference sources and posts results to the PR.
See [commands and methodology](benchmarks/README.md).

## Development

```sh
# Verify checked-in core extension inputs without network access.
python3 scripts/core_extensions_corpus.py

# Execute package tests with each supported compiler.
go test -count=1 -timeout=10m ./...
llgo test -O3 -lto=full -count=1 -timeout=10m ./...
```

Merging requires all23 required CI jobs,240 execution reports and31 per-package
coverage gates. CI distributes all32 root test packages once across eight LLGo shards per platform,
retaining `-O3 -lto=full`, with a separate CLI build, and separates Go semantics, race and quality work. It
caches compiler-specific package objects and records LLGo build traces; the
target is 5–10 minutes per job, subject to measured final-run acceptance.

Some upstream inventories are fetched and hash-verified by their corpus scripts;
follow [ci.yml](.github/workflows/ci.yml) to reproduce the complete matrix.
A selected-corpus pass, a functional implementation and milestone acceptance
remain distinct.

The development order is: close functional conformance and integration gates;
complete declared-world conformance closeout; optimize interpreter, native
helper/GC/shared-memory paths and host boundaries using phase-separated
comparisons; then satisfy performance and release gates. The
[roadmap](docs/roadmap.md) retains all deliverables and open gates.

| Documentation | Purpose |
| --- | --- |
| [Embedding](docs/embedding.md) / [CLI](docs/cli.md) | Lifetimes, capabilities and usage |
| [Architecture](docs/design/architecture.md) / [packages](docs/design/packages.md) | Runtime contracts and dependency direction |
| [Core compiler](docs/design/core-compiler.md) / [native ABI](docs/design/native-abi.md) | IR, slots, code generation, roots and resumption |
| [Core extensions](docs/design/core-extensions.md) | Feature gates, proposal pins and inventories |
| [Standards audit](docs/research/status-2026-11-01.md) / [performance audit](docs/research/performance-2026-11-01.md) | Dated evidence and comparisons |
| [Go/LLGo CI](docs/design/llgo-ci.md) / [validation](docs/validation.md) | Toolchain pins and observed results |
| [Development plan](docs/roadmap.md) | Remaining work and acceptance gates |

## License and provenance

Project code is licensed under the [MIT License](LICENSE). Third-party code,
fixtures and conformance corpora retain their respective upstream licenses,
notices and immutable manifests. See each inventory, including the
[Core 2 provenance](testdata/core2/README.md#provenance-and-notices).
