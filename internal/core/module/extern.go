package module

// LimitsSubtype compares declared Core limits. Runtime callers substitute the
// current memory/table size for Min before checking a concrete import.
func LimitsSubtype(actual, expected Limits) bool {
	return actual.Shared == expected.Shared && actual.Index64 == expected.Index64 && actual.Min >= expected.Min &&
		(!expected.HasMax || actual.HasMax && actual.Max <= expected.Max)
}

// ExternSubtype compares validated descriptors whose type IDs and reference
// types have already been canonicalized in the same registry. Ownership and
// live runtime-object checks remain the responsibility of the caller.
func ExternSubtype(registry *TypeRegistry, actual, expected Import) bool {
	if actual.Kind != expected.Kind {
		return false
	}
	switch actual.Kind {
	case 0:
		return registry.FunctionSubtype(actual.Type, expected.Type)
	case 1:
		return actual.Table.Element == expected.Table.Element && LimitsSubtype(actual.Table.Limits, expected.Table.Limits)
	case 2:
		return LimitsSubtype(actual.Memory, expected.Memory)
	case 3:
		return actual.Global.Mutable == expected.Global.Mutable && (actual.Global.Value == expected.Global.Value ||
			!expected.Global.Mutable && registry.Subtype(actual.Global.Value, expected.Global.Value))
	case 4:
		return actual.Type == expected.Type
	}
	return false
}
