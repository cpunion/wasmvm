# Held normalized host binding: 2026-11-03

Status: **held; excluded from selected PR #105 and the i31 follow-up**.
Actual PR #105 head `5ceae09` is compared with `d2ee9e7`. Registration stores
one private record with value/scalar/suspendable callback fields, replacing
repeated function-value type assertions with one record-pointer assertion and
typed fields. Driver, fresh Caller/stack, references, errors, cancellation and
continuation contracts stay shared. The experiment does not establish enough
LLGo native benefit to adopt this extra representation now.

All 31 Go packages, thirty coverage gates, pinned full-LTO LLGo VM/P1/component,
affected VM/component race and vet pass. The supplemental no-Cgo component
command explicitly requests native backends: it fails the same eight top-level
tests at unchanged base and head. Both logs/commands are retained; no skip or
product patch hides them. Supported cgo component tests pass; no-Cgo VM passes.
This does not claim complete no-Cgo component-profile validation.

Two separately predeclared campaigns run serially after every check/build and
baseline profile diagnostic finishes. Shared M4 Max/macOS arm64, Go 1.27.3,
LLGo `8e45ed5` with -O3 -lto=full. Both compilers/backends/base/head, exact base
A/A, default and separately cancellation-enabled references and ordinary C.
Five fresh interleaved blocks/two warmups/seven samples, original bytes/counts/
oracles. Full P1 retains 360 VM/twenty C observations and 72 VM/four C alternate
checks; public-host control retains ninety VM/five C and eighteen VM/one C checks.
All phases and control fluctuations and unsupported policy differences stay.

Warm ns/source iteration:

| Kernel | Go native base → head | LLGo interpreter base → head | LLGo native base → head |
| --- | ---: | ---: | ---: |
| args-sizes | 267.104 → 259.067 | 577.254 → 552.158 | 765.300 → 756.521 |
| environ-sizes | 260.225 → 253.713 | 579.363 → 556.275 | 762.271 → 754.504 |
| args-copy | 339.021 → 320.921 | 848.358 → 810.188 | 1025.938 → 1040.621 |
| integer-control | 2.763 → 3.038 | 43.967 → 43.050 | 2.592 → 2.500 |

LLGo interpreter P1 improves about 4.0–4.5%, but native sizes gains of 1.0–1.1%
remain small beside A/A and copy regresses 1.4%. Go native integer regresses
10.0% versus A/A -0.9%; LLGo integer improves 3.5% while A/A improves 7.2%.
Original public-host control: Go native 260.167→256.217 (A/A 250.654), LLGo
interpreter 399.350→393.958 (A/A 407.821), LLGo native 703.067→705.938
(A/A 712.046). Its native LLGo path is +0.4%; this does not resolve the earlier
public-host control. Keep every original and A/A observation instead of
selecting only interpreter gains.

Head args still takes 2.92x Go, 19.49x enabled wazero (38.813 ns) and 13.46x
default Wasmtime (56.200 ns). C and other providers retain their original missing
VM contracts. Startup/preparation gains cannot erase warm regressions. The
[separate function-boxing reproduction](performance-llgo-function-boxing-2026-11-03.md)
finds a real LLGo conversion allocation, while repeated extraction of a preboxed
function allocates zero. Since VM callback interfaces were already constructed
at registration, that reproduction is not an allocation attribution for this
warm-dispatch experiment. No new assertion bug or throughput claim is filed.

[All phases, exact sources/binaries, unsupported-profile diagnostic and controls](evidence/performance-host-binding-held-2026-11-03.json.gz)
