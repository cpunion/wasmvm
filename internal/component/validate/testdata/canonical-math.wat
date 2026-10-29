(component
  (core module $math
    (func (export "add") (param i32 i32) (result i32)
      local.get 0 local.get 1 i32.add))
  (core instance $math (instantiate $math))
  (alias core export $math "add" (core func $add))
  (type $add-type (func (param "left" u32) (param "right" u32) (result u32)))
  (func $add (type $add-type) (canon lift (core func $add)))
  (export "add" (func $add)))
