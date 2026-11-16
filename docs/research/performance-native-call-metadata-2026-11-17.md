# Rejected immutable native call metadata

The 2026-11-17 experiment moved caller function/PC identity from each 24-byte
private call record to immutable module metadata, leaving 16-byte Used/resume
records. Exact continuation membership and independent code leases retained the
normal lifetime and admission checks. Both Go and LLGo semantic tests passed.

The first image-offset registry improved LLGo general calls by 1.70%, but Go
16 B P3 request CPU increased 0.81% and instructions increased 0.80%, with
resolving same-binary controls. Go general-call controls were unresolved.

A separate event bound metadata to absolute code addresses at publication and
resolved the held module once per changed exit packet. General calls improved
1.84% under Go and 2.07% under LLGo, with resolving controls. Complete P3 request
CPU still increased under Go: 16 B **1.0172 [1.0116, 1.0225]**, 64 KiB
**1.0170 [1.0115, 1.0222]** head/base. Instruction ratios were 1.0082 and 1.0069;
both CPU and instruction controls resolved. LLGo request controls were unresolved.

Each event used six fixed fresh-process symmetric-AA blocks, the same pinned
`f5e246c8` SDK with `-O3 -lto=full`, default GC, full request/result checks and
separate startup/setup phases. Intervals are marginal bootstrap 95% intervals;
there was no event pooling, stopping extension or competitor acceptance claim.
The request measurements include driver and verification work. No total-allocation
comparison is inferred from the managed allocation counters.

The registry introduces host lookup work on native exits. The data rejects the
complete tradeoff; it does not isolate that lookup as the exclusive cause.
Both implementations and raw protocols remain local experimental evidence under
`artifacts/immutable-native-call-sites-20261117`, including `bound-addresses`.
Production retains the compact 24-byte record and direct identity loads. Further
work targets argument transfer and frame initialization without a new exit lookup.

LLGo subsequently merged `9e45f103` (wide SIMD rounding/conversions). Its upgrade
is a separate compiler event; none of the older measurements above is relabeled
as a result from the new SDK.
