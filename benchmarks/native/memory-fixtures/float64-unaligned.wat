(module (memory 1 1)
      (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
        (local $i i32) (local $j i32) (local $address i32) (local $x i32) (local $sum i32)
        (local $number f64) (local $total f64)
        local.get $iterations i32.const 0 i32.le_s if unreachable end
        loop $reset
          local.get $j i32.const 8 i32.mul i32.const 1 i32.add local.get $seed i32.const 255 i32.and f64.convert_i32_u f64.store align=1
          local.get $j i32.const 1 i32.add local.tee $j i32.const 2048 i32.lt_u br_if $reset
        end

        loop $again
          local.get $i i32.const 17 i32.mul i32.const 2047 i32.and
          i32.const 8 i32.mul i32.const 1 i32.add local.set $address
          local.get $address f64.load align=1 f64.const 0.125 f64.add local.set $number
          local.get $address local.get $number f64.store align=1
          local.get $total local.get $number f64.add local.set $total
          local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
        end
        local.get $total i32.trunc_f64_u))
