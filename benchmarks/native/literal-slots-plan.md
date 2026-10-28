# Shared literal slot scope

Retain local-coalescing candidate `7bfd24befb223910100084baedb37e6245de3860`
and its mixed measurements. Its CRC native regression is not accepted. Compare
the next candidate with both that intermediate and engine baseline `f1513a8`.

Pool identical numeric/vector literals only while their block-local SSA values
overlap. Reference-count active definitions so a slot is recycled only after
every alias dies. Keep cross-block values dedicated. Remove each inactive pool
entry and reset pools per block, avoiding function-wide retention of distinct
literals. Preserve exact types/bits, including NaN payloads and signed zero;
keep instructions, original PCs, fuel and cold-entry semantics unchanged.

Weight native local copies by their actual transfers: eliminated self-copies
must not compete for registers, and a coalesced tee has one destination. Reuse
the existing allocator rather than adding another execution path.

Extend independent symbolic checking for equal literal values. Check overlapping
and disjoint lifetimes, last-alias release, local writes, mixed types/bits,
vectors, edges, suspended calls, quotas and the full Go/LLGo Core 2.0 corpus.
Retain package coverage, final four-target semantics, the same twelve-family
base/intermediate/head comparison with pinned references, separate phases and
the existing performance/calibration gates. No new acceptance waiver.
