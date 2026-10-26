(module
  (tag $value (param i32))
  (func (export "run") (result i32)
    block $caught (result i32)
      try_table (catch $value $caught)
        i32.const 42
        throw $value
      end
      unreachable
    end))
