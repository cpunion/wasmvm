(component
  (type $code (component
    (import "observe" (func (param "rep" u32)))
    (export "r" (type $r (sub resource)))
    (type $make (func (result (own $r))))
    (export "make" (func (type $make)))))
  (import "observe" (func $observe (param "rep" u32)))
  (import "fail" (func $fail))
  (import "code" (component $code (type $code)))
  (instance $api (instantiate $code (with "observe" (func $observe))))
  (alias export $api "r" (type $r))
  (alias export $api "make" (func $make))
  (export "r" (type $r)))
