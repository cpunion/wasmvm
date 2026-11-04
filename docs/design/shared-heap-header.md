# Shared object headers

This candidate replaces the common 64-byte record and its separate native view
with one 32-byte header on the supported 64-bit targets. A header contains the
complete handle, a payload pointer, 32-bit length and type, and a full 64-bit
allocation serial. The pointer slot table remains 16 bytes per handle. This is
an experimental representation until complete semantics and performance pass.

Portable headers belong to typed Go pages, so payload pointers remain host GC
roots. Native headers alias the owned, stable native view allocation and contain
only owned native payload pointers. Native entries borrow this storage only
between serialized safepoints. There are no Go pointers in native storage.
Optional tags, descriptors and continuation metadata remain in lazy typed pages.

The live slot's otherwise unused free-list word holds mark and optional-cell
flags. Free slots retain the complete 32-bit next index. Handles retain all
31 generation bits, and allocation serials retain all 64 bits and their original
exhaustion/rollback behavior. A canonical-type-indexed layout cache is bounded
by registered types. Payload sizes derive from immutable layouts and lengths;
exceptional continuation charges live in optional metadata. Logical accounting
still charges the original 64 metadata bytes per object and descriptor charge.
Physical record savings do not alter the quota or collection threshold.

Native word zero remains the complete handle and word one the payload address.
Word two packs the low 32-bit length and high 32-bit type. During construction
the type is instance-local and word three is zero. Finish validates the complete
batch, reservations, types, payload bounds and accounting before publishing any
object. It then replaces local types with canonical IDs and assigns serials in
the same headers. Both amd64 and arm64 length/bounds loads read only 32 bits.
No external record/C-entry ABI changes; this private view and emitted code are
owned by the same runtime build and are never persisted independently.

Metadata slabs may extend past the currently committed view length. Native
backing must grow to the complete quota-bounded slab before it is aliased,
without clearing pending constructors or moving existing headers. Failure must
release reservations without publishing partial batches or consuming epochs.
Private native Memory growth previously copied and moved storage. Header views
now use the same owner/lifetime API with a bounded PROT_NONE anonymous mapping;
growth enables read/write access only for newly exposed pages. Reservation does not eagerly initialize the maximum payload extent. Close unmaps
the complete reservation
after active entries and bindings drain. Shared guest memory behavior is unchanged.
Payload growth and compaction update the shared pointer once. Empty-object
sweeping must distinguish published, reserved and free headers, even when a
reserved constructor has already written a nonzero handle.

Required validation includes exact layout/alias checks under both compilers,
host-pointer ownership and lifetime checks, incomplete/invalid native batches,
quota/error/close paths, generation reuse, serial exhaustion, nested rollback,
cycles and typed roots, exceptions, descriptors and continuations. Both native
tiers and the interpreter must execute the full affected pinned inventories,
including forced Stack Switching collection. Performance retains sparse and
linked heaps, access, empty objects, exceptions and ordinary controls against
the exact merged PR117 binaries and independent competitors. Regressions remain
recorded; the smaller header alone does not establish a faster runtime.
