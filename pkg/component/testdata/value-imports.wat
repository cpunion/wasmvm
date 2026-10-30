;; Authored binary: the pinned producer does not support value text syntax.
(component
  (type $list (list string))
  (type $tuple (tuple u8 $list))
  (import "first" (value $first $tuple))
  (import "same" (value $same (eq $first)))
  (instance $values (export "first" (value $first)) (export "same" (value $same)))
  (export "values" (instance $values)))
