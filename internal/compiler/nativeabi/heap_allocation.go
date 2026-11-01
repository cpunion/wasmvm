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

const HeapAllocationBatch = 64

// Bound constructor work inside one slot instruction; other shapes keep the
// complete checked helper. Payloads never move while an entry borrows them.
const MaxNativeStructFields = 32
