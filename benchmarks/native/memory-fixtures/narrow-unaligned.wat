(module (memory 1 1)
      (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
        (local $i i32) (local $j i32) (local $address i32) (local $x i32) (local $sum i32)
        (local $number f64) (local $total f64)
        local.get $iterations i32.const 0 i32.le_s if unreachable end
        loop $reset
          local.get $j i32.const 4 i32.mul i32.const 0 i32.store
          local.get $j i32.const 1 i32.add local.tee $j i32.const 2049 i32.lt_u br_if $reset
        end
        local.get $seed local.set $sum
        loop $again
          local.get $i i32.const 13 i32.mul i32.const 2047 i32.and
          i32.const 4 i32.mul i32.const 1 i32.add local.set $address
          local.get $address i32.load16_s align=1 local.get $seed i32.xor local.get $i i32.xor
          i32.const 31337 i32.add i32.const 65535 i32.and local.set $x
          local.get $address local.get $x i32.store16 align=1
          local.get $sum local.get $x local.get $address i32.load8_u offset=1 i32.xor i32.add local.set $sum
          local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
        end
        local.get $sum))
