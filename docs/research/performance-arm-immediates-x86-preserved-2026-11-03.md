# ARM immediate fusion with original amd64 placement

Status: **corrective contribution candidate; new-head CI and performance are
pending**. Original production is `a176cb5`; the first PR #107 head is `c1e37de`.
Correction `b54195d` confines the overwritten-original-result literal elision to
arm64. AMD64 retains its prior instruction placement. This responds to adverse
hosted measurements; instruction alignment is a hypothesis, not an established
root cause or an LLGo defect.

The first head's [Linux benchmark](https://github.com/cpunion/wasmvm/actions/runs/41562555892)
completes with 228 observations, 76 independently regenerated summaries and
19 configurations. Both source inventories, exact compiler pins and identical
full-LTO flags are verified. The automatic [PR report](https://github.com/cpunion/wasmvm/pull/107#issuecomment-6594142102)
retains every phase and reference. Its native warm observations are:

| Kernel | Go original → first head | LLGo original → first head | Wasmtime | wazero native | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| empty, ns/call | 1333→2775 | 5410→5310 | 191 | 50 | 301 |
| integer, ns/iteration | 1.900664→2.209596 | 1.882868→2.184348 | 1.851092 | 1.949312 | 1.860988 |
| calls, ns/iteration | 1.073267→1.040207 | 0.960053→0.960193 | 1.855993 | 2.765760 | 1.909560 |
| memory, ns/iteration | 1.264427→1.421713 | 1.276307→1.277647 | 1.099047 | 2.038267 | 0.980293 |

Integer regresses **Go 16.25% / LLGo 16.01%**. Go empty +108.2% and memory
+12.44% remain visible. LLGo interpreter API integer/calls/memory are
+2.60%/+0.23%/+0.15%; Go API +1.39%/+3.14%/+4.03%. These are three-process
shared-runner diagnostics, with short controls and variable reference A/A
observations. They do not establish controlled causality. The first head is
not merged or relabeled as the correction's performance.

The first candidate removes five bytes from each of seven captured unmetered
amd64 workloads, shifting subsequent addresses. The correction restores every
captured amd64 body to original production while retaining the arm64 immediate
selector/fusion and its checked live-slot restoration. Forty-eight paired
captures verify all amd64/metered bytes equal to original, and all arm64 bytes
equal to the preceding selected source `fbacef5` (production-identical to
`c1e37de`). Fourteen arm64 unmetered pairs change versus original. The
[earlier ARM study](performance-immediate-literals-2026-11-03.md) retains its
actual source and adverse controls; it is not rewritten as a new measurement.

Local validation executes all 31 Go packages/thirty coverage gates and pinned
full-LTO LLGo codegen/nativeexec/VM, plus affected race/no-cgo/vet. Both targets
retain complete local set/tee alias and cold-PC semantic checks; only arm64
requires elimination of the overwritten literal. No rejected heap-cache or
null-branch source enters this contribution. The new exact PR head still needs
the full dual-platform matrix and a new Linux comparison before merge. Restored
bytes alone do not establish restored throughput or satisfy the empty-call,
controlled portfolio, cancellation/allocation or expanded WASI gates.

[First-head raw Linux result/source audit and correction validation/body witnesses](evidence/performance-arm-immediates-x86-preserved-2026-11-03.json.gz)
