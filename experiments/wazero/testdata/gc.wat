(module
  (type $box (struct (field i32)))
  (func (export "run") (result i32)
    i32.const 42
    struct.new $box
    struct.get $box 0))
