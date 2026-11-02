# Linux combination preflight and allocation study: 2026-11-02

This once-only PR-scope workflow tests a combination of separately retained
integer-pair, outlined-vector-load and comparison-liveness candidates. It is not
the isolated result for PR #99 or a controlled portfolio acceptance campaign.

Base `4e6cd7bb5898a12febcd2fa00745a0559528e534`; head
`0063067b1c128fa2c165734049e2a0b166ce4c5a`. The [workflow](https://github.com/cpunion/wasmvm/actions/runs/41428619382)
passes with all 228 phase observations, 76 summaries and 19 configurations.
Archive digest, exact base/head source inventories and pinned LLGo
`8e45ed5` `-O3 -lto=full` validate. Source inventories contain 1,454/1,459
files; their contents independently match git archives. Go remains 1.27.3.
Three fresh process rounds, two warmups, seven measured calls and the unchanged
PR fixtures/iteration counts retain all exposed phases and timer controls.

Core process-median rates, ns/source iteration:

| Workload | Go interpreter base → head | LLGo interpreter base → head | LLGo native base → head | WasmEdge interpreter | Wasmtime native |
| --- | ---: | ---: | ---: | ---: | ---: |
| Integer | 55.857 → 53.872 | 57.011 → 53.425 | 1.348 → 1.345 | 47.632 | 1.322 |
| Calls | 82.056 → 88.340 | 105.253 → 107.509 | 0.688 → 0.688 | 63.307 | 1.543 |
| Memory | 110.381 → 117.300 | 114.040 → 113.511 | 0.860 → 0.830 | 97.377 | 0.622 |

The combination regresses Go interpreter calls 7.7% and memory 6.3%, and LLGo
calls 2.1%. LLGo integer improves about 6.3%, while native memory improves about
3.5%. Native memory still takes 1.33x Wasmtime's time. The complete raw table
retains Wasmer, wazero and dispatch/call controls as separate configurations.
No favorable local ARM64 rate is substituted for these x86 outcomes, and rates
from separate workers/campaigns are not pooled. This PR-scope run has no ordinary
native-C algorithm row; those controls remain in their matching-count local
campaigns, distinct from the C API adapters used by the reference engines here.

The exact-head PR #99 benchmark must isolate the first candidate before merge.
Later candidates require their own final-head checks and retained deficits.
The combined preflight does not justify attributing a regression to one candidate
or to LLGo without further isolation.

## Concrete native allocation evidence

A compile-only diagnostic captures the real scalar-memory slot IR, fixed-register
plan and generated code at `ae3a754`. The allocation implementation is unchanged
through the combination head. Diagnostic-generated amd64/arm64 functions match
the original captured bodies byte-for-byte; the foreign code is never executed.
The input function has seven locals and ten slots. ARM64 caches every scalar slot
and an immutable word, while amd64's memory function has six scalar registers.

On amd64, slots 8/9 receive weights 2,180/384; the running value in local slot 5
receives 192 and spills. The hot body writes that value at frame offset 80,
reloads it for the guest store, then reloads it again for the accumulation.
The selected literal slot has no arithmetic read in that hot sequence: its
operations use encoded immediates, and two earlier literal writes disappear.
The current weights therefore count work that the selected hot instructions
avoid while undervaluing an actual store/two-load chain. This is an owned
compiler cost-model finding, not evidence of a new LLGo defect.

The next candidate should model emitted immediate/temporary costs rather than
adding workload-specific priorities. Stable register assignments and generic
cold-PC handling must remain intact. Merely splitting sums was already rejected
in the [reduction experiment](performance-native-reduction-held-2026-11-02.md).
This assembly/weight evidence motivates the next change; it does not quantify
its unimplemented throughput benefit or explain every part of the native gap.

Only the explicitly passing `TestAllocationDiagnostic` log establishes the
allocation diagnostic's execution. The earlier no-test invocation remains a
separate diagnostic artifact. Compilation and actual guest execution are not
interchanged in this report.

[Lossless evidence](evidence/performance-linux-combined-2026-11-02.json.gz)
retains every Linux phase/configuration, source/artifact audits, allocation IR
and plans, assembly, the diagnostic/test program, source identities and logs.
