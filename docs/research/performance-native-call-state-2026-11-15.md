# Native call identity, scalar results and cached stack top

Unmetered linked bodies now know their module function identity, carry the low
word of a single integer/reference result into the caller's permanent register,
and retain private call-stack Top in a third dedicated callee-saved register.
Normal exits share one state-publication tail. Complete canonical 128-bit values,
zeroed child frames, roots, physical/logical quotas, cancellation, precise fuel
and public record layouts remain intact. Metered, unlinked, floating/vector and
multiple-result paths retain their existing contracts. See the
[ABI contract](../design/native-abi.md#checked-internal-call-entries).

## Frozen sources and measurement

Integrated base is **51778a62**, final implementation **9de037fb**. Reused base
binaries identify **0baaf03e**; Git verifies all Go/module sources are identical
to integrated base. Seven-file source snapshots, overlays, hashes and original
identities remain explicit. Intermediate sources **33da30b9** and **7b79ae65**
are separate events, never substituted for final results. Go 1.27.3, LLGo
6a482f7f, LLVM 22.1.8, `-O3 -lto=full` and default production GC are fixed.
References are wazero 1.12.0, Wasmtime 48.0.1 and Wasmer 7.3.0.

Host: M4 Max/macOS arm64. Other user tests, compiler/link and desktop activity
were observed. No owned build, test, profile, source edit, analysis or pruning
runs during timed cohorts. All load snapshots and same-binary A/A aliases remain;
absolute costs across cohorts are not directly comparable. Intervals are
individual block-bootstrap 95% intervals; controls resolve only when their
whole interval lies within 0.975–1.025. These are scoped advisory measurements,
not portfolio or release acceptance.

Final general-call comparison: twelve fixed blocks/192 processes, Go/LLGo
base/head, three references and unchecked LLVM C, each with an alias. Preparation,
compilation, instantiation, first call, two warmups and seven interleaved
five-million-iteration calls remain separate. Every output is checked. C performs
2N affine steps normalized per guest iteration, without VM safety contracts.
A separate ten-block/40-process counter audit measures twenty interleaved calls
of five million iterations plus a separate flush. Cohorts are not extended.

## Final general-call results

| Go head/base counter audit | Ratio [95% interval] |
| --- | ---: |
| CPU | 0.97419 [0.97142, 0.97728] |
| Instructions | 0.88957 [0.88952, 0.88963] |
| Cycles | 0.97351 [0.97152, 0.97579] |
| Cycles/instruction | 1.09436 [1.09218, 1.09690] |

Both own controls resolve for all four metrics. CPU falls **2.58%**, instructions
**11.04%** (about 190 to 169 per iteration). Higher CPI absorbs much of the
instruction reduction; instruction count alone does not predict elapsed speed.
The separate complete comparison gives Go warm head/base **0.98179
[0.97781, 0.98541]**, LLGo **0.97455 [0.96859, 0.98177]**, and LLGo/Go
**1.00036 [0.99581, 1.00609]**. All warm controls resolve.

| Final Go/reference warm time | Ratio [95% interval] |
| --- | ---: |
| wazero | 2.69742 [2.68099, 2.71801] |
| Wasmtime | 2.89428 [2.88444, 2.90366] |
| Wasmer | 2.88788 [2.88049, 2.89548] |
| Unchecked LLVM C | 2.89577 [2.88718, 2.90368] |

Median warm costs are about 6.20 ns for Go/LLGo, 2.30 ns for wazero and 2.14 ns
for the other references. These references are independent, with distinct safety
contracts. Competitive parity is not achieved. The preceding
[Wasmtime disassembly audit](performance-native-frame-registers-2026-11-15.md)
shows genuine register-argument BL/RET calls, not callee inlining. Remaining
explicit frame/metadata copies, entry preparation and register allocation are
compiler work; host GC cannot explain this allocation-free native kernel's gap.

## Complete P3 requests and LLGo adverse results

Each payload has eight fixed blocks/80 processes: Go/LLGo base/head and Wasmtime,
each with an alias. Four Stores/sixteen clients run twenty individually
interleaved 512-request batches plus a flush, with snapshot input and buffered
output. All bodies, statuses, headers, trailers and CRCs are checked. Cold setup,
first call, warmups, measured work and cleanup stay separate.

| Go head/base | 16 B | 64 KiB |
| --- | ---: | ---: |
| CPU | 0.99208 [0.97218, 1.01016] | 0.98864 [0.97806, 0.99797] |
| Instructions | 0.99475 [0.99378, 0.99573] | 0.99514 [0.99400, 0.99620] |
| Throughput | 1.00768 [1.00056, 1.01388] | 1.00483 [0.99400, 1.01678] |

Both 64 KiB Go CPU controls resolve; the small CPU reduction is supported in this
cohort. The 16 B head CPU/throughput controls are unresolved, so no resolved gain
is claimed there. Go/Wasmtime CPU points are **1.923x/2.089x**, with unresolved
reference CPU controls; throughput points are **0.837x/0.712x**, with reference
throughput controls resolving. Go allocation counts remain about **696/756** per
request. Foreign Rust/C allocations are not counted by Go's host counters.

LLGo normal-GC CPU changes **1.02305 [1.00124, 1.04587]** for 16 B and
**0.97796 [0.91905, 1.03464]** for 64 KiB. Both own CPU controls are unresolved
for both payloads. The adverse 16 B result is retained, not labeled a pass.
Median measured collections per 10,240 requests increase **29.5 to 39** for
16 B (ranges 22–39 and 27–39); initial heap medians are 32.36/32.78 MB. For
64 KiB both medians are 61, with ranges 35–111 and 59–111; initial heap medians
44.62/31.51 MB. Managed allocation counts remain **1165/1430**. These observations
implicate collection variability but do not isolate its root cause.

Separate bounded no-GC diagnostics use eight blocks/32 processes per payload,
five measured 512-request batches plus a flush, and require zero collections.
LLGo CPU ratios are **0.98821 [0.97325, 1.00045]** and **0.99747
[0.99002, 1.00513]**, with resolving CPU controls. Instruction ratios are
**0.99613/0.99637**, with resolving controls. These do not establish a CPU gain,
but show no resolved no-GC regression. Production GC remains unchanged. Cold
heap sizing versus conservative retention remains open in
[LLGo #3106](https://github.com/xgo-dev/llgo/issues/3106); this event does not
supply new compiler evidence sufficient to assign an additional LLGo bug.

## Image growth, shared exits and rejected experiments

The unshared intermediate 7b79ae65 increased a successfully checked P3 request's
two native module images from **7,762,468 to 8,397,716 bytes (+8.18%)**. Repeated
exit publication was identified and replaced with a common normal-exit tail.
Final image size is **7,586,656 bytes: -2.27% versus base, -9.66% versus the
unshared intermediate**. The diagnostic overlay logs image sizes; it is untimed
and supplies no speed claim. Malformed-entry BadState retains separate guarded
publication and the caller-supplied diagnostic function identity.

A separate fixed ten-block/60-process audit directly interleaves base, unshared
7b79ae65 and final shared 9de037fb, all with aliases and twenty five-million-call
samples. Shared/unshared CPU is **0.99955 [0.99731, 1.00167]**, instructions
**1.00007 [0.99997, 1.00016]**; all own CPU/instruction/cycle controls resolve.
Shared/base CPU is **0.97081 [0.96877, 0.97300]**. This does not support a hot
regression caused by sharing. The earlier unshared cohort's larger 7.26% CPU
reduction occurred with very different absolute costs/load and is not the final
source's gain. Exact hardware/load contributions remain unisolated.

Final cold Go P3 compile-wall ratios are **1.01590/0.99871**; LLGo ratios
**0.99035/0.98334**. Cold controls remain unresolved, and all intervals cross
parity. Image growth is corrected; no cold speedup or exact cold regression
magnitude is established. The intermediate adverse 16 B cold point (+9.71%)
and all its controls remain in the evidence.

All prior experiments remain distinct in the machine-readable record:

- Relative result-destination variants remove instructions but provide no CPU
  benefit, with roughly 8% higher CPI. They were discarded. An initial incorrect
  frame-base restoration failed semantics and was corrected before timing.
- Function identity/result carrier alone has a 1.58% CPU reduction in its own
  audit. Its adverse no-GC 64 KiB CPU point (+2.75%, interval crosses parity,
  unresolved head control) remains causally inconclusive, not deleted.
- Fresh private PC-zero entry and first-argument carrier pass affected semantics
  but show no resolved incremental CPU gain. They were reverted. The first
  diagnostic run omitted the interactive acknowledgement; its failure is a
  harness protocol error, not an engine failure.
- Cached Top has a favorable separate prototype audit. Final source comparisons
  supersede its larger prototype/intermediate headline gains. Existing MADD
  combination was already implemented; it is not attributed to this change.

## Validation and remaining work

All **32 Go root test packages** pass; all **31 measured executable packages**
reach at least **90% statement coverage**. LLGo executes affected codegen and
nativeexec tests with the production profile. Added tests cover permanent and
borrowed register homes, complete low/high result words and large offsets,
malformed headers with an invalid metadata pointer, function identity/import
prefixes, optional compilation and serial/parallel byte equality. Existing full
proposal, roots, quota, reentry, cancellation, cleanup and official WASI tests
stay green. Complete final-head Go/LLGo × Linux amd64/macOS arm64 CI is required
before merge; local runs do not replace that matrix.

The [machine-readable report](data/performance-native-call-state-2026-11-15.json)
links **1,748 process records in 25 separate cohorts**, source identities,
protocols, load snapshots, all intermediate/rejected source overlays, raw phases,
counters and collection observations. Only unused per-request timing vectors
become count/extrema/canonical-JSON hashes in the committed projection; full
originals remain local. Remaining work includes reducing general call ABI work
and register spills, complete component allocation/dispatch overhead, LLGo
collection stability, cross-platform performance and full release acceptance.
