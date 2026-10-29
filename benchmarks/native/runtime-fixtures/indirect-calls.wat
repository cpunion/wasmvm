(module (type $step (func (param i32) (result i32)))
          (func $a (type $step) local.get 0 i32.const 1664525 i32.mul i32.const 1013904223 i32.add)
          (func $b (type $step) local.get 0 i32.const 13 i32.rotl i32.const 0x9e3779b9 i32.xor)
          (table 2 funcref) (elem (i32.const 0) $a $b)
      (func (export "run") (param i32 i32) (result i32) (local $i i32) (local $x i32)
        local.get 0 local.set $x
        (loop $loop local.get $x local.get $i i32.const 1 i32.and call_indirect (type $step) local.set $x
          local.get $i i32.const 1 i32.add local.tee $i
          local.get 1 i32.lt_u br_if $loop)
        local.get $x))
