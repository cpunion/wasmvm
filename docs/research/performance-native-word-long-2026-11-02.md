# Longer native word-loop diagnostic: 2026-11-02

The longer follow-up again measures a scalar-memory improvement: Go native
**0.552→0.500**, LLGo native **0.551→0.492 ns/source iteration**, versus
Wasmtime **0.514**, Wasmer **0.557** and wazero compiler **0.873**. Every one of
six paired process blocks improves for both host compilers. Narrow/dependent
controls retain identical guest code and remain approximately flat.

This resolves the direction of the local candidate sufficiently to propose
cross-platform evaluation; it does not complete controlled performance
acceptance. The original [favorable twelve-kernel run](performance-native-word-loops-2026-11-02.md)
and [unfavorable short combination](performance-atomic-combined-2026-11-02.md)
remain evidence. All comparisons below use the same exact, previously built
candidate and base, rather than choosing a new source after inspecting outcomes.

Base `ffd6300e596735d2c51181897ae92499fcde347b`; candidate
`06196a9bba2e6619d400be96ac14b8b90053fd26`. Apple M4 Max/macOS arm64,
Go 1.27.3, LLGo `8e45ed5` `-O3 -lto=full`, Wasmtime 48.0.1/Wasmer 7.3.0
Cranelift and wazero 1.12.0 compiler. Fixture bytes are unchanged; every engine
receives **50,000,000 iterations** and the same runtime seed/checksum. Independent
Python oracles and alternate-input checks pass. Six fresh serially interleaved
process blocks, two warmups and seven warm calls use independent aliases of the
exact baseline Go/LLGo native binaries. Source, binary, library and fixture
hashes are checked. All preparation/first/warm/cleanup/readiness phases remain
separate. No subtraction is used.

| Workload | Configuration | ns/source iteration |
| --- | --- | ---: |
| core/memory-load-store | base-go-native | 0.552403 |
| core/memory-load-store | head-go-native | 0.499766 |
| core/memory-load-store | base-llgo-native | 0.550870 |
| core/memory-load-store | head-llgo-native | 0.491612 |
| core/memory-load-store | wasmtime-cranelift | 0.513520 |
| core/memory-load-store | wasmer-cranelift | 0.556990 |
| core/memory-load-store | wazero-compiler | 0.873445 |
| core/memory-load-store | aa-go-native | 0.556377 |
| core/memory-load-store | aa-llgo-native | 0.550957 |
| memory/narrow-unaligned | base-go-native | 0.672959 |
| memory/narrow-unaligned | head-go-native | 0.673198 |
| memory/narrow-unaligned | base-llgo-native | 0.672887 |
| memory/narrow-unaligned | head-llgo-native | 0.671059 |
| memory/narrow-unaligned | wasmtime-cranelift | 0.684860 |
| memory/narrow-unaligned | wasmer-cranelift | 0.704820 |
| memory/narrow-unaligned | wazero-compiler | 0.950139 |
| memory/narrow-unaligned | aa-go-native | 0.673371 |
| memory/narrow-unaligned | aa-llgo-native | 0.672746 |
| memory/dependent-address | base-go-native | 1.793660 |
| memory/dependent-address | head-go-native | 1.793903 |
| memory/dependent-address | base-llgo-native | 1.799155 |
| memory/dependent-address | head-llgo-native | 1.790253 |
| memory/dependent-address | wasmtime-cranelift | 2.022920 |
| memory/dependent-address | wasmer-cranelift | 2.042670 |
| memory/dependent-address | wazero-compiler | 2.927789 |
| memory/dependent-address | aa-go-native | 1.798423 |
| memory/dependent-address | aa-llgo-native | 1.780538 |

The shortest warm intervals are 22.36 ms (scalar memory), 33.04 ms (narrow)
and 84.71 ms (dependent). The old scalar C command accepts at most one million
iterations, so this run deliberately has no mismatched-count C row. Earlier C
comparisons stay in their own campaign; no C number is fabricated or rescaled
from a different count. The native memory/C gap therefore remains open.

Per-block head/base memory ratios are 0.880–0.922 (Go) and 0.872–0.927 (LLGo).
Native A/A memory ratios span 0.996–1.028 and 0.956–1.029 respectively. The
candidate improvement exceeds those observed ranges, but A/A does not resolve
the declared ±0.25% acceptance requirement. This is descriptive evidence, not
an adjusted threshold or statistical parity verdict. Host load changes
15.5→9.2 on 16 CPUs, without affinity or host isolation. The six blocks do not
replace the controlled portfolio or the planned native acceptance sample count.

The proposal uses the same bounded, proven private-memory transformation and
preserves original continuations, fuel, traps, code limits, wrap/tail behavior
and cancellation. Subword, invariant-address and dependent-address regions stay
compact. The final contribution still needs its own exact-head dual-platform
semantic checks and same-runner baseline/competitor data.

[Lossless evidence](evidence/performance-native-word-long-2026-11-02.json.gz)
retains all phases, raw calls, sample order, original and alternate checks,
independent A/A controls, complete summary and the measurement program. Build
identities are also retained in the earlier twelve-kernel evidence.
