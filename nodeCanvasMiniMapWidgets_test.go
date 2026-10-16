package dfx

import (
	"runtime"
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
)

func miniNodePosition(nc *NodeCanvas[string], screen imgui.Vec2) imgui.Vec2 {
	return canvasFromScreen(imgui.Vec2{X: screen.X - nc.style.NodePadding, Y: screen.Y - nc.style.NodePadding}, nc.frameView, nc.origin)
}

func TestMiniMapNavigationAtZoomDetents(t *testing.T) {
	for _, zoom := range []float32{.25, 1.5} {
		h := newCanvasWidgetTest(t)
		h.nc.SetView(View{Zoom: zoom})
		h.nc.SetMiniMapVisible(true)
		draw := func() {
			h.nc.Node("one", imgui.Vec2{X: 100, Y: 100}, NodeFlags{}, func(n *NodeContext[string]) { n.Label("one") })
		}
		h.frame(draw)
		h.frame(draw)
		p := h.nc.miniFrame.inner.Min
		h.point(imgui.Vec2{X: p.X + 10, Y: p.Y + 10})
		h.frame(draw)
		h.io.AddMouseButtonEvent(0, true)
		h.frame(draw)
		h.io.AddMouseButtonEvent(0, false)
		h.frame(draw)
		if h.nc.View().Zoom != zoom || h.nc.View().Pan == (imgui.Vec2{}) {
			t.Fatalf("detent %v navigation: %+v", zoom, h.nc.View())
		}
	}
}

func TestMiniMapIndependentCanvases(t *testing.T) {
	runtime.LockOSThread()
	ctx := imgui.CreateContext()
	defer func() { imgui.DestroyContextV(ctx); runtime.UnlockOSThread() }()
	io := imgui.CurrentIO()
	io.SetIniFilename("")
	io.SetDisplaySize(imgui.Vec2{X: 1000, Y: 800})
	io.SetDeltaTime(1.0 / 60)
	io.SetBackendFlags(imgui.BackendFlagsRendererHasTextures)
	left := NewNodeCanvas[int](NodeCanvasConfig{MiniMap: NodeCanvasMiniMapConfig{Visible: true}})
	right := NewNodeCanvas[int](NodeCanvasConfig{MiniMap: NodeCanvasMiniMapConfig{Visible: true}})
	defer left.Destroy()
	defer right.Destroy()
	frame := func() {
		imgui.NewFrame()
		imgui.SetNextWindowPos(imgui.Vec2{})
		imgui.SetNextWindowSize(imgui.Vec2{X: 1000, Y: 800})
		imgui.BeginV("two", nil, imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)
		for i, nc := range []*NodeCanvas[int]{left, right} {
			imgui.SetCursorScreenPos(imgui.Vec2{X: float32(i * 500), Y: 0})
			nc.Begin(&State{Size: imgui.Vec2{X: 480, Y: 700}})
			nc.Node(i+1, imgui.Vec2{X: 100, Y: 100}, NodeFlags{}, func(n *NodeContext[int]) { n.Label("one") })
			nc.End()
		}
		imgui.End()
		imgui.Render()
	}
	frame()
	frame()
	p := left.miniFrame.inner.Min
	io.AddMousePosEvent(p.X+10, p.Y+10)
	frame()
	io.AddMouseButtonEvent(0, true)
	frame()
	io.AddMouseButtonEvent(0, false)
	frame()
	if left.View().Pan == (imgui.Vec2{}) || right.View().Pan != (imgui.Vec2{}) {
		t.Fatalf("map input crossed canvases: left=%+v right=%+v", left.View(), right.View())
	}
}

func TestMiniMapOverlayBlocksCoveredWidgets(t *testing.T) {
	h := newCanvasWidgetTest(t)
	h.nc.SetMiniMapVisible(true)
	panelValue := float32(.5)
	var panelWidget canvasRect
	draw := func() {
		l := h.nc.miniFrame
		h.nc.Node("panel", miniNodePosition(h.nc, l.panel.Min), NodeFlags{}, func(n *NodeContext[string]) {
			imgui.PushItemWidth(100)
			panelValue, _ = WheelSlider("covered", panelValue, 0, 1, 100, "%.2f", imgui.SliderFlagsNone)
			panelWidget = canvasRect{Min: imgui.ItemRectMin(), Max: imgui.ItemRectMax()}
			imgui.PopItemWidth()
		})
	}
	h.frame(draw)
	h.point(panelWidget.center())
	h.frame(draw)
	h.io.AddKeyEvent(imgui.ModCtrl, true)
	h.io.AddMouseWheelEvent(0, 1)
	intents := h.frame(draw)
	if panelValue != .5 || h.nc.View().Zoom != 1 || intents.NodeRaised != nil {
		t.Fatalf("wheel leaked through panel: value=%v view=%+v intents=%+v", panelValue, h.nc.View(), intents)
	}
	h.io.AddKeyEvent(imgui.ModCtrl, false)
	h.io.AddMouseButtonEvent(0, true)
	intents = h.frame(draw)
	h.io.AddMouseButtonEvent(0, false)
	h.frame(draw)
	if panelValue != .5 || intents.NodeRaised != nil || intents.SelectionChanged != nil || h.nc.gesture.kind != gestureIdle {
		t.Fatalf("press leaked through panel: value=%v intents=%+v", panelValue, intents)
	}
	if h.nc.View().Pan == (imgui.Vec2{}) {
		t.Fatal("map press did not navigate")
	}
}

func TestMiniMapFirstAppearanceAndWidgetCapture(t *testing.T) {
	h := newCanvasWidgetTest(t)
	value := float32(.5)
	var slider canvasRect
	draw := func() {
		slider = h.slider("active", imgui.Vec2{X: 100, Y: 100}, &value)
	}
	h.frame(draw)
	h.nc.SetMiniMapVisible(true)
	// the new panel reserves input immediately, but its previous transform
	// has not yet been displayed, so this first press cannot navigate.
	panel := layoutMiniMap(h.nc.origin, h.nc.viewport, h.nc.MiniMapConfig(), h.nc.style, h.nc.zoomOverlayRect()).panel
	h.point(panel.center())
	h.io.AddMouseButtonEvent(0, true)
	h.frame(draw)
	if h.nc.View().Pan != (imgui.Vec2{}) || h.nc.gesture.kind != gestureIdle {
		t.Fatalf("first-appearance press navigated: %+v", h.nc.View())
	}
	h.io.AddMouseButtonEvent(0, false)
	h.frame(draw)
	// a slider captured outside the map keeps its active ID while the
	// pointer crosses the map; the map never takes the gesture.
	h.point(imgui.Vec2{X: slider.Min.X + 20, Y: slider.center().Y})
	h.frame(draw)
	h.io.AddMouseButtonEvent(0, true)
	h.frame(draw)
	if !imgui.IsAnyItemActive() {
		t.Fatal("slider did not capture")
	}
	h.point(panel.center())
	h.frame(draw)
	if !imgui.IsAnyItemActive() || h.nc.gesture.isMiniMap() {
		t.Fatal("map stole active widget capture")
	}
	h.io.AddMouseButtonEvent(0, false)
	h.frame(draw)
}

func TestMiniMapPopupOverPanelKeepsPopupInput(t *testing.T) {
	h := newCanvasWidgetTest(t)
	h.nc.SetMiniMapVisible(true)
	back, popupValue := float32(.5), float32(.5)
	var popupRect canvasRect
	open := true
	draw := func() {
		h.slider("back", miniNodePosition(h.nc, h.nc.miniFrame.panel.Min), &back)
		h.nc.Node("popup owner", imgui.Vec2{X: 100, Y: 100}, NodeFlags{}, func(n *NodeContext[string]) {
			n.Label("popup owner")
			if open {
				imgui.OpenPopupStr("popup")
				open = false
			}
			l := h.nc.miniFrame
			imgui.SetNextWindowPos(imgui.Vec2{X: l.panel.Min.X + 10, Y: l.panel.Min.Y + 10})
			if imgui.BeginPopup("popup") {
				imgui.PushItemWidth(100)
				popupValue, _ = WheelSlider("popup value", popupValue, 0, 1, 100, "%.2f", imgui.SliderFlagsNone)
				popupRect = canvasRect{Min: imgui.ItemRectMin(), Max: imgui.ItemRectMax()}
				imgui.PopItemWidth()
				imgui.EndPopup()
			}
		})
	}
	h.frame(draw)
	h.frame(draw)
	h.point(popupRect.center())
	h.frame(draw)
	h.io.AddMouseWheelEvent(0, 1)
	intents := h.frame(draw)
	if back != .5 || !approx32(popupValue, .51) || intents.NodeRaised != nil || h.nc.View().Pan != (imgui.Vec2{}) {
		t.Fatalf("popup input leaked or blocked: back=%v popup=%v intents=%+v view=%+v", back, popupValue, intents, h.nc.View())
	}
}

func TestMiniMapNoExtraMeasurementCallbacks(t *testing.T) {
	h := newCanvasWidgetTest(t)
	calls := 0
	draw := func() {
		h.nc.Node("one", imgui.Vec2{X: 100, Y: 100}, NodeFlags{}, func(n *NodeContext[string]) {
			calls++
			n.Label("one")
		})
	}
	h.frame(draw)
	h.nc.SetMiniMapVisible(true)
	h.frame(draw)
	if calls != 2 {
		t.Fatalf("opening map measured nodes again: %d calls", calls)
	}
	h.nc.ZoomToFit()
	h.frame(draw)
	if h.nc.fitCanvas == nil || h.nc.fitCanvas.miniRetained.drawn || h.nc.fitCanvas.miniFrame.panelShown {
		t.Fatal("hidden fit canvas prepared mini-map chrome")
	}
}

func TestMiniMapRuntimeCornerHandoverAndPanCapture(t *testing.T) {
	h := newCanvasWidgetTest(t)
	h.nc.SetMiniMapVisible(true)
	draw := func() {
		h.nc.Node("one", imgui.Vec2{X: 100, Y: 100}, NodeFlags{}, func(n *NodeContext[string]) { n.Label("one") })
	}
	h.frame(draw)
	h.frame(draw)
	c := h.nc.MiniMapConfig()
	c.Corner = NodeCanvasUpperLeft
	h.nc.SetMiniMapConfig(c)
	newPanel := layoutMiniMap(h.nc.origin, h.nc.viewport, c, h.nc.style, h.nc.zoomOverlayRect()).panel
	h.point(newPanel.center())
	h.io.AddMouseButtonEvent(0, true)
	h.frame(draw)
	if h.nc.View().Pan != (imgui.Vec2{}) || h.nc.gesture.isMiniMap() {
		t.Fatal("corner handover accepted a press before drawing")
	}
	h.io.AddMouseButtonEvent(0, false)
	h.frame(draw)
	h.io.AddMouseButtonEvent(0, true)
	h.frame(draw)
	h.io.AddMouseButtonEvent(0, false)
	h.frame(draw)
	if h.nc.View().Pan == (imgui.Vec2{}) {
		t.Fatal("new corner did not become navigable")
	}
	// existing canvas pan keeps its capture when crossing the map.
	h.nc.SetView(View{Zoom: 1})
	h.point(imgui.Vec2{X: 450, Y: 350})
	h.frame(draw)
	h.io.AddMouseButtonEvent(2, true)
	h.frame(draw)
	h.point(newPanel.center())
	h.frame(draw)
	h.io.AddMouseButtonEvent(2, false)
	h.frame(draw)
	if h.nc.View().Pan == (imgui.Vec2{}) || h.nc.gesture.isMiniMap() {
		t.Fatalf("middle pan lost capture over map: %+v", h.nc.View())
	}
}

func TestMiniMapHiddenDoesNotReserveCanvasInput(t *testing.T) {
	h := newCanvasWidgetTest(t)
	value := float32(.5)
	var rect canvasRect
	draw := func() {
		rect = h.slider("node", imgui.Vec2{X: 100, Y: 100}, &value)
	}
	h.frame(draw)
	h.point(imgui.Vec2{X: rect.Min.X + 20, Y: rect.center().Y})
	h.frame(draw)
	h.io.AddMouseButtonEvent(0, true)
	h.frame(draw)
	if value == .5 {
		t.Fatal("hidden map intercepted canvas widget")
	}
	h.io.AddMouseButtonEvent(0, false)
	h.frame(draw)
}
