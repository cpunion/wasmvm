# Checked heap slots during collection: 2026-11-08

Implementation `1d11057` compares with main `3d91471`. Go 1.27.3 and LLGo
`3571c25` (`-O3 -lto=full`), reference versions and CPU policies remain fixed.

Marking and native compaction previously validated a handle, then used its
object header to locate the same slot again for mark flags. A shared checked
lookup now returns that slot after checking the tag, nonzero/in-range index,
populated cell and generation. Marking, sweeping and release reuse it directly.
The ordinary object lookup delegates to the same validation and exposes no
object for an invalid reference. Slot pointers stay within serialized operations
that cannot grow the heap or invoke host callbacks. No object/Store field, ABI,
quota, collector framework, root or epoch rule changes. Capacity maintenance
still preserves mark bits; collection clears only marks and retains extra flags.

Go and release-LLGo tests pass for State, execute, nativeexec and the public VM;
vet passes. The new test combines stale generations, tagless/out-of-range/null
indices, wrong reference domains and duplicate live roots. Replacements remain
unreachable, survivors retain data/serial identity through relocation, and a
second generation reuse releases cleanly in portable and native Stores. Existing
empty objects, rollback, quota, deep cycles, extras and forced-GC suites remain.
Affected coverage: State **95.22% (1495/1570)**, nativeexec **92.54% (459/496)**.

Final linked LLGo symbols shrink: visitor **584→524**, native compaction
**1316→1112**, empty sweep **344→324**, release **312→284 bytes**. These are static
sizes, not dynamic counts or elapsed attribution. All 22 AMD64 images at feature
mask 7 remain identical (compile/link only on ARM). All 22 actual ARM64 images
and entry lists match base/head and Go/LLGo, with 44 fresh candidate checksum
publications. Reused baseline evidence requires complete source, compiler,
profile, executable, library and fixture identity checks.

## Affected GC comparison

The routing workflow now accepts `routing_scope=gc`: one unchanged common build
and one GC group, retaining the original four cases plus a separate sustained
cohort. Default `all` retains the existing complete portfolio. No counts, oracle,
order, runtime, reducer or policy is changed; actionlint, routing tests and the
31-package/eight-shard semantic manifest check pass.

Local ARM64 retains five process blocks, seven interleaved warm calls and two
warmups. No builds/tests/profilers overlap measurement. Allocation uses the
median of process means including every observed pause, alongside the original
medians and both calibrations. Other cases use process medians. Setup, first,
warmup, warm, cleanup and process intervals stay separate, without subtraction.

| Local cohort / workload | Go native head/base | LLGo native head/base | LLGo interpreter head/base | LLGo native / Go | LLGo native / Wasmtime |
| --- | ---: | ---: | ---: | ---: | ---: |
| Original / i31-roundtrip | 0.983 | 1.026 | 0.980 | 0.943 | 0.829 |
| Original / struct-access | 1.017 | 1.031 | 1.007 | 0.975 | 1.433 |
| Original / struct-allocation | 0.925 | 0.945 | 0.983 | 0.910 | 0.365 |
| Original / struct-indirect-access | 1.012 | 0.981 | 0.975 | 0.859 | 1.545 |
| Sustained / i31-roundtrip | 0.987 | 1.006 | 1.007 | 0.996 | 0.747 |
| Sustained / struct-access | 1.016 | 1.001 | 0.981 | 0.982 | 1.219 |
| Sustained / struct-allocation | 0.917 | 0.944 | 0.965 | 0.901 | 0.592 |
| Sustained / struct-indirect-access | 0.997 | 0.992 | 0.967 | 1.001 | 1.697 |

All local rows have incomplete A/A controls. These observations do not establish
a reliable gain or controlled acceptance. The independent [Linux GC run](https://github.com/cpunion/wasmvm/actions/runs/42438275783)
and final four-profile CI are recorded in the PR body after completion. Hosts
and original/sustained cohorts are never pooled; adverse and inconclusive data
remain. No favorable rerun or compiler/profile/count/reference sweep is used.
Controlled portfolio, steady-state GC, expanded world and release acceptance
remain open.

The [machine record](data/gc-slot-reuse-2026-11-08.json) preserves complete local
observations and audits. Immutable local evidence:
`artifacts/gc-slot-reuse-20261108/local-evidence.tar.gz`,
2469410 bytes, 338 files; SHA-256
`bf16b5dd3fcd58dd72ef5814f7fa355e4d66b0a2836b9a326f271f6716c4fa17`.
