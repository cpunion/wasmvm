(component
 (type $clock (instance
  (export "pollable" (type $p (sub resource)))
  (type $now (func (result u64)))
  (export "now" (func (type $now)))
  (export "resolution" (func (type $now)))
  (type $subscribe (func (param "when" u32) (result (own $p))))
  (export "subscribe-instant" (func (type $subscribe)))
  (export "subscribe-duration" (func (type $subscribe)))))
 (import "wasi:clocks/monotonic-clock@0.2.12" (instance (type $clock))))
