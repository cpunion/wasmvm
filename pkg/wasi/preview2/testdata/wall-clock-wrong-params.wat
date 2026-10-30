(component
 (type $clock (instance
  (type $definition (record (field "seconds" u64) (field "nanoseconds" u32)))
  (export "datetime" (type $datetime (eq $definition)))
  (type $read (func (param "unexpected" u32) (result $datetime)))
  (export "now" (func (type $read)))
  (export "resolution" (func (type $read)))))
 (import "wasi:clocks/wall-clock@0.2.12" (instance $clock (type $clock))))
