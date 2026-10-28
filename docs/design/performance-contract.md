# Core engine performance contract

Status: **proposed engineering requirements, not implemented capabilities or measured results**. This contract follows the decision to implement wasmvm's own core. It defines how the interpreter, compiler, and their optimization changes will be evaluated. The numerical budgets below are provisional acceptance requirements; they do not describe the current prototype or promise a result before measurement.

Correctness and bounded execution are prerequisites. A faster result obtained by omitting validation, bounds checks, fuel obligations, cancellation checks, or required feature semantics cannot satisfy this contract. M1 accepts the full declared Core 2.0 interpreter, M2 accepts the optimized interpreter and native compiler over that same scope, and M3 extends both backends to the declared Core 3.0 and Threads profiles. A small fast subset does not complete a milestone.

## Per-feature merge gate

Every feature-development PR compares the affected capability with pinned
relevant competitors and the own baseline, using identical guest workloads and
separate preparation/first/warm measurements. It may merge only after reaching
parity or better on the affected workloads for both host compilers and required
targets. Slower, missing or inconclusive evidence blocks merging. A tiny
noise-bound control cannot establish either a win or a loss; resolve it with
an appropriate measured workload before deciding. Do not waive a deficit by
averaging unrelated cases or weakening safety settings.

The initial native integer feature gate uses the existing sustained workload,
20 independent process blocks (one predeclared extension to 60 if inconclusive),
simultaneous 95% paired-bootstrap intervals adjusted for both planned looks, and
identical-binary A/A controls. Practical parity means the entire interval lies
within ±0.5%. An interval spanning faster and parity also passes if its upper
bound is at most 1.005; crossing the boundary between two accepted outcomes is
not a rejection. A/A intervals must resolve ±0.25% or the gate remains inconclusive. Candidate and control calls use 25 million iterations of the same sustained integer binary and seed; the three descriptive baseline calls retain five million. All warm rates and baseline ratios normalize each record by its own count. The longer candidate duration changes neither acceptance margins nor planned looks; earlier failed measurements remain visible.
These feature-level equivalence limits do not replace the stricter full-portfolio
throughput targets below. Record every methodology change with old/new evidence.

The `warm-call-interleave-v1` protocol prepares each candidate serially, then
randomizes individual warm calls across waiting processes within each block.
Only one runtime executes a timed call at a time. Handshakes and final state or
checksum inspection are outside call timers; necessary polling work stays
inside. Each block still uses fresh processes and seven warm calls after two
warmups; the process median remains the statistical unit. Raw records include
the dispatch order and protocol events. This changes temporal pairing, not the
workload, margins, planned looks or generated code. Earlier serial-call reports
remain evidence; compare protocols explicitly rather than pooling their samples.

Keep these checks proportional: reuse existing runners/fixtures and the current
PR benchmark job, scope comparisons to the changed capability, and retain the
full-profile controlled acceptance gate below. Begin native register allocation,
memory access, calls and polling comparisons during their implementation, not
after completing a slow whole backend. Compact shared stages and runtime
contracts take precedence over benchmark-specific implementations.

## 1. Measure wasmvm implementation revisions

The historical CLI harness builds official wazero executables selected by base/head dependency pins. The phase harness now adds exact wasmvm implementation builds under Go and LLGo, retaining wazero/other engines as explicitly separate references. Historical dependency-only results remain labeled as such.

The owned embedding adapter executes wasmvm's core directly. Build the core from the exact base and head commits, with the same pinned host compiler and build options. Build production CLI latency measurements from those commits as well. The artifact must identify the core commit, adapter source hashes, binary hash, host compiler, compilation options, feature profile, and runtime policy. A benchmark must fail identity verification if it accidentally links wazero as the wasmvm implementation.

Use a common versioned adapter and protocol when the API permits it. The M1 adapter requires internal-package access, so committed source is archived into isolated build trees and only the two instrumentation directories are overlaid from head. This serves the same isolation purpose as selecting a complete checkout through a local module replacement. Record the resolved local module, checkout identity, and source hashes. If an API change requires revision-specific adapters, record both adapter versions and review the changed timing boundaries before comparing results. Copying head core implementation files into the base build is forbidden.

When the baseline predates the engine, record `baseline-unavailable` and the exact reason. Run head correctness and competitor measurements, establish the first baseline through its PR, and leave the revision-regression comparison unaccepted. Never substitute a reference engine or a successful build for a missing wasmvm execution result.

### Required runtime matrix

| Series | Configurations once M2 is implemented | Meaning |
| --- | --- | --- |
| Go host, own engine | wasmvm base interpreter, head interpreter, base native, head native | Four actual implementation builds/backend configurations |
| Pinned references | wazero interpreter, wazero native, Wasmtime Cranelift, Wasmer Cranelift, WasmEdge interpreter | Five separate reference rows; no synthetic fastest-per-case reference |
| LLGo host, own engine | wasmvm base interpreter, head interpreter, base native, head native | Four additional configurations, using the same fixtures and assertions |

This gives **nine configurations in the Go/reference series and four in the LLGo series**, before multiplying the own-engine rows by required fuel/policy configurations. M1 has no native wasmvm implementation: its corresponding matrix is seven plus two configurations, with native support explicitly absent. Once a backend enters a milestone's required scope, its failures cannot be hidden by a fallback or an unsupported label.

Each own-engine host compiler is evaluated independently. Go and LLGo results must not be pooled into one median. The native compiler must expose its actual selected backend and emitted-code identity; a requested native run that silently interprets is a correctness failure. The known LLGo/wazero native boundary does not excuse failures in wasmvm's own native backend.

All participants receive the same head fixture bytes, runtime arguments, expected results, resource sizes, and shared feature profile. Record policy differences rather than asserting identical cancellation or metering semantics. Every kernel has finite inputs and an external process timeout. Pin reference versions, SDKs, backends, optimization settings, and checksums. Upgrading a reference is a separate recorded baseline event. Missing required tools, corpus files, or selected backends are infrastructure failures.

## 2. Direct, nonoverlapping measurement boundaries

Retain the existing ready/acknowledgement protocol: the child emits its ready event and waits for `run\n`; the parent records readiness before releasing API work. Keep protocol parsing, assertions, result serialization, and bookkeeping outside the API timers. Reuse argument/result storage in every adapter where the API permits it.

| Metric | Required interval and interpretation |
| --- | --- |
| Process readiness | Process launch to observed ready event, including runtime initialization and pipe delivery; not bare process-creation cost |
| Input load | Read fixture bytes; filesystem cache state is recorded separately |
| Engine/store setup | Explicit construction of runtime policy, engine, and store objects |
| Decode | Binary reading and structural decoding into owned module data |
| Validate | Type/control/index/feature/limit validation of the decoded module |
| Lower | Conversion of validated semantics into executable interpreter or compiler IR, including required metadata |
| Optimize | Optional transformation passes, with the selected pass set recorded |
| Code generation | Native instruction emission, relocations, and backend-specific layout work |
| Code publication | Executable memory allocation/protection, relocation finalization, and registration needed before execution |
| Preparation total | One outer interval covering the actual decode-to-executable pipeline; no duplicate validation API call |
| Instantiate | Bind imports and allocate/initialize instance state; pure fixtures have no start function or host imports |
| Export lookup | Resolve the callable entry point; validate the signature outside this timer |
| First call | First invocation on the new instance, including any real lazy work |
| Warm call | Direct API invocation on the same warmed instance, with result checks after the timer |
| Cleanup | Explicit instance/store/module/engine teardown through completed reclamation (`Close` plus `WaitClosed` or equivalent), not just a nonblocking close request; process shutdown remains separate |
| Clock control | Consecutive clock reads, recorded without subtracting their cost |
| Process wall time | Launch through completion and pipe draining, including all calls and untimed protocol work |

The phase schema must distinguish `measured`, `not-applicable`, and `not-exposed`. A portable interpreter has no native code-generation phase: record `not-applicable`, not an invented zero-duration compilation result. References that expose only a combined preparation API retain that aggregate. Do not infer their validation or optimization costs by subtracting unrelated calls.

Stages inside wasmvm must correspond to real internal boundaries. Record whether stages overlap or include one another; the outer preparation interval is an aggregate, not an additional summand. Aggregate observations before computing statistics: sums of phase medians do not reconstruct a measured total. Whole-process time must not be reduced by subtracting readiness, clock, or call-control medians to manufacture a "pure guest" result.

Fresh-engine preparation, cache-hit preparation, and deserialization of compiled artifacts are separate workloads. Default comparisons use fresh engines and no disk compilation cache. Record lower-level code/cache counters when a supposedly fresh operation can reuse work.

## 3. Workload portfolio and correctness evidence

The current integer, source-call, and memory loops are a starting point, not the acceptance portfolio. Before accepting a core milestone, commit a versioned portfolio with independent oracles, source and binary hashes, producers, workload sizes, normalization units, and the exact required feature set.

Primary warm-execution families must include:

- Integer and floating-point arithmetic, conversions, bit operations, and mixed dependency chains.
- Predictable and irregular control flow, nested blocks, `br_table`, locals, and multi-value transfers.
- Direct calls, indirect calls with type checks, deep guest frames, and host-call boundaries measured separately.
- Aligned and unaligned private memory, bounds checks, bulk copy/fill, and distinct small and sustained working sets.
- Globals, tables, reference operations supported by the profile, and fixed SIMD lane/shuffle/memory operations.
- Complete compute workloads that combine instructions and control flow, including realistic modules produced by at least two pinned guest toolchains.

Use representative families rather than adding many variants of one favorable loop. Predeclare primary cases and family weights. Keep instruction microbenchmarks and call controls as diagnostic series; they cannot replace application-shaped cases. A source-level call can be inlined by a native compiler, so its result is not an exact machine-call latency.

Compilation/instantiation cases must independently vary binary size, function count and size, control depth, locals, imports, tables, memory/data initialization, and valid/invalid inputs. Include cold compilation of many small functions and a few large functions. Preserve error classifications and test bounded rejection of malformed or excessive inputs; rejecting an invalid module is not successful execution.

Every first, warmup, and measured call checks the expected result. Include alternate seeds, signed/unsigned boundary bit patterns, different loop counts, and repeated calls on the same instance. Stateful fixtures reset their state explicitly; report whether reset is inside or outside the measured interval. Seed/count parameters must prevent accidental acceptance of a constant-result implementation. The corpus must be accepted independently by a pinned reference validator.

Use specification-defined result sets for NaNs and nondeterministic operations. Check trap categories and observable state, not arbitrary error strings. Correctness suites must exercise optimized and unoptimized forms, both own backends, both host compilers, and applicable instruction combinations. Required conformance cases are not removed because their timing is inconvenient.

M3 adds separate GC allocation/liveness, shared-memory contention, wait/notify, growth, and lifecycle series. Do not fold those results into an earlier Core 2.0 aggregate. A reference that does not implement a later feature stays explicitly outside that comparison; it does not become a successful zero-cost sample.

## 4. Initial numerical requirements

Ratios compare elapsed time: smaller is better. Use equally weighted workload families, with equal case weights inside each family. Compute geometric means of per-case ratios, not an arithmetic average of nanoseconds or a mix of different compiler/backend series.

| Requirement on the controlled acceptance runner | Initial budget |
| --- | --- |
| M2 interpreter warm execution against pinned wazero interpreter | Family-weighted geometric mean ≤ **1.00×**; every primary case ≤ **1.25×** |
| M2 native warm execution against each of wazero native, Wasmtime Cranelift, and Wasmer Cranelift | Geometric mean ≤ **1.20× against each named engine separately**; every primary case ≤ **1.50× against each** |
| Initial interpreter preparation against pinned wazero interpreter | Geometric mean ≤ **1.50×**; every primary case ≤ **2.00×** |
| Initial native preparation against pinned wazero native | Geometric mean ≤ **2.00×**; every primary case ≤ **3.00×** |
| First call on each non-control primary workload | ≤ **2.00×** that workload's warmed-call time; keep the original first-call samples |
| Empty exported-call control, warmed / first call | ≤ **1 microsecond / 10 microseconds**, separately on each accepted compiler/platform |
| Process readiness for the standard benchmark adapter | ≤ **10 milliseconds** on each controlled acceptance platform |
| Guest cancellation request-to-stop, excluding arbitrary blocking host callbacks | **p99 ≤ 10 milliseconds; observed maximum ≤ 100 milliseconds** in the predeclared stress run |
| Own head/base warm-execution regression | Geometric mean ≤ **1.05×**; every primary case ≤ **1.10×** |
| Own head/base preparation-total and instantiate regression | Each operation's geometric mean ≤ **1.05×**; every primary case ≤ **1.10×** |
| Own head/base startup, decode, validate, lower, optimize, codegen, publication, and cleanup diagnostics | Investigate changes over **10%** that exceed their measured noise/resolution floor; a faster total does not erase the recorded component change |
| Warm numeric/private-memory calls after declared preallocation | **Zero engine allocation events per call**; also report host allocation measurements where the host compiler exposes reliable counters |
| Own head/base IR size, emitted-code size, or peak engine-owned memory | No unexplained per-case increase over **10%**; retain absolute bytes and size denominators |

Commit a versioned budget manifest containing these initial ceilings, workload definitions, platform/compiler identity, calibrated absolute effect floors, and the cancellation sample plan before M2 performance acceptance. Each cancellation stress scenario uses at least 1,000 randomized requests per required compiler/backend/platform and retains all observations; the maximum is an observed run bound, not a real-time scheduling guarantee. A/A calibration determines measurement resolution, not an acceptable latency. Missing budgets or a first native `baseline-unavailable` may permit a bootstrap contribution, but cannot close M2.

M1 establishes the complete correctness and measurement baseline; M2 must satisfy the numerical interpreter and native performance gates. The native target is not an average across the three competitors: being fast against one must not hide a large deficit against another. WasmEdge remains a separately reported interpreter reference. Apply the own-backend requirements to Go and LLGo builds independently on accepted host platforms; successful Go results cannot waive a required LLGo result. References use their explicitly recorded Go-built or C-API hosts; an LLGo wasmvm comparison against those rows is not a claim that the reference engine runs under LLGo.

These initial ratios are demanding engineering targets, not evidence that the first implementation will meet them. Do not declare a milestone performance-complete until its controlled measurements satisfy the targets over its full primary portfolio. Compile/execute tradeoffs must remain visible: an optimization that improves warm execution but exceeds the preparation regression budget needs a reviewed contract change or another implementation.

Tiny call/clock controls, cold filesystem I/O, and cases below reliable timer resolution are not judged by percentage alone. They remain visible and cannot be silently dropped. Before a case enters a hard gate, define a meaningful absolute effect floor from its measured resolution and A/A noise, and use enough work to resolve the intended effect. Do not invent a cross-host nanosecond threshold before calibration.

Any change to weights, primary cases, reference pins, policy, or numerical budgets requires an explicit design/PR explanation with the old and new results. Correctness requirements are not weakened to meet a performance budget. Automatically rebasing the accepted baseline onto a slower head is not a resolution.

## 5. Fuel, cancellation, and optimization costs

Reference runtimes do not share an identical deterministic fuel model or interruption API. Equal numeric fuel values are not equivalent work budgets. Therefore separate two configurations:

1. **Unmetered core acceptance:** disable optional deterministic fuel, but keep wasmvm's normal bounded cancellation polling, bounds checks, trap behavior, and required safety checks enabled. The reference warm-execution ratios above use this configuration. Record each competitor's actual cancellation settings and guarantees; the policies need not have identical cost. A faster unchecked wasmvm mode cannot satisfy these targets.
2. **Metered production-policy acceptance:** additionally enable deterministic fuel and exercise its interaction with cancellation, quotas, and interruption under both host compilers and own backends. Compare against the same wasmvm revision with fuel disabled and the same mandatory cancellation/safety policy. This is an additional requirement, not an alternative acceptance path.

Initial sustained-workload overhead budgets for enabling deterministic fuel are a geometric mean of **≤1.15× for the interpreter** and **≤1.20× for the native backend**, with **≤1.35× for each primary case**. Both sides retain mandatory cancellation polling. Keep controls and genuinely tiny calls separate. If a reference offers a comparable metering policy, show it in a separately named series with its actual semantics; do not merge incomparable fuel units. Internal instrumentation diagnostics may explain costs, but cannot replace production-policy measurements or expose an unchecked guest-execution mode.

Version the fuel cost model and record it with every result. Tests must prove exact success/exhaustion boundaries, consumed fuel, trap ordering, and side effects across interpreter/native and optimized/unoptimized execution. Lowering, constant folding, and fused instructions retain the charges required by that model. Charging a whole block is only valid if it preserves the promised exhaustion semantics; insufficient-fuel paths may need precise instruction stepping.

Fuel, cancellation and deterministic polling correctness apply from M1; numerical throughput, overhead and latency acceptance starts in M2. Measure cancellation request-to-stop latency and deterministic work between checks separately from throughput. The initial policy target is at most **1,024 original guest instruction points** between polls, including in fuel-disabled operation, with long bulk operations additionally polling at intervals of at most **64 KiB**. Test long straight-line blocks, loops, recursion, bulk work, and host reentry. Report latency distributions on the controlled runner; external host callbacks need their own cancellation/close contract and cannot be made bounded by a guest polling claim.

Fuel exhaustion and cancellation tests verify observable correctness even when their overhead targets are not yet met. A benchmark configuration that bypasses these paths cannot establish production-policy acceptance.

## 6. Controlled acceptance and hosted PR reports

The PR suite retains every workload family but excludes the duplicate sustained integer/memory variants; explicit full runs retain them. Selected cases keep three processes and seven measured calls. Full and PR scope is labeled in each report, and M2 controlled acceptance still requires the complete declared portfolio.

The existing three-process, seven-call hosted CI series is advisory. Keep it as a reproducible PR signal and correctness gate. It cannot establish a narrow performance budget on shared, uncontrolled hardware.

For performance acceptance, use a recorded controlled machine/runner image, CPU model/microcode, OS, compiler, affinity policy, frequency/power policy, and memory configuration. Run no overlapping benchmark jobs. Record host load and cache policy. Avoid enabling race, coverage, profilers, sanitizers, or debug tracing in timed executables; run those checks separately.

Use balanced, randomized base/head pairs and interleave references within each block. Obtain at least **20 independent process pairs** for a gated comparison, with seven measured warm calls after first-call separation and two warmups per process. Treat the process/block as the independent sampling unit: seven calls in one process are not seven independent machine runs. Keep long-lived-instance throughput as a distinct series if added.

Apply the same balanced pairing, sampling-unit and confidence-bound procedure to head/base, head/reference, first/warm, and fuel-on/off ratios. Absolute latency gates retain per-process samples or the predeclared cancellation distribution; do not replace p99 with a median. Compute per-process medians, paired log ratios, family-weighted geometric means, and **95% confidence intervals** using a recorded paired/block bootstrap procedure and seed. Preserve the raw samples. Use simultaneous bounds or a declared family-wise correction for the portfolio's per-case gate. Do not choose the statistical method after seeing whether a PR passes.

- **Pass:** the relevant upper confidence bound satisfies its ratio or latency budget (and any observed-maximum/zero-allocation condition holds) and correctness/integrity checks pass.
- **Regression:** the lower confidence bound exceeds the budget.
- **Inconclusive:** the interval crosses the budget, the A/A control is too noisy, or the measurement is unresolved. Increase the predeclared sample budget or repeat on a controlled runner; do not convert this to a pass.

Run A/A controls to establish the instrument and environment noise floor before enabling hard gates. An initial calibration requirement is a confidence interval width no greater than **5 percentage points** for primary warm/preparation ratios; tighten it where the requested effect is smaller. A/B and A/A data use the same scheduling and grouping. Up to 60 pairs may be predeclared for an inconclusive comparison; repeated selective reruns until one passes are not acceptable evidence.

The PR comment must distinguish correctness, hosted advisory results, controlled acceptance, and missing evidence. Include exact source commits, backend/compiler identity, policies, portfolio version, all references, and raw-artifact links. Keep own head/base changes separate from competitor ratios. No report may imply that the current wazero dependency harness measures wasmvm's own implementation.

## 7. Portable representation experiments

Use small, reproducible experiments to choose the interpreter representation, then validate the selected implementation over the entire milestone scope. An experiment supporting a handful of instructions is not a completed VM.

Compare these designs under both Go and LLGo using the same semantic fixtures and preparation/execution counters:

| Candidate | Representation | Questions to measure |
| --- | --- | --- |
| A: decoded stack instructions | Flat tagged instruction records, prevalidated immediates, resolved branch targets and stack heights; explicit value/call stacks | Dispatch cost, operand traffic, code footprint, preparation cost, and reliable baseline semantics |
| B: typed virtual registers | Flat register operands, constant pools, explicit basic blocks, branch argument moves, and typed value locations | Fewer stack transfers versus larger instructions/register storage and additional lowering cost |
| C: selected fused operations | Small measured combinations derived from validated A or B, preserving original offsets and fuel charges | Dispatch reduction versus code size, compiler behavior, diagnostics, and metering complexity |

Prefer plain slices, integer indices, and a `switch` dispatch baseline. Benchmark actual instruction-record sizes and generated host code on both compilers; do not assume their bounds-check elimination, inlining, or dispatch lowering is identical. Avoid interface/closure allocation per instruction, unsafe host-pointer tagging, and reliance on computed-goto or host ABI behavior for the portable baseline.

Use explicit guest frames so guest recursion and tail-call behavior do not depend on the host call stack. Keep scalar bit storage separate from references with host-visible ownership. Out-of-line constant pools, branch tables, type metadata, and root maps must be included in compilation-memory/IR-size measurements rather than hidden outside the instruction count.

Retain a simple semantic path long enough to compare optimized execution, including traps, result/state changes, fuel, and cancellation points. Differential tests against it are supplementary: both variants also run the pinned official suites and independent references. Fuzz malformed binaries, validation/control/type transitions, lowering invariants, and optimization equivalence. Preserve failing inputs, seeds, feature sets, and compiler versions.

Report statement coverage of owned decoder, validator, lowering, runtime, and backend packages separately from adapter coverage. Initial owned-package statement coverage target is **≥90%**, accompanied by an opcode/control/type/negative-case matrix; coverage percentage alone cannot accept an instruction or optimization. The same semantic suites execute under Go and LLGo, with additional Go race/fuzz and native memory-protection/code-publication checks as applicable.
