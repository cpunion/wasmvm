package component

import "github.com/cpunion/wasmvm/internal/component/canonical"

// ByteList constructs an owned packed list<u8>. Read packed or element-wise
// byte lists with Value.ByteString. Raw bytes need not be valid UTF-8.
func ByteList(data []byte) Value { return canonical.ByteList(data) }
