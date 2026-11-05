# Direct native routing and portable primitives: 2026-11-05

**Semantic qualification and all five hosted groups pass; controlled performance
acceptance remains open.** The [comparison](https://github.com/cpunion/wasmvm/actions/runs/41941908335)
completed on 2026-11-05. Eight of 22 workloads resolve all four diagnostic A/A
controls; the other 14 remain inconclusive. These results apply to the routing
candidate below, not the later global-residency or value-coalescing branches.

Engine source: `59a76f933901e56266e618bff0a502364eca0350`. Hosted candidate:
`c92383e2b60fe75910ca921786c0288f49a11b12`. Hosted baseline:
`a47055833582068b88e9aeb51d922ef84662db4f`, the earlier fully CI-qualified
register-plan source, not a merged production release. Main remains
`4ce79899a1f3ed45163afa622cc408cc6a79a599`.

The source changes two production files after the held
[runtime-effects experiment](performance-runtime-effects-2026-11-05.md).
The native driver selects shared GC/SIMD/call/throw/return leaves directly,
avoiding the general dispatcher frame before their leaf frame. Portable SIMD,
nonallocating GC access and atomic/bulk operand routing remains direct, using
the existing canonical memory/state/SIMD operations. Shared complex effects,
source PCs, exact fuel, polls, roots, quota checks, owned arenas, frame
transfers and interruption retain their contracts. Scalar native fallback
continues to fail. No ISA, compiler pin or padding/alignment sweep occurs.

Ten complete affected packages pass pinned Go 1.27.3 and LLGo 3571
(`-O3 -lto=full`) on macOS arm64: 108 actual conformance reports, scoped Go
coverage 90.71%–100%, execute 92.16% (1046/1135). Race, cgocheck2 and no-cgo
checks pass. A dry build of the new hosted pipeline qualifies 274 source/reference
configurations across 22 fixtures; 12 GC/reference combinations retain explicit
unsupported results. These dry-run timings are not used as performance data.

The new opt-in `routing` workflow builds both source/compiler pairs and pinned
references once, then reuses that immutable artifact across five sequential
Linux amd64 jobs: Core, memory/SIMD, GC, Threads and C/Rust SHA. Every job has a
10-minute bound. C/Rust sources, exact guest bytes, producer provenance and the
independent SHA oracle are committed unchanged from the earlier application
experiment. All participants use the same bytes per fixture.

Each group predeclares five fresh process blocks, two warmups/seven serially
interleaved samples, base/head LLGo A/A for both backends, one inherited CPU
affinity and `GOMAXPROCS=1`. Each marginal 95% A/A interval must lie inside
±2.5% for this diagnostic calibration. That does not replace the controlled
milestone gate, simultaneous feature intervals, absolute effect floors or
required platform portfolio. Hosted CPU power and physical neighbor contention
remain uncontrolled; no registered self-hosted runner is currently available.
Do not pool independent groups or use component/unchecked-C costs to waive
complete API outcomes.

The six successful jobs take 5m11s (build), 2m44s (Core), 5m22s
(memory/SIMD), 29s (GC), 1m30s (Threads) and 1m59s (applications), excluding
queue time. The build compiles all four own binaries and references in 2m49s.
All 362 alternate-seed configurations validate, and all 12 declared GC/reference
rejections remain explicit. No successful cohort was rerun for a better result.

The table gives elapsed-time ratios; lower is better. Each value is the median
of five process medians of seven serially interleaved calls. It compares equal
fixtures within each job. WasmEdge runs its interpreter. Native C has a different
unchecked contract; a ratio below one is not a safety-equivalent native victory.
Pinned versions: wazero 1.12.0, Wasmtime 48.0.1, Wasmer 7.3.0, WasmEdge 0.17.1.
All remaining reference modes, own head/base values, seven independent phases,
readiness/wall intervals, controls and exact input identities are in the
[machine-readable audit](data/direct-runtime-routing-hosted-2026-11-05.json).

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 1.134× | 1.638× | 7.176× | 3.423× | 2.166× | 3.479× | 2.419× | unavailable | inconclusive |
| core/integer-loop | 0.999× | 0.746× | 1.241× | 1.194× | 1.196× | 0.691× | 0.893× | 1.241× | resolved, advisory |
| core/function-calls | 0.999× | 0.811× | 0.730× | 1.027× | 0.875× | 0.753× | 1.123× | 1.174× | resolved, advisory |
| core/memory-load-store | 0.998× | 0.765× | 0.802× | 1.463× | 1.463× | 0.784× | 0.932× | 1.819× | resolved, advisory |
| applications/sha256-clang | 0.994× | 0.829× | 0.932× | 1.397× | 1.392× | 0.650× | 0.771× | 1.793× | inconclusive |
| applications/sha256-rust | 1.001× | 0.827× | 0.922× | 1.356× | 1.343× | 0.631× | 0.781× | 1.741× | inconclusive |
| gc/i31-roundtrip | 1.041× | 0.792× | unsupported | 1.064× | unsupported | unsupported | 0.507× | unavailable | inconclusive |
| gc/struct-access | 0.839× | 0.749× | unsupported | 1.643× | unsupported | unsupported | 0.371× | unavailable | inconclusive |
| gc/struct-allocation | 1.177× | 1.351× | unsupported | 0.313× | unsupported | unsupported | 1.393× | unavailable | inconclusive |
| gc/struct-indirect-access | 0.992× | 0.788× | unsupported | 2.746× | unsupported | unsupported | 0.716× | unavailable | inconclusive |
| memory/narrow-unaligned | 0.999× | 0.833× | 1.063× | 1.740× | 1.736× | 0.996× | 1.074× | 2.569× | resolved, advisory |
| memory/dependent-address | 0.993× | 0.810× | 0.896× | 1.261× | 1.255× | 0.870× | 1.060× | 1.641× | inconclusive |
| memory/float64-unaligned | 0.993× | 0.863× | 0.769× | 1.333× | 1.310× | 1.028× | 1.133× | 1.583× | resolved, advisory |
| runtime/simd-memory-full | 0.998× | 0.840× | 0.902× | 1.075× | 1.072× | 1.055× | 1.225× | 0.887× | resolved, advisory |
| runtime/simd-memory-widen | 0.998× | 0.899× | 0.759× | 1.852× | 1.564× | 1.049× | 1.389× | 0.658× | resolved, advisory |
| runtime/simd-memory-lanes | 0.997× | 0.978× | 0.944× | 1.128× | 1.124× | 1.150× | 1.410× | 1.471× | resolved, advisory |
| threads/private-memory | 0.982× | 0.959× | 0.958× | 3.202× | 2.539× | 0.910× | 0.537× | unavailable | inconclusive |
| threads/shared-memory | 1.195× | 0.553× | 1.541× | 6.252× | 4.450× | 0.914× | 0.535× | unavailable | inconclusive |
| threads/atomic-rmw | 1.156× | 0.604× | 1.507× | 1.405× | 1.274× | 0.907× | 0.559× | unavailable | inconclusive |
| threads/notify-empty | 1.220× | 0.829× | 0.077× | 0.351× | 0.229× | 0.784× | 0.453× | unavailable | inconclusive |
| threads/wait-mismatch | 1.176× | 0.610× | 0.065× | 0.258× | 0.685× | 0.659× | 0.460× | unavailable | inconclusive |
| threads/wait-timeout | 0.895× | 0.614× | 0.005× | 0.046× | 0.023× | 0.108× | 0.001× | unavailable | inconclusive |

Relative to the own baseline, LLGo interpreter integer/call/memory loops have
ratios 0.943/0.893/0.946; the corresponding native ratios are
0.995/0.999/0.995. Native routing therefore has not closed the native compiler
gap. Shared memory and GC indirect access remain the largest observed Wasmtime
gaps. LLGo native shared memory/atomic/allocation costs are 1.195/1.156/1.177
of Go, but their A/A controls are inconclusive; this event does not isolate a
new LLGo defect. The existing LLGo issues remain separate, and #2983 stays with
its contributor.

For C/Rust SHA, median LLGo native readiness is 4.514/4.434ms,
compile+validate is 4.892/5.112ms, first call is 1.610/1.557ms and warm complete
call is 1.607/1.564ms. Go compile+validate is 2.334/2.522ms; Wasmtime is
2.364/2.457ms; wazero is 0.571/0.639ms. These phases are measured directly and
never subtracted from process wall time or from each other. Both application
A/A checks remain inconclusive. Faster preparation, hot native code, shared
memory and GC indirection all require follow-up; short-call timer noise does
not justify ignoring any of them.

The local evidence archive is
`artifacts/direct-runtime-routing-20261105/hosted-evidence.tar.gz` (2,042,255
bytes, SHA-256 `0e2da645cdfdd40e6103c00aacea6b47a400afa07cd7abc5791fce4532698a10`).
The audit records every original result checksum, all six job durations and
archive inputs. Its floating-point calibration recheck permits only 1e-12
relative rounding differences between Linux and macOS math libraries; control
outcomes are identical. Original hosted intervals remain unchanged.

Both older noisy W results and failed setup attempts remain recorded. The new
event is a different source architecture and environment, not a favorable retry.
No new PR or merge is performed while external PR #2991 remains open; upstream
merge is prohibited by owner policy. LLGo #2983 stays with its contributor.
Final-head four-profile semantics, all 31 packages/profile, 30 coverage gates,
review, controlled competitor/native acceptance, global allocation and the
broader application/WASI-world portfolio remain open. No milestone or goal is
marked complete.

The follow-up [public-call allocation diagnosis](performance-public-allocation-2026-11-05.md)
retains both a negative aggregate-passing result and the remaining Go/LLGo
allocation-byte gap. It changes no engine source or hosted candidate and makes
no timing or compiler-defect claim from allocation totals alone.
