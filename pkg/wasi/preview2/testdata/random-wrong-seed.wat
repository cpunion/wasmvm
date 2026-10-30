(component (type $t (instance
 (type (tuple u64 u32))
 (type (func (result 0)))
 (export "insecure-seed" (func (type 1)))))
 (import "wasi:random/insecure-seed@0.2.12" (instance (type $t))))
