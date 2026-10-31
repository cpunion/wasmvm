(component
  (import "observe" (func $observe (param "rep" u32)))
  (core func $observe (canon lower (func $observe)))
  (type $r (resource (rep i32) (dtor (func $observe))))
  (export "r" (type $r))
  (core func $new (canon resource.new $r))
  (core module $make
    (import "env" "new" (func $new (param i32) (result i32)))
    (func (export "make") (result i32) i32.const 42 call $new))
  (core instance $env (export "new" (func $new)))
  (core instance $make (instantiate $make (with "env" (instance $env))))
  (alias core export $make "make" (core func $make))
  (func (export "make") (result (own 2)) (canon lift (core func $make))))
