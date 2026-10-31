(component
  (import "host" (func $host (param "value" u32) (result u32)))
  (core module $m (func (export "trap") unreachable))
  (core instance $m (instantiate $m))
  (alias core export $m "trap" (core func $trap))
  (func (export "trap") (canon lift (core func $trap)))
  (export "host" (func $host)))
