(component
  (core module $state
    (global $total (mut i32) (i32.const 0))
    (func (export "destructor") (param i32)
      local.get 0 i32.const 13 i32.eq if unreachable end
      global.get $total local.get 0 i32.add global.set $total)
    (func (export "total") (result i32) global.get $total))
  (core instance $state (instantiate $state))
  (alias core export $state "destructor" (core func $destructor))
  (alias core export $state "total" (core func $total))
  (type $r (resource (rep i32) (dtor (func $destructor))))
  (type $other (resource (rep i32)))
  (core func $new (canon resource.new $r))
  (core func $rep (canon resource.rep $r))
  (core func $drop (canon resource.drop $r))
  (core func $wrong (canon resource.rep $other))
  (core module $guest
    (import "env" "new" (func $new (param i32) (result i32)))
    (import "env" "rep" (func $rep (param i32) (result i32)))
    (import "env" "drop" (func $drop (param i32)))
    (import "env" "wrong" (func $wrong (param i32) (result i32)))
    (func (export "new") (param i32) (result i32) local.get 0 call $new)
    (func (export "rep") (param i32) (result i32) local.get 0 call $rep)
    (func (export "drop") (param i32) local.get 0 call $drop)
    (func (export "wrong") (param i32) (result i32) local.get 0 call $wrong)
    (func (export "churn") (param $n i32) (result i32)
      (local $i i32) (local $handle i32)
      block $end loop $next
        local.get $i local.get $n i32.ge_u br_if $end
        i32.const 1 call $new local.tee $handle
        call $rep i32.const 1 i32.ne if unreachable end
        local.get $handle call $drop
        local.get $i i32.const 1 i32.add local.set $i br $next
      end end local.get $i))
  (core instance $env
    (export "new" (func $new)) (export "rep" (func $rep))
    (export "drop" (func $drop)) (export "wrong" (func $wrong)))
  (core instance $guest (instantiate $guest (with "env" (instance $env))))
  (alias core export $guest "new" (core func $guest-new))
  (alias core export $guest "rep" (core func $guest-rep))
  (alias core export $guest "drop" (core func $guest-drop))
  (alias core export $guest "wrong" (core func $guest-wrong))
  (alias core export $guest "churn" (core func $churn))
  (func (export "new") (param "rep" u32) (result u32) (canon lift (core func $guest-new)))
  (func (export "rep") (param "id" u32) (result u32) (canon lift (core func $guest-rep)))
  (func (export "drop") (param "id" u32) (canon lift (core func $guest-drop)))
  (func (export "wrong") (param "id" u32) (result u32) (canon lift (core func $guest-wrong)))
  (func (export "total") (result u32) (canon lift (core func $total)))
  (func (export "churn") (param "n" u32) (result u32) (canon lift (core func $churn))))
