# Store-owned private operation reuse

The scalar-call comparison in PR #142 still exposes a large public-boundary
gap. Reusable argument storage removes two Go allocations, but every ordinary
call also constructs an operation and a frame-registry slice. This change
retains one completed private operation per Store alongside its existing idle
execution workspace. It does not change guest execution or add an API mode.

Reuse is limited to the standard Background/TODO parents. All frames must have
left and quiesced, reentry depth must be zero, and no cancellation child may
have been published. Parent watchers and ordinary WithCancel contexts are never
reused. Creating any Caller permanently marks that operation exposed, even if
the callback never requests Context. This covers host callbacks, sessions and
their continuations. Public context observation also permanently marks exposure.
Expired public capabilities therefore keep their original canceled lifetime.

Store admission serializes the cache. Finish still cancels the operation and
drains frames/capabilities before releasing admission; only then can begin
reuse unobserved private storage. Fuel, options, pending-instance roots and
collection state are reset. Closing the Store discards its idle operation.
Published contexts and channels are never reset, and native interruption keeps
the existing registered-frame ownership protocol. Thread waits that observe
Done automatically prevent reuse through cancellation-child publication.

Tests cover repeated successful/fuel-limited execution, input preservation,
fresh fuel, observed context/channel/value retirement, custom/live parents,
retained Callers first requesting Context after later calls, and Store close
interrupting a private infinite guest loop after workspace reuse. Existing
panic, continuation, nested reentry, deadline, GC and retired-watcher suites
remain required under both backends and compilers.

Allocation counters and timing are separate diagnostics. Compare the unchanged
typed/scalar/session/dispatch interfaces, preserve live-parent controls, directly
time preparation/first batch/cleanup, and retain all source/base, reference
policy and A/A rows. Complete all builds and probes before a frozen-source event.
No speedup, zero-allocation claim, controlled performance or release acceptance
is established merely by this implementation.
