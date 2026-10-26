(module
  (memory i64 1)
  (func (export "run") (result i32)
    i64.const 0 i32.const 42 i32.store
    i64.const 0 i32.load))
