(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x f64) (local $y f64)
      local.get $iterations i32.const 0 i32.le_s if unreachable end
      local.get $seed i32.const 0x7fffff i32.and i64.extend_i32_u i64.const 29 i64.shl local.get $seed i32.const 0x80000000 i32.and i64.extend_i32_u i64.const 32 i64.shl i64.or i64.const 0x3ff0000000000000 i64.or f64.reinterpret_i64 local.set $x
      loop $again
        local.get $x f64.const 0.125 f64.add f64.const 2 f64.min f64.const -2 f64.max local.set $x
        local.get $x f64.const 1.5 f64.gt if local.get $x f64.neg local.set $x end
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end
      local.get $x i64.reinterpret_f64 local.get $x i64.reinterpret_f64 i64.const 32 i64.shr_u i64.xor i32.wrap_i64))
