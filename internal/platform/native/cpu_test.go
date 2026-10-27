package native

import "testing"

func TestX86FeatureSelectionRequiresOSState(t *testing.T) {
	const extensions = 1<<9 | 1<<19
	const state = 1<<26 | 1<<27 | 1<<28
	for _, test := range []struct {
		ecx1, ebx7 uint32
		xcr0       uint64
		want       uint32
	}{
		{0, 0, 0, 0},
		{1 << 9, 0, 0, 1},
		{1 << 19, 0, 0, 2},
		{extensions, 1 << 5, 6, 3},
		{state, 0, 6, 0},
		{state, 1 << 5, 6, 4},
		{extensions | state, 1 << 5, 0, 3},
		{extensions | state, 1 << 5, 1, 3},
		{extensions | state, 1 << 5, 2, 3},
		{extensions | state, 1 << 5, 4, 3},
		{extensions | state, 1 << 5, 6, 7},
		{extensions | state, 1 << 5, 0xe7, 7},
		{(extensions | state) &^ (1 << 26), 1 << 5, 6, 3},
		{(extensions | state) &^ (1 << 27), 1 << 5, 6, 3},
		{(extensions | state) &^ (1 << 28), 1 << 5, 6, 3},
	} {
		if got := x86FeatureMask(test.ecx1, test.ebx7, test.xcr0); got != test.want {
			t.Fatal(test, got)
		}
	}
}
