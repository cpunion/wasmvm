(component
  (type $echo-type (func (param "text" string) (result string)))
  (import "echo" (func $host (type $echo-type)))
  (core module $memory
    (memory (export "memory") 1)
    (global $next (mut i32) (i32.const 64))
    (func (export "realloc") (param i32 i32 i32 i32) (result i32)
      (local $ptr i32)
      global.get $next local.get 2 i32.const 1 i32.sub i32.add
      i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr
      local.get 3 i32.add global.set $next
      local.get $ptr))
  (core instance $memory (instantiate $memory))
  (alias core export $memory "memory" (core memory $memory-export))
  (alias core export $memory "realloc" (core func $realloc))
  (core func $lower (canon lower (func $host)
    (memory $memory-export) (realloc $realloc) string-encoding=utf16))
  (core module $guest
    (import "env" "memory" (memory 1))
    (import "env" "echo" (func $echo (param i32 i32 i32)))
    (global $calls (mut i32) (i32.const 0))
    (func (export "echo") (param i32 i32) (result i32)
      local.get 0 local.get 1 i32.const 8 call $echo i32.const 8)
    (func (export "post") (param i32)
      global.get $calls i32.const 1 i32.add global.set $calls)
    (func (export "calls") (result i32) global.get $calls))
  (core instance $env
    (export "memory" (memory $memory-export))
    (export "echo" (func $lower)))
  (core instance $guest (instantiate $guest (with "env" (instance $env))))
  (alias core export $guest "echo" (core func $echo))
  (alias core export $guest "post" (core func $post))
  (alias core export $guest "calls" (core func $calls))
  (func $echo (type $echo-type) (canon lift (core func $echo)
    (memory $memory-export) (realloc $realloc) string-encoding=utf16 (post-return $post)))
  (type $calls-type (func (result u32)))
  (func $calls (type $calls-type) (canon lift (core func $calls)))
  (export "echo" (func $echo))
  (export "calls" (func $calls)))
