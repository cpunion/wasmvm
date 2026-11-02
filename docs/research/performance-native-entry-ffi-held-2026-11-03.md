# Held native-entry FFI experiment: 2026-11-03

Status: **held; excluded from PR #104 and main**. Source `69cd547` follows
selected source `6a13a48`. Only the three platform adapter implementation files change.
Keep the C entry ABI and its retained code/context protocol; LLGo calls the
existing entry symbol through an explicit C declaration with two C-owned
numeric addresses, like existing shared-memory helpers. Go retains the same
Cgo call. Host-buffer Cgo checks are unchanged. This is an experiment, not an
accepted replacement for the current boundary.

All 31 Go packages/thirty coverage gates, pinned full-LTO LLGo
native/nativeexec/VM/P1, scoped race, core no-cgo and vet pass. Existing probes
cover ABI layout, results, cancellation, concurrent close, roots and native GC.
Assembly for Context.run has three AllocU calls in the base and a direct native
entry call with none in the head. Third-call LLGo native args-sizes TotalAlloc
falls 35036624 → 34240800 bytes per 10000 callbacks, approximately eighty bytes
per entry. This is consistent with known [LLGo #2958](https://github.com/xgo-dev/llgo/issues/2958);
it creates no new independently established upstream problem.

The primary timed study starts only after same-host race/no-cgo/vet checks
finish. An initial complete attempt overlapped those checks. It was excluded
before any score was inspected; all its records remain in the evidence. The
complete unchanged matrix then runs exactly once after the quality barrier.
No optional statistical repeat or favorable-case replacement is used.

Five fresh interleaved blocks on shared M4 Max/macOS arm64, Go 1.27.3 and
LLGo `8e45ed5`, `-O3 -lto=full`; all four original 10000-iteration kernels,
both compilers/backends/base/head, four exact base A/A aliases, four pinned
reference modes and ordinary C: 320 VM/twenty C observations, 68 summaries,
64 VM/four C alternate checks. All phases and original controls remain distinct.

Warm ns/source iteration:

| Kernel | Go native base → head | LLGo native base → head | LLGo native A/A | wazero default native | Wasmtime | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| args-sizes | 476.904 → 478.342 | 1509.196 → 1471.663 | 1498.717 | 24.546 | 55.300 | 0.900 |
| environ-sizes | 476.354 → 472.554 | 1502.796 → 1457.992 | 1509.592 | 24.438 | 53.600 | 0.900 |
| args-copy | 511.000 → 511.267 | 1667.921 → 1627.604 | 1641.642 | 30.363 | 65.900 | 0.800 |
| integer-control | 2.475 → 3.033 | 2.304 → 2.583 | 2.283 | 1.683 | 1.900 | 1.600 |

LLGo native P1 improves only 2.4–3.0%. Original native integer regresses Go
22.6% (A/A -6.4%) and LLGo 12.1% (A/A -0.9%). Unchanged guest compiler sources
do not waive these controls. The small P1 win does not establish a useful
regression-free integrated boundary; hold the code. All interpreter movements
and other named reference rows remain in the raw evidence. Native C omits VM
policies and its short samples are coarse. Default competitor cancellation
policies differ; the [separate cancellation study](performance-p1-cancellation-policy-2026-11-03.md)
retains both default and enabled wazero modes.

[All phases, identities, allocations, assembly, validation and excluded attempt](evidence/performance-native-entry-ffi-held-2026-11-03.json.gz)
