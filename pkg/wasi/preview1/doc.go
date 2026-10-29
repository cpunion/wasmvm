// Package preview1 implements the wasi_snapshot_preview1 linear-memory ABI
// on Linux and macOS. Process resources are separate from guest Stores and
// require explicit directory, stdio and network capabilities. Core Threads and
// legacy wasi-threads are separate adapters.
package preview1
