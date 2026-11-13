# Shared indirect dispatch and current performance gaps

Production source: base `85ea347`, candidate `e2e537c`; test-only portability
correction `72c13a9`. Performance acceptance remains open. This change removes
quadratic dispatch metadata; complete P3 execution cost is unchanged.

## Implementation and image size

Previously every native function using indirect calls emitted 32 bytes per
module definition. The compiler now emits one immutable table per module policy:
unmetered ordinary/expanded bodies share one table; metered bodies use the other.
Live function identities, structural type checks, frame quotas, fuel,
cancellation and common fallback remain. Accepted bodies reserve table bytes
inside the aggregate budget; optional expansion follows both ordinary tiers.

The arm64 address sequence uses an image-relative forward displacement, without
assuming page-aligned JIT allocations or ADR's ±1 MiB range. AMD64 uses a checked
RIP-relative address. Address templates and branches are validated before
publication. There is no new public API or native frame ABI.

| Compiled image | Base bytes | Candidate bytes | Reduction |
| --- | ---: | ---: | ---: |
| 16 indirect-calling functions | 70,912 | 55,808 | 21.3% |
| 64 | 480,256 | 223,232 | 53.5% |
| 128 | 1,484,800 | 446,464 | 69.9% |
| 256 | 5,066,752 | 892,928 | 82.4% |
| 512 | 18,522,112 | 1,785,856 | 90.4% |
| Actual P3 Rust component, two core modules | 8,477,788 | 7,512,652 | 11.4% |

Go 1.27.3, Clang 22.1.8, macOS 27.0.1/M4 Max arm64, normal GC and
GOMAXPROCS=4. The compile diagnostic uses 20 fixed blocks, base/head and two
identical-binary controls: 160 fresh processes across two cases. Core decoding,
validation and slot lowering precede the native Compile timer. P3's two compile
intervals are summed. The scaling case follows the same fixed smaller-size
sequence in every process. These images remain linked.

Compile wall-time ratios are 0.21635 [0.20783, 0.22683] for the scaling case and
0.96499 [0.94839, 0.98200] for P3. **All compile wall A/A controls are unresolved
at ±2.5%; these are diagnostics, not qualified time-speedup claims.** Byte counts
are deterministic across every process. Explicit linked compilation of the
large official Go fixture still reaches a resource/range limit; public Compile
retains ordinary-tier fallback. This is not a large-Go linking fix.

## Complete P3 execution

A separate fixed event uses 20 blocks, four roles and identical-binary aliases:
320 fresh processes across independent 16 B/64 KiB cohorts. Four Stores, sixteen
clients, 512 requests/batch, twenty measured batches plus a final flush. Every
response checks body, status, headers, trailers and CRC32. The frozen base
runner's `d87ea62` production and measurement sources match parent `85ea347`.
LLGo `56fa9ae -O3 -lto=full` runs the candidate only. The independent frozen
Wasmtime 48.0.1 bridge uses its actual Rust P3 providers.

Preparation, first call, two warmups and cleanup remain separate. Individual
batches interleave with other children stopped. Whole-process CPU/instruction
windows include the final background tail; batch throughput and public API
intervals remain distinct. No owned builds, tests, profiles, source edits or
pruning overlap timing. Independent 300 ms process-CPU calibration agrees with
the mach-time conversion within 0.008%. Individual 95% intervals use 20,000
seeded paired-block resamples; these are not simultaneous portfolio gates.

| Go candidate/base | 16 B | 64 KiB |
| --- | ---: | ---: |
| CPU ratio, individual 95% interval | 0.99441 [0.98716, 1.00136] | 1.00075 [0.99551, 1.00592] |
| Retired-instruction ratio | 0.99916 [0.99838, 0.99993] | 0.99959 [0.99866, 1.00046] |
| Batch throughput ratio | 1.00310 [0.99220, 1.01328] | 0.99301 [0.98158, 1.00354] |
| Go allocations/request | 698 → 698 | 758 → 758 |
| Go allocated bytes/request | 54,696 → 54,696 | 124,345 → 124,345 |

All Go CPU/instruction controls resolve ±2.5%. Small-request Go throughput
controls resolve; large-request candidate throughput does not. No execution
speedup is claimed. Go/Wasmtime instruction ratios are 4.04057 and 3.81562, with
all instruction controls resolved. CPU ratio points are 2.48007/2.47086, but
**both Wasmtime CPU controls and both LLGo CPU controls are unresolved**. No
formal competitive CPU ranking or LLGo source improvement is established.
Reference managed counters cover only the Go shim, excluding Rust/native heaps.

## Fresh native Linux portfolio

[Run 43425801088](https://github.com/cpunion/wasmvm/actions/runs/43425801088) uses
source `85ea347` on both sides and pinned Go 1.27.3/LLGo 56fa9ae production
profiles. It is a same-source competitor calibration, not a source improvement
or compiler upgrade. Five independent blocks, two warmups, seven randomized
interleaved samples, one inherited CPU affinity and GOMAXPROCS=1. Each row
compares runtimes on the same group host. Core uses EPYC 9V74; memory/SIMD and
Threads use EPYC 7763; applications and GC use Xeon 6975P-C. No cross-host pooling
or physical-isolation claim.

Ratios above 1 mean Go-native wasmvm is slower. Units are ns/source iteration
(SHA-256: one 64-byte block). Call-control includes a complete public API call.
GC allocation uses process means including pauses; other rows use process
medians. Native C omits Wasm validation/isolation and is an unchecked algorithm
reference. N/A denotes unavailable/unsupported, never measured zero. The last
column refers to the portfolio's four LLGo A/A controls; it is not a controlled
Go or competitive equivalence verdict. All individual intervals, other runtimes
including interpreters, phases and unsuccessful reference probes remain in the
machine record.

| Workload | Go native ns/unit | Go / Wasmtime | Go / Wasmer | Go / wazero compiler | Go / native C | Four LLGo A/A resolve |
| --- | ---: | ---: | ---: | ---: | ---: | --- |


All 22 original cases and all four sustained GC cases complete. Sustained
Threads retains five complete cases and 68 successful final wait-timeout rows
when its job reaches ten minutes. The partial case remains without final
calibration in the raw artifact, whose status is still running; the enclosing
job is cancelled. These rows are not silently completed or pooled into another
cohort. Sustained private/shared-memory Go/Wasmtime points are 1.973×/1.531×,
atomic RMW 1.027×; all calibration limitations remain recorded. Shared dispatch
does not claim to fix those paths.

Next targets are the remaining P3 host/guest instruction cost and generated
memory/SHA-256 paths. Ratios locate affected cases; they do not establish a
specific allocator, GC or compiler pass as the cause. A matched comparative
profile is needed before the next runtime optimization or a new LLGo bug
report. Controlled milestone and release acceptance remain open.

## Validation and evidence

All 32 Go test packages pass; all 31 executable packages meet 90% coverage on
actual production source with cross-package instrumentation. Codegen 97.96%,
nativeexec 93.07%, component 92.47%, P3 91.28%. Compiler/nativeexec/component/P3
race tests pass. Those four packages pass under pinned LLGo's production
profile. The final hosted matrix is required before merge.

Regression tests cover distant addresses, malformed templates, validation before
patching, both encodings, exact aggregate table budgets, architecture-independent
body scaling and ordinary fallback. Existing live retargeting/null/type/host/
foreign-instance, quota and fuel tests execute through the shared tables. The
initial Linux test failure used an arm64-specific absolute size cap; replacing
it with a bounded growth comparison changes no production code.

- [Machine results](data/performance-shared-indirect-dispatch-2026-11-13.json): identities, controls, medians and CPU calibration.
- [P3 16 B](data/performance-shared-indirect-dispatch-2026-11-13-p3-16.jsonl.gz) and [P3 64 KiB](data/performance-shared-indirect-dispatch-2026-11-13-p3-65536.jsonl.gz): all 320 process rows and every counter/phase/check. Only unused individual RequestNs/PublicAPINs arrays are projected to count/distribution/compact-JSON SHA256; lossless originals remain local. Reported metrics do not depend on these arrays.
- [Complete Linux portfolio](data/performance-shared-indirect-dispatch-2026-11-13-linux-routing.json.gz): every runtime outcome, including cancelled partial observations.
- [Exact protocols](data/performance-shared-indirect-dispatch-2026-11-13-protocol.tar.gz): compile overlays, all 160 compile rows, P3 scripts/plans and semantic logs. Small overlay fragments replace checkout snapshots; no executables are committed.
