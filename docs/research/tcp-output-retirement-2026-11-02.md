# TCP output shutdown during worker retirement: 2026-11-02

PR #101 head `5f41558` passes its runtime benchmark but fails Linux LLGo's
`TestUpstreamP2WorldConsumers/native/p2_tcp_streams`: the guest reaches the
unchanged 30-second context deadline while waiting for closure. The consumer
uses metered native execution, whose captured bodies are unchanged by the
unmetered compiler optimizations. This identifies a real matching host-output
race rather than proving those optimizations caused the timeout.

The deterministic regression reconstructs this interleaving:

1. The output worker observes an empty queue and returns from its pump.
2. Before worker retirement sets `running=false`, `shutdown` stores a completion
   callback. It sees an active worker and does not start another.
3. Retirement sees neither queued bytes nor a flush. It exits without consuming
   the pending shutdown; the remote input never receives its EOF notification.

Retirement now treats a pending stop as work for the existing successor-worker
path. The accepted prefix still drains before send shutdown. A small shared
retirement method permits a deterministic test without production test hooks.
Callbacks remain outside the stream mutex, run exactly once, and retain existing
failure/drop/cancellation and worker ownership rules. No extra execution framework,
weakened timeout, discarded bytes or skipped consumer is introduced.

The regression fails before the condition change (completion missing after two
seconds) and passes afterward. Fixed source `52a873e` passes all 31 Go packages,
30 coverage gates, pinned LLGo O3/full-LTO Preview2, full package race/vet, and
ten consecutive executions of the original native TCP consumer. These local
results do not replace the new final-head dual-platform CI. This is an owned
ioStream issue; no LLGo defect is inferred or filed from this failure.

[Evidence](evidence/tcp-output-retirement-2026-11-02.json.gz) retains exact source,
regression, baseline/fixed logs, commands, coverage, earlier CI job failure and
the separately identified earlier-head benchmark. PR #101's current 5f benchmark
records native memory Go 0.9084→0.8043 and LLGo 0.9061→0.8549 ns/iteration,
versus Wasmtime 0.7218. LLGo native integer/calls regress 1.1%/2.7%; Go native
calls 3.4%. All phases and source/artifact audits remain retained. That earlier
source is not relabelled as the final corrected contribution head.
