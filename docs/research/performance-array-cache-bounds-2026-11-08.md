# Guarded array payload caches: 2026-11-08

Implementation `34a3b29` compares with main `914be42`. Go 1.27.3,
LLGo `3571c25` (`-O3 -lto=full`), LLVM 22.1.8 and reference pins remain fixed.

Unmetered native array accesses reuse an immutable payload pointer when the
existing bounded i32 transfer supplies a useful per-class index hint. Ordinary
successful accesses promote a cache only when its actual immutable length
exceeds that hint. Every hit checks the non-null full handle and actual unsigned
index. Misses retain the original identity, length, bounds and trap checks.
Cold continuations may violate any hint and remain fully checked. Unknown
accesses disable the class specialization. Both emitters use the existing two
cache registers; no runtime field, ABI, safepoint or precise-tier change.

Go and pinned LLGo execute codegen, State, execute, nativeexec and public VM
tests; vet passes. Differential tests compare with ordinary checked execution
for all element widths, packed signed loads, vector halves, every cold PC,
empty/short arrays, high i32 bits, valid unexpected indices and null/stale/wrong
identities after a successful access. Defined exits preserve frames and payload
effects; invalid internal states reject with BadState without payload writes.

Local combined coverage: codegen **97.82% (8692/8886)**, State **95.22%
(1495/1570)** and nativeexec **92.54% (459/496)**. Full final-head platform CI
and review remain required before merging.

## Code selection and measurements

Actual ARM64 publications under both host compilers match byte for byte for
all four GC fixtures and retain their baseline code. Existing register ranking
selects the structure cache in the indirect fixture, leaving no pair for the
array. These local timings therefore do not demonstrate a guest-kernel gain.
AMD64 compilation/linking at unchanged feature mask 7 selects the array cache:
the indirect module changes **13507→13663 bytes**; the other **21/22** portfolio
modules remain identical. This is compile-only evidence on ARM64, not execution
or speed evidence. Ranking was not adjusted to favor the measured fixture.

The independent [Linux GC comparison](https://github.com/cpunion/wasmvm/actions/runs/42443718634)
records actual AMD64 execution. Its completed evidence and full final-head CI
audit are linked from the contribution PR. No local/hosted or original/sustained
cohorts are pooled. Warm intervals include the complete API call; engine setup,
compile/validate, instantiation, first call, cleanup and process times remain
separate. Five interleaved process blocks, seven warm calls and two warmups
retain fixed inputs, counts and alternate checksum checks. Allocation uses the
median of process means including every pause; other rows use process medians.

| Local cohort / workload | LLGo native head/base | LLGo native / Go | LLGo native / Wasmtime | All four A/A controls |
| --- | ---: | ---: | ---: | --- |
| original / i31-roundtrip | 1.020 | 0.968 | 0.887 | inconclusive |
| original / struct-access | 0.992 | 0.924 | 1.400 | inconclusive |
| original / struct-allocation | 0.992 | 0.937 | 0.375 | inconclusive |
| original / struct-indirect-access | 0.986 | 0.876 | 1.747 | inconclusive |
| sustained / i31-roundtrip | 1.037 | 1.022 | 0.760 | inconclusive |
| sustained / struct-access | 0.992 | 0.981 | 1.254 | inconclusive |
| sustained / struct-allocation | 1.005 | 0.931 | 0.597 | inconclusive |
| sustained / struct-indirect-access | 1.011 | 1.013 | 1.751 | resolved, advisory |

The local cohorts resolve **0/4** and **1/4** cases. Retain all adverse and
unresolved results; no favorable retry or compiler/profile/count sweep. Pinned
wazero and Wasmer do not support these GC fixtures; native C is unavailable.
Controlled portfolio, steady-state GC, broader WASI worlds, first-call gaps and
release acceptance remain open. MIT resolves the project-license decision; it
does not close these technical gates.

The [compressed machine record](data/array-cache-bounds-2026-11-08.json)
retains every local observation, source/build identity and publication audit.
Immutable local evidence: `artifacts/array-cache-bounds-20261108/local-evidence.tar.gz`,
2729237 bytes, 178 files; SHA-256
`deea422fd3a1e37a832a2d3950244c001c4eada6ecd483d435c17084f6f1ecf6`.
