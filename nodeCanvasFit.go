package dfx

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
)

// fitNode retains this frame's declaration only until End measures it.
type fitNode[ID comparable] struct {
	id      ID
	pos     imgui.Vec2
	flags   NodeFlags
	content func(*NodeContext[ID])
}

// Measuring reports a hidden zoom-to-fit pass. content runs with widgets
// disabled and a separate imgui ID scope; it must declare the same layout
// as visible content at Detent(), without performing application actions.
func (n *NodeContext[ID]) Measuring() bool {
	return n.nc.measuring
}

// measureFit measures one candidate without changing the visible view or
// retaining native draw output across frames. the isolated window renders
// no pixels, but remains in the viewport so clipped widgets can lay out.
func (nc *NodeCanvas[ID]) measureFit() {
	if nc.fitCanvas == nil {
		nc.fitCanvas = NewNodeCanvas[ID](NodeCanvasConfig{Detents: nc.detents, HideZoomOverlay: true})
		nc.fitCanvas.measuring = true
	}
	probe := nc.fitCanvas
	probe.SetStyle(nc.style)
	probe.SetView(nc.fitView)

	imgui.SetNextWindowPos(nc.origin)
	imgui.SetNextWindowSize(nc.viewport)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{})
	imgui.BeginV(fmt.Sprintf("##nodecanvasfit%p", nc), nil,
		imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoInputs|imgui.WindowFlagsNoSavedSettings|
			imgui.WindowFlagsNoFocusOnAppearing|imgui.WindowFlagsNoBringToFrontOnFocus)
	imgui.PopStyleVar()
	imgui.BeginDisabled()
	probe.Begin(&State{Size: nc.viewport})
	for _, n := range nc.fitNodes {
		if len(nc.fitIDs) == 0 || idInSlice(nc.fitIDs, n.id) {
			probe.Node(n.id, n.pos, n.flags, n.content)
		}
	}
	probe.End()
	imgui.EndDisabled()
	// hide after declarations: hiding before BeginChild can suppress layout.
	imgui.InternalCurrentWindow().SetHidden(true)
	imgui.End()

	bounds, ok := nodeBounds(probe.retained.nodes, nc.fitIDs)
	if !ok {
		nc.fitPending = false
		return
	}
	if fitsAtZoom(bounds, nc.viewport, nc.fitView.Zoom, fitMargin) || nc.fitView.Zoom == nc.detents[0] {
		nc.view = View{Pan: centeredPan(bounds, nc.viewport, nc.fitView.Zoom), Zoom: nc.fitView.Zoom}
		nc.fitPending = false
		return
	}
	next := stepDetent(nc.detents, nc.fitView.Zoom, -1)
	nc.fitView = View{Pan: centeredPan(bounds, nc.viewport, next), Zoom: next}
}
