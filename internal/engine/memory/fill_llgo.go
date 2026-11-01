//go:build llgo

package memory

// LLVM recognizes this bounded slice loop as a bulk fill. Keep address/range
// checks in the shared caller so each byte needs no additional bounds proof.
func fillNonzero(data []byte, b byte) {
	for i := range data {
		data[i] = b
	}
}
