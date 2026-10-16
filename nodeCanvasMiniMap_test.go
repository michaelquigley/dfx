package dfx

import (
	"math"
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
)

func TestMiniMapConfigAndLayout(t *testing.T) {
	n := normalizeMiniMapConfig(NodeCanvasMiniMapConfig{Corner: 99, Size: imgui.Vec2{X: float32(math.NaN()), Y: -1}})
	if n.Corner != NodeCanvasLowerLeft || n.Size != (imgui.Vec2{X: 240, Y: 160}) {
		t.Fatalf("normalization: %+v", n)
	}
	if got := NewNodeCanvas[int](NodeCanvasConfig{}).MiniMapConfig(); got.Corner != NodeCanvasLowerLeft || got.Size != (imgui.Vec2{X: 240, Y: 160}) {
		t.Fatalf("zero-config default: %+v", got)
	}
	style := NodeCanvasStyle{MiniMapMargin: 8, MiniMapGap: 6, MiniMapPadding: 6}
	for _, corner := range []NodeCanvasCorner{NodeCanvasLowerRight, NodeCanvasLowerLeft, NodeCanvasUpperRight, NodeCanvasUpperLeft} {
		cfg := NodeCanvasMiniMapConfig{Visible: true, Corner: corner, Size: imgui.Vec2{X: 240, Y: 160}}
		l := layoutMiniMap(imgui.Vec2{X: 10, Y: 20}, imgui.Vec2{X: 800, Y: 600}, cfg, style, canvasRect{})
		if !l.panelShown || !rectNonempty(l.inner) {
			t.Fatalf("corner %d: %+v", corner, l)
		}
		if (corner == NodeCanvasLowerRight || corner == NodeCanvasUpperRight) != (l.panel.center().X > 410) ||
			(corner == NodeCanvasLowerRight || corner == NodeCanvasLowerLeft) != (l.panel.center().Y > 320) {
			t.Fatalf("wrong corner %d: %+v", corner, l)
		}
	}
	pill := canvasRect{Min: imgui.Vec2{X: 740, Y: 570}, Max: imgui.Vec2{X: 800, Y: 600}}
	l := layoutMiniMap(imgui.Vec2{}, imgui.Vec2{X: 800, Y: 600}, NodeCanvasMiniMapConfig{Visible: true, Corner: NodeCanvasLowerRight, Size: imgui.Vec2{X: 240, Y: 160}}, style, pill)
	if !l.drawPill || l.panel.overlaps(pill) {
		t.Fatalf("pill reservation: %+v", l)
	}
	tiny := layoutMiniMap(imgui.Vec2{}, imgui.Vec2{X: 50, Y: 50}, NodeCanvasMiniMapConfig{Visible: true, Size: imgui.Vec2{X: 240, Y: 160}}, style, pill)
	if tiny.panelShown || rectNonempty(tiny.panel) {
		t.Fatalf("tiny viewport: %+v", tiny)
	}
	hidden := layoutMiniMap(imgui.Vec2{}, imgui.Vec2{X: 800, Y: 600}, NodeCanvasMiniMapConfig{Size: imgui.Vec2{X: 240, Y: 160}}, style, canvasRect{})
	if hidden.panelShown || rectNonempty(hidden.panel) {
		t.Fatalf("hidden panel: %+v", hidden)
	}
}

func TestMiniMapProjectionAndGrab(t *testing.T) {
	inner := canvasRect{Min: imgui.Vec2{X: 20, Y: 30}, Max: imgui.Vec2{X: 220, Y: 130}}
	for _, bounds := range []canvasRect{
		{Min: imgui.Vec2{X: -10000, Y: -5000}, Max: imgui.Vec2{X: 30000, Y: 15000}},
		{Min: imgui.Vec2{X: 7, Y: 9}, Max: imgui.Vec2{X: 7, Y: 9}},
	} {
		m := newMiniMapTransform(inner, bounds)
		if !m.valid || m.scale <= 0 {
			t.Fatalf("invalid transform: %+v", m)
		}
		p := bounds.center()
		q := m.unproject(m.project(p))
		if abs32(p.X-q.X) > .01 || abs32(p.Y-q.Y) > .01 {
			t.Fatalf("roundtrip %v -> %v", p, q)
		}
		if m.project(p) != inner.center() {
			t.Fatalf("center not mapped: %+v", m)
		}
	}
	if newMiniMapTransform(canvasRect{}, canvasRect{}).valid {
		t.Fatal("zero-area panel produced a transform")
	}
	m := newMiniMapTransform(inner, canvasRect{Min: imgui.Vec2{X: -100, Y: -50}, Max: imgui.Vec2{X: 100, Y: 50}})
	v := mapViewport(m, View{Zoom: 1, Pan: imgui.Vec2{X: 1000, Y: 0}}, imgui.Vec2{}, imgui.Vec2{X: 20, Y: 20})
	if v.Max.X <= v.Min.X || mapGrab(v, inner, 8) != (canvasRect{}) {
		t.Fatalf("off-map viewport/grab: %+v %+v", v, mapGrab(v, inner, 8))
	}
	small := canvasRect{Min: imgui.Vec2{X: 100, Y: 70}, Max: imgui.Vec2{X: 101, Y: 71}}
	grab := mapGrab(small, inner, 8)
	if grab.Max.X-grab.Min.X != 8 || grab.Max.Y-grab.Min.Y != 8 || small.Max.X-small.Min.X != 1 {
		t.Fatalf("visual grab changed actual viewport: %+v %+v", grab, small)
	}
}

func TestMiniMapGestureSequence(t *testing.T) {
	m := miniMapTransform{scale: 2, valid: true}
	g := &canvasGeometry[int]{view: View{Zoom: 1, Pan: imgui.Vec2{X: 10, Y: 20}}, viewport: imgui.Vec2{X: 100, Y: 80}}
	params := gestureParams{detents: []float32{1}}
	press := inputSnapshot{mouse: imgui.Vec2{X: 20, Y: 30}, leftPressed: true, leftDown: true,
		mapPress: miniMapPressViewport, mapPoint: imgui.Vec2{X: 7, Y: 9}, mapTransform: m, mapViewport: g.viewport}
	r := stepGesture(gestureState[int]{}, press, g, params)
	if r.state.kind != gestureMiniViewportArmed || r.mapNavigated || r.intents.NodeRaised != nil || r.intents.SelectionChanged != nil || r.intents.LinkCreated != nil || len(r.intents.NodesMoved) != 0 {
		t.Fatalf("arm: %+v", r)
	}
	move := inputSnapshot{mouse: imgui.Vec2{X: 40, Y: 50}, leftDown: true, leftDragging: true}
	r = stepGesture(r.state, move, g, params)
	if r.state.kind != gestureMiniViewportDrag || r.view.Pan != (imgui.Vec2{X: 0, Y: 10}) || !r.mapNavigated {
		t.Fatalf("drag: %+v", r)
	}
	g.view = r.view
	release := inputSnapshot{mouse: imgui.Vec2{X: 50, Y: 60}, leftReleased: true, leftDragging: true}
	r = stepGesture(r.state, release, g, params)
	if r.state.kind != gestureIdle || r.view.Pan != (imgui.Vec2{X: -5, Y: 5}) {
		t.Fatalf("release outside: %+v", r)
	}
	g.view = View{Zoom: 1, Pan: imgui.Vec2{X: 10, Y: 20}}
	r = stepGesture(gestureState[int]{}, press, g, params)
	r = stepGesture(r.state, inputSnapshot{mouse: press.mouse, leftReleased: true}, g, params)
	if r.view.Pan != (imgui.Vec2{X: 43, Y: 31}) || !r.mapNavigated {
		t.Fatalf("click without drag: %+v", r)
	}
	press.mapPress = miniMapPressPanel
	r = stepGesture(gestureState[int]{}, press, g, params)
	if r.view.Pan != (imgui.Vec2{X: 43, Y: 31}) || r.state.kind != gestureMiniHold {
		t.Fatalf("panel click: %+v", r)
	}
	// a lost pointer ends capture at the last committed pan. restored
	// button state cannot resume that drag without a new press.
	r = stepGesture(gestureState[int]{}, inputSnapshot{mouse: press.mouse, leftPressed: true, leftDown: true,
		mapPress: miniMapPressViewport, mapTransform: m, mapViewport: g.viewport}, g, params)
	r = stepGesture(r.state, inputSnapshot{mouse: imgui.Vec2{X: 40, Y: 50}, leftDown: true, leftDragging: true}, g, params)
	g.view = r.view
	r = stepGesture(r.state, inputSnapshot{mouseUnavailable: true, leftDown: true}, g, params)
	if r.state.kind != gestureIdle || r.view != g.view {
		t.Fatalf("lost input: %+v", r)
	}
	r = stepGesture(r.state, inputSnapshot{leftDown: true}, g, params)
	if r.state.kind != gestureIdle || r.view != g.view {
		t.Fatalf("restored button resumed capture: %+v", r)
	}
}

func TestMiniMapNavigationSupersession(t *testing.T) {
	nc := NewNodeCanvas[int](NodeCanvasConfig{})
	nc.gesture = gestureState[int]{kind: gestureMiniViewportDrag}
	nc.SetView(View{Zoom: 1.2})
	if nc.gesture.kind != gestureIdle || nc.View().Zoom != 1.2 {
		t.Fatal("SetView did not supersede map capture")
	}
	nc.retained = &canvasGeometry[int]{nodes: []nodeGeometry[int]{{id: 1, rect: canvasRect{Min: imgui.Vec2{X: 100, Y: 100}, Max: imgui.Vec2{X: 120, Y: 120}}}}}
	nc.viewport = imgui.Vec2{X: 800, Y: 600}
	nc.gesture = gestureState[int]{kind: gestureMiniHold}
	nc.CenterOn(1)
	if nc.gesture.kind != gestureIdle {
		t.Fatal("CenterOn did not supersede map capture")
	}
	nc.gesture = gestureState[int]{kind: gestureMiniViewportArmed}
	nc.ZoomToFit(1)
	if nc.gesture.kind != gestureIdle || !nc.fitPending {
		t.Fatal("ZoomToFit did not supersede map capture")
	}
}
