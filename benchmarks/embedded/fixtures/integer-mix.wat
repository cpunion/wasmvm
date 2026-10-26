;; Authored for wasmvm. Regenerate: python3 benchmarks/embedded/generate.py
;; Runtime parameters determine the work and checksum; no WASI/import/start code.
(module

  (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
    (local $i i32) (local $x i32)
    local.get $iterations i32.const 0 i32.le_s if unreachable end

    local.get $seed local.set $x
    loop $again
      local.get $x local.get $i i32.xor i32.const 5 i32.rotl
      i32.const 1664525 i32.mul i32.const 1013904223 i32.add local.set $x
      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    local.get $x
))
