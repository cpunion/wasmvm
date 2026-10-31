(module
  (type $node (struct (field i32)))
  (table (export "refs") 1 (ref null $node))
  (global (export "last") (mut (ref null $node)) (ref.null $node))
  (tag (export "event") (param i32))
  (func (export "make") (result (ref $node)) i32.const 0 struct.new $node)
  (func (export "add") (param i32 i32) (result i32) local.get 0 local.get 1 i32.add))
