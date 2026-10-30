(component
  (core module $math
    (func (export "add") (param f32 f32) (result f32)
      local.get 0 local.get 1 f32.add))
  (core instance $math (instantiate $math))
  (alias core export $math "add" (core func $add))
  (type $add-type (func (param "left" f32) (param "right" f32) (result f32)))
  (func $add (type $add-type) (canon lift (core func $add)))
  (export "add" (func $add)))
