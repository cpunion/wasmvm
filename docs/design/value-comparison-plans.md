# Resource-aware value comparison plans

Component schemas are compilation-owned immutable graphs. Runtime stream,
future and task-return checks repeatedly compare the same schema pairs while
resource identities belong to current namespace bindings. A bounded comparison
program reuses the structural work without retaining nominal identity results.

`validate.ValueComparisons` keys each program by the two schema pointers and
the exact node budget. The existing equality visitor compiles it with a recorder
that accepts each nominal predicate and retains the ordered ID pair. The visitor
still performs all shape/name/label comparisons, visited-pair handling and node
charges. Its first structural mismatch, success or budget failure becomes the
terminal outcome. No second graph comparison algorithm is introduced.

Execution rechecks the recorded predicates against current bindings, in their
original order, including repeats before memoized graph edges. A failed predicate
terminates before any later structural mismatch or budget failure. Recording
stops at the original visitor's first terminal outcome, so unreachable resource
checks do not run. Budget changes select a different program. A budget failure
creates the usual fresh error; error objects are not retained in the cache.

Resource-free programs need no callback or namespace access at runtime. Nominal
programs retain only numeric resource IDs, not namespaces, resource objects or
the result of earlier lookups. Missing resolvers leave obligations unresolved.
The input graphs must remain immutable throughout the cache lifetime, matching
the existing compilation and Canonical ABI builder contracts.

Each component instance retains at most 64 programs and 4096 nominal operations.
A last-pair shortcut precedes bounded map lookup. Capacity exhaustion or an
oversized nominal program uses the original checked comparator. Recording and
callback capture stay in separate cold helpers. The Store actor serializes all
lookup, construction and matching; no shared global cache or locking is added.
Shutdown drains Store calls and source cleanup before releasing program/type
roots, including cleanup panic paths.

The same plans serve stream/future endpoint matching and task-return schemas in
both backends. This optimization does not reuse conversion budgets, host Callers,
resource ownership or returned values, and it does not skip runtime validation.
General Canonical ABI conversion still has its existing checked implementation.
