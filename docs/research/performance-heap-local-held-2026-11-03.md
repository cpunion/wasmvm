# Held heap/local forwarding: 2026-11-03

Status: **held; do not include source `260c6c0` in the selected i31 contribution**.
Baseline `b99d1b6` has i31 production source `1594908`. The experiment extends
checked heap store/load forwarding to an adjacent local set/tee, preserving
observable getter temporaries and cold entries, and selects cached integer
moves directly. The existing checked store remains before all local effects;
no combination crosses a poll boundary or changes the precise tier.

Tests execute every width, packed sign extension, reference/vector/NaN bits,
local/source aliases, set/tee and live/overwritten getter slots, every original
PC, complete frames/payloads and pre-store null traps. An initial test used
noncanonical high bits for F32 input; corrected to a canonical-width NaN while
retaining the integer high-bit cases. The original test and recorded diagnostic
are retained; no product fix is inferred from that invalid initializer.
All 31 Go packages/thirty coverage gates, full-LTO LLGo codegen/nativeexec/VM,
affected race/no-cgo/vet pass.

One predeclared shared-M4 campaign uses four original GC binaries/counts,
five interleaved blocks/two warmups/seven samples, both native host compilers,
exact base A/A, Wasmtime and explicit reference GC rejections. All checks/builds
finish before timing. 140 VM/ten C observations, 28 VM/two C alternate checks,
eight unsupported records and thirty summaries. Ordinary Clang-22 O3 C supplies
matching i31 and local struct-field arithmetic, permitting ordinary scalar
replacement; it omits VM GC/quota/cancellation contracts. No C guest-GC
allocation implementation is invented.

The recorded shared-host load rises **91.3 → 307.1** during this run. Absolute
rates are substantially slower than earlier campaigns; earlier fast medians
are not substituted or pooled. Native field warm ns/iteration:

| Case | Go base → head | LLGo base → head | Wasmtime | C |
| --- | ---: | ---: | ---: | ---: |
| struct-access | 2.428208 → 2.443709 | 2.421583 → 2.425333 | 1.930000 | 1.918000 |
| i31 control | 2.843416 → 2.581709 | 2.614792 → 2.536417 | 3.411000 | 2.483000 |
| struct-allocation | 81.400000 → 81.250000 | 113.150000 → 104.679200 | 176.500000 | — |
| indirect access | 9.938750 → 9.930041 | 9.950291 → 9.948667 | 5.504000 | — |

The affected field case changes Go +0.6% / LLGo +0.2%, with A/A 2.443917 /
2.421583. Go allocation A/A moves +20.0%; i31 control rates also move despite
this source not changing its selector. This is not useful performance evidence
for adopting the extra forwarding path. Keep it held and preserve the original
proven i31 candidate separately. Remaining field/indirect gaps and all-target
controlled performance stay open. Regression tests can be retained only after
executing them successfully against the selected original implementation.

[Exact source, all phases/controls/load, C source and semantic diagnostics](evidence/performance-heap-local-held-2026-11-03.json.gz)
