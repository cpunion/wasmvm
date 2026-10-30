(component
 (core type $t (module
  (type (func (param i32) (result i32))) (type (func (param i32)))
  (export "apply" (func (type 0)))
  (export "memory" (memory i64 1 4))
  (export "table" (table 1 4 funcref))
  (export "global" (global i32))
  (export "tag" (tag (type 1)))))
 (import "module" (core module $m (type $t)))
 (core instance $m (instantiate $m))
 (alias core export $m "apply" (core func $apply))
 (func (export "apply") (param "value" u32) (result u32) (canon lift (core func $apply))))
