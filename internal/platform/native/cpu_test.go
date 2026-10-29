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
		if got := x86FeatureMask(test.ecx1, test.ebx7, 0, test.xcr0); got != test.want {
			t.Fatal(test, got)
		}
	}
}

func TestAVX512RequiresCPUAndCompleteOSState(t *testing.T) {
	const ecx1 = 1<<9 | 1<<19 | 1<<26 | 1<<27 | 1<<28
	const ebx7 = 1<<5 | 1<<16 | 1<<31
	const ecx7 = 1 << 12
	for _, test := range []struct {
		name                   string
		clear1, clearB, clearC uint32
		clearState             uint64
		want                   uint32
	}{
		{name: "complete", want: 31},
		{name: "no-foundation", clearB: 1 << 16, want: 7},
		{name: "no-short-vectors", clearB: 1 << 31, want: 7},
		{name: "no-byte-popcount", clearC: 1 << 12, want: 15},
		{name: "no-avx2", clearB: 1 << 5, want: 27},
		{name: "no-xsave", clear1: 1 << 26, want: 3},
		{name: "no-osxsave", clear1: 1 << 27, want: 3},
		{name: "no-avx", clear1: 1 << 28, want: 3},
		{name: "no-xmm-state", clearState: 1 << 1, want: 3},
		{name: "no-ymm-state", clearState: 1 << 2, want: 3},
		{name: "no-opmask-state", clearState: 1 << 5, want: 7},
		{name: "no-zmm-upper-state", clearState: 1 << 6, want: 7},
		{name: "no-upper-register-state", clearState: 1 << 7, want: 7},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := x86FeatureMask(ecx1&^test.clear1, ebx7&^test.clearB, ecx7&^test.clearC, 0xe6&^test.clearState); got != test.want {
				t.Fatal(got, test.want)
			}
		})
	}
}
