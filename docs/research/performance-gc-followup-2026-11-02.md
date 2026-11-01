# Native heap traffic follow-up: 2026-11-02

This increment reduces state-transfer overhead, repeated heap lookups and
store-to-load traffic. **Competitive acceptance remains open.** Long-loop LLGo
struct access is 1.26x Wasmtime and indirect access is 1.80x; allocation and the
interpreter still have large deficits.

## Sources and method

Base `afc8f2e2a8c308a22150b5760d6455f949947197` is the validated head of merged
PR #88. Merge `f095340` has the same source tree. Measured head `384af8e` includes
the follow-up implementation and the merged baseline. Documentation changes
after that head do not change the measured engine sources.

Both engines were archived independently, with only common phase-harness
instrumentation for features/provenance. The host, pins and ordinary fixtures
match the [preceding comparison](performance-gc-2026-11-02.md): Apple M4 Max,
macOS arm64, Go 1.27.3, LLGo `8e45ed5` with `-O3 -lto=full`, Wasmtime 48.0.1
Cranelift and WasmEdge 0.17.1 interpreter. Reference library hashes were checked
against the original SDK inventory. Wazero 1.12.0 and Wasmer 7.3.0 still reject
the GC modules; their actual errors are retained without synthetic timings.

The ordinary survey retains all four own base/head compiler/backend pairs,
three fresh process blocks, two warmups and seven interleaved warm calls per
process, plus independent alternate-seed checks. A separately declared longer
experiment repeats i31, struct and indirect fixtures for **1,000,000 iterations**
under Go/LLGo native and Wasmtime, with the same protocol. Expected checksums
come from the existing independent oracle; Wasm binaries are unchanged. Longer
and ordinary observations remain separate. The statistic is the median of three
process medians, not a pool of dependent calls.

Startup/readiness, load, setup, validate/compile, instantiate, lookup, first call,
warm calls and cleanup have separate timers; no control or startup value is
subtracted. Raw phase/sample/checksum records, process orders, source/compiler/
binary/SDK hashes and unsupported outcomes are in the
[evidence](evidence/performance-gc-followup-2026-11-02.json). The complete source
inventories, transcripts and scripts remain under
`artifacts/native-gc-cache-20261102`, with hashes recorded in that evidence.
Shared-host timing and three blocks do not establish a controlled equivalence
or interpreter non-regression verdict.

## Results

**LLGo native, ns per source iteration; lower is better.** The longer loops
reduce the influence of fixed call/allocation costs and short-call noise.

| 1,000,000 iterations | Base | Head | Wasmtime |
| --- | ---: | ---: | ---: |
| i31 round trip | 2.511 | 2.532 | 1.796 |
| Struct set/get | 2.984 | 1.306 | 1.038 |
| Indirect array/struct access | 14.983 | 4.551 | 2.534 |

The original **10,000-iteration** configurations remain separately visible:

| Workload | Base | Head | Wasmtime |
| --- | ---: | ---: | ---: |
| Struct set/get | 4.204 | 2.529 | 1.600 |
| Linked allocation | 665.021 | 598.979 | 107.900 |
| Indirect array/struct access | 14.875 | 5.717 | 2.900 |
| Private-memory control | 1.200 | 0.975 | 1.400 |

Go results, interpreter results, process-level dispersion and every preparation
phase remain in the evidence. Small changes in unchanged controls must not be
reported as calibrated wins. The allocation fixture retains a linked chain; it
does not characterize collector scaling or peak-heap behavior.

## Mechanisms and limits

The entry adapter shares one implementation for `Enter` and combined `Run`.
The latter publishes state and snapshots the result using the existing entry
mutex transitions. Lease/memory exclusion, close/cancel behavior and reference
publication remain intact. This reduces helper-transition overhead without
changing the native ABI or introducing another execution framework.

The existing arm64 register allocator caches heap-view words. Up to two spare
register pairs also cache checked handles and object addresses for one native
entry. Every hit checks the entire non-null handle; every miss checks bounds
and generation. Entry resets discard addresses after helpers, calls, collection
and relocation. Cache registers are excluded from guest and inline-leaf
allocation. Precise entries and amd64 retain checked uncached lookups.

Both emitters forward a successful struct/array store into a following read of
the same object, field/index and layout. Only no-ops or self copies may
intervene, within a bounded window and the same poll interval. Packed signed/
unsigned loads, 32-bit normalization, reference handles and all v128 bits retain
their original semantics. Each interior PC keeps a complete checked cold body;
branches/resumption cannot assume an earlier store. Opcode checks prevent shared
source locations from classifying control exits as memory effects.

Tests cover field widths, reused result slots, changed references, every cold
continuation, null/generation/index failures, source/control boundaries,
cancel/close races and forced guest collection. Full local Go/LLGo suites pass
all 31 test packages; final target CI and controlled performance acceptance
remain required. Further work includes native allocation/helper costs, residual
indirect traffic, amd64 lookup optimization and interpreter throughput.
