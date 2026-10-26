(module
  (memory 1 1)
  (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
    (local $i i32) (local $x i32) (local $tmp i32) (local $v v128) (local $wide i64)
    local.get $iterations i32.const 0 i32.le_s if unreachable end
    local.get $seed i32x4.splat local.set $v
    loop $again
      
      i32.const 1
      local.get $v local.get $i i32x4.splat i32x4.add
      v128.const i32x4 1 305419896 2654435769 -1 v128.xor
      local.tee $v local.get $v
      i8x16.shuffle 12 13 14 15 0 1 2 3 4 5 6 7 8 9 10 11
      local.tee $v i32.const 3 i32x4.shl
      local.get $v i32.const 29 i32x4.shr_u v128.or v128.store align=1
      i32.const 1 v128.load align=1 local.set $v

      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    local.get $v i32x4.extract_lane 0 local.get $v i32x4.extract_lane 1 i32.const 7 i32.rotl i32.add
        local.get $v i32x4.extract_lane 2 i32.xor local.get $v i32x4.extract_lane 3 i32.const 13 i32.rotl i32.xor))
