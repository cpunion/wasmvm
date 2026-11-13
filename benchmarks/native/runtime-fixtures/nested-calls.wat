(module (func $step (param i32) (result i32)
            local.get 0 i32.const 1664525 i32.mul i32.const 1013904223 i32.add)
          (func $wrap1 (param i32) (result i32) local.get 0 call $step)
          (func $wrap2 (param i32) (result i32) local.get 0 call $wrap1)
      (func (export "run") (param i32 i32) (result i32) (local $i i32) (local $x i32)
        local.get 0 local.set $x
        (loop $loop local.get $x call $wrap2 local.set $x
          local.get $i i32.const 1 i32.add local.tee $i
          local.get 1 i32.lt_u br_if $loop)
        local.get $x))
