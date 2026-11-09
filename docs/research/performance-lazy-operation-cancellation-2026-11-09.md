# Lazy background operation cancellation

Baseline: `0bdf2295525189e18ac8a83ab7f98e7a20598b21` (PR #135).
This experiment targets public invocation overhead. Controlled performance and
production-release acceptance remain open.

## Diagnosis and mechanism

The six original Threads cases execute 10,000 guest iterations per API call.
The last pinned Linux event observes complete private/shared API calls at
12.514/17.102 microseconds, versus Wasmtime5.630/5.480. Those intervals include
invocation ownership, cancellation, marshalling and native entry. The emitted
amd64 private/shared hot loops have the same instructions; the apparent ratios
cannot all be assigned to instruction selection or memory synchronization.

An operation whose parent is exactly standard Background/TODO now embeds a
sticky cancellation bit. A standard cancel context is materialized only when
Done, Value or the public Caller.Context is observed. Caller.Context still
returns one stable standard context per operation. Custom, value-bearing and
cancelable parents retain ordinary WithCancel. Publication precedes the sticky
bit check, so racing materialization/cancellation cannot miss cancellation.

No operation or caller context is pooled or rebound. Store close, panic cleanup,
parent cancellation, context metadata/causes, nested reentry, full guest-state
publication, budgets and native cancellation polling retain their contracts.
The guest compiler, native ABI, original fixtures and execution policy do not
change. This is not an unchecked or benchmark-selected execution mode.

## Qualification and declared measurements

Go race and pinned LLGo execute the affected VM tests. Added cases exercise
lazy retirement without a child, standard/public context identity, late context
creation after cancellation, repeated cancel, AfterFunc, derived children,
concurrent publication/cancellation and non-comparable custom parents. Existing
metadata, deadline/cause, parent isolation, native interruption, Store close,
panic and reentry tests remain mandatory. Final full platform/compiler CI,
coverage and all six verified CLI packages remain merge requirements.

Keep Go1.27.3, LLGo38de1e218 `-O3 -lto=full`, LLVM22.1.8, named references,
source/fixture hashes and the five-block/two-warmup/seven-call protocol fixed.
The full Linux original22 portfolio and separate sustained GC cohort remain.
An additional **threads-sustained-v2** cohort runs all six unchanged Threads
binaries at exactly100,000 iterations, with independent checksums and all
native/interpreter references. It does not replace the original10,000 cases or
subtract empty-call costs. A separate measurement job reuses the shared build,
keeping ordinary semantic jobs unchanged. Local arm64 covers Core/Threads and
the separate sustained Threads cohort with the same pinned profiles.

Retain every sample, failed/unsupported outcome and A/A interval. Allocation
keeps process means including pauses. Startup, preparation, first call, warm
calls and cleanup remain separate. No profile/ISA/count/source sweep, selected
retiming or platform/cohort pooling is permitted. Final source-specific results
belong in the contribution PR; five-block diagnostics do not close the required
controlled performance gate. Go-host callbacks and live-parent boundary
controls are measured separately from sustained guest execution.

## Setup correction before local timing

The first hosted million-iteration Threads event (run42551167100) completed
five cases, then the wait-timeout interleaved block exceeded its existing
180-second deadline. Preserve that entire event, including the failed sixth
case. It is not a complete cohort or an accepted performance result. Before
local timing, uniformly reduce this additional cohort to100,000 iterations
for all six unchanged binaries and label it threads-sustained-v2. The existing
original22, process protocol, references and timeout remain unchanged. This
single correction bounds CI work; do not sweep counts or select favorable
samples. Report both hosted events separately. The engine sources are unchanged.
