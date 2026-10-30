(component
 (type $environment-type (instance
  (type (list string))
  (type (func (result u32)))
  (export "get-arguments" (func (type 1)))
  (type (tuple string string))
  (type (list 2))
  (type (func (result 3)))
  (export "get-environment" (func (type 4)))
  (type (option string))
  (type (func (result 5)))
  (export "initial-cwd" (func (type 6)))))
 (import "wasi:cli/environment@0.2.12" (instance $environment (type $environment-type)))
)
