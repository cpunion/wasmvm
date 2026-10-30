(component
 (type $host-type (instance
  (core type $m (module (type (func (param i32 i32) (result i32))) (export "add" (func (type 0)))))
  (export "math" (core module (type $m)))))
 (import "host" (instance $host (type $host-type)))
 (alias export $host "math" (core module $m))
 (core instance $m (instantiate $m))
 (alias core export $m "add" (core func $add))
 (func (export "add") (param "a" u32) (param "b" u32) (result u32) (canon lift (core func $add))))
