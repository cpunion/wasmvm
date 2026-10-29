(module
 (import "wasi_snapshot_preview1" "fd_read" (func $read (param i32 i32 i32 i32) (result i32)))
 (import "wasi_snapshot_preview1" "fd_write" (func $write (param i32 i32 i32 i32) (result i32)))
 (memory (export "memory") 1)
 (data (i32.const 0) "\40\00\00\00\05\00\00\00")
 (func (export "_start")
  i32.const 0 i32.const 0 i32.const 1 i32.const 16 call $read if unreachable end
  i32.const 1 i32.const 0 i32.const 1 i32.const 16 call $write if unreachable end)
 (func (export "spam")
  loop $again
   i32.const 1 i32.const 0 i32.const 1 i32.const 16 call $write if unreachable end
   br $again
  end))
