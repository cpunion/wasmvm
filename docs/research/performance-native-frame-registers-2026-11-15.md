# Native frame usage and depth in registers

Unmetered linked modules retain logical arena usage and depth in two dedicated
callee-saved integer registers. Ordinary entries seed them before validation;
every exit publishes them. All allocation/cache/leaf plans exclude the words.
The caller derives the checked child address from its current frame and avoids
repeated context loads/stores. A return restores parent state only after its
complete result copy succeeds, preserving cancellation inside large copies.
See the [ABI contract](../design/native-abi.md#checked-internal-call-entries).

Complete 128-bit values, zeroed children, physical/logical quotas, cancellation,
roots, helper/host/foreign exits and ordinary entry checks remain. Metered and
unlinked bodies retain their context-based state. Public record layout and
hardware stack usage do not change. There is no new executor or guest shortcut.

## Sources and protocol

Measured base is integrated main **798e5443**, head **0baaf03e**. Seven-file
symmetric Git overlays freeze every production difference; report-only changes
retain that code. Go 1.27.3, LLGo 6a482f7f, LLVM 22.1.8, `-O3 -lto=full` and
default production GC are fixed. References remain wazero 1.12.0, Wasmtime
48.0.1 and Wasmer 7.3.0. Host: M4 Max/macOS arm64. Other user compiler/link and
desktop activity was observed; load snapshots and all controls remain. No owned
build, test, profile, source edit or pruning overlaps timed cohorts.

General calls use twelve fixed blocks/192 fresh processes: Go/LLGo base/head,
three references and unchecked native C, each with an identical-binary alias.
Every process retains readiness, preparation/compilation, instantiation, first
call, two warmups, seven interleaved five-million-iteration calls and checksums.
C executes 2N equivalent affine steps, normalized per guest iteration, without
VM limits/cancellation/frame contracts. A separate ten-block/40-process Go
counter audit measures twenty interleaved five-million-iteration calls plus a
separate flush. Events are not extended or pooled.

## General native calls

| Go head/base counter audit | Ratio [individual 95% interval] |
| --- | ---: |
| CPU | 0.88842 [0.88440, 0.89228] |
| Instructions | 0.93594 [0.93592, 0.93597] |
| Cycles | 0.88835 [0.88485, 0.89161] |
| Cycles/instruction | 0.94915 [0.94535, 0.95268] |

Both own controls resolve 2.5% for these metrics. CPU falls **11.16%**, from
approximately 7.29 to 6.49 ns/iteration; instructions fall about 203 to 190.
The larger CPU reduction is consistent with fewer context memory dependencies;
the exact contribution of dependencies versus layout remains unisolated.

The separate complete comparison gives Go warm head/base **0.87895
[0.87070, 0.88746]**, with both own controls resolving. Head/reference points
are **2.875x wazero, 2.969x Wasmtime, 3.006x Wasmer and 2.992x C**. The first
three reference controls resolve; C's does not. Competitive parity remains open.
LLGo head/base is **0.87287 [0.86078, 0.88490]**, with its base control narrowly
unresolved. LLGo/Go is 1.00293 [0.99721, 1.00869].

An untimed audit serializes this exact guest through the frozen Wasmtime 48
Cranelift speed profile. The callee is not inlined: it uses register arguments,
`BL`/`RET`, and a small hardware frame. Its static nonfinal outer iteration is
35 instructions. wasmvm still performs more explicit frame, metadata, full-value
copy and bounded cancellation work. This identifies internal call state/ABI and
allocation as further compiler work; host GC cannot explain the closed native
kernel's gap. The reference's interruption policy and unchecked C contract are
not equated to wasmvm's. Serializer source, identity and disassembly are archived.

## Complete P3 requests and LLGo collection sensitivity

Each payload has ten fixed blocks/100 fresh processes: Go/LLGo base/head and
independent Wasmtime, each with an alias. Twenty 512-request batches are
individually interleaved with four Stores/sixteen clients, snapshot input and
buffered output. Every response checks body/status/headers/trailers/CRC.
Preparation, first call, warmups, final flush and cleanup stay separate.

| Go head/base | 16 B | 64 KiB |
| --- | ---: | ---: |
| CPU | 0.99486 [0.98974, 0.99895] | 0.99337 [0.98529, 1.00190] |
| Instructions | 0.99391 [0.99260, 0.99508] | 0.99432 [0.99363, 0.99494] |
| Throughput | 0.99513 [0.97217, 1.01504] | 1.00973 [0.99839, 1.02339] |

Both Go CPU controls resolve. The 16 B CPU reduction is small; 64 KiB CPU and
both throughput comparisons cross parity. The 16 B head throughput control is
unresolved. Go/Wasmtime CPU points remain 2.593x/2.612x, with unresolved reference
CPU controls. Managed allocations stay approximately 696/756 per request; Rust/C
allocations are not counted by these host counters.

LLGo normal-GC CPU changes **0.94643 [0.90755, 0.98342]** and **1.11302
[1.04629, 1.18552]**, with unresolved own controls for both payloads. Both the
favorable and adverse points are retained. For 64 KiB, median measured collections
increase 50.5 to 61 per 10,240 requests; initial heap median changes 49.98 to
34.30 MB. Both sources range from 35/39 to 111 collections. CPU tracks those
buckets: about 266–270 us/request near 35/39 collections versus 414–437 at 111.
For 16 B, median collections fall 39 to 28. Allocations remain 1165/1430.

Two separate bounded no-GC diagnostics each have ten blocks/40 processes and
five measured 512-request batches plus a flush. All measured collection counts
must be zero. LLGo instruction ratios become **0.99523 [0.99354, 0.99693]** and
**0.99584 [0.99340, 0.99793]**, with resolving instruction controls. CPU ratios
are 1.00070 [0.99073, 1.01223] and 1.00616 [0.98792, 1.02415]; both cross parity
and one CPU control per payload is unresolved. Collection work materially
contributes to the normal-GC changes; precise cold heap sizing versus conservative
root retention remains open in [LLGo #3106](https://github.com/xgo-dev/llgo/issues/3106).
Production GC is unchanged, and the normal-GC regression is not relabeled a pass.

## Cold points, prototypes and validation

The general-call Go compile-wall point increases 5.78% [0.95593, 1.16842 ratio];
all cold controls are unresolved. P3 compile points are approximately flat for
Go and +1–2% for LLGo, also with unresolved controls. No cold speedup or exact
regression magnitude is established. Fewer guest registers can change allocation
plans and code layout; cold scaling and larger pressure-heavy workloads remain
follow-up work.

The first register-only prototype's separate audit gives CPU ratio 0.90656
[0.90133, 0.91104], with resolving own controls. Its two four-block P3 screens
do not establish CPU gains. Those records remain separate from final source.
The initial exact code-budget test expected the old profile; mirroring the new
production configuration restores its exact quota/fallback assertion.

All 32 Go root test packages pass; all 31 measured executable packages reach
90% statement coverage. LLGo executes affected codegen/nativeexec tests using
the production profile. New cases exercise state-register allocation exclusion,
invalid configurations, large Used increments beyond ARM immediate range, and
ordinary/invalid entry state preservation. Existing full proposal, root, mixed
tuple, reentry, quota, cancellation, cleanup and official WASI tests stay green.
The complete dual-compiler/platform CI is required on the final PR head.

The [machine-readable report](data/performance-native-frame-registers-2026-11-15.json)
links **600 process records**, all nine distinct cohorts, hashes, source overlays,
protocols, load snapshots, disassembly and raw phase/GC/counter evidence. Only
unused per-request timing vectors become count/extrema/canonical-JSON hashes in
the committed projection; full originals remain local. Milestone performance
and production-release acceptance remain open.
