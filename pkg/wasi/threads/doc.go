// Package threads implements the legacy wasi.thread-spawn ABI on Linux and
// macOS. Each worker has an independent Store, instance and execution stack;
// shared linear memories and Preview 1 Process capabilities form the group.
package threads

const DefaultMaxThreads uint32 = 256
