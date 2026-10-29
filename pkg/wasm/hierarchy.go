package wasm

// TypeHierarchy indexes a declared single-parent forest. Binary ancestor
// levels bound subtype queries by 32 steps even for hostile deep type chains.
// IDs start at one; zero is the absent-parent sentinel.
type TypeHierarchy struct {
	depth  []uint32
	levels [][]uint32
}

func (h *TypeHierarchy) Init() {
	if h.depth == nil {
		h.depth = []uint32{0}
		h.levels = [][]uint32{{0}}
	}
}
func (h *TypeHierarchy) Add(parent uint32) {
	h.Init()
	n := len(h.depth)
	d := h.depth[parent] + 1
	h.depth = append(h.depth, d)
	for uint64(1)<<uint(len(h.levels)) <= uint64(d) {
		h.levels = append(h.levels, make([]uint32, n))
	}
	h.levels[0] = append(h.levels[0], parent)
	for i := 1; i < len(h.levels); i++ {
		p := h.levels[i-1][n]
		h.levels[i] = append(h.levels[i], h.levels[i-1][p])
	}
}
func (h *TypeHierarchy) Rollback(count int) {
	h.depth = h.depth[:count+1]
	for i := range h.levels {
		h.levels[i] = h.levels[i][:count+1]
	}
}
func (h *TypeHierarchy) IsSubtype(actual, expected uint32) bool {
	if actual == expected {
		return true
	}
	if expected == 0 || uint64(actual) >= uint64(len(h.depth)) || uint64(expected) >= uint64(len(h.depth)) || h.depth[actual] < h.depth[expected] {
		return false
	}
	diff := h.depth[actual] - h.depth[expected]
	for level := 0; diff != 0; level++ {
		if diff&1 != 0 {
			actual = h.levels[level][actual]
		}
		diff >>= 1
	}
	return actual == expected
}
func (h *TypeHierarchy) ReferenceSubtype(actual, expected ValueType, labels []uint32) bool {
	return referenceSubtype(actual, expected, labels, h.IsSubtype)
}
