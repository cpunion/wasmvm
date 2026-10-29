(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x v128)
      local.get $iterations i32.eqz if unreachable end
      local.get $seed i32x4.splat local.set $x
      loop $again
        local.get $x v128.const i32x4 1 3 5 7 i32x4.add
        local.get $i i32x4.splat i32x4.add
        i8x16.popcnt local.set $x
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end local.get $x i32x4.extract_lane 0 local.get $x i32x4.extract_lane 1 i32.xor local.get $x i32x4.extract_lane 2 i32.xor local.get $x i32x4.extract_lane 3 i32.xor))
