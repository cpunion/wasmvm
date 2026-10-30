(component
  (core module $math
    (func (export "add") (param i32 i32) (result i32)
      local.get 0 local.get 1 i32.add))
  (component $child
    (alias outer 1 $math (core module $imported))
    (core instance $math (instantiate $imported))
    (alias core export $math "add" (core func $add))
    (type $f (func (param "left" u32) (param "right" u32) (result u32)))
    (func $add (type $f) (canon lift (core func $add)))
    (export "add" (func $add)))
  (instance $child (instantiate $child))
  (alias export $child "add" (func $add))
  (export "add" (func $add)))
