(module
  (type $unary (func (param i32) (result i32)))
  (func $increment (type $unary)
    local.get 0 i32.const 1 i32.add)
  (elem declare func $increment)
  (func (export "run") (result i32)
    i32.const 41 ref.func $increment call_ref $unary))
