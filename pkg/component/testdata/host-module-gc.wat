(component
 (core type $t (module
  (type $base (sub (struct (field i32))))
  (type $get (func (result i32)))
  (export "value" (global (ref null 0)))
  (export "get" (func (type 1)))))
 (import "module" (core module $m (type $t)))
 (core instance $m (instantiate $m))
 (alias core export $m "get" (core func $get))
 (func (export "get") (result u32) (canon lift (core func $get))))
