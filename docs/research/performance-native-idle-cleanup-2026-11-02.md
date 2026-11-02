# Idle native allocation cleanup: 2026-11-02

Status: **unmerged standalone candidate; full VM performance is not accepted**.
This compact workaround addresses an independently reproduced LLGo issue, but
its isolated helper win does not establish a GC-loop win. PR #101 remains the
sole active contribution and is not modified by this experiment.

## Change and provenance

Baseline source `25298a667a587da9b89663149ff92950c66c9e9d` already has eight-copy
native memory groups; its engine/test source equals PR #101 `dc580d6` (only
documentation differs). Reuse its previously verified exact binaries without
relabeling the reported revision. Candidate
`c4fef4f2c4d53244b02d8cbeea94573ba1bc6608` adds seven lines: the existing
idle/no-heap checks remain in FinishNativeAllocations, while its entire original
active-batch body and rollback defer move to one private helper. There is no
second executor, omitted error path or changed batching/collector policy.

The same 86 guest bodies on both architectures and tiers remain byte-identical.
The improvement therefore targets host cleanup, not faster generated kernels.
Foreign bodies are compile-only, not performance evidence.

## Isolated production-helper proof

Go 1.27.3, pinned LLGo `8e45ed5` / `-O3 -lto=full`, macOS arm64 / Apple M4 Max.
Five fresh processes per compiler/revision, randomized ordering,
GOMAXPROCS=1, 100 warmups then 100,000 calls per case. Both a portable Store
without a native heap and a native Store with no pending batch return success
and preserve zero guest heap objects. All raw records and build identities are
retained. TotalAlloc is approximate host-collector accounting, not a Mallocs
counter or a general no-allocation claim.

| Idle call | Go bytes/call base → head | LLGo bytes/call base → head | Go ns base → head | LLGo ns base → head |
| --- | ---: | ---: | ---: | ---: |
| Portable Store | 0 → 0 | 79.973 → 0 | 1.799 → 1.055 | 45.410 → 1.008 |
| Native Store | 0 → 0 | 80.000 → 0 | 1.661 → 1.014 | 44.970 → 0.998 |

Final-binary assembly independently shows the baseline's AllocU(8) receiver
capture and AllocU(48) exception frame before either idle guard. The candidate's
inlined wrapper tests both guards before calling the deferred helper. Both idle
branches return without an allocation call. This is a concrete mitigation of
[LLGo #2932](https://github.com/xgo-dev/llgo/issues/2932#issuecomment-6579659757),
not a compiler fix, pin change or attribution of the full host-compiler GC gap.

## Real native workload comparison

Three fresh serially interleaved process blocks, two warmups and seven calls.
Four GC kernels and three scalar native controls use both host compilers,
base/head, exact baseline A/A aliases and applicable native competitors.
The linked allocation chain preserves its original 10,000 iterations and quotas;
nonallocating GC controls use one million, guest calls five million, scalar
memory 50 million, empty call one. All participants receive the same guest bytes,
counts and oracles. Ordinary Clang 22.1.8 O3 C controls apply to scalar calls and
memory only. There is no invented C GC implementation.

165 VM observations, 6 C observations, 55 alternate checks, 57 summary rows,
and 8 explicit GC-unsupported records complete. All readiness, preparation,
first/warm execution and cleanup intervals remain separate in the raw evidence.
Wazero/Wasmer native rows are included on scalar cases; their declared GC
limitations remain explicit. The state-only native path does not need a repeated
interpreter campaign.

Warm ns/source iteration (the one-call control is ns/call):

| Workload | Go base → head | LLGo base → head | Wasmtime | C |
| --- | ---: | ---: | ---: | ---: |
| core/call-control | 5875.000000 → 2792.000000 | 5375.000000 → 4666.000000 | 1000.000000 | — |
| core/function-calls | 0.772133 → 0.763283 | 0.775692 → 0.769992 | 0.882000 | 0.747800 |
| core/memory-load-store | 0.417416 → 0.418257 | 0.413724 → 0.415965 | 0.468400 | 0.268120 |
| gc/i31-roundtrip | 2.549209 → 2.492458 | 2.507083 → 2.550042 | 1.808000 | — |
| gc/struct-access | 1.274458 → 1.310333 | 1.303000 → 1.286041 | 1.030000 | — |
| gc/struct-allocation | 46.800000 → 50.170800 | 59.308400 → 60.300000 | 103.900000 | — |
| gc/struct-indirect-access | 4.556584 → 4.608708 | 4.589459 → 4.559334 | 2.531000 | — |

GC allocation regresses Go 46.800→50.171 (7.2%) and LLGo 59.308→60.300 (1.7%).
Go paired ratios are 1.0981/1.0219/1.0628 with A/A 1.0317/0.9995/1.0485;
LLGo 1.0227/1.0222/0.9723 with A/A 1.0914/1.1277/1.0646. These observations
neither attribute the full adverse result to the wrapper nor establish a stable
benefit. LLGo head is about 1.20x Go and faster than the configured Wasmtime
allocation case; the complete compiler gap stays open.

Other adverse observations include LLGo memory +0.4%, Go struct access +2.8%
and LLGo i31 +1.7% by aggregate medians. A guest-call Go block regresses 33.95%
despite a lower aggregate median; baseline A/A in that block also regresses 6.6%.
The empty-call control varies widely between blocks and stays above its budget;
its apparent head improvement is not accepted. All observations are retained,
without dropping inconvenient controls or treating identical guest bytes as a
proof of equal host performance. LLGo scalar memory remains about 1.55x C in
this campaign; those rates are not pooled with earlier ARM64 or Linux campaigns.

## Validation and next step

All 31 Go packages and all 30 coverage gates pass, including state 95.09%.
Pinned LLGo state/nativeexec/VM and scoped race/no-cgo/vet pass. Existing tests
exercise active and idle finish, complete/partial commits, invalid batches,
quota/root/generation reuse, traps and cancellation. Cleanup remains deferred on
the active path; no timeout or correctness contract is relaxed.

Hold this standalone candidate rather than attributing its helper win to the
whole VM. The next isolated experiment increases the bounded reservation batch
to amortize host publication. It uses this exact candidate as its baseline so
batch-size effects are not mixed with the idle-wrapper effect. Controlled
cross-platform performance, allocation and expanded WASI acceptance remain open.

[Complete measurements, helper proof, source identities and validation](evidence/performance-native-idle-cleanup-2026-11-02.json.gz)

The subsequent [bounded batch combination](performance-native-allocation-batches-2026-11-02.md)
is measured against the original PR engine and this helper-only middle control
in one twenty-block campaign. It is a separate follow-up candidate; the standalone
regressions above remain recorded rather than being relabelled as gains.
