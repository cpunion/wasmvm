# LLGo escape-analysis compiler baseline

Candidate compiler source: `3d92a58e09bd8971899c6eaf63437fbecd18d432`.
Previous production pin: `38de1e218b8146ba73f9c6f2e60e3c31f29d96b6`.
Go remains1.27.3, LLVM22.1.8, guest build flags `-O3 -lto=full`.
This compiler-source event changes neither production VM code nor reference
versions, safety policy, fixtures, sample counts or acceptance budgets.

The candidate includes merged upstream
[PR2997](https://github.com/xgo-dev/llgo/pull/2997), which addresses unified
escape decisions and invocation-local defer storage. The previous pin predates
that merge. Do not report its already-tracked defer allocation issue as a new
bug, or attribute all compiler changes to a single optimization. The candidate
also includes intervening runtime, KeepAlive, intrinsic and LLVM-binding fixes.

An isolated ordinary Go bootstrap on macOS arm64 records the exact clean source,
Go/LLVM identities, command, binary hash and VCS build information; no development
build tags. Production pin and historical measurements stay unchanged until
the final contribution passes its full semantic/profile acceptance checks.

## Qualification and allocation-only diagnostic

Engine source `2eae069a69c3639ab9dd4057c2735533c76f4083` is held constant for
the old/new compiler. The candidate passes the full VM package tests on both
backends. The same allocation-only probe runs three10000-call rounds plus
zero-call controls; all48 old/new records and source/probe/compiler identities
are retained. It measures warmed pure numeric calls after workspace reuse:

| Public interface | Previous LLGo collector bytes/call | Candidate LLGo |
| --- | ---: | ---: |
| Typed values | About736 | About240 |
| Reusable scalar stack | About672 | About144 |

The numbers are approximate BDWGC TotalAlloc deltas. Mallocs is unavailable;
these are neither allocation-event counts nor equivalent Go object-byte
accounting. New compiler allocations remain and need independent investigation.
No timing or full-engine speedup is inferred from these counters.

Final-head Linux amd64/macOS arm64 semantics, all coverage/CLI checks, affected
public-boundary and complete portfolio comparisons remain required. Preserve
old/new compiler and same-compiler Go controls, original22 workloads, independent
sustained GC/Threads cohorts, named reference policies and separate phases.
Do not relabel old samples or pool hosts. Hosted/local advisory timing cannot
close controlled performance or production-release acceptance.
