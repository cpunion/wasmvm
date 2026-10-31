(component
 (import "wasi:http/types@0.2.12" (instance $types
  (export "incoming-request" (type (sub resource)))
  (export "response-outparam" (type (sub resource)))))
 (alias export $types "incoming-request" (type $request))
 (alias export $types "response-outparam" (type $response))
 (core module $code (func (export "handle") (param i32 i32)))
 (core instance $code (instantiate $code))
 (func $handle async (param "request" (own $request)) (param "response-out" (own $response)) (canon lift (core func $code "handle")))
 (instance $handler (export "handle" (func $handle)))
 (export "wasi:http/incoming-handler@0.2.12" (instance $handler)))
