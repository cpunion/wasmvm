package state

// Write directly into available capacity. In the pinned LLGo release a scalar
// append can allocate a temporary argument array even when no growth is needed.
// Keep growth in the ordinary builtin and share this path across heap metadata.
func appendValue[T any](values []T, value T) []T {
	n := len(values)
	if n == cap(values) {
		return append(values, value)
	}
	values = values[:n+1]
	values[n] = value
	return values
}
