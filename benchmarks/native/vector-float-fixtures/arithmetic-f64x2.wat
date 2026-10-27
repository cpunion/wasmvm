(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x v128) (local $y v128)
      local.get $iterations i32.eqz if unreachable end
      local.get $seed i32.const 0x7fffff i32.and i64.extend_i32_u i64.const 29 i64.shl local.get $seed i32.const 0x80000000 i32.and i64.extend_i32_u i64.const 32 i64.shl i64.or i64.const 0x3ff0000000000000 i64.or i64x2.splat v128.const f64x2 0.0 0.25 f64x2.add local.set $x
      loop $again
        local.get $x v128.const f64x2 1.0001220703125 1.000244140625 f64x2.mul v128.const f64x2 0.125 0.25 f64x2.add v128.const f64x2 1.125 1.25 f64x2.div local.set $x
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end local.get $x i32x4.extract_lane 0 local.get $x i32x4.extract_lane 1 i32.xor local.get $x i32x4.extract_lane 2 i32.xor local.get $x i32x4.extract_lane 3 i32.xor))
