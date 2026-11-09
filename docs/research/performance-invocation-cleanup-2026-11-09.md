# Invocation-local cleanup without temporary heap cells

Baseline: `19fe83a1157f26194ad6fa2b497b7ecee8cae9f1`, with LLGo
`3d92a58e09bd8971899c6eaf63437fbecd18d432`, Go 1.27.3, LLVM 22.1.8 and the
unchanged `-O3 -lto=full` release profile.

The compiler-upgrade investigation isolates three remaining allocation requests
on the ordinary numeric boundary: two eight-byte deferred-capture cells and a
104-byte continuation-map iterator, including when that map is nil. Normal
linked code and a separate allocation-only size-class diagnostic agree.
The independent reproducer and direct-method/empty-map controls are reported in
[LLGo issue #3033](https://github.com/xgo-dev/llgo/issues/3033). Inlined collector
calls matter; absence of a named AllocU call does not prove allocation freedom.

Use direct deferred methods for depth release and workspace return. Their
receiver/argument pointers are immutable, so capture time, LIFO ordering and
panic cleanup match the previous closures. The workspace still returns before
depth decreases. Only enter continuation cleanup when the map has entries;
nonempty maps retain exactly the existing lock/abort/unlock loop.

No Store admission, context/exposure lifetime, fuel, execution driver, native
ABI, compiler stage or guest instruction policy changes. Existing tests cover
success, limits, cancellation, close, reentry, continuation drain and panic
cleanup on both backends. Go race and pinned LLGo release tests pass with 92.7%
scoped VM coverage. Final exact-head CI remains required.

Allocation and affected-interface timing are separate diagnostics. Keep the
original typed/scalar/live/session/dispatch/entry and named reference policies,
all zero-call controls, fixed rounds/counts and A/A results. No performance or
zero-allocation claim is established until measured; controlled acceptance and
release remain open.
