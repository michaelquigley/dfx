package nodecanvasstress

import (
	"reflect"
	"testing"
)

func TestGenerateDeterministicAndDisjoint(t *testing.T) {
	for _, layout := range []Layout{Grid, Clusters} {
		for _, count := range []int{0, 1, 100, 1000, 5000, 10000, 20000} {
			for _, fanout := range []int{0, 1, 4} {
				s := Settings{Count: count, Layout: layout, Spacing: 1.25, Fanout: fanout, Seed: 42}
				a := Generate(s)
				if !reflect.DeepEqual(a, Generate(s)) || len(a.Nodes) != count {
					t.Fatalf("unstable scene or count: %+v", s)
				}
				wantLinks := count * fanout
				if count <= fanout {
					wantLinks = count * (count - 1)
				}
				if len(a.Links) != wantLinks {
					t.Fatalf("%+v: links=%d, want %d", s, len(a.Links), wantLinks)
				}
				ids, pins := map[int]bool{}, map[int]int{}
				for _, n := range a.Nodes {
					for _, id := range []int{n.ID, n.Inputs[0], n.Inputs[1], n.Output} {
						if ids[id] {
							t.Fatalf("duplicate ID %d", id)
						}
						ids[id] = true
					}
					pins[n.Inputs[0]], pins[n.Inputs[1]], pins[n.Output] = n.ID, n.ID, n.ID
				}
				for _, l := range a.Links {
					if ids[l.ID] || pins[l.From] == 0 || pins[l.To] == 0 || pins[l.From] == pins[l.To] {
						t.Fatalf("invalid link %+v", l)
					}
					ids[l.ID] = true
				}
			}
		}
	}
}

func TestGenerateClampsCount(t *testing.T) {
	if got := len(Generate(Settings{Count: MaxNodes + 1}).Nodes); got != MaxNodes {
		t.Fatalf("clamped count = %d, want %d", got, MaxNodes)
	}
}

func TestGenerateSettingsAffectPositionsAndPreserveIDs(t *testing.T) {
	a := Generate(Settings{Count: 500, Layout: Grid, Spacing: 1, Seed: 7, Fanout: 4})
	b := Generate(Settings{Count: 500, Layout: Clusters, Spacing: 2, Seed: 7, Fanout: 4})
	if a.Nodes[250].Pos == b.Nodes[250].Pos || a.Nodes[250].ID != b.Nodes[250].ID || !reflect.DeepEqual(a.Links, b.Links) {
		t.Fatal("layout/spacing did not affect only placement")
	}
}
