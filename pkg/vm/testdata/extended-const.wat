(module
  (type $result (func (result i32)))
  (import "base" "i32" (global $base32 i32))
  (import "base" "i64" (global $base64 i64))
  (memory (export "memory") 1)
  (table (export "table") 16 funcref)
  (global $x (export "x") i32
    (i32.add (i32.sub (i32.mul (global.get $base32) (i32.const 20)) (i32.const 22)) (i32.const 4)))
  (global $y (export "y") i64
    (i64.add (i64.sub (i64.mul (global.get $base64) (i64.const 20)) (i64.const 22)) (i64.const 5)))
  (func $target (type $result) (i32.const 7))
  (elem (i32.add (global.get $base32) (i32.const 1)) func $target)
  (data (i32.add (global.get $base32) (i32.const 4)) "\11\00\00\00")
  (func (export "run") (result i32)
    (i32.add
      (i32.add (global.get $x)
        (i32.load (i32.add (global.get $base32) (i32.const 4))))
      (call_indirect (type $result) (i32.add (global.get $base32) (i32.const 1)))))
)
