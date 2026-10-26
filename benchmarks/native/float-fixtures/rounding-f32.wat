(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x f32) (local $y f32)
      local.get $iterations i32.const 0 i32.le_s if unreachable end
      local.get $seed i32.const 0x807fffff i32.and i32.const 0x3f800000 i32.or f32.reinterpret_i32 local.set $x
      loop $again
        local.get $x f32.ceil local.get $x f32.floor f32.add
        local.get $x f32.trunc f32.add local.get $x f32.nearest f32.add local.set $y
        local.get $x local.get $x f32.mul f32.const 0.5 f32.add f32.sqrt
        local.get $y f32.const 8 f32.add f32.div f32.const 0.25 f32.add local.set $x
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end
      local.get $x i32.reinterpret_f32))
