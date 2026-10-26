(module
  (memory 1 1)
  (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
    (local $i i32) (local $x i32) (local $tmp i32) (local $v v128) (local $wide i64)
    local.get $iterations i32.const 0 i32.le_s if unreachable end
    local.get $seed local.set $x
    loop $again
      
      i32.const 0 i32.const 0 i32.const 512 memory.fill
      i32.const 1 local.get $seed local.get $i i32.add i32.const 257 memory.fill
      i32.const 3 i32.const 1 i32.const 257 memory.copy
      i32.const 253 i64.load align=1 local.tee $wide i32.wrap_i64
      local.get $wide i64.const 32 i64.shr_u i32.wrap_i64 i32.xor
      local.get $x i32.const 3 i32.rotl i32.xor local.get $i i32.xor local.set $x

      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    local.get $x))
