package component

import "github.com/cpunion/wasmvm/internal/component/model"

var byteSourceType = ValueType{typ: &model.Type{Kind: model.Stream, Element: &model.Type{Kind: model.U8}}}

// ByteStreamType returns the immutable stream<u8> schema for packed byte
// sources. It has no nominal resource bindings. NewSource still enforces the
// receiving Store's ownership, cancellation and resource limits.
func ByteStreamType() ValueType { return byteSourceType }
