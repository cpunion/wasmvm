(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x i64) (local $y f64)
      local.get $iterations i32.const 0 i32.le_s if unreachable end
      local.get $seed i64.extend_i32_u i64.const 4294967297 i64.mul local.set $x
      loop $again
        local.get $x i64.const 6364136223846793005 i64.mul i64.const 1013904223 i64.add local.get $i i64.extend_i32_u i64.add i64.const 4611686018427387903 i64.and local.get $i i64.extend_i32_u i64.const 1 i64.and i64.const 63 i64.shl i64.or f64.convert_i64_u i64.trunc_f64_u local.set $x local.get $x local.get $x local.get $seed i64.extend_i32_u i64.xor local.get $i i64.extend_i32_u i64.xor f64.reinterpret_i64 i64.trunc_sat_f64_u i64.xor local.set $x
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end
      local.get $x local.get $x i64.const 32 i64.shr_u i64.xor i32.wrap_i64))
