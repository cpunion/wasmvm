# Held native allocation cleanup scopes: 2026-11-03

Status: **both implementations held; excluded from the selected engine**.
Baseline is checked-memory production source `ceea912`. Source `6cd19f4`
extracts the shared native entry loop and keeps a deferred publication wrapper
only for WasmGC allocation batches. Source `c2b80f9` then keeps ordinary entries
in one driver frame; the allocating wrapper reenters that same driver once with
its reservation prepared. Neither duplicates the executor or weakens cleanup.

This tests the independently reported [LLGo non-loop defer allocation](https://github.com/xgo-dev/llgo/issues/2932).
Pinned baseline assembly unconditionally requests a 48-byte defer frame before
checking whether the entry allocates. Both experiments remove that request from
ordinary entries. Exact-archive third-call diagnostics reduce LLGo native
args-sizes TotalAlloc from 15848160 to 15206400 / 15202992 bytes per original
10000 callbacks. Go stays 2080808. These are approximate collector byte counts,
not allocation events or production timing; all rounds/control fluctuations stay.
This concrete allocation change does not establish an overall speedup.

Both sources pass all 31 Go packages, thirty coverage gates, pinned full-LTO
LLGo nativeexec/VM/P1 execution, affected race/no-cgo and vet. New tests check
unpublished reservations after cancellation and a custom-context panic before
native entry, preserving fuel and allowing subsequent reservation. Existing
trap/active cancellation publication tests remain. The initial first-campaign
launcher used an incorrect parent path and failed before compilation; its setup
error is retained separately from actual successful semantic runs.

## Original P1 comparison

Go 1.27.3 / LLGo `8e45ed5` with `-O3 -lto=full`, shared M4 Max/macOS arm64.
Each separate campaign uses five fresh interleaved blocks, original four guest
kernels at 10000 iterations, two warmups/seven samples, both compilers/backends,
exact base A/A aliases, four default references, two distinct cancellation-enabled
wazero modes and ordinary C. Each retains 360 VM/twenty C observations, 76
summaries and 72 VM/four C alternate checks. All test/build processes finish
before timing. Original reference policies and every preparation/first/warm/
cleanup phase remain; no percentages are pooled between campaigns.

| Kernel | First Go native base → head | First LLGo native base → head | Revised Go native base → head | Revised LLGo native base → head |
| --- | ---: | ---: | ---: | ---: |
| args-sizes | 251.667 → 248.713 | 744.346 → 762.583 | 254.913 → 244.250 | 746.621 → 757.800 |
| environ-sizes | 254.788 → 253.088 | 757.288 → 761.796 | 255.858 → 245.275 | 754.783 → 762.329 |
| args-copy | 331.258 → 327.229 | 992.133 → 1000.117 | 331.917 → 323.429 | 1015.183 → 999.017 |
| integer-control | 2.850 → 3.142 | 2.538 → 2.621 | 2.775 → 2.921 | 2.375 → 2.496 |

Units are warm ns/source iteration. First native integer regresses Go 10.2% /
LLGo 3.3%; revised 5.3% / 5.1%. Corresponding baseline A/A rows and all
interpreter regressions remain in evidence. Revised args improves Go 4.2% but
regresses LLGo 1.5%; env improves Go 4.1% but regresses LLGo 1.0%. Revised
copy improves LLGo 1.6% against A/A -1.4%, without resolving other cases.
Revised LLGo interpreter args/env/copy regress 3.5%/4.7%/4.5%. No source-identity
waiver, new timing-only LLGo issue, favorable-only rerun or throughput claim.

For revised args, enabled wazero measures 38.733 ns and default Wasmtime 55.400;
LLGo head remains 19.56x / 13.68x those configurations and 3.10x Go. Ordinary C
measures 0.8 ns on this short unchecked provider; it omits VM contracts and is
a coarse lower bound. The earlier sustained integer-vs-C result does not supply
portfolio acceptance for either new source.

## Predeclared revised GC controls

A separate exact-archive harness enables the same supported Core features on
both revisions, preserving the pure-module contract. Four original GC binaries,
10000 iterations for allocation and one million for nonallocating controls,
five fresh blocks/two warmups/seven samples. Base/head Go/LLGo native, exact
base A/A and Wasmtime: 140 observations, 28 alternate checks, eight explicit
wazero/Wasmer GC rejection records. No invented C GC or interpreter acceptance.

| Kernel | Go base → head | LLGo base → head | Wasmtime |
| --- | ---: | ---: | ---: |
| i31-roundtrip | 2.554750 → 2.557750 | 2.547041 → 2.554500 | 1.820000 |
| struct-access | 1.316125 → 1.321041 | 1.303417 → 1.305417 | 1.040000 |
| struct-allocation | 46.058400 → 46.266700 | 53.187500 → 54.262500 | 104.400000 |
| struct-indirect-access | 4.584167 → 4.628584 | 4.607875 → 4.595666 | 2.559000 |

Allocation regresses LLGo 2.0%; nonallocating kernels retain 1.25–1.80x
Wasmtime gaps. GC baseline A/A changes and startup remain explicit; LLGo
readiness is around 95 ms, still far over its 10 ms target. Reducing one known
allocation has not resolved the dominating boundary costs. Keep both source
experiments out of the selected contribution. Final CI, controlled both-platform
portfolio, allocation, startup and cancellation acceptance remain open.

[Both source archives, assembly, toolchains, all phases/A/A/controls and validation](evidence/performance-native-allocation-scope-held-2026-11-03.json.gz)
