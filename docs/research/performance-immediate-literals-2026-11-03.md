# Shared integer immediates and short-call clock precision: 2026-11-03

Status: **earlier both-target candidate; the contribution now restores original
amd64 placement after its adverse Linux result**. See the
[corrective follow-up](performance-arm-immediates-x86-preserved-2026-11-03.md).
New-head CI and controlled performance acceptance remain open. This report
retains the original measured sources and observations. Original production is `a176cb5`.
`9eda283` shares immediate encoding/eligibility and constant-pair fusion between
amd64 and arm64. `85c0210` also drops a literal overwritten by the original
arithmetic result when fusion retargets its output into a local. The compiler
restores the original result slot when it remains observable. This corrects a
missed optimization on both targets without changing the private ABI or adding
cache state. Independent cold PCs, live constants, spilled fallback, exact
frames and 256-instruction polling boundaries remain checked. Neither held heap
cache source is included.

## Source and semantic evidence

A single pure ARM selector supplies both eligibility and encoding. This avoids
predicate/encoder drift when constant fusion declines a spilled/unsupported
operation. Existing exhaustive constant, alias and local-forwarding tests now
execute the ARM selector; new tests cover boundaries, unsupported masks,
source/destination aliases, both integer widths and overwritten literals after
local set/tee. The bitfield rejection test now probes that selector directly:
independent immediate fusion is legal on either target and must not masquerade
as an unsafe bitfield transformation. Every original PC still compares complete
frames against independent numeric semantics.

Go passes all 31 packages/thirty package coverage gates; pinned full-LTO LLGo
passes codegen/nativeexec/VM, plus affected race, no-cgo and vet. The unrelated
indirect-handle tests also pass against original production. Exact archived
body witnesses use the defined `run` export, module memory/type/global config,
both targets, metered/unmetered and ordinary/expanded tiers: 48 body pairs,
28 changed unmetered pairs; **all metered bytes are identical**. Initial diagnostic
fixture-path/export assumptions and the old aggregate fusion-counter assertion
are retained separately from successful witnesses/semantic evidence.

## Explicit timer baseline event

The independent unchanged seed-return C probe reports this host's
CLOCK_MONOTONIC resolution as 1000 ns, with 982/1000 measured intervals zero.
CLOCK_MONOTONIC_RAW reports 42 ns and 691 zero intervals; such tiny C calls
remain unresolved by a single timer pair. No zero is interpreted as zero cost.
`f9f7cc0` / `fbacef5` select the finer clock in embedded C API and native C probes
on Darwin, exposing the extension before headers under strict POSIX builds.
Linux retains CLOCK_MONOTONIC. Both C reference adapters are rebuilt against
identical pinned SDKs with the same new header; native algorithms and Go timers
are unchanged. The initial strict-POSIX header failure is retained, followed by
successful real SDK builds. All 78 benchmark-tool tests pass. Component probe
clocks and batched short-call/controlled latency acceptance remain future work.

## Final same-campaign comparison

M4 Max/macOS arm64, Go 1.27.3, LLGo `8e45ed5`, -O3 -lto=full; Wasmtime 48.0.1,
pinned wazero/Wasmer and ordinary Clang-22 O3 C controls. Original versus full
selected engine `fbacef5` (production `85c0210`), with common new C timing.
Five fresh interleaved process blocks, two warmups/seven samples; unchanged
original guest bytes and finite runtime inputs, one million iterations except
GC allocation 10,000 and empty control one invocation. Independent alternate
oracles, exact source/compiler/binary hashes and baseline A/A aliases are kept.
All checks/builds finish before timing. 320 VM/35 C observations, 64 VM/seven C
alternate checks, eight explicit GC reference rejections and 71 summaries are
independently verified. Startup, compile/instantiate, first/warm calls and
completed cleanup stay separate; no overhead subtraction or cross-study pooling.

Warm ns/source iteration (empty control: ns/call):

| Kernel | Go base → head | LLGo base → head | Wasmtime | wazero | Wasmer | Ordinary C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| i31 | 1.365375 → 1.376250 | 1.350416 → 1.351917 | 1.828750 | unsupported | unsupported | 1.290625 |
| struct access | 1.327084 → 1.334750 | 1.312916 → 1.317000 | 1.035292 | unsupported | unsupported | 1.031292 |
| allocation | 45.4083 → 47.7958 | 55.5417 → 55.2375 | 107.8292 | unsupported | unsupported | — |
| indirect access | 4.604750 → 4.491792 | 4.587916 → 4.496333 | 2.552709 | unsupported | unsupported | 2.815334 |
| empty control | 4667 → 5667 | 4584 → 5125 | 1000 | 459 | 3042 | 83* |
| integer | 1.584458 → 1.558917 | 1.554750 → 1.556459 | 1.546291 | 1.695000 | 1.556833 | 1.558292 |
| calls | 0.792875 → 0.793417 | 0.781000 → 0.789417 | 0.899875 | 1.319625 | 0.779959 | 0.772458 |
| memory | 0.436167 → 0.438541 | 0.433167 → 0.436458 | 0.473875 | 0.786500 | 0.526333 | 0.267250 |

LLGo indirect improves **2.0%**, versus A/A +0.09%; Go improves 2.45%, versus
A/A +0.51%. It still takes 1.76x Wasmtime and 1.60x C. Preserve adverse LLGo
calls +1.08% (A/A +0.05%), Go allocation +5.26% (A/A -0.78%), LLGo empty +11.8%
(A/A +19.1%) and Go empty +21.4% (A/A +21.4%). I31, struct access, memory and
other series remain in full; small changes are not waived as unchanged controls.
The one-minute host load is 10.3→9.7; this is a shared-host study, not controlled
acceptance. The 2% local target improvement does not establish a regression-free
portfolio or all-platform parity.

The three gap axes remain distinct: native memory beats the three measured VM
references but takes 1.63x C; indirect access and direct struct access still trail
Wasmtime; LLGo allocation takes 1.16x Go, while its ordinary integer/call loops
are near C. No new LLGo defect is inferred from throughput ratios. Existing
independent compiler reports and startup/host-boundary deficits remain open.

Head compile/validate medians are 149–279 microseconds for Go and 159–351 for
LLGo. Process readiness is 4.1–4.5 ms versus 92.6–97.4 ms. Baseline LLGo empty
compile is 399 microseconds versus A/A 164; these variable phase observations
are retained without attributing a compiler/startup improvement to this patch.

*C control uses an ordinary volatile indirect seed-return entry plus its timer/
dispatch, only about two RAW-clock ticks here. It is not a precise pure-call
cost or a VM contract equivalent. C may optimize algorithms normally and omits
VM quotas, GC, cancellation and typed continuations. No C GC-allocation equivalent
is invented. The earlier unresolved C empty-call samples remain raw, with their
curated summary marked unresolved; no historical value is retroactively corrected.

## Separate intermediate source / old-clock campaign

`9eda283` was fully tested and measured before the forwarded-literal correction.
It does not establish a target gain: indirect LLGo 4.594959→4.625250 (+0.66%,
A/A +0.20%); Go 4.643333→4.641708 (-0.03%, A/A -0.59%). LLGo empty is +25.4%
with A/A -8.2%. Every other original control and every phase is retained.
That 320 VM/35 C observation campaign uses the old C clock; its unresolved
C empty median is not zero performance. It is kept as a separate source and
instrumentation event, never pooled into the final table or substituted for a
less favorable final observation.

[Both complete studies, clock reproduction, exact source/flags, witnesses and verification](evidence/performance-immediate-literals-2026-11-03.json.gz)
