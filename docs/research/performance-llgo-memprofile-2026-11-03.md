# LLGo allocation-profile attribution gap: 2026-11-03

Pinned LLGo 8e45ed5, -O3 -lto=full, macOS arm64; Go 1.27.3 reference.
An independent program retains 512 allocations from each of two distinct
noinline allocation functions, sets MemProfileRate=1 and collects twice before
snapshotting. Both MemProfile calls return ok=true. Go yields 23 records with
nonempty stacks and identifies both allocationA/allocationB. All fifteen LLGo
records have empty stacks: the pinned recorder groups by size without Stack0.

This prevents allocation-site attribution for the remaining VM/host overhead.
Aggregate TotalAlloc or size classes do not identify the allocating call site;
qualitative sampled stacks cannot supply that accounting. No new VM throughput
or compiler-patch speedup is inferred.

Existing [proposal #2351](https://github.com/xgo-dev/llgo/issues/2351) is closed;
its [implementation PR #2027](https://github.com/xgo-dev/llgo/pull/2027) is closed
without a merge. The concrete pinned reproduction was [added to that proposal](https://github.com/xgo-dev/llgo/issues/2351#issuecomment-6587978601)
rather than creating a duplicate. Current upstream main was not rebuilt, so
its current behavior remains unclaimed. Other independently reproduced defer,
Cgo and aggregate-parameter reports remain separate.

[Standalone source, outputs, commands/identities and upstream status evidence](evidence/performance-llgo-memprofile-2026-11-03.json.gz)
