(component
  (import "fail" (func $fail))
  (core func $fail (canon lower (func $fail)))
  (type $r (resource (rep i32)))
  (core func $new (canon resource.new $r))
  (core module $start
    (import "env" "new" (func $new (param i32) (result i32)))
    (import "env" "fail" (func $fail))
    (func $start i32.const 7 call $new drop call $fail)
    (start $start))
  (core instance $env (export "new" (func $new)) (export "fail" (func $fail)))
  (core instance (instantiate $start (with "env" (instance $env)))))
