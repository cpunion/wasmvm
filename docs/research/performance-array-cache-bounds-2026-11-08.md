# Guarded array payload caches: 2026-11-08

Final implementation `2533767` compares with main `914be42`. Go 1.27.3,
LLGo `3571c25` (`-O3 -lto=full`), LLVM 22.1.8 and reference pins remain fixed.

Unmetered native array accesses reuse an immutable payload pointer when the
existing bounded i32 transfer supplies a useful per-class index hint. Ordinary
successful accesses promote a cache only when its actual immutable length
exceeds that hint. Every hit checks the non-null full handle and actual unsigned
index. Misses retain the original identity, length, bounds and trap checks.
Cold continuations may violate any hint and remain fully checked. Unknown
accesses disable the class specialization. Both emitters use the existing two
cache registers; no runtime field, ABI, safepoint or precise-tier change.

The hit path uses the existing unsigned `lo` branch with an exclusive hint + 1
limit. Unknown branch-condition names now fail before emitting any bytes.

Go and pinned LLGo execute codegen, State, execute, nativeexec and public VM
tests; vet passes. Differential tests compare with ordinary checked execution
for all element widths, packed signed loads, vector halves, every cold PC,
empty/short arrays, high i32 bits, valid unexpected indices and null/stale/wrong
identities after a successful access, including `0x80000000`/`0x80000001`
indices that set signed-overflow flags in ordinary 32-bit comparisons.
Defined exits preserve frames and payload effects; invalid internal states reject with BadState without payload writes.

Local combined coverage: codegen **97.82% (8694/8888)**, State **95.22%
(1495/1570)** and nativeexec **92.54% (459/496)**. Full final-head platform CI
and review remain required before merging.

## Code selection and measurements

The initial source `34a3b29` used unsupported `hi`/`always` condition names;
the intermediate `e133269` removed one jump but retained an unsupported `ls`.
The old helper silently selected condition-code zero. Neither source is the
final optimization. The [initial Linux event](https://github.com/cpunion/wasmvm/actions/runs/42443718634)
observed indirect LLGo head/base **1.039 original / 1.126 sustained** (Go
sustained **1.121**), with unresolved controls. The [intermediate event](https://github.com/cpunion/wasmvm/actions/runs/42444539410)
observed indirect LLGo head/base **1.077 original / 1.157 sustained** (Go
sustained **1.140**), also with unresolved indirect controls. It is preserved
separately. The final source fixes emission instead of changing
reference, input, compiler or policy settings. Do not pool these source events
or relabel their adverse data as success.

Pilot `34a3b29` actual ARM64 publications under both host compilers match
byte for byte for
all four GC fixtures and retain their baseline code. Existing register ranking
selects the structure cache in the indirect fixture, leaving no pair for the
array. These local timings therefore do not demonstrate a guest-kernel gain.
AMD64 compilation/linking at unchanged feature mask 7 selects the array cache:
the indirect module changes **13507→13639 bytes**; the other **21/22** portfolio
modules remain identical. This is compile-only evidence on ARM64, not execution
or speed evidence. Ranking was not adjusted to favor the measured fixture.

The independent [Linux GC comparison](https://github.com/cpunion/wasmvm/actions/runs/42444694119)
records actual AMD64 execution. Its completed evidence and full final-head CI
audit are linked from [PR #128](https://github.com/cpunion/wasmvm/pull/128). No local/hosted or original/sustained
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

## Corrected Linux source event

The corrected `2533767` event retains all four cases in each cohort, 56
alternate configurations and 12 explicitly unsupported configurations per
cohort. Original controls resolve 0/4; sustained controls resolve 3/4, including
indirect access. These are advisory hosted observations, not a controlled gate.

| Sustained workload | LLGo native head/base | Go native head/base | LLGo native / Go | LLGo native / Wasmtime | LLGo interpreter / WasmEdge | All four A/A controls |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| i31-roundtrip | 0.988 | 1.015 | 0.993 | 0.726 | 0.824 | resolved, advisory |
| struct-access | 1.005 | 1.001 | 0.998 | 1.019 | 0.743 | resolved, advisory |
| struct-allocation | 1.015 | 1.001 | 1.188 | 0.812 | 1.500 | inconclusive |
| struct-indirect-access | 0.995 | 0.990 | 0.996 | 0.887 | 0.768 | resolved, advisory |

Original indirect access is 0.992 head/base under LLGo, 1.103 under Go and
1.011 versus Wasmtime, with unresolved controls. Sustained indirect access
observes parity with its own baseline; the 0.887 ratio versus Wasmtime does
not establish a source-induced gain. Hosted machines and source events remain
independent; cross-event absolute timings are not an optimization comparison.
The [three-event machine record](data/array-cache-condition-2026-11-08.json)
retains every raw observation and recomputed audit, including both failed pilots.

The local table and compressed observations belong to pilot `34a3b29`, not
the corrected source. Ratios are elapsed-time ratios, lower is better. Those
local cohorts resolve **0/4** and **1/4** cases. Retain all adverse and
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
