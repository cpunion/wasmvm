# Reference-family filters: 2026-11-08

Implementation `97b83fa` compares with main `310d9c2`. The pinned Go 1.27.3,
LLGo `3571c25` (`-O3 -lto=full`), reference versions and CPU policies are unchanged.

The preceding warm stack profiles repeatedly sample reference tracing. Linked
LLGo `CollectReferences$3` validates extended type metadata before rejecting
nonfunction families. The shared type predicates now test the family first;
`IsReference` tests its byte range before encoding metadata. Matching types still
receive complete validation. This is a reordering of pure Boolean predicates,
without changing reference domains, layouts, handles, GC roots or VM policy.
The existing malformed-type test now also checks classifiers and storage erasure
for malformed function, external, exception and continuation representations.

All affected type, State, execute, nativeexec and public-VM tests pass under Go
and release LLGo; vet passes. Combined affected coverage is State 95.21% and
nativeexec 92.54%; final CI retains every package and coverage gate. The linked
normal LLGo runner confirms family rejection now precedes encoding validation.
All 22 AMD64 compiled/linked images remain identical. All 22 actual ARM64 images
and entry lists match baseline and Go/LLGo, with 44 fresh candidate checksum
publications. Baseline reuse requires exact source/compiler/binary/fixture
identity. The host code changes, so unchanged guest bytes do not waive timing.

## Separate local ARM64 cohorts

Five fresh process blocks, seven interleaved warm calls and two warmups retain
the existing fixed workload counts and oracles. No builds, tests or profiling
ran alongside these calls. Every setup/first/warmup/warm/cleanup interval and
all alternate-seed checks remain in the raw record. Allocation uses the median
of process means including pauses; median diagnostics and both A/A reductions
remain separate. Other cases retain process medians.

| Cohort / workload | LLGo interpreter head/base | LLGo native head/base | LLGo native / Go | LLGo native / Wasmtime |
| --- | ---: | ---: | ---: | ---: |
| Original / i31-roundtrip | 0.926 | 1.009 | 0.912 | 0.827 |
| Original / struct-access | 0.922 | 0.968 | 0.856 | 1.340 |
| Original / struct-allocation | 0.993 | 1.001 | 0.944 | 0.379 |
| Original / struct-indirect-access | 0.919 | 1.007 | 0.908 | 1.793 |
| Sustained / i31-roundtrip | 0.884 | 0.995 | 0.976 | 0.741 |
| Sustained / struct-access | 0.928 | 0.993 | 0.981 | 1.250 |
| Sustained / struct-allocation | 1.000 | 0.973 | 0.880 | 0.624 |
| Sustained / struct-indirect-access | 0.909 | 1.000 | 0.979 | 1.746 |

All four A/A controls remain unresolved for each local row. Ratios are
observations, **not established gains or controlled acceptance**. The original
and sustained cohorts are independent, not pooled. GC-unsupported reference
configurations remain explicit; no native-C GC surrogate is invented.

The independent [Linux original22 and sustained-GC run](https://github.com/cpunion/wasmvm/actions/runs/42429886790)
and final four-profile semantic CI are recorded in the PR body after completion.
Their observations are not pooled with this workstation. Compiler/profile and
measurement counts are not swept or favorably retried. Controlled portfolio,
steady-state GC, expanded world and release acceptance remain open.

The [machine record](data/reference-kind-filter-2026-11-08.json) preserves both
local cohorts and audits. Immutable local evidence:
`artifacts/reference-kind-filter-20261108/local-evidence.tar.gz`,
2438456 bytes, 322 files; SHA-256
`4d7ea14d4a20625c1a7790dc9208d221bc871e14322b32ef4fc300288c288291`.
