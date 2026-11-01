# Native GC field access: 2026-11-02

Direct native heap access removes most of the previous field-access cost.
**Competitive performance acceptance remains open:** LLGo is still 2.79x
Wasmtime on struct access, 4.92x on indirect access and 6.55x on linked
allocation in this short workstation run. The interpreter remains much slower.

## Sources and measurement

Base `ea6dfc419cf64e12b8efc71e862981b1860144d3` is merged PR #87.
Head `66a931144fb61974f2702b82780043362d00b634` contains this optimization.
Both engine revisions were independently archived and hashed. Only phase-runner
instrumentation enables `wasm.SupportedFeatures` and sets provenance; no engine
implementation files are overlaid. Subsequent documentation changes do not
change these measured engine sources.

The host is an Apple M4 Max, macOS arm64, Go 1.27.3 and LLGo
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f`, `-O3 -lto=full`.
The existing pinned SDK adapters provide Wasmtime 48.0.1 Cranelift and WasmEdge
0.17.1 interpreter. Wazero 1.12.0 and Wasmer 7.3.0 actually reject each of the
four GC modules; all twelve rejection records remain separate from timings.
The private-memory control also runs successfully on these engines.

The five unchanged fixtures are the first five entries of
[`benchmarks/modern/manifest.json`](../../benchmarks/modern/manifest.json).
Three fresh process blocks use `warm-call-interleave-v1`, two warmups and seven
serially interleaved warm calls; the statistic below is the median of the three
process medians. Every call checks its checksum. Each configuration also checks
the alternate seed with 17 iterations: 212 successful processes total, plus
twelve unsupported-validation processes. Allocation retains a linked chain and
is not a collector/scaling benchmark.

Startup/readiness, load, engine setup, validate/compile, instantiate, export
lookup, first call, warm execution and cleanup have separate timers. No startup
or control value is subtracted. These short shared-host measurements have no
CPU isolation, calibrated A/A experiment or equivalence confidence interval.
Small control differences and potential interpreter regressions are unresolved.
The [evidence](evidence/performance-gc-2026-11-02.json) retains all individual
timings, process orders, compiler/binary/source/fixture hashes, rejection records
and alternate-seed checks. Full source inventories, protocol transcripts and the
orchestration script remain in `artifacts/native-gc-20261102`; their hashes are
included in the evidence. The script reuses `own_build.archive`,
`phase_ci.validate_observation` and `interleaved.invoke` from the existing suite.

## Warm native execution

**ns per source iteration, lower is better.** These are complete guest loops,
not individual Wasm instructions. The struct loop has both a get and a set.

| Workload | Go base | Go head | LLGo base | LLGo head | Wasmtime |
| --- | ---: | ---: | ---: | ---: | ---: |
| i31 round trip | 4.696 | 4.400 | 3.583 | 3.625 | 2.200 |
| Struct access | 287.050 | 5.600 | 1303.412 | 4.750 | 1.700 |
| Linked struct allocation | 378.579 | 232.588 | 1177.242 | 671.704 | 102.600 |
| Indirect array/struct access | 804.062 | 16.608 | 3376.125 | 15.258 | 3.100 |
| Private-memory control | 1.704 | 1.246 | 1.333 | 1.238 | 1.400 |

LLGo field access improves 274x and indirect access 221x. Linked allocation
improves 1.75x because its final field read no longer exits; allocation still
uses the checked runtime path. Independent candidate/reference compile and
instance timers remain visible rather than being inferred from CLI latency.

For example, struct-access median phase times are **microseconds per process**:

| Phase | LLGo base | LLGo head | Wasmtime |
| --- | ---: | ---: | ---: |
| Validate + compile | 613.9 | 629.2 | 751.0 |
| Instantiate | 12.6 | 12.9 | 19.0 |
| First call, 10,000 iterations | 13109.5 | 44.6 | 47.0 |
| Cleanup | 3.8 | 9.2 | 91.0 |

## Interpreter and remaining work

**ns per source iteration.** Portable payload ownership remains unchanged.

| Workload | Go base | Go head | LLGo base | LLGo head | WasmEdge interpreter |
| --- | ---: | ---: | ---: | ---: | ---: |
| i31 round trip | 108.979 | 109.083 | 612.242 | 611.342 | 45.200 |
| Struct access | 125.562 | 127.621 | 649.117 | 648.821 | 55.500 |
| Linked struct allocation | 207.017 | 212.504 | 630.304 | 657.633 | 106.300 |
| Indirect array/struct access | 218.896 | 224.154 | 1282.700 | 1280.700 | 95.500 |
| Private-memory control | 64.062 | 65.771 | 134.287 | 140.417 | 40.300 |

Allocation is 2.7% slower under Go and 4.3% under LLGo in this run. This limited
sample does not resolve the change as noise or a regression. It does not pass a
formal interpreter non-regression gate. Native relocation metadata stays out of
portable object size classes; dispatch/allocation profiling remains necessary.

Next native work should eliminate redundant checked lookups and field traffic
with valid alias/safepoint proofs, then reduce allocation and guest-call/host
helper costs. The existing [full audit](performance-2026-11-01.md) still governs
unremeasured shared memory, bulk memory, SIMD and WASI/component gaps. Neither
these improvements nor passing semantics close M2 or full-portfolio acceptance.
