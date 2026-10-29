(component
  (type $convert-type (func (param "input" u32) (result u32)))
  (import "convert" (func $convert (type $convert-type)))
  (value $input u32 42)
  (start $convert (value $input) (result (value $output)))
  (export "answer" (value $output)))
