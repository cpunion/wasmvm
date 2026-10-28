(module
  (global $g (mut v128) (v128.const i32x4 0 0 0 0))
  (func (export "run") (param i32 i32) (result i32) (local $i i32)
    local.get 0 i32.const 1023 i32.and i32x4.splat global.set $g
    (loop $loop
      global.get $g v128.const i32x4 1 3 5 7 i32x4.add global.set $g
      local.get $i i32.const 1 i32.add local.tee $i
      local.get 1 i32.lt_u br_if $loop)
    global.get $g i32x4.extract_lane 0 global.get $g i32x4.extract_lane 1 i32.xor global.get $g i32x4.extract_lane 2 i32.xor global.get $g i32x4.extract_lane 3 i32.xor))
