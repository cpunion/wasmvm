# Rejected prefix eligibility guard: 2026-11-02

Candidate `925011575393b81c463b8d2abbfb4fa840383cc8` moves the native/metered
eligibility check outside the interpreter's prefix loop. It is semantically
valid but rejected for performance; production source is restored exactly to
the earlier `d8a3569` prefix implementation.

The LLGo Run disassembly of original and earlier prefix binaries has the same
143 call sites and the same allocation call-site count. No new allocation site
is established by that comparison; it does not prove equal dynamic allocation
counts. No new LLGo issue is inferred from a timing difference alone.

A focused four-case comparison reuses exact verified base/old binaries and
builds only the new variant under pinned Go 1.27.3/LLGo 8e45ed5 O3/full LTO.
Original base, earlier prefix and guard variant use both backends/compilers;
exact baseline A/A aliases cover interpreter and native. All five pinned
references remain, with explicit configured GC exclusions and applicable C
controls. Three scalar loops retain one million iterations; the linked GC
chain retains ten thousand. Three fresh interleaved blocks, two warmups,
seven calls and all phases/oracles stay retained. This is a shared M4 Max host,
not controlled acceptance.

All 31 Go packages/30 coverage gates and pinned LLGo execute/VM pass at the
variant. Those checks do not establish its speed.

| Workload | LLGo interpreter base | Earlier prefix | Guard variant | Base A/A |
| --- | ---: | ---: | ---: | ---: |
| core/integer-loop | 51.6476 | 42.8903 | 43.6200 | 51.1072 |
| core/function-calls | 82.7512 | 72.8795 | 73.2252 | 83.1343 |
| core/memory-load-store | 99.4631 | 81.0275 | 101.3830 | 99.8559 |
| gc/struct-allocation | 167.2083 | 158.4250 | 160.1625 | 161.3250 |

Units: ns/source iteration. LLGo scalar memory regresses 25.1% relative to the
earlier prefix, while baseline A/A differs about 0.4%. Native guards being
logically outside the prefix scan does not guarantee a better dispatch binary.
The simpler earlier implementation is retained; no byte padding or fixture-
specific alignment is introduced to rescue this variant.

In this same follow-up, LLGo native GC is base 68.1334, earlier prefix 63.9875,
guard 69.5500 and baseline A/A 63.2958 ns/iteration. Go is 49.2958/50.0667/
48.8792, A/A 49.4917. The earlier prefix's initial +18.4% native GC observation
is not stable across subsequent +3.2% and -6.1% comparisons; baseline A/A also
moves materially. All earlier unfavorable data remain. This narrows but does
not close the controlled native regression question, and does not accept the
whole portfolio or ordinary-native target.

[Evidence](evidence/performance-prefix-guard-held-2026-11-02.json.gz) retains
all three variants' identities, all phases/configurations, A/A controls,
allocation-call-site assembly diagnostics and validation results.
