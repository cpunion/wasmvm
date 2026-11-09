# LLGo escape-analysis compiler baseline

Candidate compiler source: `3d92a58e09bd8971899c6eaf63437fbecd18d432`.
Previous production pin: `38de1e218b8146ba73f9c6f2e60e3c31f29d96b6`.
Go remains 1.27.3, LLVM 22.1.8, guest build flags `-O3 -lto=full`.
This compiler-source event changes neither production VM code nor reference
versions, safety policy, fixtures, sample counts or acceptance budgets.

The candidate includes merged upstream
[PR #2997](https://github.com/xgo-dev/llgo/pull/2997), which addresses unified
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
backends. The same allocation-only probe runs three 10000-call rounds plus
zero-call controls; all 48 old/new records and source/probe/compiler identities
are retained. It measures warmed pure numeric calls after workspace reuse:

| Public interface | Previous LLGo collector bytes/call | Candidate LLGo |
| --- | ---: | ---: |
| Typed values | About 736 | About 240 |
| Reusable scalar stack | About 672 | About 144 |

The numbers are approximate BDWGC TotalAlloc deltas. Mallocs is unavailable;
these are neither allocation-event counts nor equivalent Go object-byte
accounting. New compiler allocations remain and need independent investigation.
No timing or full-engine speedup is inferred from these counters.

## Hermetic P2 connection-failure fixture

The first full local candidate event retains 30 successful packages and a P2
consumer failure on both backends: the upstream request-timeout sample waited
for its outer 30-second context. The original harness also fails under stock Go
and the previous LLGo. An independent fixed-source socket diagnostic finds that
all three compilers successfully connect to `203.0.113.12:80` in under 1 ms on
this host. The guest's assumption that TEST-NET-3 cannot connect is invalid in
that network environment; this evidence does not establish a compiler defect.

Keep the upstream guest, hashes and assertions unchanged. For this named test
only, adapt its authority to a reserved, bound-but-not-listening loopback socket
and delegate to the ordinary WASI HTTP handler. The real socket connect then
produces the guest's explicitly permitted connection-refused alternative. The
reservation prevents another process from claiming the port; cleanup closes it
and checks that the guest reached the route. P3 already uses the same reserved
endpoint principle through its package-private transport hook. No production
transport, timeout policy or public API is changed.

The corrected P2 package executes all 184 consumers under the candidate, while
the targeted case passes Go race checks on both backends. Preserve the original
failed whole-project event and baseline controls; a corrected package result
does not relabel that event as a pass. Final exact-head CI remains required.

Final-head Linux amd64/macOS arm64 semantics, all coverage/CLI checks, affected
public-boundary and complete portfolio comparisons remain required. Preserve
old/new compiler and same-compiler Go controls, original22 workloads, independent
sustained GC/Threads cohorts, named reference policies and separate phases.
Do not relabel old samples or pool hosts. Hosted/local advisory timing cannot
close controlled performance or production-release acceptance.
