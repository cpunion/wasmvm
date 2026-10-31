(component
 (type $inner-type (component (type $f (func (param "a" u32) (param "b" u32) (result u32))) (export "add" (func (type $f)))))
 (import "inner" (component $inner (type $inner-type)))
 (instance $api (instantiate $inner))
 (alias export $api "add" (func $add))
 (export "add" (func $add)))
