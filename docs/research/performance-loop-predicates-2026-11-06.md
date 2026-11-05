# Private-loop predicate reconstruction: 2026-11-06

**Local Go/LLGo semantics pass; controlled performance acceptance remains open.**
Engine `15567dfe8068f5e279c61c67f9bd06d37ac0176e`; baseline engine
`993cbb3a59f9b19b58523ce8f239c2d6b37ec3dd`, documentation head `eab79fe`.
The [preceding frame-copy event](performance-loop-frame-copies-2026-11-05.md)
retains its separate allocation counts and Linux comparison.

## Frame and polling contract

A private, nontrapping loop can defer its final i32/i64 comparison result when
the result's entry version is unread. The existing bounded local graph and
private-loop proof retain all comparison inputs. Header/interior polls reject
the transform. The hot backedge polls before comparing, then branches directly
on flags. x86 polling may clobber flags, so the ordering is necessary.

One frame-snapshot contract covers copies and the predicate. Ordinary exit
reconstructs copies and writes the known exit condition. Cancellation first
reconstructs copies, then runs the original comparison emitter and publishes
the original branch PC. A taken backedge leaves the result word stale only
because the next iteration overwrites it before any read. Variable residency
transitions keep all raw input words; shared result homes are excluded. Original
cold PCs and the precise tier keep the complete original instruction sequence.
The existing comparison encoder supplies both materialized and flags-only forms.

## Semantic evidence

Complete codegen, nativeexec and public VM suites pass on macOS arm64 with
Go 1.27.3 and LLGo `3571c25`, using `-O3 -lto=full`. Vet passes. The three suites
cross-cover **codegen 97.66% (8342/8542)** and **nativeexec 92.54% (459/496)**.
These are Go statement-coverage measurements for the named packages.

New independent instruction models cover all 22 integer comparison/eqz opcodes,
both backedge directions, resident and pressured inputs, aliased destinations,
every original cold PC, complete raw Lo/Hi words, first generated backedge poll,
and resumption from the actual published branch snapshot. Header-poll and
entry-read cases reject deferral. Existing copy/variable-residency/shared-memory
tests also compare exact frames, late traps, memory and precise continuations.
Both targets compile; the local arm64 host executes. The initial full-suite
failure exposed a test patcher that assumed continuation publication was the
first cold-exit instruction. It now uses the existing generated poll offsets;
all precise-reference frame, memory, trap and continuation assertions remain.

## Actual generated code

All 44 fresh candidate publication witnesses pass their checksum oracles, and
Go/LLGo images and entry lists match for all 22 fixtures. The baseline's 22
existing witnesses are reused after fixture/source/image verification. Ten
baseline/candidate ARM64 images differ; all four GC images remain identical.

| SHA loop | ARM64 instructions | x86 instructions | x86 frame references | Polls |
| --- | ---: | ---: | ---: | ---: |
| Clang schedule | 29 → 28 | 47 → 44 | 0 → 0 | 1 → 1 |
| Clang compression | 47 → 46 | 88 → 85 | 20 → 20 | 1 → 1 |
| Rust schedule | 29 → 28 | 47 → 44 | 0 → 0 | 1 → 1 |
| Rust compression | 41 → 40 | 85 → 82 | 18 → 18 | 1 → 1 |

The complete x86 compression functions grow by 16 bytes each: Clang
10984 → 11000, Rust 10866 → 10882. Complete ARM64 module images grow by 24 bytes
for each producer, including cold reconstruction. ARM64 Clang compression still
has one frame reference, Rust zero. x86 uses the unchanged feature mask 7 and
production module compilation/linking, stopping before publication on this
arm64 host; it establishes no x86 guest-execution result. Static reductions
do not establish elapsed gains.

## Measurement scope

The local ARM64 event selects every group containing a changed actual image:
applications, core, memory/SIMD and threads, 18 fixtures in total. This choice is
recorded before timing. GC retains its previous evidence because all four guest
images match. Each group uses five fresh process blocks, two warmups and seven
randomly interleaved warm samples, unchanged fixture counts and seeds, and
separate preparation, first call, warm call and cleanup records. All competitors
and applicable native-C cases remain present. Exact-source baseline/reference
binaries are reused after source, harness, toolchain/profile and executable/
library identity checks; candidate executables are fresh.

The local event has no CPU affinity or controlled frequency/neighbors. Each
case retains four independent identical-binary A/A controls and the declared
marginal 95% ±2.5% diagnostic calibration. Elapsed ratios are descriptive and
lower is faster; unresolved controls remain inconclusive. No sample pooling,
phase subtraction, favorable retry, ISA/flag/alignment sweep or omitted adverse
result is allowed. Final-head four-profile CI, contribution review, expanded
world conformance, general rematerialization and controlled performance
acceptance remain open.

## Local elapsed results

All four groups complete: **18 fixtures, 306 alternate-seed/count checks and
1,585 measured process records**, retaining every runtime, native-C row and
sample order. **8/18 cases resolve all four A/A controls.** The following are
median warm elapsed ratios; lower is faster. These diagnostic controls do not
satisfy the controlled acceptance gate or establish statistical significance for
small source differences.

| Workload | LLGo head/base | LLGo / Go | / wazero | / Wasmtime | / Native C | All four A/A resolve |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 0.906 | 1.000 | 7.250 | 4.143 | — | No |
| core/integer-loop | 0.990 | 0.985 | 0.907 | 0.995 | 1.000 | No |
| core/function-calls | 1.004 | 0.982 | 0.548 | 0.869 | 1.022 | Yes |
| core/memory-load-store | 1.005 | 0.959 | 0.495 | 0.906 | 1.619 | Yes |
| threads/private-memory | 1.201 | 0.993 | 1.290 | 1.274 | — | No |
| threads/shared-memory | 0.920 | 0.809 | 1.187 | 1.177 | — | No |
| threads/atomic-rmw | 1.012 | 0.918 | 1.112 | 1.122 | — | No |
| threads/notify-empty | 0.874 | 0.831 | 0.060 | 0.176 | — | No |
| threads/wait-mismatch | 1.015 | 0.766 | 0.066 | 0.179 | — | No |
| threads/wait-timeout | 0.995 | 0.562 | 0.006 | 0.038 | — | No |
| memory/narrow-unaligned | 0.946 | 0.979 | 0.621 | 0.945 | 1.431 | Yes |
| memory/dependent-address | 0.993 | 0.987 | 0.581 | 0.873 | 1.298 | Yes |
| memory/float64-unaligned | 1.003 | 0.975 | 0.849 | 1.027 | 1.103 | No |
| runtime/simd-memory-full | 1.003 | 0.993 | 0.862 | 0.928 | 0.999 | Yes |
| runtime/simd-memory-widen | 1.006 | 0.981 | 0.903 | 1.010 | 1.009 | Yes |
| runtime/simd-memory-lanes | 1.005 | 0.993 | 0.835 | 0.867 | 1.132 | Yes |
| applications/sha256-clang | 0.993 | 0.989 | 0.685 | 1.023 | 1.255 | Yes |
| applications/sha256-rust | 1.013 | 0.990 | 0.648 | 1.130 | 1.285 | No |

Clang SHA records a 0.7% lower LLGo warm median; Rust records a 1.3% higher
median with unresolved controls. Narrow integer memory records a 5.4% lower
median. The private-memory thread case records a 20.1% higher median with
unresolved controls; it remains visible. Unchanged-image cases also move, so
instruction reductions alone cannot explain every observed difference.

Both SHA cases remain slower than Wasmtime and native C in this local event,
while faster than wazero. These are workload/target-specific observations.
Complete Wasmer and WasmEdge interpreter observations, both own interpreters,
all lifecycle phases and all adverse values remain in the machine-readable
record; a reference interpreter is not presented as its native compiler.

SHA preparation (validation plus native compilation) medians remain separate:

| Producer | Go baseline → head | LLGo baseline → head |
| --- | ---: | ---: |
| Clang | 2.209 → 2.284 ms | 4.564 → 4.673 ms |
| Rust | 2.226 → 2.221 ms | 4.832 → 4.854 ms |

The LLGo/Go preparation gap remains about 2.05–2.19×. Warm A/A calibration
does not calibrate preparation; these preparation medians are descriptive and
do not identify a new LLGo compiler bug. First-call, instantiation, loading,
export lookup and cleanup measurements are retained without subtraction.

## Reproduction and integration status

The machine-readable [record](data/loop-predicates-2026-11-06.json) contains
source/profile identities, exact package coverage, output summaries and every
local case summary. The immutable evidence archive retains raw process
records, checksums, protocol orders, generated images/disassembly, diagnostic
patches, source changes, harness snapshots and development failures.

The preceding engine’s [Linux event](https://github.com/cpunion/wasmvm/actions/runs/42051975324)
and [four-profile CI](https://github.com/cpunion/wasmvm/actions/runs/42057704970)
continue independently. They validate `993cbb3`/document head `eab79fe`, not
this new engine. Existing runs are not replaced because of queue delay. New
final-head CI and actual x86 execution remain required; the current contribution
branch is not merged.

Evidence archive: `artifacts/loop-predicate-reconstruction-20261105/evidence.tar.gz`, 4074760 bytes, 253 files. SHA-256:
`034018cec0cbc4d1fb2f473b49d9fae06c81b04d002b1f4f25f790a8431d05b3`.
