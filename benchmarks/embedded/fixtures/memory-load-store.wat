;; Authored for wasmvm. Regenerate: python3 benchmarks/embedded/generate.py
;; Runtime parameters determine the work and checksum; no WASI/import/start code.
(module
  (memory 1 1)

  (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
    (local $clear i32) (local $i i32) (local $address i32) (local $x i32) (local $sum i32)
    local.get $iterations i32.const 0 i32.le_s if unreachable end

    ;; Exactly 8192 i32 stores reset the 32 KiB working region on every call.
    ;; This cost is included in run; it is not charged to instantiation alone.
    loop $reset
      local.get $clear i32.const 0 i32.store
      local.get $clear i32.const 4 i32.add local.tee $clear
      i32.const 32768 i32.lt_u br_if $reset
    end
    local.get $seed local.set $sum
    loop $again
      local.get $i i32.const 17 i32.mul i32.const 8191 i32.and
      i32.const 4 i32.mul local.set $address
      local.get $address i32.load local.get $i i32.xor local.get $seed i32.xor
      i32.const 2654435769 i32.add local.set $x
      local.get $address local.get $x i32.store
      local.get $sum local.get $x i32.add local.set $sum
      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    local.get $sum
))
