package dfx

import "github.com/AllenDang/cimgui-go/imgui"

// prepareNodeInput chooses a single mouse owner before any content runs.
// Node callbacks stay synchronous: future declarations cannot be inspected
// yet, so routing uses the last completed frame (the graph the user saw).
// a transform change waits for a measurement at the new transform; node
// content may change size or switch between labels and widgets on zoom.
func (nc *NodeCanvas[ID]) prepareNodeInput() {
	nc.inputNode = nil
	nc.inputNodeSeen = false
	g := nc.retained
	if g == nil {
		return
	}
	hit := hitTest(g, imgui.MousePos(), nc.style.hitParams())
	if hit.kind != hitNode && hit.kind != hitPin {
		return
	}
	for i := range g.nodes {
		if g.nodes[i].id == hit.node {
			nc.inputNode = &g.nodes[i]
			return
		}
	}
}

// drawNodeContent gates hover, not active IDs: an already-active slider
// continues its drag even outside its node or behind another one. keyboard
// focus is likewise untouched. only the canvas window's hover is masked;
// popup windows own their input normally, including outside the node bounds.
func (nc *NodeCanvas[ID]) drawNodeContent(id ID, pos imgui.Vec2, index int, n *NodeContext[ID], content func(*NodeContext[ID])) {
	owner := nc.inputNode
	g := nc.retained
	allowMouse := owner != nil && owner.id == id && owner.rect.Min == pos && owner.declIndex == index &&
		g.view == nc.frameView && g.origin == nc.origin && g.viewport == nc.viewport && nc.gesture.kind == gestureIdle
	if allowMouse {
		nc.inputNodeSeen = true
	}
	ctx := imgui.CurrentContext()
	hoveredWindow := ctx.HoveredWindow()
	if !allowMouse && hoveredWindow.CData == imgui.InternalCurrentWindow().CData {
		// a zero wrapper passes a null C pointer; the generated setter cannot
		// accept a nil Go wrapper. restore before any other node/canvas runs.
		ctx.SetHoveredWindow(&imgui.Window{})
		io := imgui.CurrentIO()
		wheel, wheelH := io.MouseWheel(), io.MouseWheelH()
		io.SetMouseWheel(0)
		io.SetMouseWheelH(0)
		defer func() {
			ctx.SetHoveredWindow(hoveredWindow)
			io.SetMouseWheel(wheel)
			io.SetMouseWheelH(wheelH)
		}()
	}
	if content != nil {
		content(n)
	}
}

// nodeInputReady prevents a canvas selection/zoom from taking the place of
// a widget input suppressed during geometry handover. removed nodes release
// their old input region on the following frame, never to a covered widget
// halfway through this frame. frame declarations remain authoritative.
func (nc *NodeCanvas[ID]) nodeInputReady(g *canvasGeometry[ID], mouse imgui.Vec2) bool {
	hit := hitTest(g, mouse, nc.style.hitParams())
	if hit.kind == hitNode || hit.kind == hitPin {
		return nc.inputNodeSeen && nc.inputNode != nil && nc.inputNode.id == hit.node
	}
	return nc.inputNode == nil
}
