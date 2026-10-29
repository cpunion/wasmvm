(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x v128) (local $y v128)
      local.get $iterations i32.eqz if unreachable end
      local.get $seed i32.const 0x807fffff i32.and i32.const 0x3f800000 i32.or i32x4.splat v128.const f32x4 0.0 0.25 0.5 0.75 f32x4.add local.set $x
      loop $again
        local.get $x f64x2.promote_low_f32x4
          v128.const f64x2 0.875 0.75 f64x2.mul v128.const f64x2 0.25 0.5 f64x2.add
          f32x4.demote_f64x2_zero v128.const f32x4 0.125 0.25 0.5 0.75 f32x4.add local.set $x
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end local.get $x i32x4.extract_lane 0 local.get $x i32x4.extract_lane 1 i32.xor local.get $x i32x4.extract_lane 2 i32.xor local.get $x i32x4.extract_lane 3 i32.xor))
