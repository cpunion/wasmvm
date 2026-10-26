;; Authored for wasmvm. Regenerate with: python3 benchmarks/generate.py
;; The computed checksum is checked before any success output is written.
(module
  (import "wasi_snapshot_preview1" "fd_write"
    (func $fd_write (param i32 i32 i32 i32) (result i32)))
  (memory (export "memory") 1)
  (data (i32.const 16) "memory-sustained:2753322400\0a")

  (func $work (result i32)

    (local $i i32) (local $address i32) (local $x i32) (local $sum i32)
    loop $again
      local.get $i i32.const 17 i32.mul i32.const 8191 i32.and
      i32.const 4 i32.mul i32.const 4096 i32.add local.set $address
      local.get $address i32.load local.get $i i32.xor
      i32.const 2654435769 i32.add local.set $x
      local.get $address local.get $x i32.store
      local.get $sum local.get $x i32.add local.set $sum
      local.get $i i32.const 1 i32.add local.tee $i
      i32.const 3000000 i32.lt_u br_if $again
    end
    local.get $sum
)
  (func (export "_start")
    call $work i32.const 2753322400 i32.ne
    if unreachable end
    i32.const 0 i32.const 16 i32.store
    i32.const 4 i32.const 28 i32.store
    i32.const 1 i32.const 0 i32.const 1 i32.const 8 call $fd_write
    if unreachable end))
