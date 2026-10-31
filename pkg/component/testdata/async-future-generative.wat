(component
 (component $child
  (type $r (resource (rep i32))) (export "r" (type $r))
  (type $t (future (own 1))) (export "endpoint-type" (type $t))
  (core func $new (canon future.new $t)) (core func $drop (canon future.drop-readable $t))
  (core instance $env (export "new" (func $new)) (export "drop" (func $drop)))
  (core module $code
   (import "env" "new" (func $new (result i64))) (import "env" "drop" (func $drop (param i32)))
   (func (export "make") (result i32) call $new i32.wrap_i64)
   (func (export "take") (param i32) local.get 0 call $drop))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "make") (result $t) (canon lift (core func $code "make")))
  (func (export "take") (param "end" $t) (canon lift (core func $code "take"))))
 (instance $one (instantiate $child)) (instance $two (instantiate $child))
 (export "one" (instance $one)) (export "two" (instance $two)))
