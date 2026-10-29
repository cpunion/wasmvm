(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x v128) (local $y v128)
      local.get $iterations i32.eqz if unreachable end
      local.get $seed i32.const 0x807fffff i32.and i32.const 0x3f800000 i32.or i32x4.splat v128.const f32x4 0.0 0.25 0.5 0.75 f32x4.add local.set $x
      loop $again
        local.get $x f32x4.ceil local.get $x f32x4.floor f32x4.add
          local.get $x f32x4.trunc f32x4.add local.get $x f32x4.nearest f32x4.add local.set $y
          local.get $x local.get $x f32x4.mul v128.const f32x4 0.5 0.5 0.5 0.5 f32x4.add f32x4.sqrt
          local.get $y v128.const f32x4 8 9 10 11 f32x4.add f32x4.div v128.const f32x4 0.25 0.375 0.5 0.625 f32x4.add f32x4.abs local.set $x
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end local.get $x i32x4.extract_lane 0 local.get $x i32x4.extract_lane 1 i32.xor local.get $x i32x4.extract_lane 2 i32.xor local.get $x i32x4.extract_lane 3 i32.xor))
