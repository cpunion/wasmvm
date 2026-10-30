(component
 (type $exit-type (instance
  (type (result u32))
  (type (func (param "status" 0)))
  (export "exit" (func (type 1)))
  (type (func (param "status-code" u8)))
  (export "exit-with-code" (func (type 2)))))
 (import "wasi:cli/exit@0.2.12" (instance $exit (type $exit-type)))
)
