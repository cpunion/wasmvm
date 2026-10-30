(component
  (type $types (instance
    (type $ft (func (param "a" u32) (param "b" u32) (result u32)))
    (export "function-type" (type (eq $ft)))
    (type $it (instance))
    (export "instance-type" (type (eq $it)))
    (type $ct (component))
    (export "component-type" (type (eq $ct)))))
  (import "types" (instance $types (type $types)))
  (alias export $types "function-type" (type $ft))
  (core module $add
    (func (export "add") (param i32 i32) (result i32)
      local.get 0 local.get 1 i32.add))
  (core instance $add (instantiate $add))
  (alias core export $add "add" (core func $add))
  (func (export "add") (type $ft) (canon lift (core func $add))))
