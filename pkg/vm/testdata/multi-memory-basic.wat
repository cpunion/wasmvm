(module
  (memory (export "empty") 0 2)
  (memory $data (export "data") 1 2)
  (data (memory $data) (i32.const 0) "\2a\00\00\00")
  (func (export "run") (result i32) (i32.load $data (i32.const 0)))
)
