(module
  (global $g (mut i64) (i64.const 0))
  (func (export "run") (param i32 i32) (result i32) (local $i i32)
    local.get 0 i32.const 1023 i32.and i64.extend_i32_u global.set $g
    (loop $loop
      global.get $g i64.const 3 i64.add global.set $g
      local.get $i i32.const 1 i32.add local.tee $i
      local.get 1 i32.lt_u br_if $loop)
    global.get $g i32.wrap_i64))
