//go:build cgo && !llgo && darwin && arm64

package hostatomic

// FetchAdd32 adds delta and returns the old word. Callers must check FEAT_LSE
// and retain a naturally aligned memory word for the entire call.
//
//go:noescape
func FetchAdd32(address *uint32, delta uint32) uint32

// FetchAdd64 has the same contract for an eight-byte word.
//
//go:noescape
func FetchAdd64(address *uint64, delta uint64) uint64
