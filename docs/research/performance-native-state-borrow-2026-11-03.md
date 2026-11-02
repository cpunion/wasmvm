# Borrowed native entry state: 2026-11-03

Status: **scoped candidate; final contribution CI and controlled performance
acceptance remain open**. Source `c131587` follows cancellation-view source
`2bc8ab7`. Neither the held host-buffer implementation nor a compiler upgrade
is included. The prior PR remains the sole active contribution while this
follow-up is evaluated.

## Mechanism and contracts

The native driver previously passed its 112-byte state by value to Context.Run,
which borrowed the parameter's address. The pinned LLGo compiler emits an
unconditional AllocU(112) in that wrapper, matching the independently reported
[input-parameter problem](performance-llgo-value-parameter-2026-11-03.md)
([LLGo #2974](https://github.com/xgo-dev/llgo/issues/2974)). Its native State
contains numeric snapshots, not Go roots. Keep one state in each exclusively
owned execution workspace and borrow its address for entry. Generated code
continues to receive only the C-owned execution record; the host pointer never
crosses the C ABI. Publish state and copy the exit snapshot under the existing
active-entry/lease/memory protocol, before concurrent close can release it.

Retain the by-value internal adapter API for existing callers. The driver uses
RunState, resets its workspace field on begin/suspend/close, and preserves the
old fuel receipt on adapter errors. All exits, continuation validation, root
maps, limits, sticky cancellation, reentry and code retirement stay intact.
Compiler and Core frontend source inventories are unchanged. Final LLGo
assembly calls Context.run directly from the driver; the new borrowed wrapper
has no AllocU call. The retained old wrapper still does. Other driver allocations
remain and are not attributed to this wrapper.

All 31 Go packages and thirty coverage gates pass. Pinned full-LTO LLGo runs
native/nativeexec/VM/P1 tests. Scoped race, no-cgo core and vet pass. New tests
cover every native exit, snapshots and roots, nil/invalid/limited/closed states,
leases, concurrent cancellation/close and workspace cleanup. Initial test and
diagnostic-invocation failures are retained separately from successful evidence.

## Original P1 comparison

Shared M4 Max/macOS arm64, Go 1.27.3, LLGo `8e45ed5`, `-O3 -lto=full`.
Four unchanged original kernels, 10000 iterations, identical arguments,
environment and independent result checks. Five fresh interleaved process
blocks, two warmups/seven samples; both compilers/backends/base/head, four
exact baseline A/A aliases, four pinned reference modes and ordinary C.
There are 320 VM and twenty C observations, 68 summaries and 64 VM/four C
alternate checks. Startup, preparation, instantiation, first/warm calls and
completed cleanup remain separate; no subtraction or cross-study pooling.

Warm ns/source iteration:

| Kernel | Go native base → head | LLGo native base → head | wazero native | Wasmtime | Wasmer | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| args-sizes | 477.754 → 475.583 | 1651.675 → 1562.208 | 24.212 | 54.800 | 44.300 | 0.800 |
| environ-sizes | 483.142 → 480.121 | 1661.862 → 1560.267 | 24.517 | 55.200 | 59.300 | 0.800 |
| args-copy | 511.304 → 513.237 | 1818.729 → 1769.946 | 30.887 | 65.000 | 90.300 | 0.800 |
| integer-control | 2.283 → 2.529 | 2.325 → 2.458 | 1.667 | 1.800 | 2.100 | 1.600 |

LLGo native args/env/copy improve 5.4%/6.1%/2.7%, versus corresponding A/A
-0.6%/-0.2%/-0.3%. Go native changes -0.5%/-0.6%/+0.4%; copy A/A is +1.4%.
LLGo interpreter changes +0.1%/-1.1%/-1.0%. All original adverse arithmetic
controls stay visible: Go native +10.8% with A/A -0.5%; LLGo native +5.7%
with A/A -3.8%. These regressions are not waived by unchanged compiler sources.
They keep performance acceptance open.

A separate fixed five-block diagnostic uses the same arithmetic bytes at one
million iterations; it does not replace those short-control results. LLGo native
is 1.717584 → 1.749791 ns/iteration (+1.9%), baseline A/A 1.710791 (-0.4%).
Go is 1.852709 → 1.715958, A/A 1.778791. Wasmtime 1.838, wazero 1.810667,
Wasmer 1.746 and C 1.723. This different-duration shared-host experiment does
not establish a calibrated regression-free result or resolve its cause.

## Allocation and the three remaining gaps

Independent exact-archive test diagnostics retain three rounds for every
kernel/backend. Only exact `2bc8ab7` baseline allocation records are reused,
with identical added-test and fixture hashes; new throughput is never pooled.
Third-call LLGo native args-sizes TotalAlloc falls 39531472 → 38244480 bytes
per original 10000 callbacks, about 1.29 MB. Env/copy show the same approximate
1.28 MB reduction. Go args-sizes remains about 5.28 MB. Test executables do not
supply production timing or a Mallocs count. Small control allocations fluctuate
and are retained; the remaining LLGo native args-sizes allocation still exceeds
its interpreter by about 656 bytes/callback and Go native by about 3.30 KB.

For this campaign's args-sizes, LLGo native remains 3.28x Go, 28.51x Wasmtime,
64.52x wazero native and 35.26x Wasmer. The ordinary C provider performs the
same finite algorithm and writes through opaque separate-translation-unit
calls, with no VM bounds, cancellation, quotas or capability lifecycle. Its
0.8 ns short sample is a coarse lower-bound diagnostic, not calibrated parity.
Sustained arithmetic is much closer to ordinary C, while P1 provider/boundary
work remains a large independent deficit. No arithmetic control or startup
median is subtracted to manufacture a pure callback cost.

Continue with marshaling, capability/memory access and callback lifecycle
profiles, preserving arbitrary host semantics. Controlled Linux amd64/macOS
arm64 portfolio, startup, allocation and component/WASI acceptance remain open.

[Raw phases, all A/A blocks, identities, allocation/assembly diagnostics and validation](evidence/performance-native-state-borrow-2026-11-03.json.gz)
