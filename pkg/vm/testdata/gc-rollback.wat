(module
 (type $new (struct (field f64)))
 (type $bytes (array (mut i8)))
 (global (ref $new) (struct.new_default $new))
 (global (ref $bytes) (array.new_default $bytes (i32.const -1)))
)
