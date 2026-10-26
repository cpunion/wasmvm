(module (memory 1 65)
          (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
            i32.const 64 memory.grow i32.const -1 i32.ne
            if
              i32.const 65536 i32.load if unreachable end
            end
            memory.size i32.const 65 i32.ne if unreachable end
            i32.const 65536 local.get $seed local.get $iterations i32.xor i32.store
            i32.const 65536 i32.load))
