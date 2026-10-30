(component (type $t (instance
 (type (list s8))
 (type (func (param "len" u64) (result 0)))
 (export "get-random-bytes" (func (type 1)))))
 (import "wasi:random/random@0.2.12" (instance (type $t))))
