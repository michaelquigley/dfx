package dfx

import "testing"

// TestNodeCanvas_ConstructorDefaults pins the construction-time contract
// that is cheap to verify without an imgui context: the default detent set
// and the opening view. the constructor must not read imgui state
// (no-context-at-construction), so it is safe to call from tests.
func TestNodeCanvas_ConstructorDefaults(t *testing.T) {
	nc := NewNodeCanvas[string](NodeCanvasConfig{})

	want := []float32{0.25, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1, 1.2, 1.3, 1.4, 1.5}
	if len(nc.detents) != len(want) {
		t.Fatalf("default detents = %v, want %v", nc.detents, want)
	}
	for i := range want {
		if !approx32(nc.detents[i], want[i]) {
			t.Fatalf("default detents = %v, want %v", nc.detents, want)
		}
	}
	for i, d := range nc.detents {
		if i > 0 && d < nc.detents[i-1] {
			t.Fatalf("default detents not sorted: %v", nc.detents)
		}
	}
	if !containsDetent(nc.detents, 1.0) {
		t.Fatalf("default detents must include the editing detent 1.0: %v", nc.detents)
	}

	// the canvas opens at the editing detent, not the largest configured
	// one (which is 1.5 by default).
	if !approx32(nc.view.Zoom, 1.0) {
		t.Fatalf("initial zoom = %v, want 1.0", nc.view.Zoom)
	}

	// a non-positive WheelStepsPerZoomLevel reads as 1.0.
	nc = NewNodeCanvas[string](NodeCanvasConfig{WheelStepsPerZoomLevel: 0})
	if !approx32(nc.wheelStepsPerZoomLevel, 1.0) {
		t.Fatalf("zero wheel steps = %v, want 1.0", nc.wheelStepsPerZoomLevel)
	}
}

func containsDetent(detents []float32, want float32) bool {
	for _, d := range detents {
		if approx32(d, want) {
			return true
		}
	}
	return false
}