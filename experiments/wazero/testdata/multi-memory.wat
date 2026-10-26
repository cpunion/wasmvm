(module
  (memory 1)
  (memory 1)
  (func (export "run") (result i32)
    i32.const 0 i32.const 42 i32.store 1
    i32.const 0 i32.load 1))
