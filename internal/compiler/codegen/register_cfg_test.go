package codegen

import "testing"

func TestRegisterComponentsMatchIndependentReachability(t *testing.T) {
	for mask := 0; mask < 1<<9; mask++ {
		var reach [3][3]bool
		var edges []registerCFGEdge
		for from := 0; from < 3; from++ {
			for to := 0; to < 3; to++ {
				if mask&(1<<(from*3+to)) != 0 {
					reach[from][to] = true
					edges = append(edges, registerCFGEdge{from, to})
				}
			}
		}
		for via := 0; via < 3; via++ {
			for from := 0; from < 3; from++ {
				for to := 0; to < 3; to++ {
					reach[from][to] = reach[from][to] || reach[from][via] && reach[via][to]
				}
			}
		}
		want := map[registerLoop]bool{}
		for from := 0; from < 3; from++ {
			if !reach[from][from] {
				continue
			}
			span := registerLoop{from, from}
			for to := 0; to < 3; to++ {
				if reach[from][to] && reach[to][from] {
					span.first, span.last = min(span.first, to), max(span.last, to)
				}
			}
			want[span] = true
		}
		for _, span := range registerLoopComponents(3, edges) {
			if !want[span] {
				t.Fatal(mask, "unexpected or repeated cycle", span)
			}
			delete(want, span)
		}
		if len(want) != 0 {
			t.Fatal(mask, "missing cycle", want)
		}
	}
}

func TestRegisterComponentsExcludeBackwardReturnAndKeepTraversalBounded(t *testing.T) {
	loops := registerLoopComponents(4, []registerCFGEdge{{0, 2}, {2, 3}, {3, 2}, {3, 1}})
	if len(loops) != 1 || loops[0] != (registerLoop{2, 3}) {
		t.Fatal("backward return merged into the hot loop", loops)
	}
	var chain []registerCFGEdge
	for i := 0; i < 2047; i++ {
		chain = append(chain, registerCFGEdge{i, i + 1})
	}
	if loops := registerLoopComponents(2048, chain); len(loops) != 0 {
		t.Fatal("acyclic chain became a loop", loops)
	}
	chain = append(chain, registerCFGEdge{2047, 1024})
	loops = registerLoopComponents(2048, chain)
	if len(loops) != 1 || loops[0] != (registerLoop{1024, 2047}) {
		t.Fatal("deep cycle changed", loops)
	}
	if len(registerLoopComponents(0, nil)) != 0 {
		t.Fatal("empty graph became a loop")
	}
}
