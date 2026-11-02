# Long native atomic/C comparison: 2026-11-02

LLGo native RMW measures **1.776 ns/source iteration**, ordinary C **1.723**,
Wasmtime **1.709**, wazero compiler **1.780** and Wasmer **3.649** on Apple M4
Max/macOS arm64. LLGo is about 3.1% behind C and 3.9% behind Wasmtime here;
Go native is **1.772**. The two host compilers produce essentially equal native
guest throughput in this sample. This one-case result does not complete native
or cross-platform acceptance.

The new [standalone C11 reference](../../benchmarks/native/c/atomic_baseline.c)
uses sequentially consistent initialization, fetch-add and final load. Its count
and seed are runtime inputs; every call checks the result. A volatile indirect
kernel entry keeps calls inside their clock intervals. Optimized ARM64 assembly
contains the sequentially consistent store, `ldaddal` loop and load; the loop is
not replaced by a final arithmetic result. The reference has no guest bounds,
fuel or cancellation contract. The older six-case scalar reference is unchanged.

Clang `-std=c11 -D_POSIX_C_SOURCE=200809L -O3 -mcpu=native -Wall -Wextra -Werror`
builds the release reference. ASan/UBSan and release binaries pass 24 checked
seed/count cases plus six invalid-count/wrong-checksum cases. These short cases
validate semantics only; zero clock readings there are not performance evidence.

The long probe performs **100,000,000 iterations per call**, three fresh
serially interleaved process blocks, two warmups and seven warm calls. The
minimum observed warm interval is **165.395 ms**. Independent checksum and
alternate-input checks pass. Source, binary, compiler, SDK and fixture identities
are retained. LLGo uses pinned `8e45ed5` with `-O3 -lto=full`; Go is 1.27.3.
References are Wasmtime 48.0.1, Wasmer 7.3.0 and wazero 1.12.0 in their named
native compiler configurations. Host load changes about 6.4→5.3 on 16 CPUs.

| Configuration | ns/source iteration |
| --- | ---: |
| base-go-native | 1.779066 |
| head-go-native | 1.772194 |
| base-llgo-native | 1.771629 |
| head-llgo-native | 1.776073 |
| wasmtime-cranelift | 1.709310 |
| wasmer-cranelift | 3.648640 |
| wazero-compiler | 1.779978 |
| aa-go-native | 1.775661 |
| aa-llgo-native | 1.772814 |
| native-c | 1.723280 |

Base is `72d74d1`; head is the separately held frame-access experiment `4aa4ce7`.
That experiment does not alter generated native RMW code. The table retains both
revisions and independent aliases of the exact baseline Go/LLGo native binaries;
it does not attribute their small differences to an interpreter-only change.

Loading, setup, compilation/validation, instantiation, export lookup, first call,
warm calls, cleanup and process readiness remain separate in the raw data.
There is no startup subtraction. C has no VM phases, so those entries are absent.
Long dwell reduces clock granularity concerns but does not establish a calibrated
controlled-host threshold. The much larger interpreter RMW gap remains open;
[the helper allocation work](performance-interpreter-atomics-2026-11-02.md)
addresses a different part of that gap.

[Lossless evidence](evidence/performance-native-atomic-c-2026-11-02.json.gz)
contains complete samples, A/A controls, all phases, source/build/library hashes,
verification and measurement programs, assembly and sanitizer outcomes.
