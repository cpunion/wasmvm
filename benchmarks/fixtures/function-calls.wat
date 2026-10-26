;; Authored for wasmvm. Regenerate with: python3 benchmarks/generate.py
;; The computed checksum is checked before any success output is written.
(module
  (import "wasi_snapshot_preview1" "fd_write"
    (func $fd_write (param i32 i32 i32 i32) (result i32)))
  (memory (export "memory") 1)
  (data (i32.const 16) "function-calls:811008637\0a")

  (func $mix (param $x i32) (param $i i32) (result i32)
    local.get $x local.get $i i32.add i32.const 7 i32.rotl
    i32.const 2654435769 i32.xor)

  (func $work (result i32)

    (local $i i32) (local $x i32)
    i32.const 7 local.set $x
    loop $again
      local.get $x local.get $i call $mix local.set $x
      local.get $i i32.const 1 i32.add local.tee $i
      i32.const 150000 i32.lt_u br_if $again
    end
    local.get $x
)
  (func (export "_start")
    call $work i32.const 811008637 i32.ne
    if unreachable end
    i32.const 0 i32.const 16 i32.store
    i32.const 4 i32.const 25 i32.store
    i32.const 1 i32.const 0 i32.const 1 i32.const 8 call $fd_write
    if unreachable end))
