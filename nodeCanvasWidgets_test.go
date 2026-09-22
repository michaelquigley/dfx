package dfx

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
)

// these tests run real ImGui frames and input queues, without a window or
// renderer. in particular they don't synthesize the widget arbitration bits
// that the pure gesture tests consume. they must not run in parallel.
type canvasWidgetTest struct {
	nc *NodeCanvas[string]
	io *imgui.IO
}

func newCanvasWidgetTest(t *testing.T) *canvasWidgetTest {
	t.Helper()
	runtime.LockOSThread()
	ctx := imgui.CreateContext()
	io := imgui.CurrentIO()
	io.SetIniFilename("")
	io.SetDisplaySize(imgui.Vec2{X: 1000, Y: 800})
	io.SetDeltaTime(1.0 / 60)
	io.SetBackendFlags(imgui.BackendFlagsRendererHasTextures)
	h := &canvasWidgetTest{nc: NewNodeCanvas[string](NodeCanvasConfig{}), io: io}
	t.Cleanup(func() {
		h.nc.Destroy()
		imgui.DestroyContextV(ctx)
		runtime.UnlockOSThread()
	})
	return h
}

func (h *canvasWidgetTest) frame(draw func()) Intents[string] {
	imgui.NewFrame()
	imgui.SetNextWindowPos(imgui.Vec2{})
	imgui.SetNextWindowSize(imgui.Vec2{X: 1000, Y: 800})
	imgui.BeginV("test", nil, imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)
	h.nc.Begin(&State{Size: imgui.Vec2{X: 900, Y: 700}})
	draw()
	intents := h.nc.End()
	imgui.End()
	imgui.Render()
	return intents
}

func (h *canvasWidgetTest) slider(id string, pos imgui.Vec2, value *float32) canvasRect {
	var rect canvasRect
	h.nc.Node(id, pos, NodeFlags{}, func(n *NodeContext[string]) {
		imgui.PushItemWidth(220)
		*value, _ = WheelSlider("value", *value, 0, 1, 100, "%.3f", imgui.SliderFlagsNone)
		rect = canvasRect{Min: imgui.ItemRectMin(), Max: imgui.ItemRectMax()}
		imgui.PopItemWidth()
	})
	return rect
}

func (h *canvasWidgetTest) point(p imgui.Vec2) {
	h.io.AddMousePosEvent(p.X, p.Y)
}

func TestNodeCanvas_OverlappingWheelSliders(t *testing.T) {
	for _, zoom := range []float32{1, 1.5} {
		t.Run(fmt.Sprintf("zoom_%g", zoom), func(t *testing.T) {
			h := newCanvasWidgetTest(t)
			h.nc.SetView(View{Zoom: zoom})
			back, front := float32(0.5), float32(0.5)
			var rect canvasRect
			draw := func() {
				h.slider("back", imgui.Vec2{X: 100, Y: 100}, &back)
				rect = h.slider("front", imgui.Vec2{X: 100, Y: 100}, &front)
			}
			h.frame(draw)
			h.point(rect.center())
			h.frame(draw)
			for _, modifier := range []struct {
				key  imgui.Key
				step float32
			}{{imgui.KeyNone, 0.01}, {imgui.ModCtrl, 0.1}, {imgui.ModAlt, 0.001}} {
				if modifier.key != imgui.KeyNone {
					h.io.AddKeyEvent(modifier.key, true)
				}
				before := front
				h.io.AddMouseWheelEvent(0, 1)
				h.frame(draw)
				if back != 0.5 || !approx32(front, before+modifier.step) || h.nc.View().Zoom != zoom {
					t.Fatalf("wheel reached wrong target: back=%v front=%v view=%+v", back, front, h.nc.View())
				}
				if modifier.key != imgui.KeyNone {
					h.io.AddKeyEvent(modifier.key, false)
				}
			}
		})
	}
}

func TestNodeCanvas_CoveringChromeBlocksWidgetAndAllowsSelection(t *testing.T) {
	h := newCanvasWidgetTest(t)
	back := float32(0.5)
	var rect canvasRect
	draw := func() {
		rect = h.slider("back", imgui.Vec2{X: 100, Y: 100}, &back)
		h.nc.Node("front", imgui.Vec2{X: 100, Y: 100}, NodeFlags{}, func(n *NodeContext[string]) {
			n.Label("covering node")
			imgui.Dummy(imgui.Vec2{X: 260, Y: 50})
		})
	}
	h.frame(draw)
	h.point(rect.center())
	h.frame(draw)
	h.io.AddMouseWheelEvent(0, 1)
	h.frame(draw)
	if back != 0.5 || h.nc.View().Zoom != 1.1 {
		t.Fatalf("chrome wheel: back=%v zoom=%v", back, h.nc.View().Zoom)
	}
	// let the new zoom measure, then click the foreground chrome.
	h.frame(draw)
	h.point(rect.center())
	h.frame(draw)
	h.io.AddMouseButtonEvent(0, true)
	intents := h.frame(draw)
	if back != 0.5 || intents.NodeRaised == nil || *intents.NodeRaised != "front" || intents.SelectionChanged == nil || !idSlicesEqual(intents.SelectionChanged.Nodes, []string{"front"}) {
		t.Fatalf("chrome click reached wrong target: back=%v intents=%+v", back, intents)
	}
}

func TestNodeCanvas_OverlappingClickOnlyActivatesFrontWidget(t *testing.T) {
	h := newCanvasWidgetTest(t)
	back, front := float32(0.5), float32(0.5)
	var rect canvasRect
	draw := func() {
		h.slider("back", imgui.Vec2{X: 100, Y: 100}, &back)
		rect = h.slider("front", imgui.Vec2{X: 100, Y: 100}, &front)
	}
	h.frame(draw)
	h.point(imgui.Vec2{X: rect.Min.X + 20, Y: rect.center().Y})
	h.frame(draw)
	h.io.AddMouseButtonEvent(0, true)
	intents := h.frame(draw)
	if back != 0.5 || front == 0.5 || !imgui.IsAnyItemActive() || intents.NodeRaised == nil || *intents.NodeRaised != "front" || intents.SelectionChanged != nil {
		t.Fatalf("overlap click: back=%v front=%v intents=%+v", back, front, intents)
	}
}

func TestNodeCanvas_WidgetClickRaisesWithoutSelectingAndDragKeepsCapture(t *testing.T) {
	h := newCanvasWidgetTest(t)
	back, front := float32(0.5), float32(0.5)
	var backRect, frontRect canvasRect
	raised := false
	draw := func() {
		if raised {
			frontRect = h.slider("front", imgui.Vec2{X: 200, Y: 100}, &front)
			backRect = h.slider("back", imgui.Vec2{X: 100, Y: 100}, &back)
		} else {
			backRect = h.slider("back", imgui.Vec2{X: 100, Y: 100}, &back)
			frontRect = h.slider("front", imgui.Vec2{X: 200, Y: 100}, &front)
		}
	}
	h.frame(draw)
	h.point(imgui.Vec2{X: backRect.Min.X + 20, Y: backRect.center().Y})
	h.frame(draw)
	h.io.AddMouseButtonEvent(0, true)
	intents := h.frame(draw)
	if intents.NodeRaised == nil || *intents.NodeRaised != "back" || intents.SelectionChanged != nil || h.nc.gesture.kind != gestureIdle {
		t.Fatalf("widget click didn't independently raise: %+v", intents)
	}
	before := back
	// ignore the raise first: dragging into the covering node must retain
	// the active slider, without granting hover/wheel to its other widgets.
	h.point(frontRect.center())
	h.frame(draw)
	if back == before || front != 0.5 || !imgui.IsAnyItemActive() {
		t.Fatalf("drag lost capture: back=%v (was %v), front=%v", back, before, front)
	}
	// applying the raise reorders declarations; the same active ID survives.
	raised = true
	h.point(imgui.Vec2{X: backRect.Min.X + 50, Y: backRect.center().Y})
	h.frame(draw)
	if !imgui.IsAnyItemActive() || front != 0.5 {
		t.Fatal("raising interrupted the active control")
	}
	h.io.AddMouseButtonEvent(0, false)
	h.frame(draw)
	if imgui.IsAnyItemActive() {
		t.Fatal("slider kept capture after release")
	}
	h.frame(draw)
	before = back
	h.io.AddMouseWheelEvent(0, 1)
	h.frame(draw)
	if !approx32(back, before+0.01) || front != 0.5 {
		t.Fatal("raised slider didn't receive wheel")
	}
}

func TestNodeCanvas_NodeInputGeometryHandover(t *testing.T) {
	for _, change := range []string{"remove", "reorder", "move", "zoom", "appear"} {
		t.Run(change, func(t *testing.T) {
			h := newCanvasWidgetTest(t)
			back, front := float32(0.5), float32(0.5)
			var rect canvasRect
			changed := false
			draw := func() {
				if changed && change == "reorder" {
					h.slider("front", imgui.Vec2{X: 100, Y: 100}, &front)
				}
				rect = h.slider("back", imgui.Vec2{X: 100, Y: 100}, &back)
				if (changed && (change == "remove" || change == "reorder")) || (!changed && change == "appear") {
					return
				}
				pos := imgui.Vec2{X: 100, Y: 100}
				if changed && change == "move" {
					pos.X = 500
				}
				h.slider("front", pos, &front)
			}
			h.frame(draw)
			h.point(rect.center())
			h.frame(draw)
			changed = true
			if change == "zoom" {
				h.nc.SetView(View{Zoom: 1.1})
			}
			// appearance is measured once before the new node can receive
			// input. it can never share the old owner's wheel in that frame.
			if change == "appear" {
				h.frame(draw)
			} else {
				h.io.AddMouseWheelEvent(0, 1)
				h.frame(draw)
				if back != 0.5 || front != 0.5 {
					t.Fatalf("handover leaked input: back=%v front=%v", back, front)
				}
			}
			h.point(rect.center())
			h.frame(draw) // ImGui trickles mouse movement before wheel events.
			h.io.AddMouseWheelEvent(0, 1)
			h.frame(draw)
			wantBack, wantFront := float32(0.51), float32(0.5)
			if change == "zoom" || change == "appear" {
				wantBack, wantFront = 0.5, 0.51
			}
			if !approx32(back, wantBack) || !approx32(front, wantFront) {
				t.Fatalf("new owner not live: back=%v front=%v", back, front)
			}
		})
	}
}

func TestNodeCanvas_NodePopupKeepsItsOwnMouseInput(t *testing.T) {
	h := newCanvasWidgetTest(t)
	back, popupValue := float32(0.5), float32(0.5)
	var popupRect canvasRect
	open := true
	draw := func() {
		h.slider("back", imgui.Vec2{X: 100, Y: 100}, &back)
		h.nc.Node("front", imgui.Vec2{X: 100, Y: 100}, NodeFlags{}, func(n *NodeContext[string]) {
			n.Label("popup owner")
			if open {
				imgui.OpenPopupStr("popup")
				open = false
			}
			imgui.SetNextWindowPos(imgui.Vec2{X: 550, Y: 300})
			if imgui.BeginPopup("popup") {
				imgui.PushItemWidth(220)
				popupValue, _ = WheelSlider("popup value", popupValue, 0, 1, 100, "%.3f", imgui.SliderFlagsNone)
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
	if back != 0.5 || !approx32(popupValue, 0.51) || h.nc.View().Zoom != 1 || intents.NodeRaised != nil {
		t.Fatalf("popup input was blocked or leaked: back=%v popup=%v view=%+v intents=%+v", back, popupValue, h.nc.View(), intents)
	}
}

func TestNodeCanvas_SynchronousContentAndMouseStateRestoration(t *testing.T) {
	h := newCanvasWidgetTest(t)
	value := float32(0.5)
	var rect canvasRect
	calls := 0
	draw := func() {
		h.nc.Node("back", imgui.Vec2{X: 100, Y: 100}, NodeFlags{}, func(n *NodeContext[string]) {
			calls++
		})
		if calls == 0 {
			t.Fatal("Node deferred its callback")
		}
		rect = h.slider("front", imgui.Vec2{X: 100, Y: 100}, &value)
	}
	h.frame(draw)
	h.point(rect.center())
	h.frame(draw)
	h.io.AddMouseWheelEvent(0, 1)
	h.frame(draw)
	if calls != 3 || !approx32(value, 0.51) {
		t.Fatalf("callback or input state leaked: calls=%d value=%v", calls, value)
	}
}
