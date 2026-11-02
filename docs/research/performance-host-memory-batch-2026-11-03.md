# Checked host memory batches: 2026-11-03

Status: **local candidate; final CI and controlled acceptance open**. Source
`ceea912` follows operation-effect source `ff97ead`; the held entry-FFI source
is excluded. The public Caller.WriteMemoryBatch API prechecks every complete
range under one existing Caller capability, then copies in order. Bad ranges
cause no writes; valid overlaps keep ordering. Cancellation may interrupt a
batch, and shared memory retains individual transfer synchronization rather
than promising batch atomicity. Caller expiry and ownership stay unchanged.

P1 paired size outputs use this general API while retaining their independent
memory32 overflow checks. Other operations, private I/O contexts and scalar
callback contracts are unchanged. VM tests cover private/shared memory through
both backends, unaligned and overlapping writes, a bad later range without
partial mutation, extreme offsets, empty ranges/batches, invalid indices,
cancellation and expired/instance-free Callers. P1 tests additionally check
both invalid memory32 outputs and overlapping size-output ordering.

All 31 Go packages and thirty coverage gates pass; VM coverage is 92.85%, P1
90.87%. Pinned full-LTO LLGo VM/P1, scoped race, no-cgo VM and vet pass. The
initial targeted command ran the VM tests and selected no P1 tests; the final
full P1 suites above supply actual execution. All checks finish before timing.
Compiler/interpreter/native adapter inputs are unchanged, without waiving
runtime regressions or inferring guest byte identity from source equality.

## Independent comparison

Shared M4 Max/macOS arm64, Go 1.27.3 and LLGo `8e45ed5`, `-O3 -lto=full`.
Four original kernels at 10000 iterations, identical arguments/environment and
alternate oracles. Five fresh interleaved blocks, two warmups/seven samples,
both compilers/backends/base/head, four exact base A/A aliases, four default
reference modes plus two distinct cancellation-enabled wazero modes and C:
360 VM/twenty C observations, 76 summaries, 72 VM/four C alternate checks.
Startup, preparation, first/warm calls and completed cleanup stay separate.

Warm ns/source iteration:

| Kernel | Go native base → head | LLGo interpreter base → head | LLGo native base → head | enabled wazero native | default Wasmtime |
| --- | ---: | ---: | ---: | ---: | ---: |
| args-sizes | 269.458 → 253.204 | 640.871 → 541.575 | 850.763 → 751.288 | 38.833 | 55.800 |
| environ-sizes | 270.304 → 248.929 | 638.571 → 540.354 | 839.188 → 745.975 | 39.504 | 52.400 |
| args-copy | 304.408 → 306.350 | 798.871 → 787.108 | 1013.154 → 996.188 | 45.579 | 66.000 |
| integer-control | 2.833 → 2.900 | 43.233 → 43.288 | 2.317 → 2.483 | 15.367 | 1.800 |

Paired-size operations improve Go native 6.0%/7.9%, LLGo native 11.7%/11.1%,
interpreter 15.5%/15.4%. Native A/A changes are Go -0.1%/+1.0% and LLGo
-0.1%/+0.9%. Copy uses its unchanged provider path: Go native regresses 0.6%
(A/A -1.0%), LLGo improves 1.7% (A/A -0.2%). Retain original arithmetic
regressions: Go native +2.4% versus A/A -9.4%, LLGo native +7.2% versus A/A
+2.9%. These short controls remain inconclusive for calibrated acceptance;
other gains do not erase them or the prior public-host regression control.

Exact-archive diagnostics reuse only exact ff97ead allocation baselines with
identical added-test and fixture hashes. Third-call native args TotalAlloc falls
LLGo 18086048 → 15848160 bytes per original 10000 callbacks; interpreter
11522192 → 9278528. Go native stays 2080808. All rounds and control fluctuations
remain; there is no Mallocs or production-timing claim from test executables.

The remaining args-sizes gaps are LLGo native 2.97x Go, 19.35x enabled wazero
and 13.46x default Wasmtime in this campaign. Reference poll/lifecycle policies
still differ. Matching ordinary C and all other reference rows remain in the
raw records; C omits VM contracts and short measurements are coarse. No source's
percentage is added to earlier studies or substituted for sustained/controlled
portfolio, startup, latency, allocation or all-platform acceptance.

[All phases, A/A samples, sources, binaries, allocation diagnostics and validation](evidence/performance-host-memory-batch-2026-11-03.json.gz)
