(module (global $g (mut i32) (i32.const 0)) (func $target)
          (table 2 funcref) (elem declare func $target)
      (func (export "run") (param i32 i32) (result i32) (local $i i32) (local $x i32)
        local.get 0 global.set $g
        (loop $loop local.get $i i32.const 1 i32.and ref.func $target table.set 0
          global.get $g local.get $i i32.add
          local.get $i i32.const 1 i32.and table.get 0 ref.is_null i32.eqz
          i32.add global.set $g
          local.get $i i32.const 1 i32.add local.tee $i
          local.get 1 i32.lt_u br_if $loop)
        global.get $g))
