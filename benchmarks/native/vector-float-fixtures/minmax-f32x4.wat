(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x v128) (local $y v128)
      local.get $iterations i32.eqz if unreachable end
      local.get $seed i32.const 0x807fffff i32.and i32.const 0x3f800000 i32.or i32x4.splat v128.const f32x4 0.0 0.25 0.5 0.75 f32x4.add local.set $x
      loop $again
        local.get $x v128.const f32x4 0.125 0.25 0.375 0.5 f32x4.add
          v128.const f32x4 2 3 4 5 f32x4.min v128.const f32x4 -2 -3 -4 -5 f32x4.max
          v128.const f32x4 1.75 2.0 2.25 2.5 f32x4.pmin v128.const f32x4 -1.75 -2.0 -2.25 -2.5 f32x4.pmax local.set $x
          local.get $x f32x4.neg local.get $x local.get $x v128.const f32x4 1.5 1.75 2.0 2.25 f32x4.gt v128.bitselect local.set $x
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end local.get $x i32x4.extract_lane 0 local.get $x i32x4.extract_lane 1 i32.xor local.get $x i32x4.extract_lane 2 i32.xor local.get $x i32x4.extract_lane 3 i32.xor))
