(module
  (func (export "run") (result i32)
    v128.const i32x4 0 0 0 0
    v128.const i32x4 0 0 0 0
    i8x16.relaxed_swizzle
    i32x4.extract_lane 0))
