# Interpreter CPU profiles and LLGo channel polling: 2026-11-02

The owned interpreter dispatcher dominates integer and scalar-memory execution.
LLGo cancellation polling also has measurable runtime overhead: a separate
minimal program measures an open-channel receive/default select at **6.964 ns**
versus Go **2.931 ns** in the longer local follow-up. The independently reproduced
runtime gap is reported as [LLGo #2965](https://github.com/xgo-dev/llgo/issues/2965).
It is distinct from the already reported context.Err allocation.

## Actual CPU profiles

The isolated profile source is archived from `ad66bde3e9dcd3a4b9092b901647edff1fc696b4`.
An instrumentation-only test enables the declared feature profile, constructs
the public API kernel outside the benchmark loop and repeatedly calls it with
a cancellable context. Each call checks its independent fixture checksum.
The four unchanged kernels cover integer mixing, scalar memory, atomic RMW and
GC allocation. CPU profiling includes benchmark calibration and process/runtime
activity; this is diagnostic evidence, not a throughput campaign or a matched
competitor ratio. Timings from these instrumented binaries are not substituted
for the normal phase harness.

Apple M4 Max/macOS arm64; Go 1.27.3 and pinned LLGo `8e45ed5` with
`-O3 -lto=full`. Both produce valid profiles that Go pprof can read.
Approximate sampled-CPU shares:

| Kernel | Go Run, flat | LLGo Run, flat | Go poll, cumulative | LLGo poll, cumulative |
| --- | ---: | ---: | ---: | ---: |
| Integer | 89% | 84% | 4% | 13% |
| Scalar memory | 84% | 81% | 3% | 9% |

LLGo TrySelect and pthread mutex operations appear under cancellation polling.
Go's scalar memory helpers occupy only a small part of its sampled total;
LLGo shows the same broad shape. Atomic profiles retain helper and FFI work;
GC allocation additionally exposes guest collection and host allocation/locking.
These whole-process profiles cannot assign every sampled runtime wait to one
VM call. LLGo includes host collector worker samples; GC sample CPU can exceed
wall duration. Source-line attribution is strongest for Go and must not be
mistaken for precise cycle accounting.

The Go integer profile attributes samples to operand selection and the separate
stack-switching/wide/GC classification checks executed before the main opcode
switch. This motivates moving those classifications into the existing dispatch
structure. It is a hypothesis for a smaller hot loop; independent paired
throughput and full semantic tests must determine adoption. The held unsafe
operand-access/local-counter experiments are not revived by this profile.

## Independent select reproduction

The complete authored main.go, commands, binary hashes, source hash and every
sample are retained. It polls nil, open, empty-buffered and closed channels;
every call checks the ready count. The first busy four-case campaign uses ten
million iterations and remains retained. A separate focused run uses the
cancellation-shaped open unbuffered channel, fifty million iterations, three
fresh processes/compiler, two warmups and seven warm calls. Process compiler
order alternates; no local build/test campaign runs concurrently in this follow-up.

| Process block | Go ns/select | LLGo ns/select |
| --- | ---: | ---: |
| 0 | 2.931 | 6.964 |
| 1 | 3.006 | 6.928 |
| 2 | 2.884 | 9.573 |

The shortest warm sample is 141 ms. Host load remains 36.9→31.0 on 16 CPUs;
there is no CPU affinity or controlled-host parity claim. The descriptive ratio
of process medians is about 2.38x, not a universal factor. No completed run or
unfavorable sample is removed.

Pinned TrySelect and ChanTryRecv take a channel mutex before trying an empty
receive. The Go runtime has a synchronized unsuccessful nonblocking-receive fast
path. Current LLGo source at `08e759300a2d523d5a37de0b3b31a78b20bd2843` still
contains this TrySelect path (blob `6a547eba58cf4d638d3b0cfdf4d21f0cf0fa26c2`);
current main was inspected but not rebuilt/timed. The upstream issue proposes
a synchronized fast path while retaining close/readiness/waiter/select/synctest
semantics. No unsynchronized cancellation flag or reduced polling is introduced
in wasmvm to claim a win.

[Lossless evidence](evidence/performance-interpreter-profile-2026-11-02.json.gz)
contains the diagnostic source, raw profiles (hex-encoded), pprof outputs,
identities, minimal reproduction, all samples and the exact submitted issue.
