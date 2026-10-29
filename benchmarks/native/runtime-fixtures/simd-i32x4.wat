(module
      (func (export "run") (param i32 i32) (result i32) (local $i i32) (local $v v128)
        local.get 0 i32x4.splat local.set $v
        (loop $loop local.get $v v128.const i32x4 1 3 5 7 i32x4.add
          v128.const i32x4 3 5 7 11 i32x4.mul
          local.get $i i32x4.splat v128.xor local.set $v
          local.get $i i32.const 1 i32.add local.tee $i
          local.get 1 i32.lt_u br_if $loop)
        local.get $v i32x4.extract_lane 0 local.get $v i32x4.extract_lane 1 i32.xor local.get $v i32x4.extract_lane 2 i32.xor local.get $v i32x4.extract_lane 3 i32.xor))
