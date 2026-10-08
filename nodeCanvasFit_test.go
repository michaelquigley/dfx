package dfx

import (
	"fmt"
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
)

func TestNodeCanvas_FitMeasuresWithoutDisplayingCandidates(t *testing.T) {
	h := newCanvasWidgetTest(t)
	h.nc.detents = []float32{0.5, 1, 1.5}
	start := View{Pan: imgui.Vec2{X: 30, Y: 20}, Zoom: 0.5}
	h.nc.SetView(start)
	var probes []float32
	var probeList *imgui.DrawList
	visibleCalls := 0
	draw := func() {
		h.nc.Node("node", imgui.Vec2{X: 100, Y: 100}, NodeFlags{}, func(n *NodeContext[string]) {
			if n.Measuring() {
				probes = append(probes, n.Detent())
				probeList = imgui.WindowDrawList()
			} else {
				visibleCalls++
			}
			width := float32(100)
			if n.Detent() >= 1 {
				width = 780
			}
			imgui.Dummy(imgui.Vec2{X: width * n.Detent(), Y: 40 * n.Detent()})
		})
	}
	h.frame(draw)
	h.frame(draw)
	h.nc.ZoomToFit()
	for i := 0; i < 2; i++ {
		h.frame(draw)
		if h.nc.frameView != start {
			t.Fatalf("displayed a fit candidate: %+v", h.nc.frameView)
		}
		for _, list := range imgui.CurrentDrawData().CommandLists() {
			if list.CData == probeList.CData {
				t.Fatal("hidden measurement reached the renderer")
			}
		}
		if i == 0 && (!h.nc.fitPending || h.nc.View() != start) {
			t.Fatal("first candidate changed the public view or ended the search")
		}
	}
	if len(probes) != 2 || probes[0] != 1.5 || probes[1] != 1 {
		t.Fatalf("measured detents = %v", probes)
	}
	if h.nc.fitPending || h.nc.View().Zoom != 1 {
		t.Fatalf("fit did not choose the largest detent fitting its own content: %+v", h.nc.View())
	}
	fit := h.nc.View()
	h.frame(draw)
	if h.nc.frameView != fit || visibleCalls != 5 || len(probes) != 2 {
		t.Fatal("fit did not appear in one change or kept measuring after completion")
	}
	bounds, _ := nodeBounds(h.nc.retained.nodes, nil)
	if !fitsAtZoom(bounds, h.nc.viewport, fit.Zoom, fitMargin) ||
		!approxVec2(fit.Pan, centeredPan(bounds, h.nc.viewport, fit.Zoom)) {
		t.Fatal("visible fitted content does not fit or is not centered")
	}
}

func TestNodeCanvas_FitDoesNotActivateHiddenWidgets(t *testing.T) {
	h := newCanvasWidgetTest(t)
	clicked := false
	var hiddenButton canvasRect
	draw := func() {
		h.nc.Node("node", imgui.Vec2{X: 100, Y: 100}, NodeFlags{}, func(n *NodeContext[string]) {
			if imgui.Button("button") && n.Measuring() {
				clicked = true
			}
			if n.Measuring() {
				hiddenButton = canvasRect{Min: imgui.ItemRectMin(), Max: imgui.ItemRectMax()}
			}
		})
		h.nc.Node("far", imgui.Vec2{X: 4000}, NodeFlags{}, func(n *NodeContext[string]) { n.Label("far") })
	}
	h.frame(draw)
	h.frame(draw)
	h.nc.ZoomToFit()
	h.frame(draw)
	h.point(hiddenButton.center())
	h.io.AddMouseButtonEvent(0, true)
	h.frame(draw)
	h.io.AddMouseButtonEvent(0, false)
	h.frame(draw)
	if clicked {
		t.Fatal("hidden measurement activated a widget")
	}
	host := imgui.InternalFindWindowByName(fmt.Sprintf("##nodecanvasfit%p", h.nc))
	for _, list := range imgui.CurrentDrawData().CommandLists() {
		if list.CData == host.DrawList().CData {
			t.Fatal("measurement host reached the renderer")
		}
	}
}

func TestNodeCanvas_FitEmptyAndSubset(t *testing.T) {
	for _, ids := range [][]string{nil, {"near"}, {"missing"}} {
		t.Run(fmt.Sprint(ids), func(t *testing.T) {
			h := newCanvasWidgetTest(t)
			draw := func() {
				h.nc.Node("near", imgui.Vec2{}, NodeFlags{}, func(n *NodeContext[string]) { n.Label("near") })
				h.nc.Node("far", imgui.Vec2{X: 4000}, NodeFlags{}, func(n *NodeContext[string]) { n.Label("far") })
			}
			h.frame(draw)
			h.frame(draw)
			start := h.nc.View()
			h.nc.ZoomToFit(ids...)
			for i := 0; i < len(h.nc.detents); i++ {
				h.frame(draw)
			}
			want := float32(0.25)
			if len(ids) > 0 {
				want = 1.5
				if ids[0] == "missing" {
					want = start.Zoom
				}
			}
			if h.nc.fitPending || h.nc.View().Zoom != want {
				t.Fatalf("fit = %+v, want zoom %v", h.nc.View(), want)
			}
			h.nc.ZoomToFit()
			h.frame(func() {}) // nodes disappear during a fit
			if h.nc.fitPending {
				t.Fatal("empty graph kept fit pending")
			}
		})
	}
}

func TestNodeCanvas_ExplicitViewCancelsHiddenFit(t *testing.T) {
	h := newCanvasWidgetTest(t)
	draw := func() {
		h.nc.Node("near", imgui.Vec2{}, NodeFlags{}, func(n *NodeContext[string]) { n.Label("near") })
		h.nc.Node("far", imgui.Vec2{X: 4000}, NodeFlags{}, func(n *NodeContext[string]) { n.Label("far") })
	}
	h.frame(draw)
	h.frame(draw)
	h.nc.ZoomToFit()
	h.frame(draw)
	want := View{Pan: imgui.Vec2{X: 10, Y: 20}, Zoom: 0.7}
	h.nc.SetView(want)
	h.frame(draw)
	if h.nc.fitPending || h.nc.frameView != want || h.nc.View() != want {
		t.Fatal("hidden measurement overwrote explicit navigation")
	}
}
