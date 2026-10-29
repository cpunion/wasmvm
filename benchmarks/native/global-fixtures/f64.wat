(module
  (global $g (mut f64) (f64.const 0))
  (func (export "run") (param i32 i32) (result i32) (local $i i32)
    local.get 0 i32.const 1023 i32.and f64.convert_i32_u global.set $g
    (loop $loop
      global.get $g f64.const 3 f64.add global.set $g
      local.get $i i32.const 1 i32.add local.tee $i
      local.get 1 i32.lt_u br_if $loop)
    global.get $g i32.trunc_f64_u))
