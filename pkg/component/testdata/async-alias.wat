(component
 (import "a" (type $a (sub resource)))
 (import "b" (type $b (sub resource)))
 (import "create" (func $create (result (own $b))))
 (core func $create (canon lower (func $create)))
 (core func $return (canon task.return (result (own $b))))
 (core instance $env (export "create" (func $create)) (export "return" (func $return)))
 (core module $code
  (import "env" "create" (func $create (result i32)))
  (import "env" "return" (func $return (param i32)))
  (func (export "make") call $create call $return))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "make") async (result (own $a)) (canon lift (core func $code "make") async)))
