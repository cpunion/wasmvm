(module
  (import "p" "tail-host" (func $relay (param externref) (result externref)))
  (memory 1)
  (data (i32.const 0) "C")
  (func (export "tail-import") (param externref) (result externref) (return_call $relay (local.get 0)))
  (func (export "ordinary-import") (param externref) (result externref) (call $relay (local.get 0)))
)
