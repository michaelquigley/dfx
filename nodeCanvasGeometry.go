package dfx

import (
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
)

// this file is the pure geometry core for NodeCanvas: the view transform,
// detent math, bezier distance, rect operations, and hit-testing over a
// frame's geometry snapshot. no imgui calls live here (imgui.Vec2 is used as
// a plain struct only), so everything is unit-testable headlessly.

// screenFromCanvas applies the view transform to a canvas-space point.
func screenFromCanvas(p imgui.Vec2, v View, origin imgui.Vec2) imgui.Vec2 {
	return imgui.Vec2{
		X: (p.X+v.Pan.X)*v.Zoom + origin.X,
		Y: (p.Y+v.Pan.Y)*v.Zoom + origin.Y,
	}
}

// canvasFromScreen applies the inverse view transform to a screen-space point.
func canvasFromScreen(p imgui.Vec2, v View, origin imgui.Vec2) imgui.Vec2 {
	return imgui.Vec2{
		X: (p.X-origin.X)/v.Zoom - v.Pan.X,
		Y: (p.Y-origin.Y)/v.Zoom - v.Pan.Y,
	}
}

// nearestDetent snaps a zoom factor to the nearest configured detent.
// detents are sorted ascending; an exact midpoint tie snaps to the larger
// detent.
func nearestDetent(detents []float32, zoom float32) float32 {
	best := detents[0]
	bestDist := abs32(zoom - best)
	for _, d := range detents[1:] {
		if dist := abs32(zoom - d); dist <= bestDist {
			best = d
			bestDist = dist
		}
	}
	return best
}

// stepDetent moves one detent from the current zoom in the given direction
// (+1 in, -1 out), clamping at the ends of the detent set. the current zoom
// is snapped to its nearest detent first.
func stepDetent(detents []float32, current float32, dir int) float32 {
	idx := 0
	bestDist := abs32(current - detents[0])
	for i, d := range detents[1:] {
		if dist := abs32(current - d); dist <= bestDist {
			idx = i + 1
			bestDist = dist
		}
	}
	idx += dir
	if idx < 0 {
		idx = 0
	}
	if idx > len(detents)-1 {
		idx = len(detents) - 1
	}
	return detents[idx]
}

// zoomTowardPoint changes the view's zoom while keeping the canvas point
// under the given screen position fixed: pan' = (mouse - origin)/zoom' - c.
func zoomTowardPoint(v View, origin, mouseScreen imgui.Vec2, newZoom float32) View {
	c := canvasFromScreen(mouseScreen, v, origin)
	return View{
		Pan: imgui.Vec2{
			X: (mouseScreen.X-origin.X)/newZoom - c.X,
			Y: (mouseScreen.Y-origin.Y)/newZoom - c.Y,
		},
		Zoom: newZoom,
	}
}

// canvasRect is an axis-aligned rectangle in canvas space.
type canvasRect struct {
	Min, Max imgui.Vec2
}

// normalizedRect builds a canvasRect from two arbitrary corners.
func normalizedRect(a, b imgui.Vec2) canvasRect {
	return canvasRect{
		Min: imgui.Vec2{X: min32(a.X, b.X), Y: min32(a.Y, b.Y)},
		Max: imgui.Vec2{X: max32(a.X, b.X), Y: max32(a.Y, b.Y)},
	}
}

func (r canvasRect) contains(p imgui.Vec2) bool {
	return p.X >= r.Min.X && p.X <= r.Max.X && p.Y >= r.Min.Y && p.Y <= r.Max.Y
}

func (r canvasRect) overlaps(o canvasRect) bool {
	return r.Min.X <= o.Max.X && r.Max.X >= o.Min.X && r.Min.Y <= o.Max.Y && r.Max.Y >= o.Min.Y
}

func (r canvasRect) center() imgui.Vec2 {
	return imgui.Vec2{X: (r.Min.X + r.Max.X) / 2, Y: (r.Min.Y + r.Max.Y) / 2}
}

// rectUnion returns the smallest rect containing both inputs.
func rectUnion(a, b canvasRect) canvasRect {
	return canvasRect{
		Min: imgui.Vec2{X: min32(a.Min.X, b.Min.X), Y: min32(a.Min.Y, b.Min.Y)},
		Max: imgui.Vec2{X: max32(a.Max.X, b.Max.X), Y: max32(a.Max.Y, b.Max.Y)},
	}
}

// fitsAtZoom reports whether canvas-space bounds fit within a screen-space
// viewport at the given zoom, leaving margin screen pixels on every side.
func fitsAtZoom(bounds canvasRect, viewport imgui.Vec2, zoom, margin float32) bool {
	w := (bounds.Max.X - bounds.Min.X) * zoom
	h := (bounds.Max.Y - bounds.Min.Y) * zoom
	return w <= viewport.X-2*margin && h <= viewport.Y-2*margin
}

// centeredPan computes the pan that places the center of canvas-space bounds
// at the center of a screen-space viewport at the given zoom.
func centeredPan(bounds canvasRect, viewport imgui.Vec2, zoom float32) imgui.Vec2 {
	c := bounds.center()
	return imgui.Vec2{
		X: viewport.X/2/zoom - c.X,
		Y: viewport.Y/2/zoom - c.Y,
	}
}

// bezierSegments is the fixed cubic subdivision shared by link hit-testing
// and box-select intersection, so the two always agree on a link's shape.
const bezierSegments = 24

// linkCubic builds the control points of a horizontal-tangent cubic bezier
// from an output pin position to an input pin position.
func linkCubic(from, to imgui.Vec2, tangent float32) [4]imgui.Vec2 {
	return [4]imgui.Vec2{
		from,
		{X: from.X + tangent, Y: from.Y},
		{X: to.X - tangent, Y: to.Y},
		to,
	}
}

// bezierPoint evaluates a cubic bezier at parameter t.
func bezierPoint(c [4]imgui.Vec2, t float32) imgui.Vec2 {
	u := 1 - t
	b0 := u * u * u
	b1 := 3 * u * u * t
	b2 := 3 * u * t * t
	b3 := t * t * t
	return imgui.Vec2{
		X: b0*c[0].X + b1*c[1].X + b2*c[2].X + b3*c[3].X,
		Y: b0*c[0].Y + b1*c[1].Y + b2*c[2].Y + b3*c[3].Y,
	}
}

// bezierPolyline subdivides a cubic bezier into the shared fixed segment
// count, returning bezierSegments+1 points.
func bezierPolyline(c [4]imgui.Vec2) [bezierSegments + 1]imgui.Vec2 {
	var pts [bezierSegments + 1]imgui.Vec2
	for i := 0; i <= bezierSegments; i++ {
		pts[i] = bezierPoint(c, float32(i)/bezierSegments)
	}
	return pts
}

// pointSegmentDistance returns the distance from p to the segment ab.
func pointSegmentDistance(p, a, b imgui.Vec2) float32 {
	abx, aby := b.X-a.X, b.Y-a.Y
	apx, apy := p.X-a.X, p.Y-a.Y
	lenSq := abx*abx + aby*aby
	t := float32(0)
	if lenSq > 0 {
		t = (apx*abx + apy*aby) / lenSq
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
	}
	dx := p.X - (a.X + t*abx)
	dy := p.Y - (a.Y + t*aby)
	return sqrt32(dx*dx + dy*dy)
}

// pointBezierDistance returns the minimum distance from p to the cubic,
// measured over the shared fixed subdivision.
func pointBezierDistance(c [4]imgui.Vec2, p imgui.Vec2) float32 {
	pts := bezierPolyline(c)
	best := pointSegmentDistance(p, pts[0], pts[1])
	for i := 1; i < bezierSegments; i++ {
		if d := pointSegmentDistance(p, pts[i], pts[i+1]); d < best {
			best = d
		}
	}
	return best
}

// segmentIntersectsRect reports whether the segment ab touches the rect,
// including segments entirely inside it. it is a Liang-Barsky clip test.
func segmentIntersectsRect(a, b imgui.Vec2, r canvasRect) bool {
	dx, dy := b.X-a.X, b.Y-a.Y
	t0, t1 := float32(0), float32(1)
	// each edge clips the parametric segment; p is the direction component
	// against the edge, q the distance from a to the edge.
	clip := func(p, q float32) bool {
		if p == 0 {
			return q >= 0 // parallel: inside iff on the inner side
		}
		t := q / p
		if p < 0 {
			if t > t1 {
				return false
			}
			if t > t0 {
				t0 = t
			}
		} else {
			if t < t0 {
				return false
			}
			if t < t1 {
				t1 = t
			}
		}
		return true
	}
	return clip(-dx, a.X-r.Min.X) &&
		clip(dx, r.Max.X-a.X) &&
		clip(-dy, a.Y-r.Min.Y) &&
		clip(dy, r.Max.Y-a.Y)
}

// bezierIntersectsRect reports whether any segment of the shared cubic
// subdivision touches the rect; an endpoint inside the rect falls out of
// segment intersection for free.
func bezierIntersectsRect(c [4]imgui.Vec2, r canvasRect) bool {
	pts := bezierPolyline(c)
	for i := 0; i < bezierSegments; i++ {
		if segmentIntersectsRect(pts[i], pts[i+1], r) {
			return true
		}
	}
	return false
}

// pinSide is which node edge a pin sits on: inputs left, outputs right.
type pinSide int

const (
	pinInput pinSide = iota
	pinOutput
)

// pinGeometry is a pin's derived per-frame geometry.
type pinGeometry[ID comparable] struct {
	id        ID
	pos       imgui.Vec2 // marker center, canvas space
	side      pinSide
	declIndex int // order of declaration within the frame, across all pins
}

// nodeGeometry is a node's derived per-frame geometry.
type nodeGeometry[ID comparable] struct {
	id        ID
	rect      canvasRect // outer node rect, anchored at the declared pos
	selected  bool       // as declared this frame
	declIndex int
	pins      []pinGeometry[ID]
}

// linkGeometry is a link's derived per-frame geometry.
type linkGeometry[ID comparable] struct {
	id        ID
	cubic     [4]imgui.Vec2 // canvas-space control points
	selected  bool          // as declared this frame
	declIndex int
}

// canvasGeometry is one frame's complete derived geometry: what End resolves
// input against, and what queries arriving before the next frame's
// declarations read. it is rebuilt from declarations every frame and is
// never authoritative — derived view-side state, not graph truth.
type canvasGeometry[ID comparable] struct {
	view     View
	origin   imgui.Vec2         // canvas child screen origin
	viewport imgui.Vec2         // canvas child size, screen px
	nodes    []nodeGeometry[ID] // declaration order
	links    []linkGeometry[ID] // declaration order
}

// hitParams carries the screen-pixel hit and snap tolerances; they are
// divided by the current zoom before canvas-space comparison so targets stay
// equally grabbable at every detent.
type hitParams struct {
	pinHitRadius    float32
	linkHitDistance float32
	linkSnapRadius  float32
}

type hitKind int

const (
	hitNone hitKind = iota
	hitNode
	hitPin
	hitLink
)

// hitResult identifies what sits under a screen point.
type hitResult[ID comparable] struct {
	kind hitKind
	node ID      // hitNode, and the owning node for hitPin
	pin  ID      // hitPin
	side pinSide // hitPin
	link ID      // hitLink
}

// hitTest resolves what sits under a screen point, in visual z-order: per
// node in reverse declaration order (topmost first), each node's own pins
// then that same node's body — pins outrank their own node's body but never
// reach through a node covering them. only after every node misses are links
// tested by bezier distance, nearest first, exact ties to the later-declared
// (topmost) link.
func hitTest[ID comparable](g *canvasGeometry[ID], screenPos imgui.Vec2, params hitParams) hitResult[ID] {
	p := canvasFromScreen(screenPos, g.view, g.origin)
	pinRadius := params.pinHitRadius / g.view.Zoom

	for i := len(g.nodes) - 1; i >= 0; i-- {
		node := &g.nodes[i]

		// nearest pin within radius; exact ties to the later-declared pin.
		bestPin := -1
		var bestDist float32
		for j := range node.pins {
			pin := &node.pins[j]
			dx, dy := p.X-pin.pos.X, p.Y-pin.pos.Y
			dist := sqrt32(dx*dx + dy*dy)
			if dist <= pinRadius && (bestPin < 0 || dist <= bestDist) {
				bestPin = j
				bestDist = dist
			}
		}
		if bestPin >= 0 {
			pin := &node.pins[bestPin]
			return hitResult[ID]{kind: hitPin, node: node.id, pin: pin.id, side: pin.side}
		}

		if node.rect.contains(p) {
			return hitResult[ID]{kind: hitNode, node: node.id}
		}
	}

	linkTolerance := params.linkHitDistance / g.view.Zoom
	bestLink := -1
	var bestDist float32
	for i := range g.links {
		dist := pointBezierDistance(g.links[i].cubic, p)
		if dist <= linkTolerance && (bestLink < 0 || dist <= bestDist) {
			bestLink = i
			bestDist = dist
		}
	}
	if bestLink >= 0 {
		return hitResult[ID]{kind: hitLink, link: g.links[bestLink].id}
	}

	return hitResult[ID]{kind: hitNone}
}

// snapPin finds the nearest pin compatible with a link drag from the given
// side (output→input or input→output) within the snap radius, measured in
// screen space. exact ties break to the later-declared pin. it returns the
// winning pin, its canvas-space position, and whether a snap exists.
func snapPin[ID comparable](g *canvasGeometry[ID], screenPos imgui.Vec2, sourceSide pinSide, params hitParams) (ID, imgui.Vec2, bool) {
	var bestID ID
	var bestPos imgui.Vec2
	bestDecl := -1
	var bestDist float32
	found := false

	for i := range g.nodes {
		for j := range g.nodes[i].pins {
			pin := &g.nodes[i].pins[j]
			if pin.side == sourceSide {
				continue
			}
			pinScreen := screenFromCanvas(pin.pos, g.view, g.origin)
			dx, dy := screenPos.X-pinScreen.X, screenPos.Y-pinScreen.Y
			dist := sqrt32(dx*dx + dy*dy)
			if dist > params.linkSnapRadius {
				continue
			}
			if !found || dist < bestDist || (dist == bestDist && pin.declIndex > bestDecl) {
				bestID = pin.id
				bestPos = pin.pos
				bestDecl = pin.declIndex
				bestDist = dist
				found = true
			}
		}
	}
	return bestID, bestPos, found
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func sqrt32(v float32) float32 {
	return float32(math.Sqrt(float64(v)))
}
