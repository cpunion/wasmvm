# WasmVM LLGo performance handoff

Snapshot taken on 2026-11-24 on macOS arm64. The current objective is to bring
LLGo-built WasmVM to parity with Go-built WasmVM on the same source and workload,
while preserving semantic coverage and the performance already achieved against
other engines. Keep the project-owned interpreter and native compiler, a compact
shared architecture, and shared Go/LLGo tests. Report Core interpreter, Core
native, and Preview 3 (P3) separately. Measure preparation, instantiation, first
guest return, and warmed execution separately. P3 public API time, application
throughput, and whole-process CPU time have different boundaries.

The declared functional delivery includes Core 2.0/3.0, WasmGC, Core Threads,
the Component Model, and WASI Preview 1, 2, and 3. The project has validated the
declared target and consumer inventory on both backends and compilers.
Expanded world coverage, performance acceptance, and release acceptance are
tracked separately in the [roadmap](../../roadmap.md),
[release contract](../../release-readiness.md), and
[performance contract](../../design/performance-contract.md). The
[standards support table](../../../README.md#standards-and-proposal-support)
has the scope of the separately gated proposals. This handoff captures the
existing code and evidence so review can proceed while optimization is paused.

## Starting points

| Repository | Starting revision | Contents |
| --- | --- | --- |
| WasmVM main | [`3d5c7e4c`](https://github.com/cpunion/wasmvm/commit/3d5c7e4c1d069defdbb6f53614b1d20477c5aba1) | Includes byte-source codec placement [#209](https://github.com/cpunion/wasmvm/pull/209), native compiler word stores [#208](https://github.com/cpunion/wasmvm/pull/208), interpreter update/address prefixes [#207](https://github.com/cpunion/wasmvm/pull/207), and P3 HTTP readiness [#206](https://github.com/cpunion/wasmvm/pull/206). The source tree matches the measured #209 head `ec3b3a42`. |
| WasmVM draft review | [`7fd13e7c`](https://github.com/cpunion/wasmvm/commit/7fd13e7c1a326d41051a028835b3953cb9089bb6) plus this documentation | Reuses the previous immutable native register plan while both banks retain the same physical homes. The register-plan commit is the exact measured source. |
| LLGo upstream | [`42efacae`](https://github.com/xgo-dev/llgo/commit/42efacaeff2e196622296415540e927724f7dcef) | Includes the Darwin ARM64 backend features used by full LTO from [#3197](https://github.com/xgo-dev/llgo/pull/3197). This is the refreshed compiler baseline for the data below. |
| Compiler profile | Go 1.27.3 and LLGo `-O3 -lto=full` | The full competitor report also includes the pinned LLGo `3c6b33ec` as its own column. Keep immutable source, fixture, compiler, and collector identities when replaying. |

## Measurements and decisions

### Native register-plan candidate

The four-worker compiler census attributed 301,296 allocations per process
to register plans across three compilations. Coalescing consecutive unchanged
plans reduced subsequent P3 component compile time by **6.57% with LLGo**
[5.50%, 7.65%] and **5.03% with Go** [4.34%, 5.72%]. The remaining same-source
LLGo compile-time gap was **+32.02%** [30.99%, 33.07%]. Both Core modules in the
Rust service component emitted byte-identical native code and entry tables.
Go and LLGo ran the same codegen, nativeexec, and VM tests. Go statement coverage
was 95.6%, 89.6%, and 92.7% respectively. The focused Go race run passed for
codegen and nativeexec. [Paired compile results](evidence/p3-compile-register-plans.json) ·
[Emitted images](evidence/p3-compile-emitted-images.json)

**Keep this candidate in draft.** In an independent replay of eight warm
workloads and 576 verified processes, Go interpreter function calls took
**1.35% longer** [1.00%, 1.69%] with resolved identical-binary controls. Retired
instructions were essentially equal and cycles increased 1.28%. The hot
functions retained their mnemonic sequences while relative call distances,
address immediates, and the entry position in a cache line changed. The data
implicate linked layout. The particular microarchitectural event remains
unresolved. The other seven replayed cases had unresolved controls or confidence
intervals spanning equality. [Independent replay](evidence/warm-replay.json)

The [full competitor report](native-register-plan-portfolio.md) covers 33 Core
workloads, 7,740 verified Core records, and 192 P3 records. Go, pinned and refreshed
LLGo, wazero, Wasmtime, Wasmer, and WasmEdge have individual columns wherever a
comparable adapter exists. Each scope reports its versions, absolute timings,
percentage differences, and unresolved or unsupported cells. In that event,
refreshed LLGo ran the 64 KiB P3 case at **33.40% lower throughput** and **78.15%
more CPU per request** than same-source Go. Both controls resolved.
[Source](evidence/portfolio-source.json) · [Build identities](evidence/portfolio-builds.json) ·
[Process verification](evidence/portfolio-verification.json) ·
[Same-source comparisons](evidence/portfolio-same-source.json)

### P3 runtime and collector composition on merged WasmVM

This independent event used `3d5c7e4c` with a fixed P3 service workload.
The combined LLGo arm uses the recovered-aggregate compiler at `9806ccca`,
timer head `7db84420`, Darwin channel head `6ae45e71`, and a diagnostic
BDWGC 8.2.12 allocator-lock and fresh-refill build. The two collector arms
load different verified libraries into the same Mach-O. The combined SDK
passed the Component Model and P3 packages. The timed suite verified 400
processes and 4,096,000 warmed responses. All metrics below compare the
identified arm with Go in this event. † means identical-binary controls
were unresolved. The configurations use four workers with 16 clients and
one worker with one client.

| Payload / workers | Stock LLGo CPU | Combined CPU | Combined throughput | Combined public API median |
| --- | ---: | ---: | ---: | ---: |
| 16 B / 4 | +69.54% | +45.70% | −27.29% † | +38.54% † |
| 64 KiB / 4 | +40.41% | +24.15% | −11.29% † | +29.08% † |
| 16 B / 1 | +51.57% | +26.77% | −22.96% | +13.85% |
| 64 KiB / 1 | +50.38% † | +27.71% † | −21.24% | +16.09% |

Collector-only CPU time reductions span 5.03–9.15% with resolved controls.
The sampled extra CPU for the 64 KiB four-worker combined image is roughly
14.8 µs/request in the benchmark service outside the public API timer,
10.6 in the collector, and 7.4 in component task/call paths. Named channel
wait/wake and collector/refill kernel leaves account for 7.47 and 11.09
µs/request respectively in the combined sample. These are fixed capture
attributions. Use the paired timing blocks for effect sizes and preserve
the application/API boundary when investigating the residual.
[Paired results](evidence/p3-composition-comparison.json) ·
[Build and collector hashes](evidence/p3-composition-builds.json) ·
[Kernel leaves](evidence/p3-composition-kernel-leaves.json)

### Other candidate outcomes

| Line of work | Verified outcome | Decision |
| --- | --- | --- |
| LLGo post-inline heap placement [#3200](https://github.com/xgo-dev/llgo/issues/3200) | A shared Go/LLGo reproducer goes from 16,384 LLGo allocations to zero. WasmVM removes four allocations per request in `Codec.validateLayout`. Both measured CPU intervals include equality. Tail-marked allocator tests preserve the distinction between result and inputs. | Review the fork snapshot `8ad8fd37` independently. [Probe](evidence/llgo-post-inline-heap-probe.json) · [Timing results](evidence/llgo-post-inline-heap.json) |
| LLGo recovery aggregate [#3213](https://github.com/xgo-dev/llgo/pull/3213) | On measured compiler `9806ccca`, P3 removes 12 allocations per request and reduces public API time by 1.79% and 1.54% on 16 B and 64 KiB. Total CPU intervals span equality and 64 KiB user CPU increases 0.73% with resolved controls. | Draft head `818f43c4` adds portable test and FileCheck fixes after the timing event. Keep the runtime and compile diagnostics distinct. [Timing results](evidence/llgo-recover-aggregation.json) |
| Large pointer-free allocation zeroing [#3210](https://github.com/xgo-dev/llgo/issues/3210) | The collector prototype passes 17 BDWGC tests plus dirty reuse, concurrent and incremental collection, and weak-link checks. In the single-image replay, Rust SHA instantiation through first guest return is 69.06 µs on Go and 148.25 → 90.63 µs on LLGo. Full-write total rises by a paired estimate of 4.27%. Timing controls remain unresolved. | Experimental. Review the standalone patches and full-write cost before proposing a production change. [Comparisons](evidence/sha-cold-comparison.json) |
| WasmGC layout access | Sustained LLGo interpreter allocation time falls 15.44% and still takes 6.56% more time than Go. Native allocation takes 2.15% longer with resolved controls. Go race tests, focused LLGo tests, and 61,395 Core 3.0 assertions per execution tier pass. | Preserve `b6687305` for review and isolate the native regression before integration. [Timing results](evidence/wasmgc-layout-comparison.json) · [Linked allocation audit](evidence/wasmgc-native-allocation-audit.json) |
| Dedicated native worker queues | LLGo compile time changes by +0.05% [−0.83%, +0.99%]. Disabling automatic GC raises subsequent LLGo compile time by 19.53% and Go time by 3.32%. | Preserve `268475a9` as the queue isolation branch. Use the measured allocation sites for the next experiment. |
| Coroutine prototype | Selected Darwin TCP, UDP, and FIFO cases execute. The verified 16 B P3 diagnostic retires 26.46× stock LLGo instructions [26.11×, 26.94×]. `PrepareInlineAwaitCompiler` is the largest named user-space leaf in the sampled run. | Treat the compiler and SDK snapshots as diagnostic. |

## Existing LLGo contribution PRs

These are upstream contribution PRs from the `cpunion/llgo` fork.
The head revisions were captured on 2026-11-24. Known review threads were
addressed before handoff. Use the linked live checks when completing review.

| PR | Head | Review focus | Checks at snapshot |
| --- | --- | --- | --- |
| [#3192](https://github.com/xgo-dev/llgo/pull/3192) | `7db8442004bab489e78e259676869341fb2c3b8f` | Timer wake and one-shot cleanup | 24 pending |
| [#3199](https://github.com/xgo-dev/llgo/pull/3199) | `e7abea391744935b8ad990d11e5e9b97c403dd36` | Fixed-size byte append lowering | Complete, zero failed |
| [#3209](https://github.com/xgo-dev/llgo/pull/3209) | `6ae45e7177b1e7dc193217a5cd59ffa541ffb5aa` | Darwin address-based channel wait/wake | Complete, zero failed |
| [#3211](https://github.com/xgo-dev/llgo/pull/3211) | `a5bc2858d240d983d01459e3408ad073bfc93bde` | Scalar struct-key algorithms with a shared semantic test | Complete, zero failed |
| [#3212](https://github.com/xgo-dev/llgo/pull/3212) | `4d0b24adb1934e55182f3a991be3d944236d2b6f` | Regression coverage for already-merged #3197. Verifies final ARM64 full and thin LTO from stage-2 LLGo, with a shared Go/LLGo functional test. | 32 pending |
| [#3213](https://github.com/xgo-dev/llgo/pull/3213) | `818f43c4f7ca48f852d3603cf8819d20d4df8467` | Bounded defer/recover aggregate placement and portable regression coverage | Draft, 34 pending |

## Branch and worktree inventory

Paths identify checkouts on the original machine. The remote refs and
revisions are the durable recovery points. A dash in the worktree column
means the candidate is available as a branch without a dedicated checkout.
Historical and unrelated worktrees can be inspected with
`git worktree list --porcelain` in each root checkout.

### WasmVM

The main working copy at `/Users/lijie/source/goplus/wasmvm` stays on local
`main` at `d45d609d` and contains the ignored measurement artifacts.
Use the recorded upstream revision for baseline work.

| Branch in `cpunion/wasmvm` | Revision / purpose | Local worktree |
| --- | --- | --- |
| `handoff/llgo-parity-20261124` | Draft review branch containing the register-plan candidate and this record | `/private/tmp/wasmvm-llgo-parity-handoff-20261124` |
| `perf/reuse-native-register-plans-20261124` | `7fd13e7c`, exact measured source | `/private/tmp/wasmvm-reuse-native-register-plans-20261124` |
| `perf/gc-layout-field-access-20261124` | `b6687305`, WasmGC field-access candidate | — |
| `perf/native-worker-queues-20261124` | `268475a9`, native compile queue ablation | — |
| `perf/p3-scheduled-and-option-20261123` | `7ffad23f`, scheduled P3 work and option payload experiment | `/private/tmp/wasmvm-p3-llgo-parity-20261123` |
| `perf/expire-vm-host-caller-once-20261123` | `b0a2da95`, borrowed continuation/caller-lifetime experiment | `/private/tmp/wasmvm-direct-borrowed-continuation-20261123` |
| `perf/opaque-read-critical-section-20261123` | `9f8619c5`, opaque read experiment | `/private/tmp/wasmvm-opaque-read-critical-section-20261123` |
| `perf/component-comparison-key-width-20261123` | `a894d10a`, component key-width experiment | `/private/tmp/wasmvm-comparison-key-width-20261123` |
| `perf/reuse-byte-source-codecs-20261123` | `a0562091`, earlier codec candidate | `/private/tmp/wasmvm-byte-source-codecs-20261123` |
| `perf/byte-source-codec-stack-20261123` | `ec3b3a42`, merged #209 head | `/private/tmp/wasmvm-typed-codec-cleanup-20261123` |
| `perf/interpreter-vector-load-constant-20261123` | `d51db59a`, interpreter diagnostic and runner/fixture worktree | `/Users/lijie/source/goplus/wasmvm-source-cleanup` |

### LLGo

The original working copy is `/Users/lijie/source/goplus/llgo`,
`fix/hostlayout-followup-20261108` at `0f804925c`. Its `origin` is
`cpunion/llgo` and its `upstream` is `xgo-dev/llgo`.

| Local branch | Revision / remote recovery point | Local worktree |
| --- | --- | --- |
| `perf/3200-post-inline-heap-20261124` | `8ad8fd37`, same branch in `cpunion/llgo` | `/private/tmp/llgo-3200-late-heap-20261123` |
| `perf/darwin-lazy-zero-3210` | `42efacae`, clean scaffolding. See the collector and SDK patches below. | `/private/tmp/llgo-darwin-lazy-zero-3210-20261124` |
| `coro/loop-carried-panic-interface-20261124` | `f72bd595`, same branch in `cpunion/llgo` | — |
| `coro/darwin-dynimport-callable-shadow-20261124` | `0c738769`, intermediate ancestor of the FIFO snapshot | `/private/tmp/llgo-coro-darwin-dynimport-20261124` |
| `coro/darwin-mkfifo-worker-20261124` | `9fda0568`, same branch in `cpunion/llgo` | `/private/tmp/llgo-coro-darwin-mkfifo-20261124` |
| `review/runtime-timer-wake-3191-20261124` | `7db84420`, PR #3192, fork branch `perf/runtime-timer-wake-3191` | `/private/tmp/llgo-pr-timer-review-3191-20261124` |
| `fix/3196-fixed-byte-append-20261123` | `e7abea39`, PR #3199 | `/private/tmp/llgo-fixed-byte-append-20261123` |
| `perf/darwin-address-current-20261124` | `6ae45e71`, PR #3209 | `/private/tmp/llgo-darwin-address-current-20261124` |
| `audit/native-scalar-equal-test-20261124` | `a5bc2858`, PR #3211, fork branch `perf/scalar-struct-key-algorithms-3207`. The local branch with the fork name points at an earlier commit. | — |
| `fix/darwin-arm64-lto-features-3204` | `4d0b24ad`, PR #3212 | `/private/tmp/llgo-pr-darwin-lto-3204-20261124` |
| `perf/recover-aggregate-storage-3208` | `818f43c4`, PR #3213 | `/private/tmp/llgo-pr-recover-aggregate-3208-20261124` |

## Reproducing and handing over

From the respective existing clones, recover the exact candidate worktrees
on any machine. Substitute local paths for the example relative paths.

```sh
# In the WasmVM clone, with origin pointing to cpunion/wasmvm.
git fetch origin
git worktree add --detach ../wasmvm-register-plan-review 7fd13e7c1a326d41051a028835b3953cb9089bb6
git worktree add --detach ../wasmvm-parity-base 3d5c7e4c1d069defdbb6f53614b1d20477c5aba1

# In the LLGo clone. Add a fork remote under a free name if needed.
git remote add cpunion https://github.com/cpunion/llgo.git
git fetch cpunion
git worktree add --detach ../llgo-heap-placement-review 8ad8fd37f77c0d5f5efb68bf151aaf683ebd5c37
```

Run the shared tests and distinguish both compiler outcomes. Point LLGo
at the matching stage-2 SDK. Use the repository's pinned fixtures and
the same target and compiler profile on both sides of a comparison.

```sh
go test ./internal/compiler/codegen ./internal/engine/nativeexec ./pkg/vm
llgo test -O3 -lto=full ./internal/compiler/codegen ./internal/engine/nativeexec ./pkg/vm
```

The [benchmark guide](../../../benchmarks/README.md) and
[P3 driver guide](../../../benchmarks/p3io/README.md) describe the
in-repository entry points. Run macOS arm64 locally and Linux amd64 in
local Docker. The published timed comparisons here are macOS arm64.
The exact historical runners and raw process records are retained under
`/Users/lijie/source/goplus/wasmvm/artifacts/arm64-compact-loop-poll-20261121`
on the original machine. Start with these relative directories in that root.

| Raw directory | Contents |
| --- | --- |
| `portfolio-native-register-plans` | Complete benchmark plans, runners, independent warm replay, binary provenance, and tests |
| `p3-native-register-plan-reuse` and `p3-native-compile-allocation-census` | Compile timing, allocator site counts, and native image check |
| `p3-reviewed-runtime-composition/refresh-42efacae` | Composition builds, same-image replays, semantic logs, and stack profiles |
| `llgo-post-inline-heap-placement/refresh-42efacae` | Dual-compiler allocation reproducer and caller census |
| `sha-cold-instance-and-memory/collector-fresh-span` | Fresh-span collector, single-image selector, and full-write follow-up |
| `wasmgc-layout-field-access`, `p3-parallel-compile-coordination`, and `p3-coro-darwin-mkfifo` | Rejected or qualified alternative experiments |

The tracked JSON copies use filenames in place of absolute local
artifact paths. They preserve the recorded identities, hashes, counts,
timings and control outcomes. The diagnostic
[allocator lock](patches/bdwgc-unfair-lock.patch) and
[refill selection](patches/bdwgc-fresh-refill.patch) patches target BDWGC
8.2.12. The separate fresh-span experiment uses the
[collector patch](patches/bdwgc-fresh-span.patch),
[runtime selector](patches/llgo-fresh-span-selector.patch), and
[C probe](patches/bdwgc-fresh-span-probe.c.txt). Evaluate these in isolated
builds with a separate library path. The timed artifacts and production
SDK stay pinned to their recorded dependencies.

## Review sequence

1. Review the LLGo compiler and runtime contribution PRs independently.
   Start from their current heads and use common Go/LLGo test source where
   semantics are shared. Keep the already-merged #3197 change separate from
   the #3212 coverage follow-up. Use fork PRs for any compiler updates.
2. In the WasmVM draft, review plan immutability and bank residency.
   Diagnose the Go interpreter function-call layout regression before
   requesting merge. Rerun the exact final-code portfolio with the full
   compiler and competitor columns after a production code change.
3. For the remaining P3 difference, isolate one-worker component task/call
   CPU, then four-worker service, collector and wait/wake CPU. Compare fixed
   code and library identities before attributing a gain. Revisit
   four-worker compile allocation and the 64 KiB first-return/full-write
   tradeoff independently of the warm P3 investigation.
4. Recheck shared semantics and the controlled interpreter, native and P3
   portfolios before declaring LLGo parity. The existing validation records
   define the required platform matrix. [Validation](../../validation.md)
