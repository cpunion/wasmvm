(component (type $t (instance
 (type (func (result u32)))
 (export "get-random-u64" (func (type 0)))))
 (import "wasi:random/random@0.2.12" (instance (type $t))))
