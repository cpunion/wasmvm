(module (memory 1 1)
      (func (export "run") (param i32 i32) (result i32) (local $i i32) (local $x i32)
        local.get 0 local.set $x
        (loop $loop i32.const 256 local.get $x local.get $i i32.add i32.const 8192 memory.fill
          i32.const 272 i32.const 256 i32.const 8192 memory.copy
          local.get $x i32.const 287 i32.load8_u i32.add local.set $x
          local.get $i i32.const 1 i32.add local.tee $i
          local.get 1 i32.lt_u br_if $loop)
        local.get $x))
