(module
  (import "host" "increment" (func $increment (param i32) (result i32)))
  (func (export "run") (result i32)
    i32.const 41 call $increment))
