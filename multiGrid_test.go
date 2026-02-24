package dfx

import "testing"

type capturingLayout struct {
	handleParent  Component
	arrangeParent Component
}

func (cl *capturingLayout) HandleInput(state *State) {
	cl.handleParent = state.Parent
}

func (cl *capturingLayout) Arrange(_ map[string]Component, state *State) {
	cl.arrangeParent = state.Parent
}

func TestMultiGrid_LayoutStateParentIsMultiGrid(t *testing.T) {
	mg := NewMultiGrid()
	layout := &capturingLayout{}
	mg.SetLayout(layout)

	root := &Container{Visible: true}
	state := &State{Parent: root}
	mg.Draw(state)

	if layout.handleParent != mg {
		t.Fatalf("expected HandleInput parent to be multigrid, got '%T'", layout.handleParent)
	}
	if layout.arrangeParent != mg {
		t.Fatalf("expected Arrange parent to be multigrid, got '%T'", layout.arrangeParent)
	}
}
