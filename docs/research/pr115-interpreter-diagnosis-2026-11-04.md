# PR115 interpreter diagnosis

Date: **2026-11-04**. Status: **diagnostic evidence; controlled performance acceptance remains open**.

The original published PR115 engine was `168bf9a` (engine-file identical to measured `4a30329`). The merged baseline is `cb1e245`. All follow-ups below use separate source revisions and remain outside PR115 unless its body explicitly identifies an adopted revision. No compiler/profile upgrade or reference fallback is involved. LLGo issue [#2983](https://github.com/xgo-dev/llgo/issues/2983) is handled by another contributor; this investigation does not modify it.

## Linux Go controls

Go 1.27.3, CGO enabled, normal release compilation; three unchanged kernels each use **5,000,000 iterations**, **five fresh interleaved process blocks**, two warmups and seven measured calls. A/A aliases execute the exact baseline binary. Every series verifies 45 observations and nine alternate-input checks. Linked pprof wrappers are separate diagnostics and never replace these observations. Different CPU hosts and source revisions are separate cohorts: do not rank or pool their percentages.

| Source variant | Commit | Integer | Calls | Memory | CPU |
| --- | --- | ---: | ---: | ---: | --- |
| Original PR115 | `168bf9a` | +5.87% | +6.01% | +13.15% | EPYC 9V45 96-Core Processor |
| GC outlining | `6aafcfb` | +14.93% | +14.42% | +17.74% | EPYC 7763 |
| Exception outlining | `2908cf1` | +14.86% | +7.52% | +13.82% | Xeon 8370C CPU @ 2.80GHz |
| Numeric dispatch | `8ec1caa` | +12.46% | +7.10% | +7.95% | EPYC 7763 |
| Local counter | `f99fce2` | +17.69% | +8.18% | +22.22% | EPYC 7763 |
| Low-word operands + numeric commit | `3ce7243` | -1.98% | -0.92% | -3.55% | EPYC 7763 |
| Numeric commit only | `bd0c30c` | -11.97% | -0.71% | -4.08% | EPYC 9V45 96-Core Processor |

The original shorter hosted PR suite separately observed Go interpreter **+24.6/+19.9/+29.5%** for integer/calls/memory; its LLGo interpreter improved. These original outcomes remain retained. A green semantic workflow does not resolve the performance regression.

## Local dual-compiler observations

All prototypes pass the five affected packages and **48 pinned execution reports per compiler**, including all five mandatory extensions at three tiers and forced Stack Switching collection. This is affected-scope macOS arm64 validation, not a replacement for the final 31-package dual-platform matrix.

The following original 20-kernel cohorts each retain **864 VM/reference observations, 27 ordinary unchecked C observations, 288 alternate checks, 12 explicit unsupported-reference rejections and 297 summaries**. Native and interpreter, Go and LLGo, all five named reference configurations, preparation/first/warm phases and adverse rows remain separate. Three fresh process blocks describe this workstation; they do not establish controlled acceptance.

| Variant | Go integer / calls / memory | LLGo integer / calls / memory |
| --- | --- | --- |
| Numeric dispatch | -11.81% / -5.13% / +2.57% | -9.25% / -6.98% / -9.23% |
| Local counter | -12.01% / -4.77% / -0.77% | -11.40% / -7.73% / -11.75% |
| Low-word + numeric commit | -7.18% / -6.08% / +2.22% | -0.34% / -1.75% / +4.08% |
| Numeric commit only | -5.95% / -5.68% / -0.96% | -10.11% / -8.27% / -13.71% |

Retained adverse outcomes include GC outlining: long LLGo interpreter field access **+9.88%**, allocation **+5.03%**; low-word/commit combined: interpreter memory **+4.08%**, atomics **+6.56%**, shared memory **+4.70%**. The local-counter cohort has short native GC and notify/wait increases. Tiny-call controls have large A/A variation. No adverse row is removed or replaced by a longer-count result.

The earlier six-case long scopes, empty-object scope and all exception throw/catch checks remain distinct. The new EH fixture initially used an incorrect copied integer oracle; all runtimes agreed on the independent corrected LCG result. Both attempts and the correction are retained. Its baseline LLGo native throw/catch costs **425.61 ns/throw**, versus Go **181.69** and Wasmtime **84.78**; this is an existing gap, not a demonstrated LLGo defect.

## Sampling the production binaries

[System-sampling run](https://github.com/cpunion/wasmvm/actions/runs/41719056489) verifies **12 checksummed observations** across baseline, original PR, numeric-dispatch and local-counter sources on an EPYC 7763. The executable has the ordinary main package and compiler flags, with no linked profiler wrapper. Software `cpu-clock:u` sampling succeeds; the hardware-counter probe fails and is explicitly retained. No IPC, cache-miss or branch-miss cause is claimed.

Baseline/original `Execution.Run` have 25,135-byte bodies and identical normalized instructions; their addresses differ. Samples concentrate in the dispatch loop, including the constant-prefix mask, full-value high-word operand loads/spills and the common trap-commit test. These samples motivate experiments; sampling skid and unavailable hardware events prevent attributing a microarchitectural root cause. Re-linked pprof wrappers changed layout and reversed relative outcomes in an earlier diagnostic, so their timings are excluded.

The numeric-commit experiment keeps a contiguous 128-opcode numeric dispatch. Potentially trapping `numeric.Eval` results still enter the common decorated trap path; known nontrapping operations commit directly. The ablation preserves original full-value operands, e.steps, inline GC/EH handlers, exact fuel, polling, roots and every other feature. The operand test supplies poisoned high words and verifies zero-high scalar results.

## Evidence and remaining gates

The [evidence archive](evidence/pr115-interpreter-diagnosis-2026-11-04.json.xz) contains raw results, source/binary/toolchain identities, commands, execution reports, failed observations, system samples and annotated assembly. Decode the XZ container with Python `lzma.open` or `xz -dc`. Each file entry records its encoding, original content and SHA-256 over decoded bytes; SDKs, executables and source archives are excluded. All archived bytes have been round-trip verified.

Archive: **1005 files**, **2112688 bytes**, SHA-256 `bc3d7b4431db4e901b46f8c570811a8ea98e987f343063b64a37b623944e88bb`.

Selected source `bd0c30c` also verifies eight longer cases: 300 observations, 60 alternate checks and 75 independent oracle checks. Native memory changes Go +0.21% (A/A +0.23%) / LLGo +0.04% (A/A -0.20%); native i31/access/indirect are near flat. LLGo native allocation +0.65% is retained. All original shorter adverse observations stay visible. Final exact-head target CI and review remain required before integration. Native GC allocation remains about **5.32x Wasmtime** in the original 300,000-object cohort despite the native-sweep improvement. Expanded WASI world conformance and complete controlled portfolio/native-C performance acceptance also remain open.
