(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x i32) (local $y f32)
      local.get $iterations i32.const 0 i32.le_s if unreachable end
      local.get $seed local.set $x
      loop $again
        local.get $x i32.const 1664525 i32.mul i32.const 1013904223 i32.add local.get $i i32.add i32.const 2147483647 i32.and i32.const 1073741824 i32.sub f32.convert_i32_s i32.trunc_f32_s local.set $x local.get $x local.get $x local.get $seed i32.xor local.get $i i32.xor f32.reinterpret_i32 i32.trunc_sat_f32_s i32.xor local.set $x
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end
      local.get $x))
