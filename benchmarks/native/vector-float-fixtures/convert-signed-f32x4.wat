(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x v128) (local $y v128)
      local.get $iterations i32.eqz if unreachable end
      local.get $seed i32x4.splat v128.const i32x4 0 2654435769 1013904242 3668340011 v128.xor local.set $x
      loop $again
        local.get $x v128.const i32x4 1664525 1664527 1664529 1664531 i32x4.mul
          v128.const i32x4 1013904223 1013904224 1013904225 1013904226 i32x4.add
          local.get $i i32x4.splat i32x4.add f32x4.convert_i32x4_s
          v128.const f32x4 0.75 0.875 1.0 1.125 f32x4.add i32x4.trunc_sat_f32x4_s local.set $x
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end local.get $x i32x4.extract_lane 0 local.get $x i32x4.extract_lane 1 i32.xor local.get $x i32x4.extract_lane 2 i32.xor local.get $x i32x4.extract_lane 3 i32.xor))
