;; Authored binary: the pinned producer does not support value text syntax.
(component (value $constant u8 42) (import "input" (value $input (eq $constant)))
 (export "constant" (value $constant)) (export "input" (value $input)))
