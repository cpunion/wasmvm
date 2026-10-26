# Native integer architecture gate

This probe runs the existing `integer-mix.wasm` directly through the project
compiler and C entry adapter. It is not the public VM API: memory integration,
guest-call dispatch and runtime helpers remain incomplete. Scalar numeric and
private-memory development scopes are recorded in [float-plan.md](float-plan.md)
and [memory-plan.md](memory-plan.md). No unsupported
operation silently falls back. The one kernel measures the scalar register and
interrupt-polling changes; it cannot accept the complete native backend.

Build with `go build -o /tmp/native-probe ./benchmarks/native` or the pinned LLGo
compiler. Send `run\n` on stdin and pass
`--wasm benchmarks/embedded/fixtures/integer-mix.wasm`. `--tier unmetered` keeps
bounded interrupt checks without periodic host round trips. `--tier sliced
--budget 64` selects the original precise sliced compiler with semantic fuel
disabled. The probe records frontend, native emission/publication, context setup,
first call, seven warm calls after two extra warmups, code size and host poll
exits. Checksums, initial frame setup and printing are outside the call timer.

The existing PR phase job reuses its isolated exact-revision source trees and
pinned wazero 1.12.0, Wasmtime 48.0.1 and Wasmer 7.3.0 compiler runners. It adds
20 interleaved process blocks (one predeclared extension to 60 when inconclusive)
for 25 million iterations of the existing sustained kernel
(the identical binary and seed for all candidates and A/A controls), plus three
own-baseline blocks retaining five million iterations and publishes raw records and the
feature comparison in the same PR comment. Both host compilers must pass every
competitor comparison and their identical-binary A/A control. Linux processes
share one allowed CPU through `taskset`; macOS affinity remains uncontrolled. A paired process bootstrap uses simultaneous 95% bounds;
intervals account for both planned looks. A practical parity verdict requires the whole interval within ±0.5%, and A/A
controls must resolve ±0.25%. Slower or inconclusive results fail the feature gate. Shared-host results remain
provisional and do not replace controlled full-portfolio acceptance. The same narrow gate runs on a separate macOS arm64 CI host, reusing the
common measurement code and two pinned native SDKs. It does not repeat the
semantic matrix, interpreter portfolio or production CLI series. Both target
reports and raw artifacts are linked from the same PR comment. Future feature PRs must declare and compare their affected workload
families, not reuse this kernel to excuse an unrelated deficit.

For the bootstrap baseline predating `CompileUnmeteredFunction`, only this
probe's compiler-function reference is adapted to `CompileFunction`; its recorded
tier is `sliced`. Archived engine sources are checked unchanged and both probe
and executable hashes are retained. Later baselines use their own unmetered
compiler. Own baseline ratios remain visible; only competitor ratios decide
this feature gate, so an unchanged baseline does not fail merely due to A/A
noise. Preparation and first-call costs remain separately visible.

## Measurement revision

The initial 250,000-iteration data remain on PR #26 (heads `777cd7c` and
`9501dcd`). The first head was measurably slower; after direct register selection
the small-kernel comparisons remained blocked. Each call lasted only about
0.4–0.6 ms, with roughly 1–2 ms of first-call/warmup activity, and no affinity or
A/A calibration. That did not establish steady-state parity. The revised gate
uses the already pinned sustained fixture, the same seven samples/two warmups,
fixed CPU placement on Linux and independent A/A process controls. Margins are
fixed before this new measurement, not derived from observed competitor gaps.
An interval merely crossing 1 is never a parity pass. This is a disclosed
measurement-method change, not an instruction-throughput win. Full controlled
portfolio acceptance remains unchanged.

At `dd4fba8`, the separate macOS arm64 runner measured 2.339/2.381 ns per
iteration for Go/LLGo, against 2.362 wazero, 1.914 Wasmtime and 1.925 Wasmer.
The competitor gap and inconclusive A/A controls blocked merging. The next
compiler change moves polling exits out of line and acquires cancellation only
after a nonzero atomic observation; it does not change the workload or gate.

At `a48564a`, arm64 remained slower than Cranelift (2.093/2.107 ns per
iteration versus 1.710/1.735), with inconclusive A/A controls. Moving cancellation
exits alone did not close the gap. Immediate shift/rotate selection is evaluated
next with the same workload, continuations, cancellation policy and thresholds.

At `69c8592`, final-head Linux intervals were inconclusive and arm64 remained
slower. The next implementation adds general integer multiply/add combination
with preserved intermediate slots and independent continuations. Local M4 Max
diagnostics were about 1.54 ns/iteration across both host compilers and all three
references; this is not a replacement for the declared two-target feature gate.

The gate originally had a classification gap: an interval such as 0.993–0.998
was neither wholly inside the ±0.5% equivalence band nor wholly below 0.995,
and was incorrectly rejected despite containing only faster/equivalent outcomes.
The classifier now accepts that union when the upper bound is at most 1.005,
labelled `parity-or-faster`. The parity margin, A/A requirement, observations,
bootstrap procedure and planned looks are unchanged. Reclassifying the retained
`d123e16` Linux observations passes all eight comparisons; this does not accept
arm64 results or replace final-head checks.

At `fe3b4db`, Linux passed but macOS arm64 remained blocked by noisy A/A controls
and a slower comparison. A complete local M4 Max run also failed A/A resolution
at five million iterations. The next protocol increases every candidate and
control call to 25 million iterations, keeping the same binary, seed, seven
samples, warmups, planned looks and margins. The independent Python oracle gives
checksum `3139208734`. Descriptive baseline calls retain five million iterations
to bound CI cost; their count is disclosed and each record is normalized by its
own iterations. This improves measurement duration, not generated-code speed;
earlier failures remain evidence and cannot be replaced by selective reruns.

At `b2d16af`, Linux passed at about 1.87 ns/iteration; arm64 remained blocked
(Go/LLGo 2.130/2.177 versus Wasmer 2.032), with unresolved A/A controls even
with longer calls. The next code change branches directly on allocated i32
condition registers, eliminating scratch copies and, on arm64, the separate
zero comparison. Cancellation, long-range branches and low-32-bit semantics
remain intact. The measurement protocol is unchanged.

At `381a7ee`, all four semantic jobs passed and Linux resolved parity against
all three references (Go/LLGo 1.638/1.636 ns/iteration). The capacity-canceled
macOS job was rerun once; completed measurements remained blocked: Go/LLGo
2.195/2.169 versus Wasmtime/Wasmer 1.986/2.005, with unresolved A/A calibration.
The failed result is retained in run 40363412198.

The next implementation removes redundant control-flow jumps, relaxes near
arm64 cold branches without moving continuation offsets, and forwards local
copies and integer multiply-add results out of dependency chains. Every
original PC retains independent semantics; every original poll remains. Local
three-block diagnostics are near 1.54 ns/iteration for both host compilers and
the references, but these observations do not replace the unchanged final-head
gate. Tests cover all-slot results at every original continuation, aliasing,
edge copies, branch-range boundaries, fuel and cancellation on both ISAs.
