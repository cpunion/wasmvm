# Bounded eight-copy native memory groups: 2026-11-02

Status: measured ARM64 candidate included in PR #101; final-head CI and Linux
performance are pending. This is a descriptive shared-host campaign, not a
controlled parity or complete portfolio result.

## Change and semantic boundary

Base `872e257d71779d1d20d0b92fcaba48fb56aa4a09` already contains the native
comparison/immediate cost changes and TCP shutdown fix. Measured source
`25298a667a587da9b89663149ff92950c66c9e9d` changes the existing memory group
selection to eight copies for short proved loops, four or two otherwise. Every
group still covers at most 256 original slot instructions. The existing
nontrapping, unsigned wrap, invariant bound, varying-address, memory and
observation proofs are unchanged. Cold entry at every original PC, exact final
frame/predicate values, short remainders and metered execution remain available.
This adds no new executor or workload-specific condition.

The 86 body pairs change only four unmetered bodies: scalar memory and float
memory on each architecture. AMD64 memory grows 3554→3886 bytes (9.34%); ARM64
3772→3932 (4.24%). AMD64 float grows 5208→5536; ARM64 4728→4920. All metered
bodies and the other 82 pairs are identical. The independently retained comparison
against main `ed5a0fb` includes all prior PR changes: ten of 86 bodies differ,
AMD64 memory 3653→3886 and ARM64 3796→3932. Foreign bodies were compiled only.

## Matching-work comparison

Apple M4 Max, macOS arm64; Go 1.27.3 and LLGo `8e45ed5` with `-O3 -lto=full`.
Six unchanged scalar guest binaries, two own compilers, native baseline/head,
identical-binary baseline A/A aliases, Wasmtime/Wasmer Cranelift, wazero compiler
and ordinary Clang 22.1.8 `-O3` C controls. Source/SDK/compiler/binary hashes and
commands are retained. C is an unchecked algorithm reference, not a VM policy.

Three fresh serially interleaved process blocks, two warmups and seven calls;
54 alternate-input checks pass, with 162 VM observations, 18 C observations and
60 complete summary rows. Core memory uses 50 million iterations; each other
kernel uses 5 million, including the conservative in-range float admission cap.
No interpreter retiming is needed for this native-only source change.

Warm results, ns/source iteration (median of process medians):

| Workload | Go base → head | LLGo base → head | Wasmtime | Wasmer | wazero compiler | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| core/integer-loop | 1.559633 → 1.551692 | 1.552183 → 1.554633 | 1.552800 | 1.551000 | 1.601767 | 1.556400 |
| core/function-calls | 0.785958 → 0.784108 | 0.782433 → 0.783183 | 0.901600 | 0.778000 | 1.318167 | 0.776400 |
| core/memory-load-store | 0.449895 → 0.425542 | 0.451034 → 0.419618 | 0.474800 | 0.521280 | 0.799582 | 0.285740 |
| memory/narrow-unaligned | 0.687917 → 0.685458 | 0.683917 → 0.681842 | 0.689800 | 0.714000 | 0.985400 | 0.460400 |
| memory/dependent-address | 1.814783 → 1.815950 | 1.811458 → 1.821742 | 2.057600 | 2.068800 | 3.027225 | 1.405400 |
| memory/float64-unaligned | 0.719175 → 0.715600 | 0.714625 → 0.716633 | 0.704200 | 0.707200 | 0.785500 | 0.662800 |

LLGo memory improves 6.97% and Go 5.41%. Paired block ratios are 0.93085,
0.93035, 0.90108 for LLGo, and 0.93874, 0.94587, 0.93213 for Go. LLGo A/A ratios
are 1.00842, 1.00850, 0.97145; Go 1.00941, 0.99940, 1.01506. This supports a
useful local candidate but does not meet the narrower controlled A/A gate.
Head LLGo memory is 11.6% faster than Wasmtime in this campaign, still **1.469x
C**. Narrow and dependent memory still take **1.481x / 1.296x C**. Float remains
about 1.8% behind Wasmtime and 8.1% behind C. All six cases remain visible.

The unchanged integer/call/dependent controls include small adverse observations:
LLGo +0.16%/+0.10%/+0.57% by aggregate medians. The changed LLGo float body is
+0.28%, with paired ratios 1.01258, 1.00289, 0.99884. Its baseline A/A also
moves +0.38%; this does not prove the changed body is harmless. Native host
compiler warm rates are close on these kernels; that does not resolve their
GC, startup or interpreter differences.

## Separate latency phases

These are the core-memory process medians. Compile/validate is the adapter's
reported aggregate API interval; it is not separately observed decode/codegen.
The 50-million-iteration first call is not a call-overhead control. All original
load/setup/export samples, timer resolutions and other cases remain in evidence.

| Memory configuration | Readiness, ms | Compile/validate, µs | Instantiate, µs | First call, ms | Cleanup, µs |
| --- | ---: | ---: | ---: | ---: | ---: |
| base-go-native | 4.534 | 378.125 | 19.417 | 22.762 | 6.625 |
| head-go-native | 4.470 | 298.000 | 15.417 | 21.029 | 6.833 |
| aa-go-native | 4.286 | 301.250 | 15.750 | 22.524 | 6.500 |
| base-llgo-native | 95.276 | 651.208 | 23.416 | 22.418 | 10.167 |
| head-llgo-native | 92.868 | 401.916 | 19.792 | 21.027 | 5.958 |
| aa-llgo-native | 93.663 | 457.292 | 19.375 | 22.673 | 8.375 |

LLGo readiness remains far above the 10 ms target in this campaign; warm
throughput does not close it. No setup phase is added to the warm kernel time.
No allocation-free or cancellation-latency claim follows from throughput.

## Validation and disposition

All 31 Go packages pass, all 30 coverage gates pass (codegen 97.91%), and pinned
LLGo codegen, execution and public VM suites execute successfully. Affected Go
race, no-cgo and vet pass. Boundary fixtures cover loop lengths 32/33 and 64/65,
code-size limits, every cold entry, 7/8/9 and 15/16/17 remainders, constant and
runtime bounds, both predicates, unsigned wrap and all final slots/memory.
Cancellation exercises eight/four/two-copy groups. Existing predicate-liveness
and invalid-proof tests run with the new group size. The first local invocation
failed from the missing worktree corpus symlink; its diagnostic is retained,
and the subsequent complete run uses the existing pinned fixtures.

Keep the candidate for final cross-platform verification in the sole active
PR #101. The separate x86 immediate-cost campaign is not this eight-copy Linux
measurement. Pinned Go 1.27.3 codegen tests also pass under Linux amd64 Docker emulation
in 7.885 seconds. The initial container start fails because the read-only source
mount lacks the nested fixture mountpoint; creating that empty directory fixes
setup without changing code or skipping tests. Both logs and source identities
are retained. Emulation provides semantic evidence, never Linux throughput.

No Linux speedup, controlled portfolio completion, general native
parity or expanded WASI conformance is claimed.

[Compressed raw results, exact identities, all phases, source and logs](evidence/performance-native-eight-copy-groups-2026-11-02.json.gz)
