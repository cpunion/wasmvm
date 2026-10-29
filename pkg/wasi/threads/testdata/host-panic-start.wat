(module
 (import "env" "memory" (memory 1 1 shared))
 (import "wasi" "thread-spawn" (func $spawn (param i32) (result i32)))
 (import "wasi_snapshot_preview1" "clock_time_get" (func $clock (param i32 i64 i32) (result i32)))
 (func $init (drop (call $clock (i32.const 1) (i64.const 0) (i32.const 0))))
 (start $init)
 (func (export "_start") (drop (call $spawn (i32.const 0))) (loop $wait (br $wait)))
 (func (export "wasi_thread_start") (param i32 i32)))
