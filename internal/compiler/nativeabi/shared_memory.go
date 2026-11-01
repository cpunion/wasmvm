package nativeabi

// Shared memory has an independently owned, stable native control block.
// Only full-width atomic accesses touch these cells; guest data has its own
// mixed-size machine-access contract.
const (
	SharedMemorySizeOffset    = 0
	SharedMemoryWaitersOffset = 8
	SharedMemoryControlBytes  = 16
)
