(module
  (global $answer i32 (i32.add (i32.const 20) (i32.const 22)))
  (func $target (result i32) (global.get $answer))
  (func (export "run") (result i32) (return_call $target))
)
