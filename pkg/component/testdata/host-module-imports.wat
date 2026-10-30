(component
 (core type $t (module
  (type (func (result i32))) (type (func (param i32 i32) (result i32)))
  (import "env" "seed" (func (type 0))) (export "add" (func (type 1)))))
 (import "math" (core module $m (type $t)))
 (core module $env (func (export "seed") (result i32) i32.const 10))
 (core instance $env (instantiate $env))
 (core instance $m (instantiate $m (with "env" (instance $env))))
 (alias core export $m "add" (core func $add))
 (func (export "add") (param "a" u32) (param "b" u32) (result u32) (canon lift (core func $add))))
