# Internal host tuple storage experiment: 2026-11-03

Status: **held; no consistent material LLGo throughput improvement**. Exact
source `eac9a53` follows `f384fb7`. This production source is excluded from the
separate operation-context candidate. No whole-engine or milestone parity claim.

The common execution owns one borrowed host request and bounded raw argument/
result scratch (up to 2 KiB idle retention); large tuples remain transient. All
request paths, including tail calls and Core Stack Switching host continuations,
share publication/retirement. Public callback arguments and Caller identities
remain independent, preserving retention, reentry and capability expiry.
Aliased returns copy before retirement. Failure, exception, abort and close clear
raw handles; reference roots and quotas remain in the common engine.

All 31 Go packages, thirty coverage gates, affected LLGo execute/VM/P1, race and
vet pass. No-cgo runs the changed execute/VM packages. The initial broader P1
no-cgo run attempted its native-only conformance paths and fails explicitly;
P1 full semantics runs with cgo. An accidentally observed baseline-checkout
no-cgo run is retained and excluded from candidate validation. The initial
uncommitted Stack Switching host path failure and allocation-test comparator
error remain in logs; corrected source passes. LLGo TotalAlloc is measured
independently; its Mallocs/AllocsPerRun result is not an allocation-count proof.

## Production comparison

Shared M4 Max/macOS arm64, Go 1.27.3, LLGo `8e45ed5` with `-O3 -lto=full`.
Four original P1/control binaries, original 10000 iterations, args guest/arg and
KEY=VALUE environment. Five fresh process blocks, two warmups/seven serially
interleaved calls, both own compilers/backends/base/head, four exact baseline
A/A aliases, four pinned competitor modes and ordinary C: 320 VM and twenty C
observations, 68 summaries, 64 VM/four C alternate checks. All preparation,
readiness, first/warm calls and cleanup remain separate.

Warm ns/source iteration:

| Kernel | Go native base → head | LLGo native base → head | wazero native | Wasmtime | C |
| --- | ---: | ---: | ---: | ---: | ---: |
| args-sizes | 491.137 → 465.958 | 1781.596 → 1777.954 | 25.137 | 56.500 | 0.800 |
| environ-sizes | 548.550 → 527.788 | 2022.017 → 1968.862 | 26.450 | 58.300 | 0.900 |
| args-copy | 536.167 → 499.250 | 1931.350 → 1916.562 | 30.721 | 67.900 | 1.000 |
| integer-control | 2.383 → 2.617 | 2.383 → 2.421 | 1.729 | 1.900 | 1.600 |

LLGo native args/environ/copy change -0.2%/-2.6%/-0.8%; corresponding A/A
changes are +1.2%/+5.8%/+2.0%. LLGo interpreter args regresses 3.0% (A/A -0.8%).
Go native integer control regresses 9.8% (A/A -4.5%). These results do not justify
promoting the source. No variant, workload count or prior campaign is substituted
for an adverse result.

Independent exact-archive diagnostics report third-call Go native args-sizes
host allocation 5280808 → 4320808 bytes, versus LLGo 44635552 → 43354032 for
10000 original host calls. Three rounds for all kernels/backends/compilers remain
retained. The raw tuple savings are a small part of LLGo's recurring allocation;
its remaining 43 MB and 3.82x Go elapsed time require further driver/provider
investigation. The measured head is still 31.5x Wasmtime/70.7x wazero native on
args-sizes. Existing independent LLGo defer/context reports stay distinct from
this aggregate; no timing-only duplicate issue is filed.

The C control executes the same algorithm, memory writes, argument/environment
configuration and checksums, with opaque direct provider calls from a separate
translation unit and no LTO. It has no VM bounds, lifetime, cancellation or
admission policy. Its sub-nanosecond/near-nanosecond normalized values come from
short calls and are descriptive lower-bound controls, not calibrated parity or
an attribution of VM overhead to one method. Compiler flags and every raw sample
are retained. No startup or empty-call median is subtracted.

[Lossless phases, A/A blocks, exact identities, C source, tests and allocation diagnostics](evidence/performance-host-storage-held-2026-11-03.json.gz)
