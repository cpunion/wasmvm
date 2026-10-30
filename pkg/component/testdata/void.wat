(component
  (type $f (func))
  (import "host" (func $host (type $f)))
  (core func $host (canon lower (func $host)))
  (core module $guest
    (import "env" "host" (func $host))
    (func (export "run") call $host))
  (core instance $env (export "host" (func $host)))
  (core instance $guest (instantiate $guest (with "env" (instance $env))))
  (alias core export $guest "run" (core func $run))
  (func $run (type $f) (canon lift (core func $run)))
  (export "run" (func $run)))
