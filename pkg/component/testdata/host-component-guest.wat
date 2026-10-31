(component
  (core module $m
    (func (export "add") (param i32 i32) (result i32) local.get 0 local.get 1 i32.add))
  (core instance $m (instantiate $m))
  (alias core export $m "add" (core func $add))
  (func $add (param "a" u32) (param "b" u32) (result u32) (canon lift (core func $add)))
  (export "add" (func $add))
  (export "extra" (func $add)))
