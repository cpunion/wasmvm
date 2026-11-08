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

Performance acceptance remains open. The
[latest compiler baseline and complete comparison](docs/research/performance-llgo-interface-upgrade-2026-11-08.md)
retains all 22 Linux workloads, sustained GC and an independent ARM64 old/new
LLGo event against wazero, Wasmtime, Wasmer, WasmEdge and native C where applicable.
The LLGo pin includes upstream static-interface and concurrent-lookup fixes;
Go 1.27.3, LLVM 22.1.8 and `-O3 -lto=full` remain fixed. No broad VM speedup
is established. Public-call latency, memory/SIMD, GC, Threads and native-C gaps
remain explicit, alongside adverse and unsupported observations.

Linux controls resolve 6/22 original cases and 3/4 sustained GC cases; ARM64
resolves 2/22 and 0/4. Hosts and compiler events stay separate. Setup, compilation,
instantiation, first call, warmed calls, cleanup and process intervals are
recorded separately. Merging requires all 23 semantic CI jobs, 240 conformance
reports and 30 per-package coverage gates. Jobs target 5–10 minutes. Historical
[main results](docs/research/performance-main-inventory-2026-11-08.md),
[native-entry investigation](docs/research/performance-native-entry-2026-11-08.md)
and rejected experiments retain their original source/compiler identities.

## Project status

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

CI runs Go and LLGo on both targets; eight LLGo shards execute all 31 test
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

The [2026-11-08 Linux comparison](docs/research/performance-reference-kind-filter-2026-11-08.md)
contains the latest completed full 22-case source comparison. The subsequent
[checked-slot GC event](docs/research/performance-gc-slot-reuse-2026-11-08.md)
compares the affected four GC cases with main, including a separate sustained
cohort. Its [hosted results](https://github.com/cpunion/wasmvm/pull/127#issuecomment-6716307382)
resolve 2/4 sustained controls. Both records retain independent Go/LLGo,
backend, competitor and lifecycle results; controlled acceptance remains open.
The [guarded array-cache follow-up](docs/research/performance-array-cache-bounds-2026-11-08.md)
records subsequent compiler changes, preserved adverse pilots and separate
AMD64 measurements linked from its PR.

The [2026-11-05 source event](docs/research/performance-loop-invariant-literals-2026-11-05.md)
provides a 22-case local arm64 table and published Go/LLGo code audit.
Its A/A controls resolve **0/22** workloads, so no performance gate closes.
The [preceding hosted comparison](docs/research/performance-loop-scratch-registers-hosted-2026-11-05.md)
retains the independent amd64 cohort. Results from different sources/hosts
must be read separately.

**Performance parity is not achieved.** The 2026-11-01 audit measures engine
`4ad69c2` on Apple M4 Max, macOS arm64, Go 1.27.3 and LLGo `8e45ed5` with
`-O3 -lto=full`. It includes 68 Core fixtures, nine configurations, three fresh
processes per pair and seven warm samples. These are diagnostic workstation
results, not controlled release benchmarks or measurements of the new extensions.

Representative warm medians, **ns per source iteration; lower is better**.
All columns below use native execution:

| Workload | wasmvm / Go | wasmvm / LLGo | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| Integer loop | 1.545 | 1.551 | 1.589 | 1.534 | 1.549 |
| Function calls | 0.949 | 0.852 | 1.499 | 0.927 | 0.820 |
| Memory loop | 0.512 | 0.507 | 0.831 | 0.470 | 0.520 |
| SIMD shuffle | 2.312 | 2.346 | 1.654 | 1.800 | 1.900 |
| 8-KiB fill + overlapping copy | 2,597 | 3,388 | 184.58 | 155.00 | 159.00 |
| GC mutable struct access | 288.57 | 1,342 | Unsupported | 1.600 | Unsupported |
| Shared-memory access | 251.09 | 870.67 | 2.179 | 1.500 | 1.800 |
| P1 argument-size host call | 454.32 | 1,808 | 22.33 | 51.70 | 41.40 |

Scalar native loops are competitive in these cases; bulk memory, guest GC,
shared memory and host/component boundaries have large gaps. Across measured
families, LLGo interpreter time is **1.27–3.13x** WasmEdge interpreter time and
**0.81–1.68x** wazero interpreter time. Family averages do not establish parity
for every case; tiny near-parity differences are not statistically calibrated.

Startup, load, validate/compile, instantiate, first call, warm calls and cleanup
are measured separately. Startup is excluded from warm rows; no empty-call
median is subtracted. Component/P2/async paths use Wasmtime as their reference
and retain substantial deficits in resources, canonical transfer, HTTP and
async boundaries. See the [full audit](docs/research/performance-2026-11-01.md),
[all 68 Core cases](docs/research/performance-2026-11-01-core.md) and
[raw evidence](docs/research/evidence/performance-2026-11-01.json).

The [2026-11-02 GC follow-up](docs/research/performance-gc-2026-11-02.md) measures
direct native field access against merged PR #87. LLGo struct access improves
from 1303 to **4.75 ns/iteration** and indirect access from 3376 to **15.26**;
Wasmtime measures 1.70 and 3.10 on the same run. Allocation and interpreter
deficits remain open. This targeted update does not replace the dated full audit.

A [further heap-traffic comparison](docs/research/performance-gc-followup-2026-11-02.md)
uses separately declared million-iteration loops: LLGo struct access reaches
1.31 ns/iteration versus Wasmtime's 1.04, and indirect access reaches 4.55 versus
2.53. Allocation, interpreter and full-portfolio performance acceptance remain open.

The [bulk-fill follow-up](docs/research/performance-bulk-2026-11-02.md) reduces the
8-KiB fill plus overlapping-copy workload from 2,638 to 431 ns/iteration under
the LLGo interpreter and from 3,261 to 1,208 under LLGo native. Wasmtime measures
155; native runtime-helper transitions and competitive acceptance remain open.

The [direct native bulk comparison](docs/research/performance-native-bulk-2026-11-02.md)
removes those private fill/copy helper transitions. With 100,000 iterations,
LLGo measures 142.6 ns/iteration versus Wasmtime's 146.9 and a separate native
C/libc loop's 146.2. These shared-host results do not establish controlled
superiority or full-portfolio acceptance.

The [native allocation follow-up](docs/research/performance-native-allocation-2026-11-02.md)
uses 63 warm calls of 100,000 linked allocations and includes every GC stall.
LLGo improves from 582.3 to 60.2 ns/object versus Wasmtime's 60.17; Go reaches
43.5. These are process-mean diagnostics, with raw latency distributions retained.
Interpreter, shared-memory and full-portfolio performance acceptance remain open.

The [GC interpreter follow-up](docs/research/performance-gc-interpreter-2026-11-02.md)
reduces LLGo i31, struct and indirect access from 552/591/1,200 to
62/78/129 ns/iteration; WasmEdge measures 45/55/95. Linked allocation including
all GC pauses improves from 588 to 171 ns/object versus WasmEdge's 104.
These targeted workstation results retain the remaining interpreter deficit.

The [native Threads follow-up](docs/research/performance-native-threads-2026-11-02.md)
adds direct shared accesses, all atomic widths and checked wait/notify paths on
amd64 and arm64, with stable shared growth and concurrent-entry ownership.
Targeted comparisons retain baseline and interpreter regressions; controlled
full-portfolio acceptance remains open.

The [interpreter memory follow-up](docs/research/performance-interpreter-memory-2026-11-02.md)
removes an unconditional LLGo result allocation and reduces scalar frame storage.
Private reads/writes improve from 90 to 59 ns/iteration and SIMD widening from
163 to 129; WasmEdge measures 40 and 63. Remaining deficits and both comparison
campaigns are retained.

The [dispatch and three-way gap audit](docs/research/performance-interpreter-dispatch-2026-11-02.md)
compares named engines, Go versus LLGo, and six ordinary native C kernels.
Reusable frames and cancellation polling reduce LLGo guest calls from 133 to
93 ns/iteration; WasmEdge remains faster at 62. Memory/compiler gaps, empty-call
regressions and rejected experiments remain recorded, with independent LLGo
reproductions reported upstream.

Benchmark CI is opt-in through the `run-benchmarks` PR label or
[workflow dispatch](https://github.com/cpunion/wasmvm/actions/workflows/benchmarks.yml).
It compares exact base/head sources and pinned competitors and posts results
to the PR. See [benchmark commands and methodology](benchmarks/README.md).

## Development

```sh
# Verify checked-in core extension inputs without network access.
python3 scripts/core_extensions_corpus.py

# Execute package tests with each supported compiler.
go test -count=1 -timeout=10m ./...
llgo test -O3 -lto=full -count=1 -timeout=10m ./...
```

CI distributes all 31 test packages once across seven LLGo shards per platform,
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
