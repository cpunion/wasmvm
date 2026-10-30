(component
  (type $r (resource (rep i32)))
  (component $roundtrip
    (import "r" (type $r (sub resource)))
    (export "r" (type $r)))
  (instance $roundtrip (instantiate $roundtrip (with "r" (type $r))))
  (alias export $roundtrip "r" (type $roundtripped))
  (core func $new (canon resource.new $roundtripped))
  (core func $rep (canon resource.rep $roundtripped))
  (core func $drop (canon resource.drop $roundtripped))
  (core module $guest
    (import "env" "new" (func $new (param i32) (result i32)))
    (import "env" "rep" (func $rep (param i32) (result i32)))
    (import "env" "drop" (func $drop (param i32)))
    (func (export "run") (param i32) (result i32)
      (local $id i32) (local $rep i32)
      local.get 0 call $new local.tee $id call $rep local.set $rep
      local.get $id call $drop local.get $rep))
  (core instance $env (export "new" (func $new)) (export "rep" (func $rep)) (export "drop" (func $drop)))
  (core instance $guest (instantiate $guest (with "env" (instance $env))))
  (alias core export $guest "run" (core func $run))
  (func (export "run") (param "rep" u32) (result u32) (canon lift (core func $run))))
