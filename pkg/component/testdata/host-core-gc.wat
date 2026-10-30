(module
 (type $base (sub (struct (field i32))))
 (type $child (sub $base (struct (field i32) (field i64))))
 (global $value (export "value") (ref $child) (struct.new $child (i32.const 42) (i64.const 9)))
 (func (export "get") (result i32) global.get $value struct.get $base 0))
