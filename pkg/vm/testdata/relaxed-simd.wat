(module
  (global $seed i32 (i32.add (i32.const 20) (i32.const 22)))
  (memory 0)
  (memory $data 1)
  (data (memory $data) (i32.const 0)
    "\01\02\03\04\01\02\03\04\01\02\03\04\01\02\03\04")
  (func $sum (result i32)
    (i32x4.extract_lane 0
      (i32x4.relaxed_dot_i8x16_i7x16_add_s
        (v128.load $data (i32.const 0))
        (i8x16.splat (i32.const 7))
        (i32x4.splat (global.get $seed)))))
  (func (export "run") (result i32) (return_call $sum))
  (func (export "repeat") (param $n i32) (result i32) (local $acc v128)
    (loop $loop
      (local.set $acc
        (i32x4.relaxed_dot_i8x16_i7x16_add_s
          (i8x16.splat (i32.const 1))
          (i8x16.splat (i32.const 7))
          (local.get $acc)))
      (local.set $n (i32.sub (local.get $n) (i32.const 1)))
      (br_if $loop (local.get $n)))
    (i32x4.extract_lane 0 (local.get $acc)))
)
