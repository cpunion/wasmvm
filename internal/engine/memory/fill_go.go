//go:build !llgo

package memory

// Standard Go's bulk copy outperforms its scalar nonzero-byte loop. Expand an
// initialized prefix; the returned copy count bounds the final partial block.
func fillNonzero(data []byte, b byte) {
	data[0] = b
	for initialized := 1; initialized < len(data); {
		initialized += copy(data[initialized:], data[:initialized])
	}
}
