# Native guest call frame foundation: 2026-11-04

**Local architecture work; native guest call/return emission is unfinished.** The common runtime can now adopt a validated batch of native guest frames at a runtime boundary. The real native backend uses the borrowed invocation packet and current per-invocation budgets, while generated calls still return to the host. No new performance improvement or production acceptance is claimed.

The qualified engine source is `4862cc8b94eb5e922a00ba010b9f4d241516f77f`, based on the qualified memory-boundary optimization. Both Go 1.27.3 and pinned LLGo 3571 (`-O3 -lto=full`) pass all eight affected packages and 96 actual Core/proposal reports on macOS arm64. All scoped Go coverage gates remain above 90%. Race, cgocheck2 and no-cgo checks also pass. Exact compiler commands, source revisions and working directories are recorded for every run.

The frame-transfer models cover complete scalar/reference/SIMD tuples, exact fuel against the interpreter, nested host suspension and caller roots, arena relocation, error cleanup, different arguments on reuse, current frame/slot budgets and rejection of malformed transfers before stack mutation. These model tests establish the shared contract, not the unimplemented machine-code branch mechanism. Existing official-suite runs exercise the real native backend's packet path without hidden native calls.

| Package | Scoped Go coverage |
| --- | ---: |
| internal/compiler/codegen | 97.44% |
| internal/compiler/slots | 99.68% |
| internal/engine/execute | 96.63% |
| internal/engine/nativeexec | 92.52% |
| internal/engine/state | 95.21% |
| internal/platform/native | 95.03% |
| pkg/component | 91.20% |
| pkg/vm | 93.00% |

The shared machine validates the complete incoming chain, retains the root's result destination, refreshes the active function/PC/value view, and services the already charged boundary exactly once. It publishes all caller roots before any observable runtime effect. Borrowed invocation pointers are cleared on success, failure and abort. If an error follows arena relocation, reuse retains the current owned storage view.

The packet also fixes invocation-budget enforcement for existing optional native expansion: a cached workspace may have more capacity than its current invocation can use after suspended parents share their limits. Actual native entry tests check both original workspace limits and smaller per-invocation limits.

[Emission design](../design/native-calls.md) specifies the remaining amd64/arm64 body transfers, bounded native parent records, relative code relocations, stable runtime views, helper/GC/exception/continuation materialization and full qualification. Those mechanisms and the complete same-cohort competitor/native comparison must be implemented before this branch is a performance candidate. Final four-profile semantic CI, coverage and PR review remain required. No parallel contribution PR was opened; LLGo #2983 remains with its existing contributor.

[Lossless evidence](evidence/native-call-frames-2026-11-04.json.xz) retains the exact patch, sources, Go/LLGo reports, coverage and additional checks. The earlier memory-departure report now includes its qualification-provenance correction and correct-source rerun; its original raw timing and semantic records are preserved.
