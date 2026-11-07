# Integer words in spare FP registers: 2026-11-06

**Rejected and reverted.** Engine `0e32719ab934ea2246e320da777303a808b363f7`
compares with `5099a76ddad5910dd2cfff195ca8a4bca9be9cf0`. The completed Linux
event regresses both SHA cases by about 22% under LLGo and 23% under Go, with
all four advisory controls resolved for each. Commit `4a744df` removes this
production allocation strategy. The following design, semantic and static
observations describe the rejected experiment; its code and evidence remain in
history. Current development uses [mixed scalar forwarding](performance-mixed-scalar-forwarding-2026-11-06.md).

## Shared register and frame contracts

Bounded private loops with fixed register mappings may retain spilled i32/i64
Lo words in spare FP/SIMD registers. These are bit moves, with no floating-point
conversion. Typed FP/vector homes, existing raw homes and immutable vector
constants reserve their physical registers first. A deterministic, bounded
slot-use ranking includes entry/exit transfers; it is a heuristic rather than a
measured cost model. Variable mappings and deferred-copy targets remain excluded.

The implementation shares existing slot reads/writes, register transitions,
copy forwarding and cold-exit snapshots. Every raw home preserves all 64 Lo bits,
including noncanonical upper i32 bits; each slot keeps its independent frame Hi
word. Cross-bank moves consume inputs before stores. Mixed x86 arithmetic reads
raw operands through existing bit moves instead of stale frame backing. No new
execution framework, physical register inventory, ISA requirement or API mode is
introduced. Precise fuel, original cold PCs, polling, cancellation and traps keep
the existing contracts.

## Semantic and encoding evidence

Complete codegen/nativeexec/public-VM suites pass under Go 1.27.3 and pinned LLGo
`3571c25` with `-O3 -lto=full`; vet passes. Exact package coverage is
**codegen 97.68% (8535/8738)** and **nativeexec 92.54% (459/496)**. These are Go
statement measurements, not specification coverage.

Tests exercise independent bank reservations, rejected mappings, full-width
cross-bank copies, FP/V128 coexistence, transitions, original cold entries,
forced private polls and late exits. Pressure fixtures actually select raw
homes on ARM64 and compare every Lo/Hi word with an independent frame model.
The initial insufficient-pressure fixtures and their failed assertions remain
recorded before correction.

Both compilers emit identical bytes for **1,980 complete x86 kernels** covering
all twelve arithmetic forms, GPR/raw/frame operands, destination aliases and
forwarded copies. Numeric execution checks 64 operand pairs, every frame word,
fuel/budget and continuation. This ARM64 host emits those x86 kernels only;
actual final-source numeric execution is required in Linux CI. Separately,
**280 cross-bank instruction encodings** match LLVM 22.1.8's assembler exactly.

## Actual generated code

All 44 fresh ARM64 publication witnesses pass their checksum oracles. Head
Go/LLGo images and entry lists match across all 22 fixtures. Twenty-one
baseline/head images are unchanged; Clang SHA grows by 24 bytes. Its compression
loop uses one raw home, retains 45 instructions and removes its one frame
reference. Rust SHA remains at 39 instructions with zero frame references.

Production x86 compilation/linking uses the unchanged feature mask 7 and stops
before publication on this host. Only spans containing known private poll
locations count as hot loops; embedded dispatch data is excluded.

| x86 compression loop | Raw homes | Instructions | Frame references | Complete function bytes |
| --- | ---: | ---: | ---: | ---: |
| Clang SHA | 8 | 70 → 68 | 20 → 0 | 10864 → 11086 |
| Rust SHA | 7 | 67 → 67 | 18 → 0 | 10746 → 10950 |

Both x86 schedule loops retain 43 instructions, zero frame references and one
poll. Compression polling also remains one per iteration. More entry/exit code
and cross-bank dependencies may offset removed frame traffic; static counts
establish no elapsed gain.

## Separate source-pinned measurements

The local macOS ARM64 event measures only the affected application group, with
unchanged Rust code as a declared negative control. It retains original counts,
five fresh process blocks, two warmups, seven serially interleaved calls, all
reference engines, native C, 34 alternate checksum checks and four A/A controls
per case. Builds, tests and profiling do not run concurrently with this event.
Both cases resolve the advisory ±2.5% controls; this does not satisfy controlled
milestone acceptance. Ratios below compare elapsed time; lower is faster.

| ARM64 workload | Go head/base | LLGo head/base | LLGo / Go | / wazero JIT | / Wasmtime | / native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Clang SHA | 1.019 | 1.015 | 0.984 | 0.693 | 1.093 | 1.254 |
| Rust SHA | 1.001 | 0.999 | 0.983 | 0.653 | 1.125 | 1.308 |

The Clang observation is adverse and remains recorded. Native preparation
(compile/validate) also stays separate: Clang Go **2.182 → 2.265 ms**, LLGo
**4.433 → 4.923 ms**; Rust Go **2.263 → 2.200 ms**, LLGo **4.973 → 5.081 ms**.
All load/setup/instantiate/export/first/warm/cleanup and process intervals remain
in the raw records. No timer subtraction, favorable retry or ISA/flag/alignment
sweep is used. These observations do not identify a new LLGo compiler defect;
LLGo #2983 remains with its existing contributor.

The completed [Linux event 42088602463](https://github.com/cpunion/wasmvm/actions/runs/42088602463)
compares this exact engine against `5099a76`, with one shared build and all 22
original-count workloads. All six required jobs pass: build 314s, longest measurement group 392s. The audit
checks 22 cases, 362 alternate configurations and 12 expected GC rejections.
Nine cases resolve all four advisory controls, including both SHA cases.

| Rejected Linux SHA | Go head/base | LLGo head/base | LLGo / Go | / wazero JIT | / Wasmtime | / native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Clang | 1.234 | 1.217 | 0.987 | 0.859 | 1.630 | 1.957 |
| Rust | 1.236 | 1.218 | 0.984 | 0.813 | 1.604 | 1.911 |

The measured regression withdraws this allocation policy. Removed frame accesses
provide no elapsed acceptance; the hardware-level cause remains unisolated.
All other rows, unresolved controls and lifecycle phases remain recorded.
It remains independent of local and prior events.
The completed [preceding Linux event](performance-frame-operands-2026-11-06.md)
for `5099a76` has 22 cases, 362 alternate checks and nine cases resolving all
four advisory controls. Its SHA LLGo head/base ratios are 0.970/0.919; ratios to
Wasmtime remain 1.372/1.320 and to native C 1.649/1.575. This leaves clear gaps.

## CI and integration

[CI 42083219131](https://github.com/cpunion/wasmvm/actions/runs/42083219131)
passes the preceding source's Linux Go/LLGo codegen jobs, but its Go component
job fails `TestSourceReadWithIndependentContext`: a five-millisecond deadline
can expire before a source worker starts, while the test still waits for that
worker. Commit `73427e0` replaces this scheduling assumption with a startup gate
followed by cancellation. Both complete component suites, 30 repetitions per
compiler and 20 Go race repetitions pass. The parent-context and joined-worker
checks remain; an unexpected nil read error is now rejected too.

The original CI jobs remain live until terminal status. Final-head CI, all four
semantic profiles, review/integration, controlled competitive/native acceptance,
general allocation/rematerialization, preparation costs and broader world/release
conformance remain open. There is no new parallel contribution PR while external
LLGo #2991 remains open.

The [machine-readable record](data/raw-integer-registers-2026-11-06.json) includes
source/compiler identities and separate events. Immutable local evidence:
`artifacts/loop-raw-registers-20261106/evidence.tar.gz`, 3327140 bytes, 343 files;
SHA-256 `7b09e2b7d5cfd1bd1e840b18169103701177105b1f2c535fcc4c40002240e61f`.
The completed preceding Linux archive is 2106484 bytes, 16 files;
SHA-256 `2088852c440b1918313daaaedf4b51b168b6f5bef6cebbfb932ab2c737a0aeca`.

Completed rejected Linux evidence: `artifacts/loop-raw-registers-20261106/hosted-evidence-final.tar.gz`,
2112890 bytes, 16 files; SHA-256
`62d5ae960295bd1ff14e9524103aceed593cb21cb54ab966f37432c4cf5325e3`.
