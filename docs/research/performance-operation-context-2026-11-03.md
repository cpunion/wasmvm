# Shared operation cancellation view: 2026-11-03

Status: **selected local candidate; final contribution CI and performance
acceptance remain open**. Exact source `2bc8ab7` follows `f384fb7`; the intervening
CI-only rebalance has identical engine inputs. The held host-request-buffer
source is excluded. Its negative report is retained separately rather than
cherry-picking a favorable timing from that implementation.

The execution engine already caches an immutable cancellation view for pinned
LLGo because its live context.Err path allocates (#2959, fixed upstream after
the pin). Share that same helper with the Store operation driver, which polls
at every host/capability boundary. Go keeps its existing cheaper standard path.
One immutable view belongs to each operation; retained contexts preserve their
values, deadline, cause and original cancellation binding. No mutable context
pool or weaker poll interval is introduced. Child cancellation, expired Callers,
reentry, fuel and cleanup retain their contracts.

All 31 Go packages/thirty coverage gates, affected LLGo execute/VM/P1, race,
no-cgo core scope and vet pass. Public Caller metadata/independent expiry and
existing cancellation/cause/rebinding tests execute under both compilers.
Compiler, native adapter and nativeexec source inventories are unchanged.

## Independent scoped comparison

Shared M4 Max/macOS arm64, Go 1.27.3 and pinned LLGo `8e45ed5`, `-O3 -lto=full`.
The four original P1/control binaries keep 10000 iterations, guest/arg argv and
KEY=VALUE environment. Five fresh serially interleaved process blocks, two
warmups/seven calls, both own compilers/backends/base/head, four exact baseline
A/A aliases, four pinned named competitors and matching C: 320 VM/twenty C
observations, 68 summaries and 64 VM/four C alternate checks. Input hashes,
preparation, startup, first/warm calls and cleanup are retained separately.

Warm ns/source iteration:

| Kernel | Go native base → head | LLGo interpreter base → head | LLGo native base → head | wazero native | Wasmtime | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| args-sizes | 481.025 → 483.892 | 1415.025 → 1342.425 | 1731.467 → 1655.283 | 24.017 | 55.400 | 0.800 |
| environ-sizes | 485.704 → 490.238 | 1422.513 → 1358.412 | 1734.317 → 1644.608 | 24.292 | 53.000 | 0.800 |
| args-copy | 520.700 → 519.983 | 1647.521 → 1503.404 | 1908.058 → 1823.179 | 30.742 | 68.400 | 0.800 |
| integer-control | 2.538 → 2.758 | 44.550 → 43.379 | 2.487 → 2.654 | 1.700 | 1.800 | 1.600 |

LLGo native P1 improves 4.4%/5.2%/4.4%, versus A/A +0.8%/-0.1%/-0.4%.
Its interpreter improves 5.1%/4.5%/8.7%, versus A/A +0.1%/+0.7%/-0.2%.
Go native P1 changes +0.6%/+0.9%/-0.1%. Every adverse control stays visible:
Go interpreter/native integer regress 3.6%/8.7%, versus A/A +3.0%/+3.9%;
LLGo native integer regresses 6.7%, versus A/A +9.2%. Those short original
controls are not replaced or dismissed as harmless. This is descriptive local
evidence, not the controlled portfolio gate.

Exact-archive test diagnostics reuse only the independently verified original
baseline allocation records (same added source and fixture hashes), and newly
measure this head. Third-call LLGo native args-sizes TotalAlloc decreases
44635552 → 39531472 bytes per original 10000 calls; interpreter decreases
36481568 → 31681520. Go remains about 5.28 MB. Cold and warm rounds for all
kernels/backends remain distinct. No Mallocs count or production timing is
inferred from test executables.

The three gaps remain explicit: LLGo args-sizes is still about 29.9x Wasmtime,
68.9x wazero native and 3.42x Go. C's 0.8 ns control represents the same finite
algorithm and opaque direct provider calls across separate translation units,
with no VM bounds/lifetime/cancellation/admission. Its short samples are a
coarse lower-bound diagnostic, not calibrated VM parity. No control median is
subtracted. Native-only LLGo allocation still exceeds interpretation by about
785 bytes per callback in the diagnostic; that narrows further investigation
without assigning the complete difference to one function.

[#2959](https://github.com/xgo-dev/llgo/issues/2959) is closed by #2963, while
these results retain the old fixed pin. The new independent input-parameter
reproduction is recorded separately; no compiler upgrade or unmeasured patch
speedup is claimed. Host/provider boundaries and controlled all-target Core,
component and WASI performance acceptance remain open.

[All phases, A/A blocks, source/binary identities, diagnostics and validation](evidence/performance-operation-context-2026-11-03.json.gz)
