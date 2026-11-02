# P1 immutable process cancellation view: 2026-11-03

Status: **selected scoped follow-up; final contribution CI and controlled
performance acceptance remain open**. Source `6a13a48` follows native-state
source `c131587`, which already includes the separate VM operation view.
The held raw-host-buffer source remains excluded. This is a new isolated
comparison, not a pooled continuation of earlier timings.

A P1 Process already owns a fixed cancellation binding. Its live context.Err
check appears after every host import and still allocates with the pinned LLGo
compiler (#2959, fixed upstream after the pin). Share the existing immutable
CancellationView helper once at process construction. Child callback contexts,
process admission/active counts, resource cancellation and shutdown draining
retain their existing paths. Go's constant branch keeps its original context.
No mutable context pool, optional lifetime check or new execution framework is
introduced. Native driver/compiler/VM sources are unchanged from this baseline.

All 31 Go packages/thirty coverage gates, pinned full-LTO LLGo P1, P1 race and
vet pass. The existing helper's deadline/value/cause/rebinding tests and P1's
pure-guest process cancellation, callback/resource closure and concurrent
shutdown suites retain their coverage. Initial benchmark environment rejection
is recorded separately; corrected identities and full-LTO flags are verified.

## Independent comparison

Same shared M4 Max/macOS arm64; Go 1.27.3, LLGo `8e45ed5`, `-O3 -lto=full`.
The four original P1/control binaries retain 10000 iterations, guest/arg argv,
KEY=VALUE and independent result oracles. Five fresh interleaved blocks,
two warmups/seven samples, both compilers/backends/base/head, four exact base
A/A aliases, four named pinned reference modes and matching ordinary C.
All 320 VM/twenty C observations, 68 summaries and 64 VM/four C alternate
checks preserve startup, preparation, first/warm calls and completed cleanup.
No timing subtraction or optional repeat is used.

Warm ns/source iteration:

| Kernel | Go native base → head | LLGo interpreter base → head | LLGo native base → head | wazero native | Wasmtime | Wasmer | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| args-sizes | 479.679 → 484.450 | 1352.692 → 1303.804 | 1578.333 → 1501.083 | 24.542 | 54.100 | 44.100 | 0.800 |
| environ-sizes | 482.733 → 480.962 | 1343.537 → 1304.817 | 1593.162 → 1521.646 | 24.221 | 54.100 | 59.500 | 0.800 |
| args-copy | 514.679 → 509.875 | 1503.354 → 1441.833 | 1787.608 → 1671.225 | 31.262 | 67.100 | 91.200 | 0.900 |
| integer-control | 2.279 → 2.362 | 42.583 → 43.429 | 2.263 → 2.371 | 1.837 | 1.900 | 2.000 | 1.600 |

LLGo native args/env/copy improve 4.9%/4.5%/6.5%, versus A/A -1.0%/-1.6%/-0.02%.
Its interpreter improves 3.6%/2.9%/4.1%, versus A/A +2.2%/+0.5%/+0.2%.
Go native changes +1.0%/-0.4%/-0.9%. Adverse original integer controls remain:
LLGo native +4.8% with A/A -1.3%; LLGo interpreter +2.0% with A/A +0.9%;
Go native +3.7% with A/A +5.7%; Go interpreter +1.6% with A/A -0.9%.
These shared-host observations do not establish calibrated parity or assign
control movements to the process view. The preceding state's separate short
and sustained regressions remain evidence, without replacement.

Exact-archive allocation diagnostics retain three rounds for every kernel and
backend, reusing only exact c131 baseline allocation records with the identical
added diagnostic and fixture hashes. Third-call LLGo native args-sizes
TotalAlloc falls 38244480 → 35036624 bytes per 10000 callbacks; interpreter
31680240 → 28487024. All three P1 kernels save about 3.2 MB. Go native remains
about 5.28 MB. These are approximate TotalAlloc diagnostics, not Mallocs counts
or production timing. Small control fluctuations remain retained.

The three comparisons still have open gaps. Args-sizes LLGo native takes
3.10x Go, 27.75x Wasmtime, 61.16x wazero native and 34.04x Wasmer here. Ordinary
C's matching separate-TU provider retains the same finite algorithm and writes,
but omits VM bounds/cancellation/quotas/capability lifecycle. Its 0.8 ns short
sample is an unchecked coarse lower bound. Marshaling, memory/capability access,
callback lifetime and remaining LLGo allocations need further investigation.
The preceding independent parameter issue remains #2974; this known #2959
workaround does not justify another upstream issue or a compiler pin upgrade.

[All phases, A/A blocks, source identities, diagnostic records and checks](evidence/performance-p1-process-context-2026-11-03.json.gz)
