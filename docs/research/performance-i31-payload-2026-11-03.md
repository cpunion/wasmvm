# Native i31 payload forwarding: 2026-11-03

Status: **merged in PR #106; final dual-platform semantic CI passed; controlled
performance acceptance remains open**. The [exact-head audit](pr106-final-ci-2026-11-03.md)
retains all reports and adverse hosted controls. The checked-memory baseline
is actual production source `ceea912`.
Sources `3f27587`, `1677123` and `1594908` incrementally forward an adjacent
`ref.i31` → `i31.get_s/u` pair, select the original integer as extraction source,
and select cached-register payload operations directly. The final source uses
ARM64 AND/SBFX and the existing AMD64 unsigned immediate-AND selector, with
generic spilled/signed-AMD64 fallbacks. This is one general native selector,
without benchmark identities, quota changes or a second execution framework.

A freshly constructed i31 is non-null. The fused fallthrough removes its
redundant null test, but keeps any intermediate boxed value that a later exit
can observe. It reads the original integer so preserving that value does not
extend extraction's dependency chain. Every interior PC retains its ordinary
independent checked continuation; metered code is untouched, and combinations
never cross a 256-instruction poll boundary. Original polls remain intact.

## Semantic and generated-code evidence

Both signed/unsigned forms execute under Go/LLGo with input/output/reference
slot aliases, cached/source-spilled/destination-spilled/both-spilled cases,
all original entry PCs, full frame comparison, signed boundaries/high input bits,
cold null-reference traps and source offsets. Tests cover kept/overwritten boxed
slots, metered tiers, mismatched producers and poll boundaries. Both targets
compile; this local host executes arm64, while final Linux CI must execute amd64.

Every source passes 31 Go packages, thirty package coverage gates, pinned
full-LTO LLGo codegen/nativeexec/VM execution, affected race/no-cgo and vet.
Final codegen coverage is 97.85%. Each source's test/build processes finish
before its production timing. Initial targeted test compilation used an int
expected counter against uint32; that test-only type error is retained. The
first separate byte witness omitted the decoder's GC feature gate and rejected
the module before compilation; corrected diagnostic source/logs are retained.
Neither is presented as a VM or LLGo failure.

An exact-archive witness compiles the original unchanged i31 module for both
architectures and tiers. Each unmetered target gains one selector hit, while
**metered bytes stay identical**. All body hashes, sizes, slot IR and counters
are retained. The first variant still used the tagged reference as dependency;
its modest result prompted an explicit source change, followed by a new
three-source campaign. No remeasurement of an unchanged source replaces a failure.

## Same-campaign final comparison

Shared M4 Max/macOS arm64, Go 1.27.3 and LLGo `8e45ed5`, `-O3 -lto=full`.
The final campaign measures baseline `ceea912`, middle `1677123` and head
`1594908` simultaneously with base A/A aliases, Wasmtime 48.0.1 and an ordinary
Clang-22 O3 C i31 kernel. The original GC guest bytes/inputs stay fixed: 10000
iterations for allocation, one million for other kernels. Five fresh process
blocks, two warmups/seven interleaved samples, independent alternate oracles.
180 VM/five C observations, 36 VM/one C alternate checks, eight explicit pinned
wazero/Wasmer GC rejections, 37 summaries. Source/compiler/binary identities
are checked before timing. Preparation, first/warm work and cleanup stay separate.

Warm ns/source iteration:

| Kernel | Go native base → head | LLGo native base → head | Wasmtime | Ordinary C |
| --- | ---: | ---: | ---: | ---: |
| i31-roundtrip | 2.559459 → 1.362583 | 2.543084 → 1.349500 | 1.811000 | 1.286000 |
| struct-access | 1.320917 → 1.318958 | 1.308458 → 1.313167 | 1.044000 | — |
| struct-allocation | 48.466700 → 48.495900 | 56.700000 → 54.779200 | 107.500000 | — |
| struct-indirect-access | 4.617708 → 4.594833 | 4.596541 → 4.583375 | 2.584000 | — |

For i31, Go improves **46.8%**, LLGo **46.9%**. Same-campaign middle source is
1.907041 / 1.892875 ns for Go/LLGo; direct cached-register selection is a further
28.5% / 28.7% below that middle. Baseline A/A is 2.553917 / 2.549083 ns
(-0.2% / +0.2%). Final LLGo is **25.5% faster than this Wasmtime**, 1.0% faster
than Go, and **4.9% above ordinary C**. These are one kernel's descriptive
ratios, not a controlled result or general native/competitor parity.

Preserve adverse controls: LLGo struct-access is +0.4% versus A/A +0.7%; Go
allocation is +0.1% versus A/A -1.7%. LLGo allocation improves 3.4% against
A/A -1.4%; this unchanged control is not a claimed i31 optimization benefit.
Struct access and indirect access still take about 1.26x / 1.77x Wasmtime.
Other series, including every middle/base control, remain in the evidence.

Clang's ordinary C implements the same finite wrapping multiply/add/mask at
matching inputs/counts, using an opaque volatile entry to prevent hoisting
between timed calls. Normal within-kernel optimization is allowed. It omits VM
cancellation, quotas, typed continuations and other runtime contracts. No C GC
allocation equivalent is invented. The two reference engines' unsupported GC
outcomes remain explicit rather than being replaced by another engine.

## Earlier source-changing campaigns

The first source `3f27587` uses 140 VM/five C observations and 28 VM/one C
alternate checks: i31 Go 2.556417 → 2.488625, LLGo 2.540708 → 2.489167;
Wasmtime 1.812000 and C 1.289000. Its Go allocation control regresses 2.6%
versus A/A -2.1%. The following campaign compares original, that source and
`1677123` together (180 VM/five C observations): i31 Go 2.560875 → 1.915000,
LLGo 2.550958 → 1.901000, Wasmtime 1.810000, C 1.287000. Its Go allocation
regresses 3.6% versus A/A -1.0%. All source stages, counts, A/A samples,
unsupported records and controls are retained without pooling percentages or
substituting a more favorable old measurement for the final campaign.

Final LLGo process readiness is about 90–97 ms across these kernels, versus
Go about 3.6–4.4 ms. The existing independently reproduced LLGo issues remain
separate from native guest instruction selection; no timing-only upstream issue
is filed. Earlier P1 boundary deficits, including roughly 3x LLGo/Go and 19x
enabled-wazero args-sizes, remain open at their actual measured revisions.
No interpreter-throughput, allocation-free, startup, cancellation-latency,
all-platform or full-portfolio acceptance is claimed from this native study.

[All three campaigns, exact archives/flags, C source, witnesses and validation](evidence/performance-i31-payload-2026-11-03.json.gz)
