package dfx

import "testing"

// the MinWidth clamp is applied last: no sequence of drags in no container, however narrow, takes a panel below it.
func TestHCollapseClampKeepsMinWidth(t *testing.T) {
	h := NewHCollapse(nil, HCollapseConfig{ExpandedWidth: 200, MinWidth: 36, Expanded: true, Resizable: true})
	container := h.MinWidth + 30 // narrower than MinWidth+50, where the container clamp alone would go below MinWidth
	for _, delta := range []float32{-500, 10, -3, 400, -1000, 0, 25} {
		h.applyResize(delta, container)
		if h.CurrentWidth < h.MinWidth || h.ExpandedWidth < h.MinWidth {
			t.Fatalf("after a drag of %v in a %v container: current %v, expanded %v, below MinWidth %v",
				delta, container, h.CurrentWidth, h.ExpandedWidth, h.MinWidth)
		}
	}
	for _, c := range []float32{0, 1, 20} {
		h.applyResize(-50, c)
		if h.CurrentWidth < h.MinWidth {
			t.Errorf("container %v: width %v below MinWidth", c, h.CurrentWidth)
		}
	}
}

func TestHCollapseClampBounds(t *testing.T) {
	h := NewHCollapse(nil, HCollapseConfig{ExpandedWidth: 200, MinWidth: 36, MaxWidth: 400, Expanded: true})
	if w := h.clampWidth(1000, 2000); w != 400 {
		t.Errorf("max clamp = %v", w)
	}
	if w := h.clampWidth(1000, 300); w != 250 {
		t.Errorf("container clamp = %v", w)
	}
	if w := h.clampWidth(10, 2000); w != 36 {
		t.Errorf("min clamp = %v", w)
	}
}

// a right-anchored panel's handle is on its left edge: a drag toward the interior (right) shrinks it, away (left)
// widens it; a left-anchored panel does the opposite.
func TestHCollapseAnchorDragDirection(t *testing.T) {
	right := NewHCollapse(nil, HCollapseConfig{ExpandedWidth: 300, Expanded: true, Anchor: AnchorRight})
	right.applyResize(20, 2000)
	if right.CurrentWidth != 280 || right.ExpandedWidth != 280 {
		t.Errorf("right-anchored drag right: %v, want 280", right.CurrentWidth)
	}
	right.applyResize(-50, 2000)
	if right.CurrentWidth != 330 {
		t.Errorf("right-anchored drag left: %v, want 330", right.CurrentWidth)
	}

	left := NewHCollapse(nil, HCollapseConfig{ExpandedWidth: 300, Expanded: true})
	left.applyResize(20, 2000)
	if left.CurrentWidth != 320 {
		t.Errorf("left-anchored drag right: %v, want 320", left.CurrentWidth)
	}
	if left.Anchor != AnchorLeft {
		t.Error("the default anchor is not AnchorLeft")
	}
}
