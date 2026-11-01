# Performance gap audit: 2026-11-01

The owned engine is functionally broad, but **performance parity is not achieved**.
LLGo native scalar loops are competitive in this local portfolio; GC, shared
memory, bulk-memory and several component/WASI paths have large deficits.
A favorable overall average would conceal these failures and is not an
acceptance verdict. The [standards audit](status-2026-11-01.md) independently
records implementation scope and conformance evidence.

## Snapshot and measurement boundaries

- Engine: `4ad69c24de32a96310799ad76bd10af118afed42` (merged PR #84), archived
  unchanged for both host compilers. Harness-only modern-feature/P1 adapters are
  retained separately, with engine source hashes checked unchanged.
- Host: Apple M4 Max, macOS arm64, shared workstation. Core starting/ending load
  averages and full platform identity are retained. No CPU affinity, isolation,
  frequency lock or controlled A/A calibration is claimed.
- Go 1.27.3; LLGo `8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f`, `-O3 -lto=full`.
- References: wazero 1.12.0 interpreter/native; Wasmtime 48.0.1 Cranelift speed;
  Wasmer 7.3.0 Cranelift; WasmEdge 0.17.1 **interpreter**. SDK pins, linked library
  hashes, commands, compiler identities and notices are retained. The system's
  older Wasmer CLI was not used. WasmEdge AOT/JIT was not timed.
- Core: 68 existing fixtures, eight source-manifest families, nine configurations,
  three fresh processes per pair, first call + two warmups + seven warm samples:
  **1,836 processes / 18,360 checked guest calls**. Warm calls are serially
  randomized/interleaved; setup never runs concurrently with a timed call.
- Modern: four GC and six private/shared/atomic/wait fixtures; three process
  blocks plus alternate-seed checks. Twelve actual GC validation rejections
  from wazero/Wasmer are retained as unsupported, without synthetic timings.
- P1: three actual WASI host-call loops plus a compute control, eight
  configurations, three processes: 96 successful process records and 1,280
  checked calls including alternate inputs. Components/P2/async use three fresh
  processes per reported configuration. Final native comparisons randomize
  Go/LLGo/reference process order within each workload/block; initial whole-suite
  rounds are retained separately. Existing Go benchmark and C/Rust adapters
  retain their explicit boundary differences.

Core tables use public embedding API calls, including unavoidable admission,
marshalling and result handling. Every compute call executes the same binary,
seed and iteration count and checks its output. Stateful fixtures reset inside
the timed call. Preparation, instantiation, first use, warm execution and cleanup
are measured independently; no startup/control median is subtracted.

**Units:** core/modern rows are ns per source iteration, not per Wasm/native
instruction. Source-call loops may be inlined. Component/WASI units are identified
per row; batch values are divided by the actual batch count. Small call controls
are clock-resolution limited, particularly the C clock on this macOS host.
Near-parity differences do not establish statistically calibrated parity.

## Core families and named competitors

Ratios are wasmvm LLGo / named reference; **lower is better**, 2x means twice the
elapsed time. Each cell is a geometric mean within that existing source-manifest
family. Empty-call and growth-only controls are excluded from family aggregates.
These diagnostic families are not the still-unaccepted application-weighted M2
portfolio; there is deliberately no synthetic fastest-engine aggregate.

| Family | Cases | Native / wazero | Native / Wasmtime | Native / Wasmer | Interpreter / wazero | Interpreter / WasmEdge |
| --- | --- | --- | --- | --- | --- | --- |
| core | 5 | 0.724x | 1.045x | 1.016x | 1.133x | 1.782x |
| float | 10 | 0.718x | 1.003x | 1.004x | 0.962x | 1.606x |
| global | 5 | 1.653x | 0.619x | 0.550x | 0.810x | 1.417x |
| memory | 3 | 0.683x | 0.973x | 0.954x | 1.313x | 2.220x |
| runtime | 7 | 1.072x | 1.194x | 1.337x | 1.679x | 2.992x |
| table | 3 | 0.135x | 0.195x | 0.232x | 0.831x | 1.270x |
| vector-float | 11 | 1.059x | 1.093x | 1.035x | 1.081x | 3.130x |
| vector-integer | 22 | 1.093x | 1.191x | 1.107x | 0.956x | 2.667x |

Native scalar arithmetic is close to Cranelift and often ahead of wazero. SIMD
still has systematic gaps. Fast table cases coexist with a severe bulk-memory
outlier in the runtime family; averages must not excuse that outlier. The LLGo
interpreter trails WasmEdge across many cases and also loses individual cases
to wazero. All 68 cases and both own compilers/backends are retained in the
[complete Core table](performance-2026-11-01-core.md).

Representative native rows, ns/source iteration:

| Workload | Go native | LLGo native | wazero native | Wasmtime | Wasmer |
| --- | --- | --- | --- | --- | --- |
| core/integer-sustained | 1.545 | 1.551 | 1.589 | 1.534 | 1.549 |
| core/memory-sustained | 0.512 | 0.507 | 0.831 | 0.470 | 0.520 |
| core/function-calls | 0.949 | 0.852 | 1.499 | 0.927 | 0.820 |
| runtime/indirect-calls | 1.719 | 1.642 | 2.194 | 1.400 | 1.533 |
| memory/narrow-unaligned | 0.710 | 0.697 | 1.051 | 0.692 | 0.711 |
| runtime/bulk-overlap | 2597 | 3388 | 184.58 | 155.00 | 159.00 |
| table/table-roundtrip | 1.440 | 1.305 | 14.79 | 10.61 | 5.060 |
| vector-integer/shuffle | 2.312 | 2.346 | 1.654 | 1.800 | 1.900 |
| vector-float/rounding-f32x4 | 9.829 | 8.504 | 9.062 | 7.800 | 8.000 |

The bulk case performs an 8 KiB fill and overlapping 8 KiB copy on every source
iteration. Its deficit is not explained by process startup or first use.
The corresponding LLGo interpreter result is also slow; switching backend alone
is not a solution.

## GC and Core Threads

The WasmEdge column remains an interpreter comparison. Unsupported GC is not a
competitive speed result. All own native cases actually select the owned native
backend; runtime helpers execute individual operations without interpreting a
whole guest function. Nonetheless, native helper traffic can make a compiled
function slower than the interpreter.

| Workload | Go native | LLGo native | wazero native | Wasmtime | Wasmer | WasmEdge interp |
| --- | --- | --- | --- | --- | --- | --- |
| gc/i31-roundtrip | 3.600 | 3.433 | unsupported | 2.100 | unsupported | 45.40 |
| gc/struct-access | 288.57 | 1342 | unsupported | 1.600 | unsupported | 55.20 |
| gc/struct-allocation | 375.89 | 1259 | unsupported | 105.90 | unsupported | 106.70 |
| gc/struct-indirect-access | 809.10 | 3460 | unsupported | 3.100 | unsupported | 95.50 |
| threads/private-memory | 1.183 | 1.137 | 1.242 | 1.400 | 1.700 | 39.50 |
| threads/shared-memory | 251.09 | 870.67 | 2.179 | 1.500 | 1.800 | 40.00 |
| threads/atomic-rmw | 129.51 | 433.00 | 2.004 | 2.600 | 4.600 | 33.50 |
| threads/notify-empty | 127.11 | 433.55 | 21.26 | 7.000 | 5.500 | 40.50 |
| threads/wait-mismatch | 126.70 | 425.61 | 25.41 | 9.100 | 3.100 | 44.10 |
| threads/wait-timeout | 128.25 | 422.65 | 246.21 | 43.30 | 169.40 | 96.40 |

The mutable-field and indirect GC cases expose deficits approaching three orders
of magnitude against Wasmtime. Linked allocation is a different workload with
a smaller, but still substantial, gap. It does not measure full-GC pause times
or memory reclamation under sustained pressure. Shared ordinary accesses and
uncontended atomics likewise have large per-operation overhead. Wait mismatch,
zero timeout and empty notify are separate operations; none establishes
contended scaling, wakeup tail latency or thread-creation throughput.

## WASI P1 host-call boundary

Each iteration calls the actual provider, checks errno, reads the guest-memory
result and updates a checksum. Args are `[guest,arg]`, environment is
`KEY=VALUE`; no ambient environment is inherited. Times are ns/source iteration,
with compile/instantiate/first-call/close retained separately.

| Workload | Go native | LLGo native | wazero native | Wasmtime | Wasmer |
| --- | --- | --- | --- | --- | --- |
| args-sizes | 454.32 | 1808 | 22.33 | 51.70 | 41.40 |
| environ-sizes | 452.35 | 1792 | 22.14 | 50.80 | 57.80 |
| args-copy | 481.06 | 1994 | 30.70 | 65.50 | 87.70 |

This confirms substantial P1 call-boundary/provider overhead, particularly under
LLGo, even for in-memory argument/environment operations. It is not a filesystem
or network benchmark. The first Wasmer compute-control adapter incorrectly
created an uninitialized WASI environment; its cleanup aborted. That harness
failure is retained. The corrected adapter creates the environment only for
importing modules; the entire final matrix, including the control, passed.
No engine crash is classified as an unsupported capability or successful sample.

## Component, WASI P2 and async

Times below are ns per named operation, or per complete API call where indicated.
The reference is Wasmtime 48.0.1. The byte transfer uses its **typed Rust API**,
not the much slower dynamic C byte-list expansion. TCP/UDP/HTTP and composed
commands use the Rust reference with actual providers; simple operations use
the pinned C component API. Go and LLGo results use the public owned API.

| Operation / unit | Go native | LLGo native | Wasmtime | LLGo / reference | Reference min–max |
| --- | --- | --- | --- | --- | --- |
| Component scalar API | 813.00 | 2664 | 267.82 | 9.95x | 246.13–269.68 |
| Guest resource new/rep/drop | 1450 | 3488 | 27.23 | 128.09x | 26.91–27.54 |
| Owned resource, four public calls | 5666 | 16641 | 1258 | 13.23x | 1242–1272 |
| Packed 16 KiB bytes, typed reference | 5991 | 14351 | 479.82 | 29.91x | 454.48–585.10 |
| P2 arguments | 5643 | 15623 | 480.00 | 32.55x | 467.00–499.00 |
| P2 secure random, 16 KiB | 9537 | 71456 | 39835 | 1.79x | 39412–39986 |
| P2 wall clock | 1467 | 5049 | 357.00 | 14.14x | 346.00–368.00 |
| P2 expired clock subscription | 4270 | 28566 | 272.23 | 104.93x | 268.88–275.30 |
| P2 output-stream write, 4 KiB | 3768 | 9013 | 8275 | 1.09x | 8179–8310 |
| P2 filesystem stat | 5061 | 13546 | 6017 | 2.25x | 5965–6364 |
| P2 numeric-IP lookup/resource lifetime | 11961 | 36973 | 977.66 | 37.82x | 951.86–1005 |
| P2 TCP round trip, 4 KiB each way | 49011 | 64834 | 29519 | 2.20x | 29051–29777 |
| P2 UDP round trip, 4 KiB each way | 68000 | 108381 | 46171 | 2.35x | 45271–48580 |
| P2 HTTP fields lifetime | 36864 | 94056 | 2317 | 40.59x | 2299–2347 |
| P2 HTTP 4 KiB body lifetime | 19271 | 49082 | 1274 | 38.53x | 1268–1277 |
| P2 C command, warm reusable instance | 44131 | 139218 | 7587 | 18.35x | 7552–8600 |
| P2 composed Rust command, warm instance | 54957 | 191142 | 11830 | 16.16x | 8028–22491 |
| Async direct (0 callbacks/call) | 1824 | 6496 | 775.17 | 8.38x | 706.46–775.50 |
| Async run (129 callbacks/call) | 131889 | 504851 | 40609 | 12.43x | 40574–41113 |

The generic per-element 16 KiB byte API is additionally much slower than the
packed API; both raw results are retained. Different entropy algorithms/provider
policies apply to random generation, so that row does not rank cryptographic
primitive speed. TCP/UDP use loopback, 128 round trips/batch and 4 KiB in each
direction; connection setup/teardown is excluded. HTTP field/body lifetimes are
not full network requests. Warm command rows reuse instances; they are not cold
CLI startup. Async `run` includes 129 callbacks/call and the public/task boundary;
no empty-call subtraction is used to invent a pure callback latency.

Go and LLGo have different host costs; the measured native guest arithmetic
does not predict the speed of a provider or component call. Resource, poll, numeric-IP lookup, canonical transfer and
HTTP lifetime paths still fail the desired across-family parity target. Reference sample ranges remain visible: workstation/process scheduling makes
some rows noisy, and three processes do not calibrate small differences. These
are selected boundary measurements, not exhaustive P2/P3 performance coverage.

## Startup, preparation, first use and cleanup

The sustained integer fixture below executes 5,000,000 iterations in each call.
All entries are **microseconds**; the warm column is one complete API call, not
ns/iteration. Preparation is the actual combined validate/compile API interval;
no second standalone validator call is added. Cleanup includes explicit owned
Store reclamation. A first call is measured separately from warm execution.

| Configuration | Ready | Load | Engine/store | Prepare | Instantiate | Lookup | First | Warm call | Cleanup |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| wasmvm-go-interpreter | 4552 | 40.71 | 6.292 | 68.67 | 7.500 | 0.417 | 327882 | 331228 | 7.375 |
| wasmvm-go-native | 4513 | 35.42 | 5.666 | 186.46 | 7.000 | 0.333 | 7698 | 7725 | 9.834 |
| wasmvm-llgo-interpreter | 88504 | 46.88 | 3.833 | 57.29 | 8.209 | 0.625 | 325729 | 327828 | 1.667 |
| wasmvm-llgo-native | 104566 | 45.42 | 2.375 | 508.67 | 10.46 | 0.958 | 7726 | 7756 | 10.83 |
| wazero-interpreter | 4852 | 34.38 | 2.250 | 41.17 | 14.42 | 0.209 | 409627 | 412357 | 14.42 |
| wazero-compiler | 4741 | 34.08 | 51.46 | 107.17 | 13.54 | 1.834 | 7651 | 7944 | 11.08 |
| wasmtime-cranelift | 4613 | 30.00 | 113.00 | 559.00 | 7.000 | 4.000 | 7698 | 7672 | 112.00 |
| wasmer-cranelift | 6390 | 33.00 | 35.00 | 1058 | 16.00 | 4.000 | 7785 | 7747 | 45.00 |
| wasmedge-interpreter | 16676 | 40.00 | 25.00 | 34.00 | 22.00 | 1.000 | 243330 | 243776 | 8.000 |

Internal dispatch diagnostics separately retain decode, validate, lower, slot
allocation and native compilation/publication groups. They reuse frames and
bypass public admission/marshalling, so they cannot replace public-API results.
Native optimization/emission/publication are not individually observable through
all peer APIs in this audit; those finer costs remain unmeasured rather than
being inferred by subtracting totals. OS file/executable caches are not flushed.

## Why the gaps are architectural

The following source observations explain plausible causes; they are not a
profile-based attribution of every measured nanosecond.

1. **Per-operation native exits.** The owned
   [native execution loop](../../internal/engine/nativeexec/execution.go)
   publishes state, enters generated code, reads state and returns a runtime
   helper request for many GC/shared/bulk operations. The common execution loop
   executes that operation and reenters native code. Repeating this for cheap
   field loads or atomics magnifies ABI/state/cancellation costs. Go-to-LLGo
   differences on the same guest instruction selection show that host helpers
   and transitions also matter.
2. **Shared-memory synchronization.**
   [Shared-access lowering](../../internal/compiler/codegen/threads.go) marks
   each shared memory access for a helper; one Engine domain lock orders all
   shared access. This is a deliberate conservative memory-model contract, but
   is expensive even without contention. Wasmtime lowers normal accesses and
   atomics directly through its compiler IR; it calls runtime support for waits
   and other operations that actually need it.
3. **GC field access.** Wasmtime's
   [GC lowering](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/cranelift/src/func_environ/gc.rs)
   derives field offsets, emits checked heap addresses and machine loads/stores,
   with barriers where required. It does not return to a host dispatcher for
   every scalar struct field. The indirect fixture prevents explaining the
   entire deficit solely as elimination of a single nonescaping struct.
4. **Bulk primitives.** The owned
   [bulk implementation](../../internal/engine/execute/bulk.go) preserves bounds,
   overlap and bounded cancellation through 4 KiB chunks and polls. Reference
   engines can lower copy/fill to efficient native primitives. The measured
   deficit includes these different interruption policies; optimizing it must
   preserve the declared cancellation/overlap behavior.
5. **Canonical/host paths and LLGo.** Resource ownership, marshalled Values,
   per-call admission, lifecycle and readiness/provider boundaries are visible
   in the component results. Their amplification under LLGo warrants profiles
   and allocation/transition diagnostics, not an assumption that native guest
   instructions alone determine end-to-end throughput.

## Priority and remaining measurements

1. Implement direct native GC field/array fast paths and precise root/barrier
   contracts; reduce helper exits and repeated state traffic.
2. Replace per-access shared-memory host exits with a correct native memory-model
   strategy, including mixed widths, grow, atomics and cancellation; then measure
   contention/scaling and wait/notify tails on both ISAs.
3. Improve overlap-safe bulk copy/fill and P1 host boundaries without weakening
   resource limits, cancellation or errno/trap semantics.
4. Reduce Canonical ABI/resource/poll/lookup overhead, profile LLGo helper/runtime
   costs and compare the same signatures/ownership policies. Optimize SIMD
   selection/register allocation and interpreter dispatch with the complete
   affected family visible.
5. Close performance acceptance with realistic compute applications from at least
   two pinned guest toolchains, broader large/small compile portfolios, GC pressure,
   cold CLI/application startup, RSS/code size, scaling and cancellation latency.
   The current C/Rust command consumers are integration workloads, not a substitute
   for those compute/scaling portfolios.

Wago, WAMR, browser engines, Wasmer's other compilers and WasmEdge AOT/JIT are
not timed here. Neither their published speed claims nor missing samples are
substituted into the measured matrix. Linux amd64 results must be kept separate
from this Mac table; the opt-in CI survey produces a fresh platform-specific
artifact. Controlled acceptance and calibrated parity on both accepted platforms
remain open. No performance requirement is relaxed by this audit.

## Reproduce and inspect evidence

The existing Runtime benchmarks workflow accepts `suite=survey`: it reuses its
pinned Go/LLGo and SDK setup, runs the 68-case Core and 10-case modern matrix,
and uploads phase records and complete Markdown tables. It does not run the
unrelated sustained integer acceptance gate. Existing PR and full scopes remain
available. This broad survey is opt-in, not an additional default semantic job.

Locally, use the SDK roots from the checksum-pinned tool manifests:

```sh
export WASMVM_LLGO=/path/to/pinned/llgo
export WASMVM_LLGO_REVISION=8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f
python3 benchmarks/survey.py --output artifacts/survey \
  --wasmtime-sdk "$WASMTIME_SDK" --wasmer-sdk "$WASMER_SDK" \
  --wasmedge-sdk "$WASMEDGE_SDK"
python3 benchmarks/survey_modern.py --core-output artifacts/survey
python3 benchmarks/modern/verify.py
```

Component/P2/async probe sources and boundaries are documented in
[the component benchmark guide](../../benchmarks/components/README.md).
The compressed [evidence bundle](evidence/performance-2026-11-01.tar.gz) retains
raw Core/modern/P1/dispatch samples, all component logs, build/source/library
identities, protocol orders, manifests, reproduction drivers and failed initial
harness diagnostics. It excludes executable binaries and credentials. Its
adjacent checksum manifest records the archive identity. Every complete matrix
and checksum is audited before publication; no passing timing is inferred from
an incomplete run.
