(module
  (import "wasi" "thread-spawn" (func $spawn (param i32) (result i32)))
  (func (export "run") (result i32)
    i32.const 0 call $spawn))
