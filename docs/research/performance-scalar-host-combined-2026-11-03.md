# Combined scalar-host contribution: 2026-11-03

Status: **review candidate; final contribution-head CI and controlled acceptance
remain open**. Actual measured baseline `6a13a48` is source-identical through
internal/pkg/cmd/go.mod/go.sum to merged PR #104 main `be7ed71`. Actual head
`54a0143` contains scalar host callbacks, operation-scoped private P1 contexts,
checked Caller memory batches, retained diagnostics and pre-entry cleanup tests.
Its production implementation is `ceea912`; no engine is relabeled. Held direct
FFI, host-request storage, type-dispatch and native-cleanup sources are excluded.
The later i31 selector is on its own branch and is not part of this comparison.

The scalar callback stack remains fresh and independently retainable, while
Caller expiry, reentry, exceptions, roots, cancellation and cleanup retain their
contracts. Four immutable configuration/memory-only P1 operations avoid private
I/O contexts; callback-capable operations preserve independent canceled child
contexts. Generic memory batches precheck every range and preserve ordered
copies, invalid-range non-mutation and individual shared-memory synchronization.

## Final combined measurement

One new campaign compares the complete selected contribution against PR #104
production, instead of composing earlier per-change percentages. Shared M4 Max,
Go 1.27.3 and pinned LLGo `8e45ed5` with `-O3 -lto=full`. Four original 10000-
iteration kernels/inputs, both compilers/backends/base/head, four exact base A/A
aliases, four default reference modes, two separately cancellation-enabled
wazero modes and ordinary C. Five fresh interleaved blocks, two warmups/seven
samples: 360 VM/twenty C observations, 76 summaries, 72 VM/four C alternate
checks. Source and binary hashes are verified. Startup, preparation, first/warm
work and completed cleanup remain separate. All builds/tests and the PR #104
artifact audit/local synchronization finish before timing.

Warm ns/source iteration:

| Kernel | Go native base → head | LLGo interpreter base → head | LLGo native base → head | enabled wazero native | default Wasmtime |
| --- | ---: | ---: | ---: | ---: | ---: |
| args-sizes | 560.513 → 252.892 | 1286.829 → 545.813 | 1512.038 → 749.588 | 39.150 | 55.400 |
| environ-sizes | 567.679 → 261.554 | 1295.108 → 554.938 | 1514.175 → 756.229 | 38.679 | 53.700 |
| args-copy | 600.242 → 332.379 | 1442.704 → 786.146 | 1659.967 → 1001.658 | 44.613 | 66.900 |
| integer-control | 2.808 → 2.850 | 43.725 → 43.375 | 2.538 → 2.583 | 15.138 | 1.900 |

LLGo native improves **50.4%/50.1%/39.7%**, LLGo interpreter
**57.6%/57.2%/45.5%**, Go native **54.9%/53.9%/44.6%**. Native args/env/copy
A/A changes are Go -1.6%/+0.3%/+0.5%, LLGo -0.4%/-0.1%/+1.0%.
These are descriptive shared-host results, without controlled acceptance.

Retain the adverse original short arithmetic control: Go native +1.5% versus
A/A -5.2%, LLGo native +1.8% versus A/A -2.6%. Its broad control movement is
unresolved, not waived. Earlier public HostFunction control regressions remain
in the [scalar-host study](performance-scalar-host-abi-2026-11-03.md), together
with the later rejected type-dispatch and native-cleanup alternatives. No
favorable control substitutes for these observations.

Remaining args-sizes gaps are **LLGo 2.96x Go, 19.15x enabled wazero and 13.53x
default Wasmtime**. Reference interruption/lifecycle policies still differ.
Ordinary C args is 0.9 ns at this short count; it performs the same finite
unchecked provider arithmetic but omits VM capability/cancellation contracts,
and remains a coarse lower bound. All original default-reference/C rows and
phase data remain. These results do not establish interpreter/whole-portfolio,
startup, zero-allocation, cancellation-latency or both-platform acceptance.

## Source-specific validation

Actual head `54a0143` passes all 31 Go packages and thirty coverage gates. The
first supplemental run lacked the worktree's read-only fixture symlink and
failed to open corpus files. Both original logs/exits remain. In that first
LLGo run nativeexec and VM succeed; after restoring fixtures only the failed P1
package is rerun, with the same pinned full-LTO flags, and succeeds. Full Go and
coverage also rerun successfully. No unavailable-fixture result is called a VM
or compiler semantic defect. Original per-change race/no-cgo/vet evidence keeps
its actual revisions; these checks are not silently attributed to a newer hash.
The final PR matrix still must execute its final head on both platforms.

[Exact final combined sources, binaries, all phases/A/A/controls and validation](evidence/performance-scalar-host-combined-2026-11-03.json.gz)
