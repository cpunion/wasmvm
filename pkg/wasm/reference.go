package wasm

const (
	referenceExtended  ValueType = 1 << 8
	referenceIndexed   ValueType = 1 << 9
	referenceNullable  ValueType = 1 << 10
	referenceCanonical ValueType = 1 << 11
)

// FunctionReference constructs a reference to a module-local function type.
func FunctionReference(index uint32, nullable bool) ValueType {
	t := FuncRef | referenceExtended | referenceIndexed | ValueType(index)<<16
	if nullable {
		t |= referenceNullable
	}
	return t
}

// CanonicalFunctionReference identifies a function type in one Store. Obtain
// canonical types from that Store's exports; indices are not portable handles.
func CanonicalFunctionReference(index uint32, nullable bool) ValueType {
	return FunctionReference(index, nullable) | referenceCanonical
}

func (t ValueType) referenceEncoding() bool {
	if t <= 255 {
		return true
	}
	if t>>48 != 0 {
		return false
	}
	flags := t & 0xff00
	if flags&referenceExtended == 0 || flags & ^(referenceExtended|referenceIndexed|referenceNullable|referenceCanonical) != 0 {
		return false
	}
	if flags&referenceIndexed == 0 {
		return t>>16 == 0 && flags&referenceCanonical == 0 && flags&referenceNullable == 0
	}
	return byte(t) == byte(FuncRef)
}

func (t ValueType) IsFunctionReference() bool { return t.IsReference() && byte(t) == byte(FuncRef) }
func (t ValueType) IsExternalReference() bool { return t.IsReference() && byte(t) == byte(ExternRef) }
func (t ValueType) Nullable() bool {
	return t.IsReference() && (t&referenceExtended == 0 || t&referenceNullable != 0)
}
func (t ValueType) Defaultable() bool { return t.Valid() && (!t.IsReference() || t.Nullable()) }

// Storage erases static reference refinements for the shared physical frame.
// Full types remain in validation, SSA, signatures and operation schemas.
func (t ValueType) Storage() ValueType {
	if t.IsReference() {
		return ValueType(byte(t))
	}
	return t
}
func (t ValueType) Canonical() bool { return t&referenceCanonical != 0 }
func (t ValueType) TypeIndex() (uint32, bool) {
	return uint32(t >> 16), t.IsFunctionReference() && t&referenceIndexed != 0
}
func (t ValueType) NonNull() ValueType {
	if !t.IsReference() {
		return t
	}
	return t&^referenceNullable | referenceExtended
}
func (t ValueType) WithNull() ValueType {
	if !t.IsReference() {
		return t
	}
	if t&referenceIndexed == 0 {
		return ValueType(byte(t))
	}
	return t | referenceNullable
}

// ReferenceSubtype implements nullability and invariant function-reference
// types. labels optionally normalizes equivalent module-local type indices.
func ReferenceSubtype(actual, expected ValueType, labels []uint32) bool {
	if actual == expected {
		return true
	}
	if !actual.IsReference() || !expected.IsReference() || byte(actual) != byte(expected) || actual.Nullable() && !expected.Nullable() {
		return false
	}
	e, indexed := expected.TypeIndex()
	if !indexed {
		return true
	}
	a, indexed := actual.TypeIndex()
	if !indexed || actual.Canonical() != expected.Canonical() {
		return false
	}
	if !actual.Canonical() && labels != nil {
		if uint64(a) >= uint64(len(labels)) || uint64(e) >= uint64(len(labels)) {
			return false
		}
		a, e = labels[a], labels[e]
	}
	return a == e
}
