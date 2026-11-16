# Native Full-LTO switch dispatch diagnostic, 2026-11-16

This records the investigation for LLGo [#3094](https://github.com/xgo-dev/llgo/issues/3094),
[#3098](https://github.com/xgo-dev/llgo/issues/3098), and the independently verified
[#3097](https://github.com/xgo-dev/llgo/issues/3097). It includes rejected hypotheses,
all completed performance cohorts and unresolved same-binary controls. Results
are focused hosted diagnostics, not universal or full-portfolio acceptance.

## Compiler policy and code checks

The final default-policy experiment compares LLGo main
`5860fb6f9c13371f6cdb456a987a8e95844066d4` with
`6b1da473fc50374348d9968f56f226c7e7cfc492`, Go 1.27.3, LLVM 22.1.8, native
Linux amd64, normal GC, unchanged `-O3 -lto=full`. The contribution head
`a7dfe24cf` additionally preserves explicitly selected alternate linkers; default
policy is unchanged. Its complete source diff is supplied in the data directory.

LLVM 22 can classify each arm of an unprofiled large switch as cold and stop its
inline-cost walk early. The implementation gives eligible calls in cyclic
switch regions the ordinary O3 budget of 250 and enables complete cost analysis.
It preserves explicit inline policies, PGO, cold/noreturn barriers and LLVM's
cost/viability decisions. Cache-line loop alignment and 32-byte branch-boundary
padding accompany the inline budget because inline-only output still regressed
some unchanged dispatchers through code placement. Calls/returns are excluded
from branch padding to retain adjacent PC metadata labels.

The policy applies only to Linux amd64 O3 Full-LTO executable links through
ELF LLD. Other profiles and explicit linker/inline/layout overrides retain their
own policy. Full-LTO native tests pass for Caller, CallersFrames, FuncForPC,
function-value entry PCs and external PCLN metadata on both Linux runner images.
Unit tests verify IR scope, unreachable/acyclic exclusion, PGO/attribute
preservation and driver option precedence.

The frozen compact-dispatch body has all four extra hot helper calls on main:
Memory.Load, Execution.poll, Memory.Store and runtime.NewSlice2. All four are
absent in the final fixed disassembly; guest CRCs are unchanged. This does not
make slicing unchecked: NewSlice2 performs checks/header construction rather
than allocation. Its checks remain in the inlined body.

## Protocol and identities

The counter, compact/dense and scalar interpreter sources are the exact frozen
artifacts from runs 43537347012, 43543803052 and 43545680097. Each compiled
engine/harness source hash is checked against its manifest. The dependent-address
guest is checked against result 509818616, seed 305419896, 100000 iterations.
Each cohort uses 20 fixed blocks, four fresh process labels (base, base-aa, fix,
fix-aa), eight checked warm calls after two warmups. A/A labels use identical
binary bytes. Source/compiler revisions, executable hashes, commands and every
observation survive in the results. Ratios are geometric means of paired block
mean-time ratios, with a 10000-resample block bootstrap (seed 20261116).
No timing-dependent early stopping or pooling across hosts/cohorts is used.
The two replay batches each declared four host jobs in advance. The second batch
expanded hardware coverage after the first landed entirely on AMD; all its
results are retained, including regressions and unresolved controls.

The combined-build event completed all engine cohorts and semantic/PCLN tests,
but hit the job's 20-minute limit during subsequent portfolio compilation. Those
completed engine results remain valid separate events. The portfolio was split
into bounded jobs and completed successfully in run 43785100450. Frozen-binary
replays avoid recompilation and reuse the hash-checked Ubuntu 22.04 artifacts
from combined-build run 43782740347 on new native Ubuntu 22.04/24.04 hosts.
Runtime libraries follow each host image; code bytes do not change on replay.

## Final combined policy: every engine cohort

A ratio below 1 is faster. Controls require both A/A 95% intervals to lie inside
[0.975, 1.025] for that row; unresolved rows are observations, not calibrated
performance acceptance. In particular the Intel scalar result has an unresolved
head A/A control, while its counter and dense rows are calibrated.

| Event/host job | CPU | Source | Fixed/main time ratio, 95% interval | A/A calibrated |
|---|---|---|---|---|
| combined-build/ubuntu-22.04 | AMD EPYC 9V74 80-Core Processor | counter | 0.9921 [0.9883, 0.9958] | unresolved |
| combined-build/ubuntu-22.04 | AMD EPYC 9V74 80-Core Processor | dense | 0.8140 [0.8069, 0.8256] | yes |
| combined-build/ubuntu-22.04 | AMD EPYC 9V74 80-Core Processor | scalar | 0.9225 [0.9127, 0.9320] | yes |
| combined-build/ubuntu-24.04 | AMD EPYC 7763 64-Core Processor | counter | 1.0320 [1.0031, 1.0826] | unresolved |
| combined-build/ubuntu-24.04 | AMD EPYC 7763 64-Core Processor | dense | 0.7776 [0.7709, 0.7841] | yes |
| combined-build/ubuntu-24.04 | AMD EPYC 7763 64-Core Processor | scalar | 0.9322 [0.9255, 0.9378] | yes |
| combined-replay-1/ubuntu-22.04-1 | AMD EPYC 9V74 80-Core Processor | counter | 1.0014 [0.9931, 1.0132] | yes |
| combined-replay-1/ubuntu-22.04-1 | AMD EPYC 9V74 80-Core Processor | dense | 0.8071 [0.7928, 0.8172] | yes |
| combined-replay-1/ubuntu-22.04-1 | AMD EPYC 9V74 80-Core Processor | scalar | 0.9131 [0.9056, 0.9208] | unresolved |
| combined-replay-1/ubuntu-22.04-2 | AMD EPYC 7763 64-Core Processor | counter | 1.0220 [1.0108, 1.0383] | unresolved |
| combined-replay-1/ubuntu-22.04-2 | AMD EPYC 7763 64-Core Processor | dense | 0.7746 [0.7632, 0.7840] | yes |
| combined-replay-1/ubuntu-22.04-2 | AMD EPYC 7763 64-Core Processor | scalar | 0.9376 [0.9326, 0.9429] | yes |
| combined-replay-1/ubuntu-24.04-1 | AMD EPYC 7763 64-Core Processor | counter | 1.0093 [0.9970, 1.0207] | yes |
| combined-replay-1/ubuntu-24.04-1 | AMD EPYC 7763 64-Core Processor | dense | 0.7977 [0.7825, 0.8210] | unresolved |
| combined-replay-1/ubuntu-24.04-1 | AMD EPYC 7763 64-Core Processor | scalar | 0.9457 [0.9365, 0.9559] | yes |
| combined-replay-1/ubuntu-24.04-2 | AMD EPYC 9V74 80-Core Processor | counter | 0.9900 [0.9823, 0.9958] | yes |
| combined-replay-1/ubuntu-24.04-2 | AMD EPYC 9V74 80-Core Processor | dense | 0.8032 [0.7983, 0.8079] | yes |
| combined-replay-1/ubuntu-24.04-2 | AMD EPYC 9V74 80-Core Processor | scalar | 0.9265 [0.9202, 0.9333] | yes |
| combined-replay-2/ubuntu-22.04-1 | AMD EPYC 7763 64-Core Processor | counter | 1.0037 [0.9749, 1.0197] | unresolved |
| combined-replay-2/ubuntu-22.04-1 | AMD EPYC 7763 64-Core Processor | dense | 0.7796 [0.7736, 0.7861] | yes |
| combined-replay-2/ubuntu-22.04-1 | AMD EPYC 7763 64-Core Processor | scalar | 0.9310 [0.8876, 0.9730] | unresolved |
| combined-replay-2/ubuntu-22.04-2 | AMD EPYC 9V45 96-Core Processor | counter | 0.9620 [0.9494, 0.9725] | unresolved |
| combined-replay-2/ubuntu-22.04-2 | AMD EPYC 9V45 96-Core Processor | dense | 0.7670 [0.7556, 0.7779] | yes |
| combined-replay-2/ubuntu-22.04-2 | AMD EPYC 9V45 96-Core Processor | scalar | 0.9636 [0.9517, 0.9754] | yes |
| combined-replay-2/ubuntu-24.04-1 | Intel(R) Xeon(R) 6975P-C | counter | 1.0136 [1.0088, 1.0184] | yes |
| combined-replay-2/ubuntu-24.04-1 | Intel(R) Xeon(R) 6975P-C | dense | 0.8642 [0.8605, 0.8679] | yes |
| combined-replay-2/ubuntu-24.04-1 | Intel(R) Xeon(R) 6975P-C | scalar | 0.8553 [0.8517, 0.8586] | unresolved |
| combined-replay-2/ubuntu-24.04-2 | AMD EPYC 7763 64-Core Processor | counter | 1.0092 [0.9901, 1.0254] | yes |
| combined-replay-2/ubuntu-24.04-2 | AMD EPYC 7763 64-Core Processor | dense | 0.7896 [0.7832, 0.7962] | unresolved |
| combined-replay-2/ubuntu-24.04-2 | AMD EPYC 7763 64-Core Processor | scalar | 0.9319 [0.9113, 0.9465] | unresolved |

The compact source improves on the tested Intel and AMD machines. Counter
performance has a small real regression on Intel (about 1.36% in its calibrated
cohort); the policy does not improve every workload. The scalar path has fewer
spills already on unmodified main; its result must not be attributed to a new
spill-elimination fix here.

## Existing LLGo JSON/template/field benchmarks

Both SDK roots compile the existing, hash-identical benchmark/crossinline,
benchmark/crossinline/json and benchmark/crossinline/template source files.
Their semantic tests run successfully before timing. There are 20 fixed blocks
and four fresh process labels, GOGC=100/GOMAXPROCS=1, with fixed iteration counts:
50000000 for fields, 200 JSON round trips and 1000 template executions per sample.
No new application fast path or custom workload implementation is introduced.
All benchmark output and same-binary controls are retained.

| CPU | Benchmark | Fixed/main ns/op ratio, 95% interval | A/A calibrated |
|---|---|---|---|
| AMD EPYC 9V74 80-Core Processor | BenchmarkFunctionFields | 0.9962 [0.9915, 1.0007] | yes |
| AMD EPYC 9V74 80-Core Processor | BenchmarkValueMethodFields | 0.9845 [0.9628, 0.9978] | unresolved |
| AMD EPYC 9V74 80-Core Processor | BenchmarkPointerMethodFields | 0.9952 [0.9886, 1.0000] | yes |
| AMD EPYC 7763 64-Core Processor | BenchmarkFunctionFields | 1.0091 [1.0006, 1.0156] | yes |
| AMD EPYC 7763 64-Core Processor | BenchmarkValueMethodFields | 1.0147 [1.0094, 1.0205] | yes |
| AMD EPYC 7763 64-Core Processor | BenchmarkPointerMethodFields | 1.0073 [1.0018, 1.0139] | yes |
| INTEL(R) XEON(R) PLATINUM 8573C | BenchmarkJSONRoundTrip | 0.9872 [0.9696, 1.0050] | yes |
| AMD EPYC 7763 64-Core Processor | BenchmarkJSONRoundTrip | 0.9872 [0.9704, 1.0002] | unresolved |
| INTEL(R) XEON(R) PLATINUM 8573C | BenchmarkTemplateExecute | 0.9917 [0.9886, 0.9945] | yes |
| Intel(R) Xeon(R) 6975P-C | BenchmarkTemplateExecute | 1.0009 [0.9982, 1.0038] | yes |

JSON and template show no clear regression in these events. EPYC 7763 field
microbenchmarks regress about 0.7–1.5%; these small measured regressions are
retained, not called noise. Some JSON/field A/A controls remain unresolved.

## Code size and compilation cost

The native interpreter ELF grows about 3.0% (9.26 to 9.55 MB decimal). Its .text
grows about 9.2% (2.67 to 2.92 MB), largely from alignment/padding. The compact
loop grows from 6803 to 7661 bytes, combining inline bodies and padding. Existing
JSON/template/field ELF files grow about 2.8%; this is an O3 speed/size tradeoff.
Os/Thin/off profiles are outside the policy.

Build timing surrounds the compiler invocation only, ending before symbol/section
inspection. On the initial AMD builds, engine base/fix compiles are about
88–96 seconds on 9V74 and 112–122 seconds on 7763; paired differences are small.
Portfolio builds likewise show no clear compile-time increase. These are a single
serial base/fix compile per package/host, with evolving caches, not a statistically
controlled compiler-speed claim. Earlier inline-only timings included inspection
and must not be pooled with the newer pure-build-time records.

## Rejected alternatives and independent spill result

- Heap/slice lowering plus known-nil panic noreturn annotations passed semantic
  tests but regressed dense dispatch about 5.6–8.9% on two AMD hosts. That branch
  is not proposed. Its independent run 43766943202 also verifies that the scalar
  operand spill is already absent on unmodified main 113daaad5: 7320-byte loop,
  680-byte stack reservation, first operand retained in r10 while decoding the
  second address. The original report's loop was 7494 bytes/696-byte stack.
- Inline-budget-only removes the four calls but is placement-sensitive: its Intel
  counter regresses about 8.5%; its AMD dense case regresses about 11.3% with an
  unresolved control. The combined policy is a separate measured event.
- ELF page-separated text fails to restore the original AMD layout baseline;
  it can slow the original fast source too. It is not adopted.
- On the original #3094 frozen compiler/source pair, branch-boundary padding
  alone regresses both AMD and Intel. Loop-64 plus branch-32, as a combination,
  beats the original fast source on AMD and Intel; neither standalone setting
  supplies that result. No function alignment, pinned address or source padding
  workaround is introduced.

The old layout-only event keeps its original compiler 56fa9ae and source pair
b546ea76/95943842, whereas final combined events use main 5860fb6f9 and the three
later frozen roles. Those distinct baselines are not mixed into one ratio.

## Artifacts and reproducibility

[Data](data/llgo-native-dispatch-20261116/) contains every captured result JSON,
disassembly, source diff and the executed diagnostic scripts. JSON files are
gzip-compressed without dropping observations; the manifest hashes each original
uncompressed capture. Scripts retain original checkout paths and artifact run
IDs. Recreate the recorded directory structure or adjust paths before rerunning.
The original build/replay artifacts also preserve executable bytes and hashes.

Runs: [rejected lowering](https://github.com/cpunion/wasmvm/actions/runs/43766943202),
[inline only](https://github.com/cpunion/wasmvm/actions/runs/43769960378),
[page separation](https://github.com/cpunion/wasmvm/actions/runs/43768981184),
[original layout combinations](https://github.com/cpunion/wasmvm/actions/runs/43772290928),
[combined build](https://github.com/cpunion/wasmvm/actions/runs/43782740347),
[portfolio](https://github.com/cpunion/wasmvm/actions/runs/43785100450),
[replay batch 1](https://github.com/cpunion/wasmvm/actions/runs/43785280364),
[replay batch 2](https://github.com/cpunion/wasmvm/actions/runs/43785877143).
