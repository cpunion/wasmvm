# General scalar host ABI: 2026-11-03

Status: **implementation candidate; final CI and controlled performance
acceptance open**. Source `115535b` follows selected `6a13a48`. Subsequent
operation-effect changes and the held direct-FFI source are excluded here.
[Design and lifetime contracts](../design/scalar-host-abi.md) extend the
existing host driver with independent numeric callback stacks. Public reference
callbacks retain their encoding/lifetimes. All 46 P1 imports share the scalar
interface, while Process admission and private context construction remain
unchanged in this isolated study.

All 31 Go packages/thirty coverage gates pass; pinned full-LTO LLGo VM/P1,
race, no-cgo VM and vet pass. Tests execute mixed scalar types through actual
guests, exact I32/F32 bit widths and NaNs, retained stack/result isolation,
nested calls, temporary roots with collection, errors/panics, cancellation,
continuations and structured exception identity. References/vectors/invalid
signatures are rejected at registration. Compiler/interpreter source inputs
are unchanged; no guest byte identity or calibrated parity is inferred.

## Original P1 comparison

Shared M4 Max/macOS arm64; Go 1.27.3 and LLGo `8e45ed5`, `-O3 -lto=full`.
Five fresh interleaved blocks, two warmups/seven samples, unchanged four original
10000-iteration kernels and independent oracles. Both own compilers/backends,
four exact baseline A/A aliases, four default reference modes and two separately
identified cancellation-enabled wazero modes, plus matching ordinary C: 360 VM,
twenty C observations, 76 summaries and 72 VM/four C alternate checks.
All semantic and quality checks finish before timed processes; allocation
diagnostics run separately.
Startup, preparation, first/warm calls and completed cleanup remain separate.

Warm ns/source iteration:

| Kernel | Go native base → head | LLGo interpreter base → head | LLGo native base → head | cancellation-enabled wazero native | C |
| --- | ---: | ---: | ---: | ---: | ---: |
| args-sizes | 478.308 → 445.842 | 1278.533 → 1218.446 | 1494.850 → 1439.333 | 39.454 | 0.800 |
| environ-sizes | 470.783 → 438.933 | 1277.563 → 1192.867 | 1490.067 → 1416.471 | 39.179 | 1.000 |
| args-copy | 509.063 → 476.842 | 1420.946 → 1356.838 | 1649.292 → 1610.408 | 44.946 | 0.900 |
| integer-control | 2.508 → 3.058 | 44.279 → 44.196 | 2.525 → 2.467 | 16.158 | 1.600 |

Go native P1 improves 6.3–6.8%; LLGo native 2.4–4.9%, interpreter 4.5–6.6%.
Corresponding native A/A changes are Go +0.2%/+1.2%/+0.1% and LLGo
+0.1%/+0.1%/-0.7%. Original Go native arithmetic regresses 21.9% (A/A -0.2%);
LLGo native improves 2.3% while A/A improves 7.1%. These short controls stay
visible and are not waived by unchanged compiler sources or other wins.

Exact-archive allocation diagnostics reuse only exact 6a13a48 baseline records
with identical added-test/fixture hashes. Third-call native args-sizes TotalAlloc
falls Go 5280808 → 4480824 and LLGo 35036624 → 33444192 bytes per 10000
callbacks. All rounds/backends remain retained; no Mallocs count or production
timing is inferred from test executables. Args-sizes still takes LLGo 3.23x Go,
36.48x enabled wazero and 26.41x default Wasmtime in this campaign. C lacks VM
policies and its short samples are a coarse lower bound.

## Existing public-host regression control

Before inspecting primary candidate scores, a separate predeclared control
builds both exact sources with the same minimal public HostFunction provider:
two checked memory writes and a numeric result, without P1 Process admission or
private contexts. It uses original args-sizes bytes/oracles, all eighteen VM
modes, five fresh blocks and matching C. Ninety VM/five C observations and every
phase are retained. Other engines keep their builtin P1 providers; their
lifecycle policies differ. This is an existing-API regression control, not a
replacement P1 implementation, production core patch or equal-policy claim.

| Existing public callback path | base → head | exact baseline A/A |
| --- | ---: | ---: |
| Go interpreter | 221.950 → 221.225 | 219.550 |
| Go native | 243.721 → 248.129 | 244.046 |
| LLGo interpreter | 387.517 → 394.313 | 390.592 |
| LLGo native | 693.592 → 705.146 | 700.212 |

The legacy path regresses Go native 1.8%, LLGo interpreter 1.8% and native
1.7%, with respective A/A +0.1%/+0.8%/+1.0%. Keep those deficits and the
unresolved dispatch/type-assertion hypothesis. The scalar gain does not erase
them. The remaining P1 context/admission cost motivates a separately measured
operation-effect follow-up; it is not part of these timings. Controlled
cross-platform portfolio and latency/allocation acceptance remain open.

[Every phase, oracle, A/A block, source/binary identity and diagnostic](evidence/performance-scalar-host-abi-2026-11-03.json.gz)
