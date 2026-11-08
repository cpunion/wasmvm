# LLGo interface baseline upgrade: 2026-11-08

This event advances the production compiler from `3571c25` to
`38de1e218b8146ba73f9c6f2e60e3c31f29d96b6`, canonical version
`v1.0.10-0.20261108045529-38de1e218b81`. The owned VM source is held constant
at integrated main `f2aa4f7`; Go 1.27.3, LLVM 22.1.8 and LLGo
`-O3 -lto=full` stay fixed. Guest compiler sources, resource/ownership contracts,
feature policy and benchmark inputs are not changed for the upgrade.
This is a compiler baseline event, not an own-VM implementation speedup.

## Existing upstream work

The old public-call profiles exposed runtime `NewItab` calls. A fixed standalone
reproduction with complete checksums and linked disassembly was reported in
[LLGo #3020](https://github.com/xgo-dev/llgo/issues/3020). Fresh upstream
verification then found merged [#2704](https://github.com/xgo-dev/llgo/pull/2704)
for static interface lowering and
[#3005](https://github.com/xgo-dev/llgo/pull/3005) for dynamic concurrent lookup.
The former emits the same four-instruction conversion shape as Go, removing
the repeated lookup for the known static pair. The latter separately replaces
the mutex/linear search with a published growable table.

A separate three-process/seven-trial diagnostic retains the original source,
one-million-iteration count and 1,024 warmups. It observes Go/old LLGo/upstream
static conversion at 1.756/8.365/1.618 ns per loop iteration; preconstructed
interface controls are 1.518/1.237/1.272. These include the non-inlined helper
and method invocation. They are local, uncalibrated observations from a clean
ordinary bootstrap of upstream, separate from the canonical production
compiler event below. No subtraction, broad speedup attribution or equivalence
claim follows. The issue retains both data sets and the verified upstream
resolution. No duplicate cache patch or upstream issue closure was performed.

Other compiler work, including the still-separate escape/defer reports, is
not assumed fixed. No unmerged contribution is included in this pin.

## Qualification and measurement boundaries

The compiler is built from the exact clean upstream tree with its canonical
version; source identity and executable hashes are retained. Four fresh runner
builds prove identical production VM source maps under base/head Go and LLGo.
Go uses the same executable and flags on both sources. LLGo uses its respective
exact old/new executable and runtime tree with identical release flags.
Logical source labels remain distinct; elapsed differences cannot be attributed
to VM changes because those source maps are identical.

All 91 measurement-tool tests pass. Targeted new-pin VM/native tests pass.
The first component run omitted the local worktree's corpus path and failed
to open inputs; its full log remains. After verifying the pinned corpus,
only that failed package reran with its explicit corpus directory and passed.
This was an input-preparation failure, not a hidden skip or compiler workaround.
The full final-head dual-compiler/platform semantic matrix remains mandatory.

The independent Apple M4 event retains all 22 original workloads and a separate
four-case sustained GC cohort. Five process blocks, two warmups, seven
interleaved calls, seed/counts, all named competitors and unchecked native C
where applicable remain fixed. Allocation statistics include every pause.
Complete setup/validation/compile/instantiate/first/warm/cleanup/process
intervals stay separate. Darwin affinity and physical power/background
isolation are unavailable, so this is advisory. Failed, unsupported, adverse
and unresolved outcomes are retained; there is no favorable retry.

Ordinary source-base/head benchmark jobs use the selected head compiler on
both source revisions. Such jobs can compare sources, but cannot stand in for
this explicit old/new compiler event. Historical reports and guest producers
keep their original pins. Controlled performance and production release stay open.


## Local original22 compiler event

Complete checksums/protocol pass. Complete controls resolve **2/22**. Lower elapsed ratios are better; unsupported is not zero.

| Workload | Native new/old | Interpreter new/old | Native / Go | / wazero | / Wasmtime | / Wasmer | Interpreter / Go | / wazero interpreter | / WasmEdge | Native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 0.954× | 1.085× | 1.277× | 9.221× | 3.458× | 1.509× | 1.333× | 3.038× | 1.246× | unavailable | inconclusive |
| core/integer-loop | 1.009× | 0.977× | 0.997× | 0.937× | 1.035× | 0.984× | 0.798× | 0.339× | 0.569× | 1.004× | inconclusive |
| core/function-calls | 0.967× | 1.010× | 0.992× | 0.534× | 0.851× | 0.987× | 0.860× | 0.547× | 0.781× | 1.025× | inconclusive |
| core/memory-load-store | 1.004× | 0.999× | 0.967× | 0.501× | 0.924× | 0.828× | 0.809× | 0.436× | 0.685× | 1.636× | inconclusive |
| memory/narrow-unaligned | 0.995× | 0.988× | 0.978× | 0.614× | 0.939× | 0.894× | 0.851× | 0.452× | 0.693× | 1.409× | inconclusive |
| memory/dependent-address | 1.003× | 0.988× | 0.993× | 0.580× | 0.895× | 0.881× | 0.851× | 0.395× | 0.703× | 1.309× | inconclusive |
| memory/float64-unaligned | 0.996× | 0.981× | 0.980× | 0.842× | 1.050× | 1.057× | 0.838× | 0.466× | 0.749× | 1.111× | inconclusive |
| runtime/simd-memory-full | 1.005× | 1.013× | 0.994× | 0.888× | 0.924× | 0.956× | 0.858× | 0.470× | 0.888× | 0.941× | inconclusive |
| runtime/simd-memory-widen | 1.005× | 0.999× | 0.993× | 0.929× | 1.017× | 1.004× | 0.907× | 0.417× | 0.985× | 1.013× | resolved, advisory |
| runtime/simd-memory-lanes | 1.020× | 1.016× | 1.002× | 0.821× | 0.860× | 0.863× | 0.876× | 0.470× | 0.974× | 1.132× | inconclusive |
| gc/i31-roundtrip | 0.979× | 1.035× | 0.943× | unavailable | 0.860× | unavailable | 0.861× | unavailable | 0.762× | unavailable | inconclusive |
| gc/struct-access | 0.980× | 1.037× | 0.842× | unavailable | 1.306× | unavailable | 0.881× | unavailable | 0.726× | unavailable | inconclusive |
| gc/struct-allocation | 1.050× | 0.977× | 0.929× | unavailable | 0.365× | unavailable | 1.299× | unavailable | 1.172× | unavailable | inconclusive |
| gc/struct-indirect-access | 0.999× | 1.037× | 0.932× | unavailable | 1.797× | unavailable | 0.898× | unavailable | 0.765× | unavailable | inconclusive |
| threads/private-memory | 1.020× | 0.990× | 0.947× | 1.243× | 1.290× | 1.139× | 0.782× | 0.576× | 0.948× | unavailable | inconclusive |
| threads/shared-memory | 0.985× | 0.919× | 0.881× | 1.204× | 1.238× | 1.136× | 0.423× | 0.569× | 0.901× | unavailable | inconclusive |
| threads/atomic-rmw | 0.958× | 1.014× | 0.919× | 1.069× | 1.079× | 0.544× | 0.610× | 0.760× | 1.105× | unavailable | inconclusive |
| threads/notify-empty | 0.991× | 0.982× | 1.000× | 0.069× | 0.203× | 0.275× | 0.829× | 0.751× | 0.775× | unavailable | inconclusive |
| threads/wait-mismatch | 0.920× | 0.968× | 0.797× | 0.054× | 0.151× | 0.476× | 0.639× | 0.558× | 0.893× | unavailable | inconclusive |
| threads/wait-timeout | 1.053× | 0.962× | 0.739× | 0.006× | 0.039× | 0.010× | 0.604× | 0.120× | 0.393× | unavailable | inconclusive |
| applications/sha256-clang | 1.001× | 0.964× | 0.982× | 0.686× | 1.086× | 1.046× | 0.840× | 0.434× | 0.661× | 1.227× | resolved, advisory |
| applications/sha256-rust | 1.021× | 0.971× | 0.994× | 0.669× | 1.157× | 1.134× | 0.847× | 0.418× | 0.707× | 1.286× | inconclusive |


## Separate local sustained GC

Complete checksums/protocol pass. Complete controls resolve **0/4**. Lower elapsed ratios are better; unsupported is not zero.

| Workload | Native new/old | Interpreter new/old | Native / Go | / wazero | / Wasmtime | / Wasmer | Interpreter / Go | / wazero interpreter | / WasmEdge | Native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| gc/i31-roundtrip | 1.006× | 1.050× | 1.000× | unavailable | 0.749× | unavailable | 0.876× | unavailable | 0.758× | unavailable | inconclusive |
| gc/struct-access | 1.008× | 1.028× | 0.995× | unavailable | 1.253× | unavailable | 0.912× | unavailable | 0.733× | unavailable | inconclusive |
| gc/struct-allocation | 1.094× | 0.989× | 1.018× | unavailable | 0.640× | unavailable | 1.242× | unavailable | 1.154× | unavailable | inconclusive |
| gc/struct-indirect-access | 1.002× | 1.021× | 0.995× | unavailable | 1.759× | unavailable | 0.886× | unavailable | 0.752× | unavailable | inconclusive |


The original event has 362 alternate checks and 12 explicit GC rejections;
sustained has 56 and 12. Public-call medians observe old/new LLGo at
3,625/3,458 ns, but the new alias is 3,959 ns and unchanged Go source roles
observe 3,209/2,708 ns. This spread does not establish a compiler speedup.
Native allocation observes new/old 1.050× original and 1.094× sustained,
including pauses; those adverse observations remain with unresolved controls.
Interpreter allocation remains behind Go. Direct/indirect GC, private/shared
Threads memory, public calls, atomics, some SIMD/applications and native-C gaps
remain. The verified static-conversion improvement is not broad VM parity.

## Independent Linux snapshot

[Workflow 42458772600](https://github.com/cpunion/wasmvm/actions/runs/42458772600)
uses common new compiler `38de1e218` on both source roles. Production VM source
maps remain identical. Its five groups have independent hosts, with a separate
sustained GC cohort on the GC host. It is a new-pin competitor snapshot,
not a measurement of old→new compiler benefit. It must not be pooled with
this local source-upgrade event or older Linux hosts.


All six Linux jobs pass; shared build is 396 seconds and the longest group is
494 seconds. Original controls resolve **6/22**, sustained **3/4**. Original
checks verify 362 alternate configurations and 12 explicit GC rejections;
sustained verifies 56 and 12. All 20 source/compiler identities and reuse of
the GC binaries in its separate sustained cohort are independently checked.

| Group | CPU model |
| --- | --- |
| core | Intel(R) Xeon(R) Platinum 8370C CPU @ 2.80GHz |
| memory-simd | AMD EPYC 7763 64-Core Processor |
| gc | AMD EPYC 9V74 80-Core Processor |
| threads | AMD EPYC 7763 64-Core Processor |
| applications | AMD EPYC 7763 64-Core Processor |

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control / median-v1 | 1.501× | 1.428× | 9.112× | 4.242× | 2.460× | 3.044× | 1.858× | unavailable | inconclusive |
| core/integer-loop / median-v1 | 1.002× | 0.773× | 1.076× | 1.039× | 1.038× | 0.726× | 0.925× | 1.077× | resolved, advisory |
| core/function-calls / median-v1 | 0.997× | 0.877× | 0.581× | 0.810× | 0.689× | 0.802× | 1.224× | 0.934× | resolved, advisory |
| core/memory-load-store / median-v1 | 1.000× | 0.824× | 0.641× | 1.176× | 1.170× | 0.809× | 0.976× | 1.455× | inconclusive |
| memory/narrow-unaligned / median-v1 | 0.999× | 0.973× | 0.868× | 1.312× | 1.307× | 0.485× | 0.814× | 2.073× | resolved, advisory |
| memory/dependent-address / median-v1 | 0.989× | 0.892× | 0.975× | 1.307× | 1.304× | 0.494× | 0.391× | 1.528× | resolved, advisory |
| memory/float64-unaligned / median-v1 | 0.932× | 0.924× | 0.632× | 1.058× | 1.049× | 0.578× | 0.437× | 1.076× | resolved, advisory |
| runtime/simd-memory-full / median-v1 | 0.995× | 0.881× | 0.857× | 1.065× | 1.027× | 0.487× | 0.984× | 0.888× | inconclusive |
| runtime/simd-memory-widen / median-v1 | 0.995× | 0.905× | 0.677× | 1.527× | 1.506× | 0.539× | 0.633× | 0.393× | inconclusive |
| runtime/simd-memory-lanes / median-v1 | 0.998× | 1.069× | 0.939× | 1.135× | 1.141× | 0.517× | 1.164× | 1.318× | resolved, advisory |
| gc/i31-roundtrip / median-v1 | 1.244× | 0.837× | unsupported | 1.134× | unsupported | unsupported | 0.471× | unavailable | inconclusive |
| gc/struct-access / median-v1 | 1.049× | 0.819× | unsupported | 1.784× | unsupported | unsupported | 0.416× | unavailable | inconclusive |
| gc/struct-allocation / mean-including-pauses-v1 | 1.657× | 1.503× | unsupported | 0.603× | unsupported | unsupported | 1.292× | unavailable | inconclusive |
| gc/struct-indirect-access / median-v1 | 0.978× | 0.838× | unsupported | 1.082× | unsupported | unsupported | 0.476× | unavailable | inconclusive |
| threads/private-memory / median-v1 | 1.148× | 1.002× | 0.763× | 2.223× | 1.853× | 0.643× | 0.551× | unavailable | inconclusive |
| threads/shared-memory / median-v1 | 1.269× | 0.583× | 0.865× | 3.121× | 2.348× | 0.662× | 0.548× | unavailable | inconclusive |
| threads/atomic-rmw / median-v1 | 1.058× | 0.602× | 1.604× | 1.486× | 1.438× | 0.557× | 0.598× | unavailable | inconclusive |
| threads/notify-empty / median-v1 | 1.085× | 0.872× | 0.070× | 0.290× | 0.205× | 0.520× | 0.484× | unavailable | inconclusive |
| threads/wait-mismatch / median-v1 | 0.987× | 0.630× | 0.051× | 0.204× | 0.541× | 0.487× | 0.485× | unavailable | inconclusive |
| threads/wait-timeout / median-v1 | 0.854× | 0.621× | 0.005× | 0.046× | 0.023× | 0.103× | 0.001× | unavailable | inconclusive |
| applications/sha256-clang / median-v1 | 0.996× | 0.910× | 0.655× | 1.240× | 1.208× | 0.595× | 0.853× | 1.489× | inconclusive |
| applications/sha256-rust / median-v1 | 0.986× | 0.890× | 0.661× | 1.303× | 1.230× | 0.587× | 0.855× | 1.543× | inconclusive |


### Separate Linux sustained GC

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| gc/i31-roundtrip / median-v1 | 0.997× | 0.837× | unsupported | 0.723× | unsupported | unsupported | 0.812× | unavailable | resolved, advisory |
| gc/struct-access / median-v1 | 1.002× | 0.817× | unsupported | 1.267× | unsupported | unsupported | 0.547× | unavailable | resolved, advisory |
| gc/struct-allocation / mean-including-pauses-v1 | 1.289× | 1.433× | unsupported | 0.821× | unsupported | unsupported | 1.390× | unavailable | inconclusive |
| gc/struct-indirect-access / median-v1 | 0.993× | 0.812× | unsupported | 0.954× | unsupported | unsupported | 0.668× | unavailable | resolved, advisory |


The Linux public-call control observes LLGo 9.285 μs, Go 6.185 μs, wazero
1.019 μs, Wasmtime 2.189 μs and Wasmer 3.774 μs, with unresolved controls.
Narrow memory and widening SIMD are 1.312×/1.527× Wasmtime. Sustained GC direct
access is 1.267× Wasmtime with calibrated controls; allocation interpretation
is 1.433× Go with unresolved controls. These gaps remain. Every group keeps its
own CPU identity; older absolute times on other hosts cannot isolate compiler
benefit. The common new-compiler source roles have identical VM code, so their
head/base times are controls rather than an implementation optimization.

The [complete machine record](data/performance-llgo-interface-upgrade-2026-11-08.json)
retains original and sustained data from both independent events, phases,
orders, compiler/SDK/source identities, all controls and failed input/audit
preflight evidence. The first hosted audit omitted the job-metadata file;
that failed log is retained, and the audit then consumed the actual REST job
record without changing observations or measurements. Controlled machines,
complete performance acceptance and production release remain outstanding.
