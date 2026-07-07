package dfx

import "github.com/AllenDang/cimgui-go/imgui"

// this file is the NodeCanvas gesture state machine: pure functions over a
// sampled input snapshot and a frame's geometry snapshot. no imgui calls live
// here, so press-drag-release sequences are unit-testable headlessly. the
// canvas samples real input into inputSnapshot at End and applies the
// results; completed gestures produce the Intents the app consumes.

// Intents is what a frame's completed gestures produced, returned by End.
// the canvas never mutates the graph: the app applies intents (typically as
// undo commands), and the next frame's declarations are the only truth the
// canvas renders.
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

// inputSnapshot is one frame's sampled input, in screen space. the canvas
// builds it from imgui at End; tests build it by hand.
type inputSnapshot struct {
	mouse imgui.Vec2

	leftPressed, leftDown, leftReleased       bool
	middlePressed, middleDown, middleReleased bool

	// leftDragging is true once the drag threshold has been exceeded at any
	// point during the current left press, and stays true through the
	// release frame (sticky — not imgui's IsMouseDragging, which drops on
	// the release frame).
	leftDragging bool

	// wheel is this frame's vertical wheel movement; only its sign is used,
	// stepping one detent per frame.
	wheel float32

	ctrl, shift bool

	// canvasHovered is the idle gate: the canvas child window (or a
	// descendant) is hovered. without it a sidebar click would clear the
	// canvas selection.
	canvasHovered bool

	// itemHoveredInCanvas and itemActiveInCanvas are the canvas-scoped
	// arbitration flags of the widget-first cascade — never the raw global
	// queries, which would let a text field active in a sidebar suppress
	// every canvas gesture.
	itemHoveredInCanvas bool
	itemActiveInCanvas  bool

	// anyPopupOpen blocks all gesture initiation and detent stepping while
	// any popup is open, whoever owns it; in-flight gestures run to
	// completion.
	anyPopupOpen bool
}

type gestureKind int

const (
	gestureIdle gestureKind = iota

	// pending states: left button down, drag threshold not yet crossed.
	gesturePressedNode
	gesturePressedEmpty
	gesturePressedPin

	// active gestures.
	gestureDragNodes
	gestureBoxSelect
	gestureLinkDrag
	gesturePan
)

// gestureState is the canvas's in-flight gesture: the only input state
// retained across frames. anchors are absolute — every derived value (drag
// offset, box rect, pan) is recomputed from the current mouse position
// against them, never accumulated per-frame, so drawn and committed values
// agree by construction.
type gestureState[ID comparable] struct {
	kind gestureKind

	pressScreen imgui.Vec2 // mouse at press, screen space
	pressCanvas imgui.Vec2 // mouse at press, canvas space

	// gesturePressedNode / gestureDragNodes
	pressedNode       ID
	collapseOnRelease bool              // plain press on an already-selected node: collapse to {N} on release-without-drag
	dragSet           []ID              // post-transition selected nodes at press, declaration order
	dragStart         map[ID]imgui.Vec2 // declared positions at gesture start

	// gesturePressedPin / gestureLinkDrag
	sourcePin  ID
	sourceSide pinSide
	sourcePos  imgui.Vec2 // source pin position, canvas space

	// gesturePan
	panStart imgui.Vec2 // view.Pan at middle press
}

// dragOffset is the in-flight canvas-space offset of a node drag, derived
// from the absolute mouse position against the press anchor.
func (st *gestureState[ID]) dragOffset(mouseCanvas imgui.Vec2) imgui.Vec2 {
	return imgui.Vec2{X: mouseCanvas.X - st.pressCanvas.X, Y: mouseCanvas.Y - st.pressCanvas.Y}
}

// gestureParams is the configuration the state machine steps under.
type gestureParams struct {
	hit     hitParams
	detents []float32
	locked  bool // suppresses dragNodes arming only
}

// gestureResult is one step's outcome: the successor state, the (possibly
// mutated) view, and any intents completed gestures produced. viewChanged
// reports explicit navigation (wheel detent step, middle-drag pan) so the
// canvas can cancel a pending zoom-to-fit.
type gestureResult[ID comparable] struct {
	state       gestureState[ID]
	view        View
	viewChanged bool
	intents     Intents[ID]
}

// stepGesture advances the gesture state machine by one frame. it is the
// single place input commits: state transitions, view mutations, and intents
// all resolve here, against the current frame's geometry.
func stepGesture[ID comparable](st gestureState[ID], in inputSnapshot, g *canvasGeometry[ID], params gestureParams) gestureResult[ID] {
	res := gestureResult[ID]{state: st, view: g.view}

	switch st.kind {
	case gestureIdle:
		stepIdle(&res, in, g, params)
	case gesturePressedNode, gesturePressedEmpty, gesturePressedPin:
		stepPending(&res, in, g, params)
	case gestureDragNodes, gestureBoxSelect, gestureLinkDrag:
		stepActive(&res, in, g, params)
	case gesturePan:
		stepPan(&res, in)
	}

	return res
}

// stepIdle handles gesture initiation and detent stepping. initiation is
// gated: nothing starts unless the canvas is hovered and no popup is open;
// the widget-first cascade then arbitrates per input — left needs no item
// hovered or active in the canvas, wheel needs no item hovered, middle pan
// needs neither.
func stepIdle[ID comparable](res *gestureResult[ID], in inputSnapshot, g *canvasGeometry[ID], params gestureParams) {
	if !in.canvasHovered || in.anyPopupOpen {
		return
	}

	if in.leftPressed && !in.itemHoveredInCanvas && !in.itemActiveInCanvas {
		pressLeft(res, in, g, params)
		return
	}

	if in.middlePressed {
		res.state = gestureState[ID]{
			kind:        gesturePan,
			pressScreen: in.mouse,
			panStart:    g.view.Pan,
		}
		return
	}

	if in.wheel != 0 && !in.itemHoveredInCanvas {
		dir := 1
		if in.wheel < 0 {
			dir = -1
		}
		newZoom := stepDetent(params.detents, g.view.Zoom, dir)
		if newZoom != g.view.Zoom {
			res.view = zoomTowardPoint(g.view, g.origin, in.mouse, newZoom)
			res.viewChanged = true
		}
	}
}

// pressLeft routes a left press by what it hit: pin arms a link drag, node
// runs a selection transition and may arm a node drag, link runs a selection
// transition, empty arms a box select.
func pressLeft[ID comparable](res *gestureResult[ID], in inputSnapshot, g *canvasGeometry[ID], params gestureParams) {
	hit := hitTest(g, in.mouse, params.hit)
	mouseCanvas := canvasFromScreen(in.mouse, g.view, g.origin)

	switch hit.kind {
	case hitPin:
		res.state = gestureState[ID]{
			kind:        gesturePressedPin,
			pressScreen: in.mouse,
			pressCanvas: mouseCanvas,
			sourcePin:   hit.pin,
			sourceSide:  hit.side,
			sourcePos:   pinPosition(g, hit.pin),
		}

	case hitNode:
		pressNode(res, in, g, params, hit.node, mouseCanvas)

	case hitLink:
		// selection transitions on links emit on press and arm nothing.
		res.intents.SelectionChanged = linkSelectionTransition(g, hit.link, in.ctrl, in.shift)

	case hitNone:
		// empty presses emit nothing: they arm a pending gesture, and the
		// clear (click), box-select replacement (drag), or nothing (cancel)
		// resolves on release.
		res.state = gestureState[ID]{
			kind:        gesturePressedEmpty,
			pressScreen: in.mouse,
			pressCanvas: mouseCanvas,
		}
	}
}

// pressNode runs the node selection transition and arms a node drag when the
// pressed node is selected in the post-transition set. ctrl wins over shift.
func pressNode[ID comparable](res *gestureResult[ID], in inputSnapshot, g *canvasGeometry[ID], params gestureParams, nodeID ID, mouseCanvas imgui.Vec2) {
	declaredSelected := nodeSelected(g, nodeID)

	var postNodes, postLinks []ID
	collapseOnRelease := false

	switch {
	case in.ctrl:
		postNodes = selectNodes(g, func(n *nodeGeometry[ID]) bool { return n.selected != (n.id == nodeID) })
		postLinks = selectLinks(g, func(l *linkGeometry[ID]) bool { return l.selected })
	case in.shift:
		postNodes = selectNodes(g, func(n *nodeGeometry[ID]) bool { return n.selected || n.id == nodeID })
		postLinks = selectLinks(g, func(l *linkGeometry[ID]) bool { return l.selected })
	case declaredSelected && !params.locked:
		// a plain press on an already-selected node preserves the current
		// set so the whole selection can be dragged from any member; the
		// collapse to {N} emits on release-without-drag instead. locked mode
		// bypasses this — the delay exists solely to keep a multi-drag
		// startable.
		postNodes = selectNodes(g, func(n *nodeGeometry[ID]) bool { return n.selected })
		postLinks = selectLinks(g, func(l *linkGeometry[ID]) bool { return l.selected })
		collapseOnRelease = true
	default:
		postNodes = selectNodes(g, func(n *nodeGeometry[ID]) bool { return n.id == nodeID })
		postLinks = nil
	}

	res.intents.SelectionChanged = maybeSelectionChange(g, postNodes, postLinks)

	// arm the drag only when the pressed node is selected in the
	// post-transition set: ctrl-toggling a node off is a selection-only
	// interaction. locked mode blocks the arming and no pending state
	// survives the press.
	if params.locked || !idInSlice(postNodes, nodeID) {
		return
	}

	dragStart := make(map[ID]imgui.Vec2, len(postNodes))
	for i := range g.nodes {
		if idInSlice(postNodes, g.nodes[i].id) {
			dragStart[g.nodes[i].id] = g.nodes[i].rect.Min
		}
	}
	res.state = gestureState[ID]{
		kind:              gesturePressedNode,
		pressScreen:       in.mouse,
		pressCanvas:       mouseCanvas,
		pressedNode:       nodeID,
		collapseOnRelease: collapseOnRelease,
		dragSet:           postNodes,
		dragStart:         dragStart,
	}
}

// stepPending advances a pressed-but-not-yet-dragging state: crossing the
// drag threshold starts the armed gesture, release before it is a click, and
// lost button state cancels with no intent.
func stepPending[ID comparable](res *gestureResult[ID], in inputSnapshot, g *canvasGeometry[ID], params gestureParams) {
	switch {
	case in.leftReleased:
		if in.leftDragging {
			// the threshold and the release landed in the same frame:
			// complete the armed gesture rather than treating it as a click.
			completeActive(res, in, g, params, res.state.kind)
		} else {
			clickPending(res, g)
		}
		res.state = gestureState[ID]{kind: gestureIdle}

	case in.leftDragging:
		switch res.state.kind {
		case gesturePressedNode:
			res.state.kind = gestureDragNodes
		case gesturePressedEmpty:
			res.state.kind = gestureBoxSelect
		case gesturePressedPin:
			res.state.kind = gestureLinkDrag
		}

	case !in.leftDown:
		// button state vanished without a release event: cancel, no intent.
		res.state = gestureState[ID]{kind: gestureIdle}
	}
}

// clickPending resolves a release-before-threshold: a click. empty presses
// emit their clear here, and the already-selected-node exception collapses
// to the pressed node here.
func clickPending[ID comparable](res *gestureResult[ID], g *canvasGeometry[ID]) {
	switch res.state.kind {
	case gesturePressedNode:
		if res.state.collapseOnRelease {
			nodeID := res.state.pressedNode
			postNodes := selectNodes(g, func(n *nodeGeometry[ID]) bool { return n.id == nodeID })
			res.intents.SelectionChanged = maybeSelectionChange(g, postNodes, nil)
		}
	case gesturePressedEmpty:
		res.intents.SelectionChanged = maybeSelectionChange(g, nil, nil)
	case gesturePressedPin:
		// a click on a pin has no meaning: no intent.
	}
}

// stepActive advances an in-flight left-button gesture. a release is a
// release wherever the mouse is: the gesture completes with the last sampled
// position, even outside the canvas. lost button state cancels, no intent.
func stepActive[ID comparable](res *gestureResult[ID], in inputSnapshot, g *canvasGeometry[ID], params gestureParams) {
	switch {
	case in.leftReleased:
		completeActive(res, in, g, params, res.state.kind)
		res.state = gestureState[ID]{kind: gestureIdle}
	case !in.leftDown:
		res.state = gestureState[ID]{kind: gestureIdle}
	}
}

// completeActive resolves a released gesture into its intent.
func completeActive[ID comparable](res *gestureResult[ID], in inputSnapshot, g *canvasGeometry[ID], params gestureParams, kind gestureKind) {
	mouseCanvas := canvasFromScreen(in.mouse, g.view, g.origin)

	switch kind {
	case gestureDragNodes, gesturePressedNode:
		offset := res.state.dragOffset(mouseCanvas)
		var moves []NodeMove[ID]
		for _, id := range res.state.dragSet {
			from := res.state.dragStart[id]
			to := imgui.Vec2{X: from.X + offset.X, Y: from.Y + offset.Y}
			if to != from {
				moves = append(moves, NodeMove[ID]{ID: id, From: from, To: to})
			}
		}
		res.intents.NodesMoved = moves

	case gestureBoxSelect, gesturePressedEmpty:
		rect := normalizedRect(res.state.pressCanvas, mouseCanvas)
		postNodes := selectNodes(g, func(n *nodeGeometry[ID]) bool { return n.rect.overlaps(rect) })
		postLinks := selectLinks(g, func(l *linkGeometry[ID]) bool { return bezierIntersectsRect(l.cubic, rect) })
		res.intents.SelectionChanged = maybeSelectionChange(g, postNodes, postLinks)

	case gestureLinkDrag, gesturePressedPin:
		pin, _, ok := snapPin(g, in.mouse, res.state.sourceSide, params.hit)
		if !ok {
			return // release off any compatible pin: no intent.
		}
		if res.state.sourceSide == pinOutput {
			res.intents.LinkCreated = &LinkCreate[ID]{FromPin: res.state.sourcePin, ToPin: pin}
		} else {
			res.intents.LinkCreated = &LinkCreate[ID]{FromPin: pin, ToPin: res.state.sourcePin}
		}
	}
}

// stepPan recomputes the pan from the absolute mouse position against the
// press anchor and commits it to the view each frame; release (or lost
// button state) returns to idle with the last applied pan kept.
func stepPan[ID comparable](res *gestureResult[ID], in inputSnapshot) {
	if in.middleDown || in.middleReleased {
		pan := imgui.Vec2{
			X: res.state.panStart.X + (in.mouse.X-res.state.pressScreen.X)/res.view.Zoom,
			Y: res.state.panStart.Y + (in.mouse.Y-res.state.pressScreen.Y)/res.view.Zoom,
		}
		if pan != res.view.Pan {
			res.view.Pan = pan
			res.viewChanged = true
		}
	}
	if !in.middleDown {
		res.state = gestureState[ID]{kind: gestureIdle}
	}
}

// linkSelectionTransition applies the selection table's link column: plain
// replaces the whole set with {L}, ctrl toggles L keeping all else, shift
// adds L keeping all else. ctrl wins when both modifiers are held.
func linkSelectionTransition[ID comparable](g *canvasGeometry[ID], linkID ID, ctrl, shift bool) *SelectionChange[ID] {
	var postNodes, postLinks []ID
	switch {
	case ctrl:
		postNodes = selectNodes(g, func(n *nodeGeometry[ID]) bool { return n.selected })
		postLinks = selectLinks(g, func(l *linkGeometry[ID]) bool { return l.selected != (l.id == linkID) })
	case shift:
		postNodes = selectNodes(g, func(n *nodeGeometry[ID]) bool { return n.selected })
		postLinks = selectLinks(g, func(l *linkGeometry[ID]) bool { return l.selected || l.id == linkID })
	default:
		postNodes = nil
		postLinks = selectLinks(g, func(l *linkGeometry[ID]) bool { return l.id == linkID })
	}
	return maybeSelectionChange(g, postNodes, postLinks)
}

// maybeSelectionChange compares post-transition sets against the frame's
// declared Selected flags and returns a replacement intent only when they
// differ — the canvas compares against what the app declared this frame,
// never its own memory of past selections.
func maybeSelectionChange[ID comparable](g *canvasGeometry[ID], postNodes, postLinks []ID) *SelectionChange[ID] {
	declNodes := selectNodes(g, func(n *nodeGeometry[ID]) bool { return n.selected })
	declLinks := selectLinks(g, func(l *linkGeometry[ID]) bool { return l.selected })
	if idSlicesEqual(postNodes, declNodes) && idSlicesEqual(postLinks, declLinks) {
		return nil
	}
	return &SelectionChange[ID]{Nodes: postNodes, Links: postLinks}
}

// selectNodes collects node IDs matching the predicate, in declaration order.
func selectNodes[ID comparable](g *canvasGeometry[ID], include func(*nodeGeometry[ID]) bool) []ID {
	var ids []ID
	for i := range g.nodes {
		if include(&g.nodes[i]) {
			ids = append(ids, g.nodes[i].id)
		}
	}
	return ids
}

// selectLinks collects link IDs matching the predicate, in declaration order.
func selectLinks[ID comparable](g *canvasGeometry[ID], include func(*linkGeometry[ID]) bool) []ID {
	var ids []ID
	for i := range g.links {
		if include(&g.links[i]) {
			ids = append(ids, g.links[i].id)
		}
	}
	return ids
}

func nodeSelected[ID comparable](g *canvasGeometry[ID], id ID) bool {
	for i := range g.nodes {
		if g.nodes[i].id == id {
			return g.nodes[i].selected
		}
	}
	return false
}

func pinPosition[ID comparable](g *canvasGeometry[ID], id ID) imgui.Vec2 {
	for i := range g.nodes {
		for j := range g.nodes[i].pins {
			if g.nodes[i].pins[j].id == id {
				return g.nodes[i].pins[j].pos
			}
		}
	}
	return imgui.Vec2{}
}

func idInSlice[ID comparable](ids []ID, id ID) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

// idSlicesEqual compares two declaration-ordered ID slices.
func idSlicesEqual[ID comparable](a, b []ID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
