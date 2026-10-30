(component
  (core module $state
    (global $total (mut i64) (i64.const 0))
    (func (export "destructor") (param i64)
      local.get 0 i64.const 13 i64.eq if unreachable end
      global.get $total local.get 0 i64.add global.set $total)
    (func (export "total") (result i64) global.get $total))
  (core instance $state (instantiate $state))
  (alias core export $state "destructor" (core func $destructor))
  (alias core export $state "total" (core func $total))
  (type $r (resource (rep i64) (dtor (func $destructor))))
  (export "r" (type $r))
  (type $other (resource (rep i64)))
  (core func $new (canon resource.new $r))
  (core func $rep (canon resource.rep $r))
  (core func $drop (canon resource.drop $r))
  (core func $wrong (canon resource.rep $other))
  (core module $guest
    (import "env" "new" (func $new (param i64) (result i32)))
    (import "env" "rep" (func $rep (param i32) (result i64)))
    (import "env" "drop" (func $drop (param i32)))
    (import "env" "wrong" (func $wrong (param i32) (result i64)))
    (func (export "new") (param i64) (result i32) local.get 0 call $new)
    (func (export "rep") (param i32) (result i64) local.get 0 call $rep)
    (func (export "drop") (param i32) local.get 0 call $drop)
    (func (export "wrong") (param i32) (result i64) local.get 0 call $wrong)
    (func (export "consume-owned") (param i32) (result i64)
      (local $rep i64) local.get 0 call $rep local.set $rep
      local.get 0 call $drop local.get $rep)
    (func (export "borrow-rep") (param i32) (result i32) local.get 0)
    (func (export "churn") (param $n i32) (result i32)
      (local $i i32) (local $handle i32)
      block $end loop $next
        local.get $i local.get $n i32.ge_u br_if $end
        i64.const 1 call $new local.tee $handle
        call $rep i64.const 1 i64.ne if unreachable end
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
  (alias core export $guest "consume-owned" (core func $consume-owned))
  (alias core export $guest "borrow-rep" (core func $borrow-rep))
  (alias core export $guest "churn" (core func $churn))
  (func (export "new") (param "rep" u64) (result u32) (canon lift (core func $guest-new)))
  (func (export "rep") (param "id" u32) (result u64) (canon lift (core func $guest-rep)))
  (func (export "drop") (param "id" u32) (canon lift (core func $guest-drop)))
  (func (export "wrong") (param "id" u32) (result u64) (canon lift (core func $guest-wrong)))
  (func (export "total") (result u64) (canon lift (core func $total)))
  (func (export "churn") (param "n" u32) (result u32) (canon lift (core func $churn)))
  (func (export "make-owned") (param "rep" u64) (result (own 1)) (canon lift (core func $guest-new)))
  (func (export "consume-owned") (param "value" (own 1)) (result u64) (canon lift (core func $consume-owned)))
 (func (export "borrow-rep") (param "self" (borrow 1)) (result u32) (canon lift (core func $borrow-rep))))
