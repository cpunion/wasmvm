;; Authored binary: the pinned producer does not support value text syntax.
(component (type $bytes (list u8))
 (value $constant $bytes (list 0 255 128 1))
 (import "input" (value $input (eq $constant)))
 (export "constant" (value $constant)) (export "input" (value $input)))
