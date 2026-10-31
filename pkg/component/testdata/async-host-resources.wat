(component
 (import "r" (type $r (sub resource)))
 (import "pass" (func $pass async (param "arg" (own $r)) (result (own $r))))
 (import "borrow" (func $borrow async (param "arg" (borrow $r)) (result u32)))
 (import "make" (func $make async (result (own $r))))
 (type $blob (list u8)) (export "blob-type" (type $blob))
 (import "blob" (func $blob async (param "arg" $blob) (result $blob)))
 (export "pass" (func $pass)) (export "borrow" (func $borrow))
 (export "make" (func $make)) (export "blob" (func $blob)))
