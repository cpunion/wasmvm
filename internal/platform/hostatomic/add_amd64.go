//go:build cgo && !llgo && linux && amd64

package hostatomic

import "sync/atomic"

func FetchAdd32(address *uint32, delta uint32) uint32 {
	return atomic.AddUint32(address, delta) - delta
}

func FetchAdd64(address *uint64, delta uint64) uint64 {
	return atomic.AddUint64(address, delta) - delta
}
