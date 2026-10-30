(component
 (type $clock (instance
  (type $definition (record (field "nanoseconds" u32) (field "seconds" u64)))
  (export "datetime" (type $datetime (eq $definition)))
  (type $read (func  (result $datetime)))
  (export "now" (func (type $read)))
  (export "resolution" (func (type $read)))))
 (import "wasi:clocks/wall-clock@0.2.12" (instance $clock (type $clock))))
