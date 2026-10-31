(component
  (type $host (instance
    (type $code (component
      (type $f (func (param "a" u32) (param "b" u32) (result u32)))
      (export "add" (func (type $f)))))
    (export "code" (component (type $code)))))
  (import "host" (instance $host (type $host)))
  (alias export $host "code" (component $code))
  (instance $api (instantiate $code))
  (alias export $api "add" (func $add))
  (export "add" (func $add)))
