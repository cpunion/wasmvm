# Reusable call execution scope

Baseline: `f09d952220ccacdcfa7aa0659750244f00893cb1`. The public Store call path
currently creates an execution, native context, frame arena and cancellation
registration for each call. Retain one bounded idle workspace per Store for
outermost calls, using the existing frame driver and native backend. Reentrant
calls keep independent workspaces and the existing shared quotas/fuel policy.

An idle workspace must release instance frames and memory bindings, and must
not root guest references or a former invocation's context. Stop the old native
cancellation callback before reuse. Cancellation remains sticky: if its callback
has started, discard that native context rather than clearing its control word.
Store close releases the idle workspace before retiring Store-owned state/code.
Keep return values valid until encoding and preserve panic/host-reentry cleanup.

Test repeated calls across functions, zeroed locals, traps, metering/fuel policy
changes, cancellation followed by a fresh call, panic cleanup, host reentry,
resource limits, root collection, memory binding release and close races.
Execute the complete Go/LLGo interpreter/native matrix and keep every owned
executable package above 90% statement coverage. No whole-function fallback or
new workload-specific execution path is permitted.

Compare both own compilers with the pinned three references through the public
API. Add the existing one-invocation call-control fixture to the 33 SIMD and
eight regression families. Its four alternate seeds always use one source
invocation, and report ns/call without fictitious loop normalization. Other
families retain 100,000 source iterations, three process blocks, seven calls,
two warmups and randomized interleaving for diagnostics. Keep all startup,
compile, instantiate, first-call and warm fields; separate reused internal
dispatch diagnostics when attributing residual boundary cost. Retain failed
measurements. Diagnostic completion does not close the sustained calibrated
performance or PR #26 gate.
