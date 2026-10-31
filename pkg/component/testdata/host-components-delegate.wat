(component
 (type $inner-type (component (type $f (func (param "a" u32) (param "b" u32) (result u32))) (export "add" (func (type $f)))))
 (type $outer-type (component
   (alias outer 1 $inner-type (type $inner-type))
   (import "inner" (component (type $inner-type)))
   (type $f (func (param "a" u32) (param "b" u32) (result u32)))
   (export "add" (func (type $f)))))
 (import "inner" (component $inner (type $inner-type)))
 (import "outer" (component $outer (type $outer-type)))
 (instance $api (instantiate $outer (with "inner" (component $inner))))
 (alias export $api "add" (func $add))
 (export "add" (func $add)))
