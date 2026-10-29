(module
  (import "spectest" "tail_echo" (func $echo (param externref) (result externref)))
  (func (export "run") (param externref) (result externref) (return_call $echo (local.get 0)))
  (func $big (result i32) (local v128) (i32.const 42))
  (func (export "too-wide") (result i32) (return_call $big))
)
