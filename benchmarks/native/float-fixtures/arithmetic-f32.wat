(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x f32) (local $y f32)
      local.get $iterations i32.const 0 i32.le_s if unreachable end
      local.get $seed i32.const 0x807fffff i32.and i32.const 0x3f800000 i32.or f32.reinterpret_i32 local.set $x
      loop $again
        local.get $x f32.const 1.0001220703125 f32.mul f32.const 0.125 f32.add f32.const 1.125 f32.div local.set $x
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end
      local.get $x i32.reinterpret_f32))
