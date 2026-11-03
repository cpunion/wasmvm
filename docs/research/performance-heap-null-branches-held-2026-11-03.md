# Held ARM64 heap null-branch experiment

Status: **held; production source is unchanged**. Baseline is final PR #107
`c1e37de`; experiment `dbb92fe` replaces the cached heap hit's dependent
TST/CCMP with an independent CBZ followed by a complete-handle CMP. It also
removes the repeated miss-path null test. Cache policy, registers, native ABI,
generation/index checks and null trap ordering remain unchanged.

The candidate passes all 31 Go packages/thirty coverage gates, pinned LLGo
full-LTO codegen/nativeexec/VM, affected race, no-cgo and vet. The first targeted
Go run lacked the worktree corpus link and failed to open two files; its
nativeexec package passed. Complete validation follows restoration of the link.
Initial source-preparation and verifier working-directory assumptions also fail
before correction; these are retained separately from measurements.

Forty-eight paired body witnesses change only six ARM64 unmetered ordinary/
expanded captures: struct access 2100→2084 bytes, allocation 1848→1840 and
indirect access 3368→3336. Every amd64 and metered capture is identical. Existing
complete cold-PC/null/stale/forged/index/frame tests execute the changed paths.
Smaller code alone does not establish higher throughput.

## One complete campaign

M4 Max/macOS arm64, Go 1.27.3, LLGo `8e45ed5` with `-O3 -lto=full`.
Original eight GC/Core guests, unchanged bytes/inputs/counts, five fresh serially
interleaved blocks, two warmups/seven samples, base A/A aliases for both compilers,
three pinned native references and ordinary C. C probes/references are the same
verified high-resolution binaries as the preceding study. All tests/builds finish
before timing. Independent checks verify 320 VM/35 C observations, 64 VM/seven C
alternate checks, eight explicit GC rejections and all 71 summaries. Startup,
preparation, first/warm calls and completed cleanup are retained independently.

Warm ns/source iteration; empty control is ns/call:

| Kernel | Go base → candidate | LLGo base → candidate | Wasmtime | wazero | Wasmer | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| i31 | 1.857875→1.850584 | 1.842625→1.852000 | 2.477000 | unsupported | unsupported | 1.751000 |
| struct access | 1.793583→1.786209 | 1.784542→1.772625 | 1.406708 | unsupported | unsupported | 1.399041 |
| allocation | 70.6084→74.9750 | 88.9791→86.7875 | 177.6250 | unsupported | unsupported | — |
| indirect access | 6.730417→7.148208 | 6.851791→6.930375 | 3.821250 | unsupported | unsupported | 4.195167 |
| empty control | 8791→8417 | 8667→9417 | 1708 | 750 | 5166 | 84* |
| integer | 2.941875→2.950000 | 2.933459→2.926375 | 2.904417 | 3.209959 | 2.912416 | 2.899291 |
| calls | 1.185792→1.171958 | 1.166792→1.169000 | 1.342208 | 1.961458 | 1.160750 | 1.149417 |
| memory | 0.805708→0.803125 | 0.811208→0.800416 | 0.881334 | 1.459250 | 0.978709 | 0.498792 |

LLGo struct access improves only 0.67% (A/A -0.14%), while indirect access
regresses 1.15% (A/A -1.16%). Go indirect regresses 6.21% (A/A +2.28%) and
allocation 6.18% (A/A +2.35%). LLGo allocation's -2.46% is similar to A/A
-2.27%. Preserve LLGo empty +8.65% versus A/A -5.77%, calls +0.19% versus
A/A +0.65%, and every other control. Load averages are 49.8→32.5; there is no
controlled or statistically established portfolio gain. No unchanged-source
retiming replaces this adverse campaign.

The three deficits stay distinct. Candidate LLGo struct access is 1.26x
Wasmtime; indirect is 1.81x Wasmtime and 1.65x ordinary C. Native memory is
faster than the VM references here but 1.60x C. LLGo allocation is 1.16x Go.
Integer/calls are near C. None of these throughput ratios establishes an LLGo
defect; the independently reproduced startup problem is tracked separately.
Replacing a condition-code dependency with another predicted branch does not
close the heap gap. Keep the simpler original implementation and investigate
general value/alias proofs before adding cache state or more branch variants.

*The C empty control includes timer/dispatch and is about two RAW-clock ticks,
not a precise pure-call cost. Ordinary C may optimize the algorithm and omits
VM quotas, GC, cancellation and continuation contracts; no C allocation equivalent
is invented. This campaign is not pooled with earlier lower-load observations.

[Sources, full flags/identities, all observations/phases/A/A, coverage and verification](evidence/performance-heap-null-branches-held-2026-11-03.json.gz)
