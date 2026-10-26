;; Authored for wasmvm. Regenerate with: python3 benchmarks/generate.py
;; The computed checksum is checked before any success output is written.
(module
  (import "wasi_snapshot_preview1" "fd_write"
    (func $fd_write (param i32 i32 i32 i32) (result i32)))
  (memory (export "memory") 1)
  (data (i32.const 16) "startup:0\0a")

  (func $work (result i32)
    i32.const 0
)
  (func (export "_start")
    call $work i32.const 0 i32.ne
    if unreachable end
    i32.const 0 i32.const 16 i32.store
    i32.const 4 i32.const 10 i32.store
    i32.const 1 i32.const 0 i32.const 1 i32.const 8 call $fd_write
    if unreachable end))
