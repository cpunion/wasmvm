(module
  
  (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
    (local $i i32) (local $x i32) (local $tmp i32) (local $v v128) (local $wide i64)
    local.get $iterations i32.const 0 i32.le_s if unreachable end
    local.get $seed local.set $x
    loop $again
      
      local.get $x i32.const 8388607 i32.and i32.const 1065353216 i32.or
      f32.reinterpret_i32 f32.const 1.5 f32.mul f32.const 0.25 f32.add i32.reinterpret_f32
      local.get $i i32.const 255 i32.and i32.const 1 i32.add
      f64.convert_i32_u f64.sqrt f64.const 4096 f64.mul i32.trunc_sat_f64_u
      i32.xor local.get $i i32.xor local.set $x

      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    local.get $x))
