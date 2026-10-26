// Package value defines storage shared by portable execution primitives.
package value

// Value stores exact scalar/vector bits. Scalar values use Lo with Hi zero;
// references will use Store-owned handles, never encoded Go pointers. The
// corresponding slot/signature carries the Wasm type for root identification.
type Value struct{ Lo, Hi uint64 }
