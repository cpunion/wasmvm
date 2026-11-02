package nativeabi

// The private allocation view contains only numeric words in owned native
// memory. A driver prepares one batch and commits it after every native exit.
const (
	HeapAllocationCursor = iota * 8
	HeapAllocationCapacity
	HeapAllocationHandles
	HeapAllocationBase
	HeapAllocationUsed
	HeapAllocationLimit
	HeapAllocationBudget
	HeapAllocationDescriptors
	HeapAllocationHeaderBytes
)

// Amortize host publication while retaining bounded reservation/rollback work.
// Generated constructors still check their byte budget and every native poll.
const HeapAllocationBatch = 256

// Acyclic, small constructor sequences avoid reserving the full loop batch.
const HeapAllocationSmallBatch = 64

// Bound constructor work inside one slot instruction; other shapes keep the
// complete checked helper. Payloads never move while an entry borrows them.
const MaxNativeStructFields = 32
