package dfx

import (
	"fmt"
	"math"
	"reflect"
	"sort"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/df/dl"
)

// NodeCanvas is a zoomable, pannable node-graph editing surface. the app
// owns all graph truth — nodes, links, positions, and selection are declared
// every frame between Begin and End — while the canvas owns only view and
// gesture state. completed gestures come back as Intents from End; the
// canvas never mutates the graph, and the next frame's declarations are the
// only truth it renders.
//
// NodeCanvas is a widget, not a Component: it is driven from inside an
// owning component's Draw. node, pin, and link IDs share one ID space and
// the app guarantees uniqueness across all three populations; the canvas
// does not police it. IDs must also be unique and stable frame-to-frame in
// their fmt.Sprint form (pointer-kind IDs format by address instead), since
// that form scopes per-node imgui widget state.
type NodeCanvas[ID comparable] struct {
	detents     []float32
	gridSpacing float32
	locked      bool
	style       NodeCanvasStyle
	styleSet    bool

	// wheelStepsPerZoomLevel is the accumulated wheel travel (imgui wheel
	// units) that produces one detent step.
	wheelStepsPerZoomLevel float32

	view        View
	pendingView *View
	gesture     gestureState[ID]

	// last begun canvas rect: transform helpers and queries arriving before
	// this frame's Begin (keyboard action handlers run before component
	// drawing) compute against it — one frame stale, visually
	// indistinguishable. zero before the first Begin.
	origin   imgui.Vec2
	viewport imgui.Vec2

	// frameView is the transform everything drawn this frame uses: the
	// applied view plus any in-flight pan, computed once at Begin so a
	// frame's transform is stable across everything it draws.
	frameView View

	// frameDragOffset is the in-flight node-drag offset in canvas space,
	// computed once at Begin from the absolute mouse position against the
	// gesture anchor; zero outside an active node drag.
	frameDragOffset imgui.Vec2

	// per-frame declaration buffers, reset at Begin.
	nodeIndex  int
	frameNodes []nodeGeometry[ID]
	framePins  map[ID]framePin
	frameLinks []linkDecl[ID]
	pinCounter int
	itemActive bool // OR of IsItemActive over the frame's node content groups

	// splitter is the owned drawlist splitter: channel 0 for links, then a
	// chrome/content channel pair per node. owned — never the drawlist's
	// ChannelsSplit convenience API, which cannot nest with the splits imgui
	// widgets (tables in particular) perform internally.
	splitter      *imgui.DrawListSplitter
	splitCap      int32
	lastNodeCount int

	// retained is the last completed frame's geometry: derived view-side
	// state backing queries that arrive before this frame's declarations
	// complete. it is never authoritative — declarations rebuild it every
	// frame.
	retained *canvasGeometry[ID]

	// hover is the hit resolved at the last End, consumed by node chrome the
	// following frame (the normal one-frame immediate-mode cadence).
	hover hitResult[ID]

	// sticky drag tracking: leftDragging must hold through the release
	// frame (imgui's IsMouseDragging drops there), or a fast flick —
	// threshold and release in one frame — would misread as a click.
	leftPressPos   imgui.Vec2
	leftDragSticky bool

	// wheelAccum is the sub-threshold wheel travel carried across frames:
	// raw wheel deltas accumulate here until they cross
	// wheelStepsPerZoomLevel, at which point the snapshot reports one
	// detent step (per frame at most) and the remainder keeps carrying.
	wheelAccum float32

	// pending zoom-to-fit: node bounds are detent-dependent (apps simplify
	// content below 1.0), so a fit resolves over frames, walking down from
	// the top detent until the bounds declared at the current detent fit —
	// the first detent that fits its own declared content is by construction
	// the largest such detent. most recent navigation wins: any other
	// explicit navigation cancels the pending fit.
	fitPending  bool
	fitStarting bool // next Begin jumps to the top detent
	fitIDs      []ID // empty: all nodes
}

// fitMargin is the screen-pixel margin zoom-to-fit leaves on every side of
// the fitted bounds.
const fitMargin = 40

// framePin carries a pin's per-frame positions: declared-anchored for hit
// geometry, draw-space (in-flight drag offset applied) for link rendering.
type framePin struct {
	declared  imgui.Vec2
	draw      imgui.Vec2
	side      pinSide
	declIndex int
}

// linkDecl is a link declaration, recorded by Link and resolved at End —
// pins may belong to nodes declared later.
type linkDecl[ID comparable] struct {
	id       ID
	from, to ID
	selected bool
}

// NodeFlags carries a node's per-frame declared state.
type NodeFlags struct{ Selected bool }

// LinkFlags carries a link's per-frame declared state.
type LinkFlags struct{ Selected bool }

// NodeCanvasConfig configures a NodeCanvas at construction.
type NodeCanvasConfig struct {
	// Detents are the only zoom levels, sorted ascending; the last entry is
	// the maximum and must be 1.0. defaults to {0.25, 0.5, 0.75, 1.0}.
	Detents []float32

	// GridSpacing is the grid cell size in canvas units, exposed for
	// app-side snap logic. defaults to 50.
	GridSpacing float32

	// WheelStepsPerZoomLevel is the wheel travel — in imgui wheel units,
	// one classic notch being 1.0 — that must accumulate before the wheel
	// produces a single detent step. fine-scroll devices report fractional
	// ticks, so the same value governs every wheel type. raising it makes
	// the wheel feel less sensitive; default 1.0 is one notch, one detent.
	WheelStepsPerZoomLevel float32

	// Locked suppresses node dragging only; selection, panning, zooming,
	// and link creation remain live.
	Locked bool

	// Style is a complete style value. the zero value derives
	// DefaultNodeCanvasStyle() lazily at the first Begin, when an imgui
	// context and theme are guaranteed live.
	Style NodeCanvasStyle
}

// View is the NodeCanvas view state: plain, persistable data. the view
// transform is screen = (canvas + pan) * zoom + origin, where origin is the
// canvas child region's screen-space top-left.
type View struct {
	Pan  imgui.Vec2
	Zoom float32
}

// Intents is what a frame's completed gestures produced, returned by End.
// the app applies intents (typically as undo commands); the canvas never
// assumes an intent was accepted.
type Intents[ID comparable] struct {
	// NodesMoved is emitted once, on drag release: one entry per selected
	// node that moved, in declaration order.
	NodesMoved []NodeMove[ID]

	// LinkCreated is emitted when a link drag releases on a compatible pin;
	// the app validates and applies it (or ignores it).
	LinkCreated *LinkCreate[ID]

	// SelectionChanged carries full replacement sets for nodes and links,
	// emitted only when an interaction produced sets differing from the
	// Selected flags declared this frame.
	SelectionChanged *SelectionChange[ID]
}

// NodeMove records one node's movement across a completed drag gesture:
// From is the declared position at gesture start, To is From plus the
// gesture's canvas-space offset.
type NodeMove[ID comparable] struct {
	ID       ID
	From, To imgui.Vec2
}

// LinkCreate records a completed link drag, normalized so FromPin is always
// the output-side pin and ToPin the input-side pin.
type LinkCreate[ID comparable] struct {
	FromPin, ToPin ID
}

// SelectionChange carries the full replacement selection: one combined set
// spanning both populations, split by kind, each in declaration order.
type SelectionChange[ID comparable] struct {
	Nodes, Links []ID
}

// NewNodeCanvas creates a NodeCanvas. it must not read imgui state:
// components are constructed before app.Run creates the context, so style
// defaulting is deferred to the first Begin.
func NewNodeCanvas[ID comparable](cfg NodeCanvasConfig) *NodeCanvas[ID] {
	detents := make([]float32, len(cfg.Detents))
	copy(detents, cfg.Detents)
	if len(detents) == 0 {
		detents = []float32{0.25, 0.5, 0.75, 1.0}
	}
	sort.Slice(detents, func(i, j int) bool { return detents[i] < detents[j] })
	if detents[len(detents)-1] != 1.0 {
		dl.Debugf("NodeCanvasConfig.Detents last entry is %v; the contract requires 1.0", detents[len(detents)-1])
	}

	gridSpacing := cfg.GridSpacing
	if gridSpacing <= 0 {
		gridSpacing = 50
	}

	wheelSteps := cfg.WheelStepsPerZoomLevel
	if wheelSteps <= 0 {
		wheelSteps = 1
	}

	return &NodeCanvas[ID]{
		detents:                detents,
		gridSpacing:            gridSpacing,
		locked:                 cfg.Locked,
		style:                  cfg.Style,
		styleSet:               cfg.Style != (NodeCanvasStyle{}),
		wheelStepsPerZoomLevel: wheelSteps,
		view:                   View{Zoom: detents[len(detents)-1]},
	}
}

// Begin establishes the canvas child region, draws the grid, and pushes the
// detent-scaled font. call it once per frame inside the owning component's
// Draw, declare nodes and links, then call End.
func (nc *NodeCanvas[ID]) Begin(state *State) {
	if nc.pendingView != nil {
		nc.view = *nc.pendingView
		nc.pendingView = nil
	}
	if nc.fitStarting {
		// a fresh zoom-to-fit descends from the top: the calling frame sets
		// the largest configured detent; End measures what actually got
		// declared there.
		nc.view.Zoom = nc.detents[len(nc.detents)-1]
		nc.fitStarting = false
	}
	if !nc.styleSet {
		nc.style = DefaultNodeCanvasStyle()
		nc.styleSet = true
	}

	imgui.BeginChildStrV(fmt.Sprintf("##nodecanvas%p", nc), state.Size, imgui.ChildFlagsNone,
		imgui.WindowFlagsNoScrollbar|imgui.WindowFlagsNoScrollWithMouse|imgui.WindowFlagsNoMove)
	nc.origin = imgui.WindowPos()
	nc.viewport = imgui.WindowSize()

	// frame-local draw values only: an in-flight pan renders from the
	// absolute mouse position against its anchor so the view doesn't trail
	// the mouse, but nothing commits here — End is the only phase that
	// commits, and the anchor-relative math makes the two agree by
	// construction.
	nc.frameView = nc.view
	nc.frameDragOffset = imgui.Vec2{}
	switch nc.gesture.kind {
	case gesturePan:
		mouse := imgui.MousePos()
		nc.frameView.Pan = imgui.Vec2{
			X: nc.gesture.panStart.X + (mouse.X-nc.gesture.pressScreen.X)/nc.view.Zoom,
			Y: nc.gesture.panStart.Y + (mouse.Y-nc.gesture.pressScreen.Y)/nc.view.Zoom,
		}
	case gestureDragNodes:
		mouseCanvas := canvasFromScreen(imgui.MousePos(), nc.frameView, nc.origin)
		nc.frameDragOffset = nc.gesture.dragOffset(mouseCanvas)
	}

	nc.drawGrid()

	// split the drawlist: channel 0 for links, two channels per node.
	// channel count is not knowable up front, so capacity comes from last
	// frame's node count with headroom of eight spare nodes.
	if nc.splitter == nil {
		nc.splitter = imgui.NewDrawListSplitter()
	}
	nc.splitCap = int32(1 + 2*(nc.lastNodeCount+8))
	nc.splitter.Split(imgui.WindowDrawList(), nc.splitCap)

	// the scaled font rides the whole declaration window. PushFont takes
	// the unscaled base size — never imgui.FontSize(), which is
	// post-global-scale and would double-apply DPI scaling.
	imgui.PushFont(imgui.CurrentFont(), imgui.CurrentStyle().FontSizeBase()*nc.frameView.Zoom)

	// reset the per-frame declaration buffers.
	nc.nodeIndex = 0
	nc.frameNodes = nil
	nc.framePins = make(map[ID]framePin)
	nc.frameLinks = nil
	nc.pinCounter = 0
	nc.itemActive = false
}

// Node declares one node for this frame. pos is the outer node rect's
// top-left in canvas space — the app-owned anchor: padding and title metrics
// offset content inward from it, never the card away from it, and content
// growth extends the rect right/down while the anchor stays fixed. the
// content closure runs with the canvas's scaled font pushed; at detent 1.0
// imgui/dfx widgets behave normally, below 1.0 content must not emit
// anything interactive — use NodeContext.Label and the pin rows.
func (nc *NodeCanvas[ID]) Node(id ID, pos imgui.Vec2, flags NodeFlags, content func(n *NodeContext[ID])) {
	drawList := imgui.WindowDrawList()
	idx := nc.nodeIndex
	nc.nodeIndex++
	chromeCh, contentCh := nc.nodeChannels(idx)

	drawPos := pos
	if nc.gesture.kind == gestureDragNodes && idInSlice(nc.gesture.dragSet, id) {
		drawPos = imgui.Vec2{X: pos.X + nc.frameDragOffset.X, Y: pos.Y + nc.frameDragOffset.Y}
	}

	zoom := nc.frameView.Zoom
	padding := nc.style.NodePadding

	// content flows inside its channel; chrome draws after content measures
	// it, and the channel split puts it visually behind.
	nc.splitter.SetCurrentChannel(drawList, contentCh)
	contentStart := screenFromCanvas(imgui.Vec2{X: drawPos.X + padding, Y: drawPos.Y + padding}, nc.frameView, nc.origin)
	imgui.SetCursorScreenPos(contentStart)

	// per-node ID scoping keeps imgui widget state stable across frames
	// regardless of declaration-order changes.
	imgui.PushIDStr(nodeScope(id))
	imgui.BeginGroup()
	ctx := &NodeContext[ID]{nc: nc, groupStart: contentStart}
	if content != nil {
		content(ctx)
	}
	imgui.EndGroup()
	// EndGroup propagates item status flags within the window: this is the
	// canvas-scoped active-item arbitration sample.
	nc.itemActive = nc.itemActive || imgui.IsItemActive()
	imgui.PopID()

	groupMin, groupMax := imgui.ItemRectMin(), imgui.ItemRectMax()
	width := (groupMax.X-groupMin.X)/zoom + 2*padding
	height := (groupMax.Y-groupMin.Y)/zoom + 2*padding

	// the recorded rect is anchored at the declared pos — graph truth stays
	// app-owned; the in-flight offset is view-side only.
	rect := canvasRect{Min: pos, Max: imgui.Vec2{X: pos.X + width, Y: pos.Y + height}}
	drawRect := canvasRect{Min: drawPos, Max: imgui.Vec2{X: drawPos.X + width, Y: drawPos.Y + height}}

	nc.drawNodeChrome(drawList, chromeCh, ctx, flags, drawRect, id)

	// record geometry: rect and pins anchored at the declared pos.
	offY := drawPos.Y - pos.Y
	pins := make([]pinGeometry[ID], 0, len(ctx.pins))
	for _, p := range ctx.pins {
		edgeX := rect.Min.X
		drawEdgeX := drawRect.Min.X
		if p.side == pinOutput {
			edgeX = rect.Max.X
			drawEdgeX = drawRect.Max.X
		}
		declared := imgui.Vec2{X: edgeX, Y: p.drawY - offY}
		pins = append(pins, pinGeometry[ID]{id: p.id, pos: declared, side: p.side, declIndex: p.declIndex})
		nc.framePins[p.id] = framePin{
			declared:  declared,
			draw:      imgui.Vec2{X: drawEdgeX, Y: p.drawY},
			side:      p.side,
			declIndex: p.declIndex,
		}
	}
	nc.frameNodes = append(nc.frameNodes, nodeGeometry[ID]{
		id:        id,
		rect:      rect,
		selected:  flags.Selected,
		declIndex: idx,
		pins:      pins,
	})
}

// Link declares one link for this frame. it only records the declaration —
// pins may belong to nodes declared later, so links resolve and draw at End.
// pin presence is a per-frame invariant: a link whose endpoint pin was not
// declared this frame is skipped for rendering and hit-testing.
func (nc *NodeCanvas[ID]) Link(id, fromPin, toPin ID, flags LinkFlags) {
	nc.frameLinks = append(nc.frameLinks, linkDecl[ID]{id: id, from: fromPin, to: toPin, selected: flags.Selected})
}

// End finishes the frame: it resolves and draws links, merges the splitter,
// samples input, runs the gesture state machine against this frame's
// geometry, commits view mutations, and returns the frame's intents.
func (nc *NodeCanvas[ID]) End() Intents[ID] {
	drawList := imgui.WindowDrawList()

	geomLinks, drawCubics := nc.resolveLinks()
	g := &canvasGeometry[ID]{
		view:     nc.frameView,
		origin:   nc.origin,
		viewport: nc.viewport,
		nodes:    nc.frameNodes,
		links:    geomLinks,
	}

	in := nc.sampleInput()

	// hover resolves against this frame's geometry, gated by the
	// widget-first cascade. links drawn below use it this frame; node chrome
	// consumed the previous frame's resolution.
	nc.hover = hitResult[ID]{}
	if in.canvasHovered && !in.anyPopupOpen && !in.itemHoveredInCanvas {
		nc.hover = hitTest(g, in.mouse, nc.style.hitParams())
	}

	nc.drawLinks(drawList, geomLinks, drawCubics)

	// merge first, then previews: in-flight gesture visuals are true
	// foreground, above all merged channels.
	nc.splitter.Merge(drawList)
	nc.drawGesturePreviews(drawList, in, g)
	imgui.PopFont()

	res := stepGesture(nc.gesture, in, g, gestureParams{
		hit:     nc.style.hitParams(),
		detents: nc.detents,
		locked:  nc.locked,
	})
	nc.gesture = res.state
	nc.view = res.view

	// most recent navigation wins: a wheel detent step or middle-drag pan
	// cancels a pending fit; otherwise the fit resolves one step against the
	// bounds actually declared this frame.
	if res.viewChanged {
		nc.fitPending = false
		nc.fitStarting = false
	} else if nc.fitPending && !nc.fitStarting {
		nc.resolveFit(g)
	}

	nc.retained = g
	nc.lastNodeCount = nc.nodeIndex

	imgui.EndChild()
	return res.intents
}

// resolveFit advances a pending zoom-to-fit by one frame: if the bounds
// declared at the current detent fit the viewport (or the smallest detent is
// reached), center and finish; otherwise step down one detent and stay
// pending. bounded by len(detents) frames.
func (nc *NodeCanvas[ID]) resolveFit(g *canvasGeometry[ID]) {
	bounds, ok := nodeBounds(g.nodes, nc.fitIDs)
	if !ok {
		nc.fitPending = false
		return
	}
	zoom := nc.view.Zoom
	if fitsAtZoom(bounds, nc.viewport, zoom, fitMargin) || zoom == nc.detents[0] {
		nc.view.Pan = centeredPan(bounds, nc.viewport, zoom)
		nc.fitPending = false
		return
	}
	next := stepDetent(nc.detents, zoom, -1)
	nc.view = View{Pan: centeredPan(bounds, nc.viewport, next), Zoom: next}
}

// nodeBounds unions the rects of the nodes matching ids (all nodes when ids
// is empty); non-node IDs are ignored.
func nodeBounds[ID comparable](nodes []nodeGeometry[ID], ids []ID) (canvasRect, bool) {
	var bounds canvasRect
	found := false
	for i := range nodes {
		if len(ids) > 0 && !idInSlice(ids, nodes[i].id) {
			continue
		}
		if !found {
			bounds = nodes[i].rect
			found = true
		} else {
			bounds = rectUnion(bounds, nodes[i].rect)
		}
	}
	return bounds, found
}

// NodeContext is the per-node declaration surface handed to a node's content
// closure.
type NodeContext[ID comparable] struct {
	nc         *NodeCanvas[ID]
	groupStart imgui.Vec2

	titleUsed   bool
	titleHeight float32 // canvas units, measured from the title closure's flow extent

	pins []pinRecord[ID]
}

// pinRecord is a pin row noted during content flow; marker X positions
// resolve at chrome time, once the node's final width is known.
type pinRecord[ID comparable] struct {
	id        ID
	side      pinSide
	drawY     float32 // row center in canvas space, draw-anchored
	declIndex int
}

// TitleBar declares the node's title region. when used it must be the first
// call in the content closure: its closure runs and is measured in flow like
// everything else, and its measured extent is the recorded title-band
// height — no fixed title metric exists.
func (n *NodeContext[ID]) TitleBar(content func()) {
	if n.titleUsed || imgui.CursorScreenPos() != n.groupStart {
		dl.Debugf("NodeContext.TitleBar must be the first call in the content closure")
	}
	imgui.BeginGroup()
	if content != nil {
		content()
	}
	imgui.EndGroup()
	if !n.titleUsed {
		n.titleHeight = (imgui.ItemRectMax().Y - n.groupStart.Y) / n.nc.frameView.Zoom
		n.titleUsed = true
	}
	// advance body content below the band: the band extends one padding
	// past the measured title bottom so the title sits vertically centered
	// (node padding above, the same below).
	imgui.Dummy(imgui.Vec2{Y: n.nc.style.NodePadding * n.nc.frameView.Zoom})
}

// Input declares an input pin row: a marker on the node's left edge anchored
// to this content row, with a label. pins must be declared at every detent —
// anything that should render, hit-test, snap, or anchor a link this frame
// must be declared this frame.
func (n *NodeContext[ID]) Input(pinID ID, label string) {
	n.pinRow(pinID, label, pinInput)
}

// Output declares an output pin row: a marker on the node's right edge
// anchored to this content row, with a label.
func (n *NodeContext[ID]) Output(pinID ID, label string) {
	n.pinRow(pinID, label, pinOutput)
}

// Label draws text through the drawlist and advances layout with an ID-less
// Dummy spacer: it measures like an item but stays invisible to the
// hover/active arbitration flags, so it is safe at every detent — the
// primitive reduced-detent content is built from.
func (n *NodeContext[ID]) Label(text string) {
	pos := imgui.CursorScreenPos()
	size := imgui.CalcTextSize(text)
	imgui.WindowDrawList().AddTextVec2(pos, n.nc.textColor(), text)
	imgui.Dummy(size)
}

// Detent returns the zoom factor the frame is drawing at, so content
// closures can branch without reaching back to the canvas.
func (n *NodeContext[ID]) Detent() float32 {
	return n.nc.frameView.Zoom
}

func (n *NodeContext[ID]) pinRow(pinID ID, label string, side pinSide) {
	rowTop := imgui.CursorScreenPos()
	size := imgui.CalcTextSize(label)
	imgui.WindowDrawList().AddTextVec2(rowTop, n.nc.textColor(), label)
	imgui.Dummy(size)

	centerY := canvasFromScreen(imgui.Vec2{Y: rowTop.Y + size.Y/2}, n.nc.frameView, n.nc.origin).Y
	n.pins = append(n.pins, pinRecord[ID]{id: pinID, side: side, drawY: centerY, declIndex: n.nc.pinCounter})
	n.nc.pinCounter++
}

// nodeChannels maps a node index to its chrome/content channel pair. nodes
// declared beyond the splitter's capacity share the final pair for the
// frame — a one-frame z-order artifact only when more than eight nodes
// appear at once.
func (nc *NodeCanvas[ID]) nodeChannels(idx int) (int32, int32) {
	chrome := int32(1 + 2*idx)
	if chrome+1 >= nc.splitCap {
		dl.Debugf("NodeCanvas node %d exceeds splitter capacity %d; sharing the final channel pair this frame", idx, nc.splitCap)
		chrome = nc.splitCap - 2
	}
	return chrome, chrome + 1
}

// drawNodeChrome draws the node card behind its measured content: body,
// title band, border, and pin markers.
func (nc *NodeCanvas[ID]) drawNodeChrome(drawList *imgui.DrawList, channel int32, ctx *NodeContext[ID], flags NodeFlags, drawRect canvasRect, id ID) {
	nc.splitter.SetCurrentChannel(drawList, channel)

	zoom := nc.frameView.Zoom
	screenMin := screenFromCanvas(drawRect.Min, nc.frameView, nc.origin)
	screenMax := screenFromCanvas(drawRect.Max, nc.frameView, nc.origin)
	rounding := nc.style.NodeRounding * zoom

	drawList.AddRectFilledV(screenMin, screenMax, imgui.ColorConvertFloat4ToU32(nc.style.NodeBodyColor), rounding, imgui.DrawFlagsRoundCornersAll)

	if ctx.titleUsed {
		bandColor := nc.style.TitleBandColor
		if flags.Selected {
			bandColor = nc.style.TitleBandColorSelected
		}
		bandBottom := screenFromCanvas(imgui.Vec2{Y: drawRect.Min.Y + 2*nc.style.NodePadding + ctx.titleHeight}, nc.frameView, nc.origin).Y
		drawList.AddRectFilledV(screenMin, imgui.Vec2{X: screenMax.X, Y: bandBottom},
			imgui.ColorConvertFloat4ToU32(bandColor), rounding, imgui.DrawFlagsRoundCornersTop)
	}

	borderColor := nc.style.NodeBorderColor
	thickness := nc.style.BorderThickness
	if flags.Selected {
		borderColor = nc.style.NodeBorderColorSelected
		thickness = nc.style.BorderThicknessSelected
	} else if nc.hover.kind == hitNode && nc.hover.node == id {
		borderColor = nc.style.NodeBorderColorHovered
	}
	drawList.AddRectV(screenMin, screenMax, imgui.ColorConvertFloat4ToU32(borderColor), rounding, imgui.DrawFlagsRoundCornersAll, thickness*zoom)

	for _, p := range ctx.pins {
		edgeX := drawRect.Min.X
		if p.side == pinOutput {
			edgeX = drawRect.Max.X
		}
		center := screenFromCanvas(imgui.Vec2{X: edgeX, Y: p.drawY}, nc.frameView, nc.origin)
		pinColor := nc.style.PinColor
		if nc.hover.kind == hitPin && nc.hover.pin == p.id {
			pinColor = nc.style.PinColorHovered
		}
		drawList.AddCircleFilled(center, nc.style.PinRadius*zoom, imgui.ColorConvertFloat4ToU32(pinColor))
	}
}

// resolveLinks resolves this frame's link declarations against the declared
// pins, producing hit geometry (declared-anchored) and draw cubics
// (in-flight drag offset applied). links with an undeclared endpoint are
// skipped — retained geometry never backs a declaration gap.
func (nc *NodeCanvas[ID]) resolveLinks() ([]linkGeometry[ID], [][4]imgui.Vec2) {
	var geom []linkGeometry[ID]
	var draw [][4]imgui.Vec2
	for _, ld := range nc.frameLinks {
		from, okFrom := nc.framePins[ld.from]
		to, okTo := nc.framePins[ld.to]
		if !okFrom || !okTo {
			dl.Debugf("NodeCanvas link %v skipped: endpoint pin not declared this frame", ld.id)
			continue
		}
		// orient by the from pin's side so a reversed declaration still
		// renders with sane tangents.
		declared := linkCubic(from.declared, to.declared, nc.style.LinkTangent)
		drawn := linkCubic(from.draw, to.draw, nc.style.LinkTangent)
		if from.side == pinInput {
			declared = linkCubic(to.declared, from.declared, nc.style.LinkTangent)
			drawn = linkCubic(to.draw, from.draw, nc.style.LinkTangent)
		}
		geom = append(geom, linkGeometry[ID]{id: ld.id, cubic: declared, selected: ld.selected, declIndex: len(geom)})
		draw = append(draw, drawn)
	}
	return geom, draw
}

// drawLinks draws the frame's links into channel 0, under every node.
func (nc *NodeCanvas[ID]) drawLinks(drawList *imgui.DrawList, geom []linkGeometry[ID], drawCubics [][4]imgui.Vec2) {
	nc.splitter.SetCurrentChannel(drawList, 0)
	zoom := nc.frameView.Zoom

	for i := range geom {
		color := nc.style.LinkColor
		thickness := nc.style.LinkThickness
		if geom[i].selected {
			color = nc.style.LinkColorSelected
			thickness = nc.style.LinkThickness * 1.5
		} else if nc.hover.kind == hitLink && nc.hover.link == geom[i].id {
			color = nc.style.LinkColorHovered
		}

		c := drawCubics[i]
		drawList.AddBezierCubic(
			screenFromCanvas(c[0], nc.frameView, nc.origin),
			screenFromCanvas(c[1], nc.frameView, nc.origin),
			screenFromCanvas(c[2], nc.frameView, nc.origin),
			screenFromCanvas(c[3], nc.frameView, nc.origin),
			imgui.ColorConvertFloat4ToU32(color), thickness*zoom)
	}
}

// drawGesturePreviews draws the in-flight box-select rect and link-drag
// preview from the gesture's absolute anchors and the current mouse
// position — the same math End commits with, so drawn and committed values
// agree by construction. pending states past the drag threshold preview too,
// so the visuals never trail the gesture by a frame.
func (nc *NodeCanvas[ID]) drawGesturePreviews(drawList *imgui.DrawList, in inputSnapshot, g *canvasGeometry[ID]) {
	kind := nc.gesture.kind
	if kind == gesturePressedEmpty && in.leftDragging {
		kind = gestureBoxSelect
	}
	if kind == gesturePressedPin && in.leftDragging {
		kind = gestureLinkDrag
	}

	switch kind {
	case gestureBoxSelect:
		a := screenFromCanvas(nc.gesture.pressCanvas, nc.frameView, nc.origin)
		r := normalizedRect(a, in.mouse)
		drawList.AddRectFilledV(r.Min, r.Max, imgui.ColorConvertFloat4ToU32(nc.style.BoxSelectFillColor), 0, imgui.DrawFlagsNone)
		drawList.AddRectV(r.Min, r.Max, imgui.ColorConvertFloat4ToU32(nc.style.BoxSelectBorderColor), 0, imgui.DrawFlagsNone, 1)

	case gestureLinkDrag:
		// the preview locks to a compatible pin within the snap radius.
		endpoint := canvasFromScreen(in.mouse, nc.frameView, nc.origin)
		if _, pos, ok := snapPin(g, in.mouse, nc.gesture.sourceSide, nc.style.hitParams()); ok {
			endpoint = pos
		}
		cubic := linkCubic(nc.gesture.sourcePos, endpoint, nc.style.LinkTangent)
		if nc.gesture.sourceSide == pinInput {
			cubic = linkCubic(endpoint, nc.gesture.sourcePos, nc.style.LinkTangent)
		}
		drawList.AddBezierCubic(
			screenFromCanvas(cubic[0], nc.frameView, nc.origin),
			screenFromCanvas(cubic[1], nc.frameView, nc.origin),
			screenFromCanvas(cubic[2], nc.frameView, nc.origin),
			screenFromCanvas(cubic[3], nc.frameView, nc.origin),
			imgui.ColorConvertFloat4ToU32(nc.style.LinkColorHovered), nc.style.LinkThickness*nc.frameView.Zoom)
	}
}

// textColor is the current imgui text color, used by drawlist text.
func (nc *NodeCanvas[ID]) textColor() uint32 {
	return imgui.ColorConvertFloat4ToU32(imgui.CurrentStyle().Colors()[imgui.ColText])
}

// View returns the view state, pending-first: a SetView not yet applied by
// Begin is returned so persistence paths that write then read round-trip
// correctly, while the draw transform stays frame-stable regardless.
func (nc *NodeCanvas[ID]) View() View {
	if nc.pendingView != nil {
		return *nc.pendingView
	}
	return nc.view
}

// SetView replaces the view state, with Zoom snapped to the nearest
// configured detent. callable any time — a call made mid-frame takes effect
// at the next frame's Begin, never mid-declaration. as an explicit
// navigation it cancels any pending zoom-to-fit.
func (nc *NodeCanvas[ID]) SetView(v View) {
	v.Zoom = nearestDetent(nc.detents, v.Zoom)
	nc.pendingView = &v
	nc.fitPending = false
	nc.fitStarting = false
}

// ZoomToFit fits nodes into the viewport: all nodes when ids is empty,
// non-node IDs ignored; a no-op when ids filter to zero nodes or before the
// first drawn frame. the fit resolves over the next frames (bounded by the
// detent count), descending from the top detent until the content declared
// at a detent fits — so app-side per-detent simplification is measured, not
// guessed. a new call replaces any pending fit; any other explicit
// navigation cancels it.
func (nc *NodeCanvas[ID]) ZoomToFit(ids ...ID) {
	if nc.retained == nil {
		return
	}
	if len(ids) > 0 {
		if _, ok := nodeBounds(nc.retained.nodes, ids); !ok {
			return
		}
	}
	nc.fitIDs = append([]ID(nil), ids...)
	nc.fitPending = true
	nc.fitStarting = true
	nc.pendingView = nil // the fit owns the view until it lands or is canceled
}

// CenterOn pans the bounds of the given nodes (all nodes when empty) to the
// viewport center; the detent is unchanged. same argument contract as
// ZoomToFit; cancels any pending fit.
func (nc *NodeCanvas[ID]) CenterOn(ids ...ID) {
	if nc.retained == nil {
		return
	}
	bounds, ok := nodeBounds(nc.retained.nodes, ids)
	if !ok {
		return
	}
	v := nc.View()
	v.Pan = centeredPan(bounds, nc.viewport, v.Zoom)
	nc.pendingView = &v
	nc.fitPending = false
	nc.fitStarting = false
}

// Detent returns the current zoom factor.
func (nc *NodeCanvas[ID]) Detent() float32 {
	return nc.View().Zoom
}

// GridSpacing returns the effective grid spacing in canvas units, including
// the applied default; exposed for app-side snap logic.
func (nc *NodeCanvas[ID]) GridSpacing() float32 {
	return nc.gridSpacing
}

// CanvasFromScreen converts a screen-space point to canvas space, against
// the pending-first view and the last begun canvas rect. before the first
// Begin the transform applies with a zero origin.
func (nc *NodeCanvas[ID]) CanvasFromScreen(p imgui.Vec2) imgui.Vec2 {
	return canvasFromScreen(p, nc.View(), nc.origin)
}

// ScreenFromCanvas converts a canvas-space point to screen space, against
// the pending-first view and the last begun canvas rect.
func (nc *NodeCanvas[ID]) ScreenFromCanvas(p imgui.Vec2) imgui.Vec2 {
	return screenFromCanvas(p, nc.View(), nc.origin)
}

// SetStyle replaces the style, e.g. when the app rebuilds it on theme
// change.
func (nc *NodeCanvas[ID]) SetStyle(s NodeCanvasStyle) {
	nc.style = s
	nc.styleSet = true
}

// SetLocked toggles locked mode live; like the config flag, it gates only
// node-drag arming.
func (nc *NodeCanvas[ID]) SetLocked(locked bool) {
	nc.locked = locked
}

// drawGrid draws the line grid directly into the window drawlist, under
// everything the frame declares.
func (nc *NodeCanvas[ID]) drawGrid() {
	drawList := imgui.WindowDrawList()
	// the grid fades at reduced detents: cells shrink toward busy, so alpha
	// tracks the zoom down.
	gridColor := nc.style.GridColor
	gridColor.W *= 0.25 + 0.75*nc.frameView.Zoom
	col := imgui.ColorConvertFloat4ToU32(gridColor)
	spacing := nc.gridSpacing

	canvasMin := canvasFromScreen(nc.origin, nc.frameView, nc.origin)
	canvasMax := canvasFromScreen(imgui.Vec2{X: nc.origin.X + nc.viewport.X, Y: nc.origin.Y + nc.viewport.Y}, nc.frameView, nc.origin)

	for x := float32(math.Floor(float64(canvasMin.X/spacing))) * spacing; x <= canvasMax.X; x += spacing {
		sx := screenFromCanvas(imgui.Vec2{X: x}, nc.frameView, nc.origin).X
		drawList.AddLineV(imgui.Vec2{X: sx, Y: nc.origin.Y}, imgui.Vec2{X: sx, Y: nc.origin.Y + nc.viewport.Y}, col, 1)
	}
	for y := float32(math.Floor(float64(canvasMin.Y/spacing))) * spacing; y <= canvasMax.Y; y += spacing {
		sy := screenFromCanvas(imgui.Vec2{Y: y}, nc.frameView, nc.origin).Y
		drawList.AddLineV(imgui.Vec2{X: nc.origin.X, Y: sy}, imgui.Vec2{X: nc.origin.X + nc.viewport.X, Y: sy}, col, 1)
	}
}

// nodeScope formats an ID for imgui ID scoping. pointer-kind IDs format by
// address — stable and unique — because fmt.Sprint renders a struct pointer
// as its field values, which mutate when the model mutates and collide when
// two models compare equal. non-pointer IDs must be unique and stable across
// frames in their fmt.Sprint form.
func nodeScope(id any) string {
	if v := reflect.ValueOf(id); v.Kind() == reflect.Pointer {
		return fmt.Sprintf("%p", id)
	}
	return fmt.Sprint(id)
}

// sampleInput builds the frame's input snapshot. it runs inside the canvas
// child window, before EndChild, so window-scoped queries answer for the
// canvas.
func (nc *NodeCanvas[ID]) sampleInput() inputSnapshot {
	io := imgui.CurrentIO()
	mouse := imgui.MousePos()
	hovered := imgui.IsWindowHoveredV(imgui.HoveredFlagsChildWindows)

	leftPressed := imgui.IsMouseClickedBool(imgui.MouseButtonLeft)
	leftDown := imgui.IsMouseDown(imgui.MouseButtonLeft)
	leftReleased := imgui.IsMouseReleased(imgui.MouseButtonLeft)

	// wheel travel is accumulated canvas-side and reported as one detent
	// step (sign only) once the threshold crosses. accumulation happens
	// only where the state machine could actually step — canvas hovered,
	// no popup open, no canvas item hovered, no gesture in flight — and
	// resets otherwise, so scrolling elsewhere in the app or mid-gesture
	// never banks travel that would later produce a phantom step. the
	// gesture sample is one frame stale (it is last frame's End result),
	// the same accepted staleness as origin and viewport.
	itemHovered := hovered && imgui.IsAnyItemHovered()
	anyPopup := imgui.IsPopupOpenStrV("", imgui.PopupFlagsAnyPopupId|imgui.PopupFlagsAnyPopupLevel)
	var wheel float32
	if hovered && !anyPopup && !itemHovered && nc.gesture.kind == gestureIdle {
		var dir int
		nc.wheelAccum, dir = advanceWheel(nc.wheelAccum, io.MouseWheel(), nc.wheelStepsPerZoomLevel)
		wheel = float32(dir)
	} else {
		nc.wheelAccum = 0
	}

	// sticky drag tracking against the press anchor: once the threshold is
	// exceeded at any point during the press, leftDragging holds through
	// the release frame.
	if leftPressed {
		nc.leftPressPos = mouse
		nc.leftDragSticky = false
	}
	if (leftDown || leftReleased) && !nc.leftDragSticky {
		threshold := io.MouseDragThreshold()
		dx, dy := mouse.X-nc.leftPressPos.X, mouse.Y-nc.leftPressPos.Y
		nc.leftDragSticky = dx*dx+dy*dy > threshold*threshold
	}

	return inputSnapshot{
		mouse:          mouse,
		leftPressed:    leftPressed,
		leftDown:       leftDown,
		leftReleased:   leftReleased,
		leftDragging:   nc.leftDragSticky,
		middlePressed:  imgui.IsMouseClickedBool(imgui.MouseButtonMiddle),
		middleDown:     imgui.IsMouseDown(imgui.MouseButtonMiddle),
		middleReleased: imgui.IsMouseReleased(imgui.MouseButtonMiddle),
		wheel:          wheel,
		ctrl:           io.KeyCtrl(),
		shift:          io.KeyShift(),
		canvasHovered:  hovered,
		// the canvas-scoped arbitration flags: item-hovered gates on the
		// canvas child (or a descendant) being hovered, so items elsewhere
		// in the app never suppress canvas input.
		itemHoveredInCanvas: itemHovered,
		itemActiveInCanvas:  nc.itemActiveInCanvas(),
		anyPopupOpen:        anyPopup,
	}
}

// itemActiveInCanvas reports whether the active imgui item lives in the
// canvas: the OR of per-node-group IsItemActive samples, plus a
// window-hierarchy check so an item active in a descendant window still
// counts — and a text field active in a sidebar never does.
func (nc *NodeCanvas[ID]) itemActiveInCanvas() bool {
	if nc.itemActive {
		return true
	}
	if !imgui.IsAnyItemActive() {
		return false
	}
	ctx := imgui.CurrentContext()
	if ctx == nil {
		return false
	}
	activeWindow := ctx.ActiveIdWindow()
	if activeWindow == nil {
		return false
	}
	return imgui.InternalIsWindowChildOf(activeWindow, imgui.InternalCurrentWindow(), false, false)
}
