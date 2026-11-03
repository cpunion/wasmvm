# LLGo named-result compiler migration: 2026-11-03

This contribution moves the host compiler pin from `8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f`
to `983465e779527ff496857a9a1915c1b57131091f`. The candidate includes
[merged LLGo #2963](https://github.com/xgo-dev/llgo/pull/2963), which fixes the
nonescaping named-result allocation reported in [#2959](https://github.com/xgo-dev/llgo/issues/2959).
The candidate is 79 commits ahead of the old pin, with that pin as merge base:
whole-engine differences cannot all be attributed to #2963. This is an explicit
toolchain baseline event, with no production VM implementation changes.

## Method and validation

Apple M4 Max/macOS arm64, Go 1.27.3, LLVM 22.1.8 and both LLGo versions with
`-O3 -lto=full`. All three engines use exactly the same VM revision,
`f2ec5947beeb20bfef2efb1ad5024f9c406e66fd`, with 1475 independently verified
engine files. Only shared measurement instrumentation and compiler provenance
are overlaid. The existing LLGo cancellation view remains unchanged. The
candidate executes all 31 semantic packages and all 60 expected conformance
reports locally, including the five required extensions on three tiers and
forced collection for Stack Switching. Contribution-head dual-platform CI is
still required before merge; these local results do not prove Linux acceptance.

Independent probes use 100,000 operations, five fresh serially interleaved
process blocks, two warmups and seven samples, with `GOMAXPROCS=1`. Engine
measurements use the same block/sample protocol and batches of 10,000 calls:
empty guest and 64-iteration integer guest, original and alternate-input
oracles, public/session/internal dispatch and raw native entry. There are 120
probe observations/24 summaries and 400 engine observations/80 summaries/80
alternate checks. Allocation sampling is separate: 72 probe and 54 engine
records, with all three rounds retained. Reported LLGo TotalAlloc is approximate;
zero Mallocs means unavailable instrumentation, not zero allocations.

All preparation, first call, warm execution, cleanup and process readiness
remain separately retained. No timing subtraction or cross-campaign pooling
is used. The shared workstation and five-process A/A controls do not establish
controlled performance acceptance. Initial probe metadata inspection and
missing semantic test-input setup failures are retained; corrected setup is
completed before timing. The successful semantics include committed benchmark,
testdata and script inputs, whose hashes are checked after execution.

## Independent host-compiler probes

Median ns/operation, lower is better:

| Probe | Go | LLGo old | LLGo candidate |
| --- | ---: | ---: | ---: |
| Nil atomic.Value.Load | 1.508 | 17.143 | 1.047 |
| Nonescaping named result | 1.805 | 10.813 | 1.570 |
| Live context.Err | 2.195 | 18.073 | 2.358 |
| Canceled context.Err | 8.887 | 96.252 | 67.478 |
| Existing live cancellation view | 3.922 | 7.855 | 7.873 |
| Existing canceled cancellation view | 13.062 | 107.222 | 75.014 |
| Explicit mutex lock/unlock | 2.816 | 7.999 | 7.874 |
| Deferred mutex unlock | 4.699 | 52.251 | 50.660 |

Third-round approximate bytes/operation fall from 31.968 to zero for the nil
atomic path, 15.993 to zero for the named result and 32.012 to zero for live
context.Err. Canceled context.Err remains about 112 bytes/operation; deferred
unlock remains about 96. The candidate therefore repairs the reproduced
named-result/nil-return case without resolving defer or canceled-context costs.
Existing [#2932](https://github.com/xgo-dev/llgo/issues/2932) and
[#2965](https://github.com/xgo-dev/llgo/issues/2965) remain distinct upstream
work. A later VM change may remove the obsolete live-view workaround, but no
unmeasured workaround-removal benefit is claimed here.

## Same-source VM and competitor comparison

Warm ns/call; integer means one invocation containing 64 guest iterations,
including its call boundary. It is not a sustained ns/iteration result.

| Configuration | Empty | Integer, 64 iterations |
| --- | ---: | ---: |
| Go native public | 357.513 | 434.938 |
| LLGo old native public | 1088.483 | 1205.413 |
| LLGo candidate native public | 1035.025 | 1126.942 |
| LLGo old native public, live parent | 1923.454 | 2008.596 |
| LLGo candidate native public, live parent | 1619.408 | 1750.738 |
| LLGo old retained native session | 574.608 | 676.788 |
| LLGo candidate retained native session | 590.054 | 673.488 |
| LLGo old interpreter public | 899.325 | 3946.238 |
| LLGo candidate interpreter public | 849.700 | 3909.292 |
| wazero compiler, default termination/background | 14.242 | 83.954 |
| wazero compiler, enabled termination/background | 349.971 | 1509.075 |
| wazero interpreter, default termination/background | 74.617 | 5787.188 |
| wazero interpreter, enabled termination/background | 464.467 | 6209.342 |
| Wasmtime Cranelift | 62.817 | 154.725 |
| Wasmer Cranelift | 140.108 | 226.763 |

References remain pinned wazero 1.12.0, Wasmtime 48.0.1 and Wasmer 7.3.0.
Default and enabled wazero termination policies are separate configurations;
none establishes an exactly equivalent VM contract. Both parent modes and
backends, every omitted internal component row and all phases remain in the
raw evidence. Public API numbers are not substituted by cheaper internal rows.

Candidate native public empty/integer costs decrease 4.9%/6.5%, and live-parent
costs decrease 15.8%/12.8%. Retain the 2.7% retained-session empty regression and
0.2% integer-dispatch regression in the interpreter. Background old-LLGo A/A
empty is 1122.492 versus 1088.483, and live-parent A/A is 1844.658 versus
1923.454, demonstrating shared-host variation. These are descriptive outcomes,
not a controlled significance or parity verdict.

Native public third-round allocation is Go 408.000, old LLGo 2096.075 and
candidate 1968.176 reported bytes/call. Retained session is respectively
64.041/816.869/816.446. Large compiler/runtime and VM-boundary gaps remain.
Native public process readiness is about 5.43 ms for Go, 96.54 ms for old LLGo
and 94.95 ms for candidate; this combined readiness phase is not pure runtime
startup. The independently reported [startup problem](https://github.com/xgo-dev/llgo/issues/2973)
is not resolved by this migration.

## Ordinary native boundary and remaining work

The same campaign measures a volatile indirect ordinary C identity call at
0.833 ns/call and raw native entry at Go 82.392, old LLGo 114.817 and candidate
112.813. These component rows lack complete public VM setup, quotas, roots,
cancellation and continuation contracts; differences are not subtracted to
invent isolated overhead. C has no integer workload in this experiment. The
ordinary-native portfolio gate remains open, as do sustained competitor,
startup, allocation/cancellation and expanded WASI-world acceptance.

The PR source-base/head benchmark uses the candidate compiler for both sources.
It checks migration integration and references, but is not the old/new compiler
comparison above. Historical benchmark identities and guest-producer versions
remain unchanged. The [previous final CI audit](pr109-final-ci-2026-11-03.md)
retains the old production baseline and its adverse hosted outcomes.

[Lossless evidence](evidence/performance-llgo-result-upgrade-2026-11-03.json.gz)
contains compiler/source identities, all measured attempts, summaries, allocation
rounds, semantic reports and independent source/oracle verification scripts.
