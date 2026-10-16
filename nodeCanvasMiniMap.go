package dfx

import (
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
)

// NodeCanvasCorner selects the corner occupied by the mini-map panel.
type NodeCanvasCorner uint8

const (
	NodeCanvasLowerLeft NodeCanvasCorner = iota
	NodeCanvasLowerRight
	NodeCanvasUpperRight
	NodeCanvasUpperLeft
)

// NodeCanvasMiniMapConfig is persistable presentation state. size is the
// panel's screen-pixel size.
type NodeCanvasMiniMapConfig struct {
	Visible bool
	Corner  NodeCanvasCorner
	Size    imgui.Vec2
}

func normalizeMiniMapConfig(c NodeCanvasMiniMapConfig) NodeCanvasMiniMapConfig {
	if c.Corner > NodeCanvasUpperLeft {
		c.Corner = NodeCanvasLowerLeft
	}
	if c.Size.X <= 0 || math.IsNaN(float64(c.Size.X)) || math.IsInf(float64(c.Size.X), 0) {
		c.Size.X = 240
	}
	if c.Size.Y <= 0 || math.IsNaN(float64(c.Size.Y)) || math.IsInf(float64(c.Size.Y), 0) {
		c.Size.Y = 160
	}
	return c
}

func (nc *NodeCanvas[ID]) MiniMapConfig() NodeCanvasMiniMapConfig {
	if nc.pendingMiniMap != nil {
		return *nc.pendingMiniMap
	}
	return nc.miniMap
}

func (nc *NodeCanvas[ID]) SetMiniMapConfig(c NodeCanvasMiniMapConfig) {
	c = normalizeMiniMapConfig(c)
	nc.pendingMiniMap = &c
}

func (nc *NodeCanvas[ID]) MiniMapVisible() bool { return nc.MiniMapConfig().Visible }

func (nc *NodeCanvas[ID]) SetMiniMapVisible(visible bool) {
	c := nc.MiniMapConfig()
	c.Visible = visible
	nc.SetMiniMapConfig(c)
}

type miniMapLayout struct {
	panel, inner         canvasRect // screen space
	panelShown, drawPill bool
}

func rectNonempty(r canvasRect) bool { return r.Max.X > r.Min.X && r.Max.Y > r.Min.Y }

func intersectRect(a, b canvasRect) canvasRect {
	r := canvasRect{Min: imgui.Vec2{X: max32(a.Min.X, b.Min.X), Y: max32(a.Min.Y, b.Min.Y)},
		Max: imgui.Vec2{X: min32(a.Max.X, b.Max.X), Y: min32(a.Max.Y, b.Max.Y)}}
	if !rectNonempty(r) {
		return canvasRect{}
	}
	return r
}

func layoutMiniMap(origin, viewport imgui.Vec2, cfg NodeCanvasMiniMapConfig, style NodeCanvasStyle, pill canvasRect) miniMapLayout {
	var out miniMapLayout
	if viewport.X <= 0 || viewport.Y <= 0 {
		return out
	}
	bounds := canvasRect{Min: origin, Max: imgui.Vec2{X: origin.X + viewport.X, Y: origin.Y + viewport.Y}}
	margin := max32(0, style.MiniMapMargin)
	gap := max32(0, style.MiniMapGap)
	right := cfg.Corner == NodeCanvasLowerRight || cfg.Corner == NodeCanvasUpperRight
	bottom := cfg.Corner == NodeCanvasLowerRight || cfg.Corner == NodeCanvasLowerLeft
	out.drawPill = rectNonempty(pill)
	if !cfg.Visible {
		return out
	}
	mx := min32(margin, viewport.X/2)
	my := min32(margin, viewport.Y/2)
	panelW := min32(cfg.Size.X, max32(0, viewport.X-2*mx))
	panelH := min32(cfg.Size.Y, max32(0, viewport.Y-2*my))
	var px, py float32
	if right {
		px = bounds.Max.X - mx - panelW
	} else {
		px = bounds.Min.X + mx
	}
	if bottom {
		bottomEdge := bounds.Max.Y - my
		if cfg.Corner == NodeCanvasLowerRight && rectNonempty(pill) {
			bottomEdge = min32(bottomEdge, pill.Min.Y-gap)
		}
		panelH = min32(panelH, max32(0, bottomEdge-(bounds.Min.Y+my)))
		py = bottomEdge - panelH
	} else {
		py = bounds.Min.Y + my
	}
	if panelW < 48 || panelH < 48 {
		return out
	}
	out.panel = canvasRect{Min: imgui.Vec2{X: px, Y: py}, Max: imgui.Vec2{X: px + panelW, Y: py + panelH}}
	pad := max32(0, style.MiniMapPadding)
	out.inner = canvasRect{Min: imgui.Vec2{X: px + pad, Y: py + pad}, Max: imgui.Vec2{X: px + panelW - pad, Y: py + panelH - pad}}
	out.panelShown = rectNonempty(out.inner)
	return out
}

type miniMapTransform struct {
	inner  canvasRect
	graph  canvasRect
	scale  float32
	offset imgui.Vec2
	valid  bool
}

func newMiniMapTransform(inner, bounds canvasRect) miniMapTransform {
	if !rectNonempty(inner) {
		return miniMapTransform{}
	}
	c := bounds.center()
	w := max32(1, bounds.Max.X-bounds.Min.X)
	h := max32(1, bounds.Max.Y-bounds.Min.Y)
	s := min32((inner.Max.X-inner.Min.X)/w, (inner.Max.Y-inner.Min.Y)/h)
	if s <= 0 || math.IsInf(float64(s), 0) || math.IsNaN(float64(s)) {
		return miniMapTransform{}
	}
	center := inner.center()
	return miniMapTransform{inner: inner, graph: bounds, scale: s,
		offset: imgui.Vec2{X: center.X - c.X*s, Y: center.Y - c.Y*s}, valid: true}
}

func (m miniMapTransform) project(p imgui.Vec2) imgui.Vec2 {
	return imgui.Vec2{X: p.X*m.scale + m.offset.X, Y: p.Y*m.scale + m.offset.Y}
}

func (m miniMapTransform) unproject(p imgui.Vec2) imgui.Vec2 {
	return imgui.Vec2{X: (p.X - m.offset.X) / m.scale, Y: (p.Y - m.offset.Y) / m.scale}
}

func (m miniMapTransform) clamp(p imgui.Vec2) imgui.Vec2 {
	return imgui.Vec2{X: max32(m.inner.Min.X, min32(m.inner.Max.X, p.X)),
		Y: max32(m.inner.Min.Y, min32(m.inner.Max.Y, p.Y))}
}

func mapViewport(m miniMapTransform, view View, origin, viewport imgui.Vec2) canvasRect {
	return canvasRect{Min: m.project(canvasFromScreen(origin, view, origin)),
		Max: m.project(canvasFromScreen(imgui.Vec2{X: origin.X + viewport.X, Y: origin.Y + viewport.Y}, view, origin))}
}

func mapGrab(actual, inner canvasRect, minSize float32) canvasRect {
	clipped := intersectRect(actual, inner)
	if !rectNonempty(clipped) {
		return canvasRect{}
	}
	c := clipped.center()
	w := max32(minSize, clipped.Max.X-clipped.Min.X)
	h := max32(minSize, clipped.Max.Y-clipped.Min.Y)
	return intersectRect(canvasRect{Min: imgui.Vec2{X: c.X - w/2, Y: c.Y - h/2},
		Max: imgui.Vec2{X: c.X + w/2, Y: c.Y + h/2}}, inner)
}

type miniMapRendered struct {
	layout           miniMapLayout
	transform        miniMapTransform
	actual, grab     canvasRect
	origin, viewport imgui.Vec2
	drawn            bool
}

func (nc *NodeCanvas[ID]) zoomOverlayRect() canvasRect {
	if nc.hideZoomOverlay || nc.viewport.X <= 0 || nc.viewport.Y <= 0 {
		return canvasRect{}
	}
	size := imgui.CalcTextSize(zoomLabel(nc.frameView.Zoom))
	pad, margin := nc.style.ZoomOverlayPadding, nc.style.ZoomOverlayMargin
	max := imgui.Vec2{X: nc.origin.X + nc.viewport.X - margin, Y: nc.origin.Y + nc.viewport.Y - margin}
	return canvasRect{Min: imgui.Vec2{X: max.X - size.X - 2*pad, Y: max.Y - size.Y - 2*pad}, Max: max}
}

func (nc *NodeCanvas[ID]) miniCaptureValid() bool {
	return nc.gesture.mapOrigin == nc.origin && nc.gesture.mapViewport == nc.viewport &&
		nc.gesture.mapLayout == nc.miniFrame && nc.miniFrame.panelShown
}

func (nc *NodeCanvas[ID]) cancelMiniMapGesture() {
	if nc.gesture.isMiniMap() {
		if nc.inFrame {
			nc.cancelMiniPending = true
		} else {
			nc.gesture = gestureState[ID]{kind: gestureIdle}
		}
	}
}

func (nc *NodeCanvas[ID]) miniMapBlocksNodeInput() bool {
	if nc.measuring || !imgui.IsMousePosValid() || imgui.CurrentIO().AppFocusLost() ||
		imgui.IsPopupOpenStrV("", imgui.PopupFlagsAnyPopupId|imgui.PopupFlagsAnyPopupLevel) {
		return false
	}
	p := imgui.MousePos()
	return nc.miniFrame.panelShown && nc.miniFrame.panel.contains(p)
}

func (nc *NodeCanvas[ID]) drawMiniMap(dl *imgui.DrawList, g *canvasGeometry[ID]) {
	l := nc.miniFrame
	nc.miniRetained = miniMapRendered{layout: l, origin: nc.origin, viewport: nc.viewport, drawn: true}
	if !l.panelShown {
		return
	}
	color := func(v imgui.Vec4) uint32 { return imgui.ColorConvertFloat4ToU32(v) }
	dl.AddRectFilledV(l.panel.Min, l.panel.Max, color(nc.style.MiniMapBgColor), 3, imgui.DrawFlagsNone)
	dl.AddRectV(l.panel.Min, l.panel.Max, color(nc.style.MiniMapBorderColor), 3, 1, imgui.DrawFlagsNone)
	if len(g.nodes) == 0 {
		return
	}
	var bounds canvasRect
	for i := range g.nodes {
		r := g.nodes[i].rect
		if nc.gesture.kind == gestureDragNodes && idInSlice(nc.gesture.dragSet, g.nodes[i].id) {
			r.Min.X += nc.frameDragOffset.X
			r.Min.Y += nc.frameDragOffset.Y
			r.Max.X += nc.frameDragOffset.X
			r.Max.Y += nc.frameDragOffset.Y
		}
		if i == 0 {
			bounds = r
		} else {
			bounds = rectUnion(bounds, r)
		}
	}
	m := newMiniMapTransform(l.inner, bounds)
	if nc.gesture.isMiniMap() && nc.gesture.mapTransform.valid {
		m = nc.gesture.mapTransform
	}
	if !m.valid {
		return
	}
	nc.miniRetained.transform = m
	dl.PushClipRectV(l.inner.Min, l.inner.Max, true)
	for i := range g.nodes {
		n := &g.nodes[i]
		r := n.rect
		if nc.gesture.kind == gestureDragNodes && idInSlice(nc.gesture.dragSet, n.id) {
			r.Min.X += nc.frameDragOffset.X
			r.Min.Y += nc.frameDragOffset.Y
			r.Max.X += nc.frameDragOffset.X
			r.Max.Y += nc.frameDragOffset.Y
		}
		min, max := m.project(r.Min), m.project(r.Max)
		max.X = max32(max.X, min.X+1)
		max.Y = max32(max.Y, min.Y+1)
		accent := n.accent
		if accent == (imgui.Vec4{}) {
			accent = nc.style.MiniMapNodeColor
		}
		dl.AddRectFilledV(min, max, color(accent), 0, imgui.DrawFlagsNone)
	}
	actual := mapViewport(m, nc.frameView, nc.origin, nc.viewport)
	nc.miniRetained.actual = actual
	nc.miniRetained.grab = mapGrab(actual, l.inner, nc.style.MiniMapGrabSize)
	if clipped := intersectRect(actual, l.inner); rectNonempty(clipped) {
		dl.AddRectFilledV(clipped.Min, clipped.Max, color(nc.style.MiniMapViewportFill), 0, imgui.DrawFlagsNone)
		dl.AddRectV(actual.Min, actual.Max, color(nc.style.MiniMapViewportOutline), 0, 1, imgui.DrawFlagsNone)
		if actual.Max.X-actual.Min.X < nc.style.MiniMapGrabSize || actual.Max.Y-actual.Min.Y < nc.style.MiniMapGrabSize {
			c := clipped.center()
			dl.AddCircleFilledV(c, 2, color(nc.style.MiniMapViewportOutline), 8)
		}
	} else {
		// a rim marker points toward a viewport entirely outside the map.
		tip := m.clamp(actual.center())
		v := imgui.Vec2{X: actual.center().X - l.inner.center().X, Y: actual.center().Y - l.inner.center().Y}
		length := float32(math.Hypot(float64(v.X), float64(v.Y)))
		if length > 0 {
			v.X /= length
			v.Y /= length
			base := imgui.Vec2{X: tip.X - 7*v.X, Y: tip.Y - 7*v.Y}
			perp := imgui.Vec2{X: -3 * v.Y, Y: 3 * v.X}
			dl.AddTriangleFilled(tip,
				imgui.Vec2{X: base.X + perp.X, Y: base.Y + perp.Y},
				imgui.Vec2{X: base.X - perp.X, Y: base.Y - perp.Y},
				color(nc.style.MiniMapViewportOutline))
		}
	}
	dl.PopClipRect()
}
