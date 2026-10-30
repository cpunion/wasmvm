(component
  (core module $math
    (func (export "add") (param f64 f64) (result f64)
      local.get 0 local.get 1 f64.add))
  (core instance $math (instantiate $math))
  (alias core export $math "add" (core func $add))
  (type $add-type (func (param "left" f64) (param "right" f64) (result f64)))
  (func $add (type $add-type) (canon lift (core func $add)))
  (export "add" (func $add)))
