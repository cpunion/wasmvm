(component (type $t (instance
 (type (list u8))
 (type (func (param "len" u32) (result 0)))
 (export "get-random-bytes" (func (type 1)))))
 (import "wasi:random/random@0.2.12" (instance (type $t))))
