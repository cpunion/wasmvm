(component
 (type $library (component
  (export "make" (func (param "message" string) (result error-context)))
  (export "echo" (func (param "context" error-context) (result error-context)))
  (export "inspect" (func (param "context" error-context) (result string)))
  (export "release" (func))))
 (import "code" (component $code (type $library)))
 (instance $api (instantiate $code))
 (alias export $api "make" (func $make))
 (alias export $api "echo" (func $echo))
 (alias export $api "inspect" (func $inspect))
 (alias export $api "release" (func $release))
 (export "make" (func $make)) (export "echo" (func $echo))
 (export "inspect" (func $inspect)) (export "release" (func $release)))
