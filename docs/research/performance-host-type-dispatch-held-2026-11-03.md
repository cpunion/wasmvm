# Held type-only host dispatch: 2026-11-03

Status: **held; excluded from the selected contribution branch and main**.
Source `50a6177` follows checked-memory source `ceea912`. The only production
change classifies the host callback by a type-only switch, extracting its scalar
function value only after selecting that path. Shared Caller cleanup, values,
roots and cancellation remain unchanged. Targeted Go/LLGo VM/P1 execution,
VM race/no-cgo and vet pass; no final full-platform integration is claimed.

The hypothesis was that avoiding an unused scalar function value on the public
HostFunction path might remove a cost behind its earlier small regression.
This study does not establish that cause or a useful benefit. Reject the source
instead of adding a second type query based on speculation.

Both experiments use exact source archives, pinned Go 1.27.3/LLGo 8e45ed5 with
-O3 -lto=full, the original guest bytes/counts/oracles and five fresh interleaved
blocks with two warmups/seven samples. All semantic/quality checks and both
build sets complete before timed execution. Default and cancellation-enabled
references remain distinct; their different lifecycle/poll policies are recorded.
Preparation, first/warm execution and cleanup remain separate.

## Existing public-host control

The same minimal public HostFunction performs two checked writes without P1
Process admission. This is a regression control, not full WASI conformance or
a replacement production provider. Eighteen VM modes, ninety VM/five C
observations and eighteen VM/one C alternate checks are retained.

| Path | base → head, ns/iteration | exact baseline A/A |
| --- | ---: | ---: |
| Go native | 244.125 → 243.237 | 244.958 |
| LLGo interpreter | 403.321 → 399.592 | 398.038 |
| LLGo native | 687.771 → 691.392 | 702.583 |

LLGo native remains +0.5% slower; its baseline A/A moves +2.2%. Interpretation
changes -0.9% against A/A -1.3%. These shared-host observations do not resolve
a regression-free classification improvement or identify the original cause.

## Original P1/control scope

All four original kernels, eighteen VM modes, 360 VM/twenty C observations,
76 summaries and 72 VM/four C alternate checks remain. Warm ns/iteration:

| Kernel | Go native base → head | LLGo native base → head | LLGo baseline A/A |
| --- | ---: | ---: | ---: |
| args-sizes | 251.825 → 247.896 | 745.213 → 745.338 | 746.792 |
| environ-sizes | 244.650 → 250.071 | 743.250 → 756.271 | 747.850 |
| args-copy | 301.604 → 306.875 | 996.479 → 1005.829 | 1019.888 |
| integer-control | 2.800 → 2.858 | 2.633 → 2.646 | 2.583 |

Env regresses Go 2.2%, LLGo 1.8%; copy regresses Go 1.7%, LLGo 0.9%.
Every corresponding A/A and interpreter/reference/C row stays in the evidence.
The candidate does not provide a clear improvement to either dispatch scope.
No timing-only LLGo issue or source workaround is introduced. The selected
memory-batch source and all earlier negative/public-host controls remain visible;
controlled all-target portfolio acceptance stays open.

[Exact sources, harnesses, binaries, all phases/A/A observations and validation](evidence/performance-host-type-dispatch-held-2026-11-03.json.gz)
