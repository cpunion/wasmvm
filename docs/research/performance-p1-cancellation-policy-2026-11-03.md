# Separate reference cancellation policies: 2026-11-03

Status: **descriptive configuration comparison; controlled acceptance open**.
The earlier P1 wazero reference uses default RuntimeConfig, with
WithCloseOnContextDone not enabled. Keep those results explicitly identified;
compare an additional build whose identical Go 1.27.3/wazero 1.12.0 provider
harness differs only by config.WithCloseOnContextDone(true). Native and
interpreter modes use that same additional binary. This is a configuration
event, not a reference version upgrade or a source-code speedup for wasmvm.

Use selected own source `6a13a48` (engine-equivalent to PR #104 head), excluding
the held FFI source. Four own compiler/backend modes, four exact own A/A aliases,
four default pinned reference modes and two enabled wazero modes; the same four
original 10000-iteration kernels, arguments, environment and checked alternate
oracles. Five fresh interleaved process blocks, two warmups/seven samples give
280 VM/twenty C observations, sixty summaries and 56 VM/four C alternate checks.
Startup, preparation, first/warm execution and cleanup stay separate. No earlier
study's throughput is pooled into these records. Short arithmetic own A/A moves
Go -19.4% and LLGo +1.1%; these original controls are not calibrated parity.

Warm ns/source iteration on shared M4 Max/macOS arm64:

| Kernel | Go native | LLGo native | wazero native default | wazero native cancellation enabled | Wasmtime default | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| args-sizes | 487.050 | 1507.958 | 24.217 | 40.692 | 56.900 | 0.800 |
| environ-sizes | 481.671 | 1537.408 | 25.042 | 40.308 | 54.000 | 0.900 |
| args-copy | 525.887 | 1766.125 | 31.050 | 47.254 | 66.700 | 1.000 |
| integer-control | 2.983 | 2.367 | 1.837 | 15.592 | 1.800 | 1.600 |

Wazero interpreter default → enabled is 93.867 → 101.950 for args,
95.300 → 100.408 for environment, 103.025 → 112.662 for copy and
83.100 → 84.092 for arithmetic. The native configuration adds about 15–16 ns
per original P1 iteration in this study; arithmetic changes much more
proportionally because its default body is small. These within-study
configuration differences are not subtracted from other engines to invent a
policy-free guest or callback time.

The [primary-source mechanism audit](performance-p1-boundary-profile-2026-11-03.md)
finds a Go termination trampoline at each enabled wazero loop header. Our native
poll protocol has a different granularity. Both are intended to support bounded
cancellation, but this finite-throughput comparison does not prove equivalent
latency/preemption bounds or full safety-policy equivalence. Native arithmetic
being faster than this enabled configuration does not establish a full-portfolio
win. The slower default-configuration comparison stays visible.

Three gaps remain distinct. Args-sizes LLGo native is 3.10x Go and 37.06x
cancellation-enabled wazero (62.27x its default), with other lifecycle-policy
differences still recorded. The absolute host-boundary deficit remains large;
cancellation configuration alone does not explain it. Ordinary C still omits
VM bounds, ownership, quotas and cancellation, and its short samples are coarse.
The original arithmetic control, every own A/A movement, Wasmer/default rows
and preparation/first-call costs remain in the evidence.

## Separate sustained arithmetic diagnostic

A fixed additional five-block experiment uses the same original arithmetic
bytes at one million iterations, unchanged selected own source and eight
native VM modes plus C. Two own A/A aliases, both wazero cancellation modes
and all three named native competitors are included: forty VM/five C
observations, eight VM/one C alternate checks. The original short controls
stay above; this diagnostic does not replace or pool them. Warm rates are
normalized by each record's actual count.

| Runtime | ns/source iteration |
| --- | ---: |
| Go native | 1.582500 |
| LLGo native | 1.563542 |
| Exact Go native A/A | 1.592708 |
| Exact LLGo native A/A | 1.565208 |
| Wazero default native | 1.692125 |
| Wazero cancellation-enabled native | 13.894500 |
| Wasmtime default | 1.561000 |
| Wasmer default | 1.561000 |
| Ordinary C | 1.560000 |

LLGo is about 0.23% above C on this one kernel and 1.2% faster than Go, while
its exact A/A is about 0.11% slower. Go A/A differs about 0.65%, so the
controlled calibration gate is not established. The enabled/default wazero
configuration ratio remains large at this sustained count. Different polling
granularity and unmeasured cancellation latency prevent declaring equivalent
policies or full-portfolio superiority. Native arithmetic is close to the
unchecked C lower bound here; P1 host-boundary work remains a separate large
deficit. No startup or policy median is subtracted. No new LLGo issue is filed
from this configuration-level timing difference.

[Complete records, A/A samples, patched harness/flags, input and binary identities](evidence/performance-p1-cancellation-policy-2026-11-03.json.gz)
