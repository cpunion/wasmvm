(component
  (type $ft (func (param "a" u32) (param "b" u32) (result u32)))
  (import "function-type" (type $imported-ft (eq $ft)))
  (type $it (instance
    (alias outer 1 $imported-ft (type $ft))
    (export "add" (func (type $ft)))))
  (import "instance-type" (type $imported-it (eq $it)))
  (type $ct (component
    (alias outer 1 $imported-it (type $it))
    (export "api" (instance (type $it)))))
  (import "component-type" (type $imported-ct (eq $ct)))
  (core module $add
    (func (export "add") (param i32 i32) (result i32)
      local.get 0 local.get 1 i32.add))
  (component $child
    (alias outer 1 $add (core module $add))
    (alias outer 1 $imported-ft (type $ft))
    (alias outer 1 $imported-it (type $it))
    (core instance $add (instantiate $add))
    (alias core export $add "add" (core func $add))
    (func $add (type $ft) (canon lift (core func $add)))
    (instance $api (export "add" (func $add)))
    (export "api" (instance $api) (instance (type $it))))
  (export "child" (component $child) (component (type $imported-ct)))
  (instance $child (instantiate $child))
  (alias export $child "api" (instance $api))
  (alias export $api "add" (func $add))
  (export "add" (func $add))
  (export "api" (instance $api)))
