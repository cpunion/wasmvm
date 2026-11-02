# P1 context construction by operation effect: 2026-11-03

Status: **selected local candidate; final CI and controlled acceptance open**.
Source `ff97ead` follows the independently measured scalar ABI `115535b`.
The held entry-FFI source remains excluded. Source changes are confined to P1;
VM, interpreter/compiler and native adapter inputs are unchanged.

The complete P1 operation table marks its four configuration/guest-memory-only
operations: args_get, args_sizes_get, environ_get and environ_sizes_get. Their
bounded implementation reads immutable configuration and checked guest memory;
it never publishes a context to injected callbacks or host I/O. Keep the same
active-call admission, cancellation/result checks and shutdown drain, without
creating a private child binding. Every other import and socket admission keeps
its independent merged context and cancels it before releasing its active count.
No context is pooled or rebound. This is a provider effect classification, not
an engine import-name shortcut or a change to I/O cancellation contracts.

All 31 Go packages/thirty coverage gates, pinned full-LTO LLGo P1, P1 race and
vet pass. The scalar ABI's prior VM race/no-cgo checks remain separate. Added
scope tests check metadata, independent I/O retirement, process-close
interruption, active-call draining and closed-process rejection. Initial Go
compilation found the remaining socket begin call needed its explicit I/O mode;
that failure and its correction are retained. Successful checks precede timing.

## Independent comparison

Shared M4 Max/macOS arm64; Go 1.27.3, LLGo `8e45ed5`, `-O3 -lto=full`.
All four original 10000-iteration kernels, arguments/environment and independent
alternate oracles. Five fresh interleaved blocks, two warmups/seven samples,
both own compilers/backends/base/head, four exact base A/A aliases, four default
reference modes, two separate cancellation-enabled wazero modes and ordinary C:
360 VM/twenty C observations, 76 summaries and 72 VM/four C alternate checks.
Startup, preparation, first/warm execution and completed cleanup stay separate.

Warm ns/source iteration:

| Kernel | Go native base → head | LLGo interpreter base → head | LLGo native base → head | enabled wazero native | default Wasmtime | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| args-sizes | 433.804 → 270.337 | 1199.092 → 639.075 | 1427.954 → 834.837 | 39.492 | 54.100 | 0.900 |
| environ-sizes | 436.733 → 267.517 | 1213.833 → 629.771 | 1427.917 → 819.754 | 39.538 | 51.600 | 0.900 |
| args-copy | 475.837 → 298.058 | 1358.858 → 780.396 | 1580.904 → 985.983 | 45.129 | 66.300 | 0.800 |
| integer-control | 2.429 → 2.592 | 43.404 → 43.188 | 2.142 → 2.346 | 16.088 | 1.900 | 1.800 |

Go native P1 improves 37.3–38.7%, LLGo native 37.6–42.6%, LLGo interpreter
42.6–48.1%. Native P1 A/A moves Go +0.5%/+1.4%/-1.3% and LLGo
-1.3%/-0.2%/+1.3%. All original adverse arithmetic controls remain: Go native
+6.7% with A/A +2.4%; LLGo native +9.5% with A/A +4.9%; Go interpreter
+0.7% with A/A -0.7%. Unchanged guest compiler inputs do not waive those results.
The prior scalar-only study's public-host and arithmetic regressions also remain.
No percentages from those separate studies are summed or pooled.

Exact-archive diagnostics reuse only the exact 115535b head allocation records
with the same added test/fixture hashes. Third-call native args-sizes TotalAlloc
falls Go 4480824 → 2080808 and LLGo 33444192 → 18086048 bytes per original
10000 callbacks. LLGo interpreter falls 26882832 → 11522192. All rounds,
kernels and control fluctuations remain; these are approximate TotalAlloc
measurements, not Mallocs counts or production timings.

The three remaining gaps are distinct. Args-sizes LLGo native still takes 3.09x
Go, 21.14x cancellation-enabled wazero and 15.43x default Wasmtime. Reference
poll granularity and lifecycle policies remain different. Native C lacks VM
bounds, ownership, cancellation and quotas; its short samples are a coarse lower
bound. This source has no new sustained arithmetic acceptance result; earlier
source/configuration studies remain explicitly separate. Do not infer complete
native parity, low latency or a portfolio win from any one kernel.

The remaining native-only LLGo allocation is about 656 bytes per callback above
interpretation in these diagnostics. Existing public memory/capability helpers,
deferred cleanup and native snapshot/entry work are next profiling targets.
Known upstream defer/Cgo/parameter issues retain their independent reports;
this improvement introduces no new timing-only LLGo issue. Controlled
Linux amd64/macOS arm64 portfolio and expanded WASI acceptance remain open.

[All observations, A/A blocks, identities, allocation diagnostics and validation](evidence/performance-p1-call-effects-2026-11-03.json.gz)
