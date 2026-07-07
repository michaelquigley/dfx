package dfx

import (
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
)

// the input tests drive the gesture state machine headlessly over the
// synthetic fixture from nodeCanvasGeometry_test.go, exercising the
// interaction grammar: press/click/ctrl/shift/drag/box/link-snap/cancel and
// locked mode.

func testGestureParams() gestureParams {
	return gestureParams{
		hit:     testHitParams(),
		detents: []float32{0.25, 0.5, 0.75, 1.0},
	}
}

func snapPress(x, y float32) inputSnapshot {
	return inputSnapshot{
		mouse:         imgui.Vec2{X: x, Y: y},
		leftPressed:   true,
		leftDown:      true,
		canvasHovered: true,
	}
}

func snapDrag(x, y float32) inputSnapshot {
	return inputSnapshot{
		mouse:         imgui.Vec2{X: x, Y: y},
		leftDown:      true,
		leftDragging:  true,
		canvasHovered: true,
	}
}

func snapRelease(x, y float32, dragging bool) inputSnapshot {
	return inputSnapshot{
		mouse:         imgui.Vec2{X: x, Y: y},
		leftReleased:  true,
		leftDragging:  dragging,
		canvasHovered: true,
	}
}

func withMods(in inputSnapshot, ctrl, shift bool) inputSnapshot {
	in.ctrl = ctrl
	in.shift = shift
	return in
}

// run steps the machine through a sequence of snapshots against a fixed
// geometry, returning the final result; intermediate intents fail the test.
func run(t *testing.T, g *canvasGeometry[string], params gestureParams, ins ...inputSnapshot) gestureResult[string] {
	t.Helper()
	var st gestureState[string]
	var res gestureResult[string]
	for i, in := range ins {
		res = stepGesture(st, in, g, params)
		g.view = res.view
		st = res.state
		if i < len(ins)-1 {
			if res.intents.SelectionChanged != nil || res.intents.NodesMoved != nil || res.intents.LinkCreated != nil {
				t.Fatalf("unexpected intents at step %d: %+v", i, res.intents)
			}
		}
	}
	return res
}

func wantSelection(t *testing.T, res gestureResult[string], nodes, links []string) {
	t.Helper()
	sel := res.intents.SelectionChanged
	if sel == nil {
		t.Fatalf("expected SelectionChanged{%v, %v}, got none", nodes, links)
	}
	if !idSlicesEqual(sel.Nodes, nodes) || !idSlicesEqual(sel.Links, links) {
		t.Fatalf("expected SelectionChanged{%v, %v}, got {%v, %v}", nodes, links, sel.Nodes, sel.Links)
	}
}

func wantNoIntents(t *testing.T, res gestureResult[string]) {
	t.Helper()
	if res.intents.SelectionChanged != nil || res.intents.NodesMoved != nil || res.intents.LinkCreated != nil {
		t.Fatalf("expected no intents, got %+v", res.intents)
	}
}

func TestNodeCanvas_IdleGate(t *testing.T) {
	g := testGraph([]string{"a"}, nil)
	params := testGestureParams()

	// press while not hovered: ignored.
	in := snapPress(150, 140)
	in.canvasHovered = false
	res := stepGesture(gestureState[string]{}, in, g, params)
	wantNoIntents(t, res)
	if res.state.kind != gestureIdle {
		t.Fatalf("expected idle, got %v", res.state.kind)
	}

	// press while a popup is open: ignored.
	in = snapPress(150, 140)
	in.anyPopupOpen = true
	res = stepGesture(gestureState[string]{}, in, g, params)
	wantNoIntents(t, res)
	if res.state.kind != gestureIdle {
		t.Fatalf("expected idle, got %v", res.state.kind)
	}

	// wheel while not hovered: no detent step.
	in = inputSnapshot{mouse: imgui.Vec2{X: 150, Y: 140}, wheel: -1}
	if res := stepGesture(gestureState[string]{}, in, g, params); res.viewChanged {
		t.Fatal("wheel stepped the detent while not hovered")
	}
}

func TestNodeCanvas_WidgetFirstArbitration(t *testing.T) {
	g := testGraph(nil, nil)
	params := testGestureParams()

	// a hovered item blocks left initiation and wheel, but not middle pan.
	in := snapPress(150, 140)
	in.itemHoveredInCanvas = true
	res := stepGesture(gestureState[string]{}, in, g, params)
	wantNoIntents(t, res)
	if res.state.kind != gestureIdle {
		t.Fatal("left press initiated over a hovered item")
	}

	in = inputSnapshot{mouse: imgui.Vec2{X: 150, Y: 140}, wheel: -1, canvasHovered: true, itemHoveredInCanvas: true}
	if res := stepGesture(gestureState[string]{}, in, g, params); res.viewChanged {
		t.Fatal("wheel stepped the detent over a hovered item")
	}

	in = inputSnapshot{mouse: imgui.Vec2{X: 150, Y: 140}, middlePressed: true, middleDown: true, canvasHovered: true, itemHoveredInCanvas: true, itemActiveInCanvas: true}
	if res := stepGesture(gestureState[string]{}, in, g, params); res.state.kind != gesturePan {
		t.Fatal("middle pan blocked by item state")
	}

	// an active item blocks left initiation but not wheel.
	in = snapPress(150, 140)
	in.itemActiveInCanvas = true
	res = stepGesture(gestureState[string]{}, in, g, params)
	if res.state.kind != gestureIdle {
		t.Fatal("left press initiated over an active item")
	}

	in = inputSnapshot{mouse: imgui.Vec2{X: 150, Y: 140}, wheel: -1, canvasHovered: true, itemActiveInCanvas: true}
	if res := stepGesture(gestureState[string]{}, in, g, params); !res.viewChanged {
		t.Fatal("wheel blocked by an active (not hovered) item")
	}
}

func TestNodeCanvas_PlainClickUnselectedNode(t *testing.T) {
	g := testGraph(nil, nil)
	res := run(t, g, testGestureParams(), snapPress(150, 140))
	wantSelection(t, res, []string{"a"}, nil)
	if res.state.kind != gesturePressedNode {
		t.Fatalf("expected pressedNode, got %v", res.state.kind)
	}

	// release without drag: selection already applied on press, no more intents.
	res2 := stepGesture(res.state, snapRelease(150, 140, false), g, testGestureParams())
	wantNoIntents(t, res2)
	if res2.state.kind != gestureIdle {
		t.Fatalf("expected idle, got %v", res2.state.kind)
	}
}

func TestNodeCanvas_PlainClickReplacesLinksToo(t *testing.T) {
	g := testGraph([]string{"b"}, []string{"l1"})
	res := run(t, g, testGestureParams(), snapPress(150, 140))
	wantSelection(t, res, []string{"a"}, nil)
}

func TestNodeCanvas_PlainPressSelectedNodePreservesThenCollapses(t *testing.T) {
	g := testGraph([]string{"a", "b"}, nil)
	params := testGestureParams()

	// press emits nothing: the set is preserved so a multi-drag can start
	// from any member.
	res := run(t, g, params, snapPress(150, 140))
	wantNoIntents(t, res)
	if res.state.kind != gesturePressedNode || !res.state.collapseOnRelease {
		t.Fatalf("expected pressedNode with pending collapse, got %+v", res.state)
	}

	// release without drag: the collapse to {a} emits now.
	res = stepGesture(res.state, snapRelease(150, 140, false), g, params)
	wantSelection(t, res, []string{"a"}, nil)
}

func TestNodeCanvas_PlainClickSoleSelectedNodeEmitsNothing(t *testing.T) {
	g := testGraph([]string{"a"}, nil)
	params := testGestureParams()
	res := run(t, g, params, snapPress(150, 140))
	wantNoIntents(t, res)
	// the collapse target equals the declared selection: no intent.
	res = stepGesture(res.state, snapRelease(150, 140, false), g, params)
	wantNoIntents(t, res)
}

func TestNodeCanvas_DragSelectedNodesMovesWholeSelection(t *testing.T) {
	g := testGraph([]string{"a", "b"}, nil)
	res := run(t, g, testGestureParams(),
		snapPress(150, 140),
		snapDrag(160, 150),
		snapRelease(170, 160, true),
	)
	if res.intents.SelectionChanged != nil {
		t.Fatalf("unexpected selection change on drag: %+v", res.intents.SelectionChanged)
	}
	moves := res.intents.NodesMoved
	if len(moves) != 2 {
		t.Fatalf("expected 2 moves, got %+v", moves)
	}
	// declaration order, From = declared position, To = From + offset (20,20).
	if moves[0].ID != "a" || !approxVec2(moves[0].From, imgui.Vec2{X: 100, Y: 100}) || !approxVec2(moves[0].To, imgui.Vec2{X: 120, Y: 120}) {
		t.Fatalf("unexpected move[0]: %+v", moves[0])
	}
	if moves[1].ID != "b" || !approxVec2(moves[1].From, imgui.Vec2{X: 300, Y: 100}) || !approxVec2(moves[1].To, imgui.Vec2{X: 320, Y: 120}) {
		t.Fatalf("unexpected move[1]: %+v", moves[1])
	}
}

func TestNodeCanvas_DragUnselectedNodeSelectsThenMovesIt(t *testing.T) {
	g := testGraph([]string{"b"}, nil)
	params := testGestureParams()

	// press replaces the selection with {a} immediately...
	res := run(t, g, params, snapPress(150, 140))
	wantSelection(t, res, []string{"a"}, nil)

	// ...and the drag moves only a.
	res = stepGesture(res.state, snapDrag(160, 150), g, params)
	res = stepGesture(res.state, snapRelease(170, 160, true), g, params)
	if len(res.intents.NodesMoved) != 1 || res.intents.NodesMoved[0].ID != "a" {
		t.Fatalf("expected a alone to move, got %+v", res.intents.NodesMoved)
	}
}

func TestNodeCanvas_DragBackToStartEmitsNoMoves(t *testing.T) {
	g := testGraph([]string{"a"}, nil)
	params := testGestureParams()
	res := run(t, g, params, snapPress(150, 140), snapDrag(180, 170))
	res = stepGesture(res.state, snapRelease(150, 140, true), g, params)
	if res.intents.NodesMoved != nil {
		t.Fatalf("expected no moves for a zero offset, got %+v", res.intents.NodesMoved)
	}
}

func TestNodeCanvas_CtrlToggle(t *testing.T) {
	params := testGestureParams()

	// toggle on: added to the existing set, arms a drag of the whole
	// post-transition selection.
	g := testGraph([]string{"b"}, []string{"l1"})
	res := run(t, g, params, withMods(snapPress(150, 140), true, false))
	wantSelection(t, res, []string{"a", "b"}, []string{"l1"})
	if res.state.kind != gesturePressedNode || !idSlicesEqual(res.state.dragSet, []string{"a", "b"}) {
		t.Fatalf("expected armed drag of {a b}, got %+v", res.state)
	}

	// toggle off: selection-only; crossing the drag threshold afterward
	// does nothing.
	g = testGraph([]string{"a", "b"}, nil)
	res = run(t, g, params, withMods(snapPress(150, 140), true, false))
	wantSelection(t, res, []string{"b"}, nil)
	if res.state.kind != gestureIdle {
		t.Fatalf("expected no pending state after toggle-off, got %v", res.state.kind)
	}
	res = stepGesture(res.state, snapDrag(180, 170), g, params)
	res = stepGesture(res.state, snapRelease(200, 190, true), g, params)
	wantNoIntents(t, res)
}

func TestNodeCanvas_ShiftAdd(t *testing.T) {
	params := testGestureParams()

	g := testGraph([]string{"b"}, []string{"l1"})
	res := run(t, g, params, withMods(snapPress(150, 140), false, true))
	wantSelection(t, res, []string{"a", "b"}, []string{"l1"})

	// shift on an already-selected node changes nothing (no intent) but
	// still arms the drag.
	g = testGraph([]string{"a", "b"}, nil)
	res = run(t, g, params, withMods(snapPress(150, 140), false, true))
	wantNoIntents(t, res)
	if res.state.kind != gesturePressedNode || !idSlicesEqual(res.state.dragSet, []string{"a", "b"}) {
		t.Fatalf("expected armed drag of {a b}, got %+v", res.state)
	}
}

func TestNodeCanvas_CtrlWinsOverShift(t *testing.T) {
	g := testGraph([]string{"a"}, nil)
	res := run(t, g, testGestureParams(), withMods(snapPress(150, 140), true, true))
	// ctrl toggles a off; shift would have kept it.
	wantSelection(t, res, nil, nil)
}

func TestNodeCanvas_LinkClick(t *testing.T) {
	params := testGestureParams()

	// plain: replace the whole set with {l1}.
	g := testGraph([]string{"a"}, nil)
	res := run(t, g, params, snapPress(250, 141))
	wantSelection(t, res, nil, []string{"l1"})
	if res.state.kind != gestureIdle {
		t.Fatalf("link press should arm nothing, got %v", res.state.kind)
	}

	// ctrl: toggle l1 keeping nodes.
	g = testGraph([]string{"a"}, []string{"l1"})
	res = run(t, g, params, withMods(snapPress(250, 141), true, false))
	wantSelection(t, res, []string{"a"}, nil)

	// shift: add l1 keeping nodes.
	g = testGraph([]string{"a"}, nil)
	res = run(t, g, params, withMods(snapPress(250, 141), false, true))
	wantSelection(t, res, []string{"a"}, []string{"l1"})
}

func TestNodeCanvas_ClickEmptyClears(t *testing.T) {
	params := testGestureParams()

	// press emits nothing; the clear lands on release-before-threshold.
	g := testGraph([]string{"a"}, []string{"l1"})
	res := run(t, g, params, snapPress(600, 400))
	wantNoIntents(t, res)
	if res.state.kind != gesturePressedEmpty {
		t.Fatalf("expected pressedEmpty, got %v", res.state.kind)
	}
	res = stepGesture(res.state, snapRelease(600, 400, false), g, params)
	wantSelection(t, res, nil, nil)

	// with nothing selected, the clear equals the declared sets: no intent.
	g = testGraph(nil, nil)
	res = run(t, g, params, snapPress(600, 400))
	res = stepGesture(res.state, snapRelease(600, 400, false), g, params)
	wantNoIntents(t, res)
}

func TestNodeCanvas_BoxSelect(t *testing.T) {
	g := testGraph(nil, nil)
	params := testGestureParams()

	res := run(t, g, params, snapPress(50, 50), snapDrag(150, 100))
	if res.state.kind != gestureBoxSelect {
		t.Fatalf("expected boxSelect, got %v", res.state.kind)
	}

	// the rect (50,50)-(250,150) touches node a and link l1 but not node b.
	res = stepGesture(res.state, snapRelease(250, 150, true), g, params)
	wantSelection(t, res, []string{"a"}, []string{"l1"})
}

func TestNodeCanvas_BoxSelectReplacesExistingSelection(t *testing.T) {
	// a box catching nothing replaces a non-empty selection with empty sets.
	g := testGraph([]string{"b"}, []string{"l1"})
	params := testGestureParams()
	res := run(t, g, params, snapPress(500, 400), snapDrag(550, 450))
	res = stepGesture(res.state, snapRelease(600, 500, true), g, params)
	wantSelection(t, res, nil, nil)
}

func TestNodeCanvas_BoxSelectCancelEmitsNothing(t *testing.T) {
	g := testGraph([]string{"a"}, nil)
	params := testGestureParams()
	res := run(t, g, params, snapPress(50, 50), snapDrag(150, 100))
	// button state vanished without a release event: cancel, no intent.
	res = stepGesture(res.state, inputSnapshot{mouse: imgui.Vec2{X: 150, Y: 100}}, g, params)
	wantNoIntents(t, res)
	if res.state.kind != gestureIdle {
		t.Fatalf("expected idle after cancel, got %v", res.state.kind)
	}
}

func TestNodeCanvas_DragCancelEmitsNothing(t *testing.T) {
	g := testGraph([]string{"a"}, nil)
	params := testGestureParams()
	res := run(t, g, params, snapPress(150, 140), snapDrag(180, 170))
	res = stepGesture(res.state, inputSnapshot{mouse: imgui.Vec2{X: 180, Y: 170}}, g, params)
	if res.intents.NodesMoved != nil {
		t.Fatalf("cancel emitted moves: %+v", res.intents.NodesMoved)
	}
	if res.state.kind != gestureIdle {
		t.Fatalf("expected idle after cancel, got %v", res.state.kind)
	}
}

func TestNodeCanvas_FlickCompletesDragInOneFrame(t *testing.T) {
	// threshold and release land in the same frame: the sticky leftDragging
	// flag makes it a drag completion, not a click.
	g := testGraph(nil, nil)
	params := testGestureParams()
	res := run(t, g, params, snapPress(150, 140))
	res = stepGesture(res.state, snapRelease(190, 180, true), g, params)
	if len(res.intents.NodesMoved) != 1 || !approxVec2(res.intents.NodesMoved[0].To, imgui.Vec2{X: 140, Y: 140}) {
		t.Fatalf("expected flick to complete the drag, got %+v", res.intents.NodesMoved)
	}
}

func TestNodeCanvas_ReleaseOffCanvasStillCompletes(t *testing.T) {
	// a release is a release wherever the mouse is.
	g := testGraph([]string{"a"}, nil)
	params := testGestureParams()
	res := run(t, g, params, snapPress(150, 140), snapDrag(180, 170))
	off := snapRelease(1200, 900, true)
	off.canvasHovered = false
	res = stepGesture(res.state, off, g, params)
	if len(res.intents.NodesMoved) != 1 {
		t.Fatalf("expected the drag to complete off-canvas, got %+v", res.intents.NodesMoved)
	}
}

func TestNodeCanvas_LinkDragSnapsAndNormalizes(t *testing.T) {
	params := testGestureParams()

	// output -> input.
	g := testGraph(nil, nil)
	res := run(t, g, params, snapPress(200, 140), snapDrag(250, 140))
	if res.state.kind != gestureLinkDrag {
		t.Fatalf("expected linkDrag, got %v", res.state.kind)
	}
	res = stepGesture(res.state, snapRelease(295, 138, true), g, params)
	if lc := res.intents.LinkCreated; lc == nil || lc.FromPin != "a.out" || lc.ToPin != "b.in" {
		t.Fatalf("expected LinkCreated{a.out, b.in}, got %+v", res.intents.LinkCreated)
	}

	// input -> output normalizes to the same shape.
	res = run(t, g, params, snapPress(300, 140), snapDrag(250, 140))
	res = stepGesture(res.state, snapRelease(203, 142, true), g, params)
	if lc := res.intents.LinkCreated; lc == nil || lc.FromPin != "a.out" || lc.ToPin != "b.in" {
		t.Fatalf("expected normalized LinkCreated{a.out, b.in}, got %+v", res.intents.LinkCreated)
	}
}

func TestNodeCanvas_LinkDragReleaseOffPinEmitsNothing(t *testing.T) {
	g := testGraph(nil, nil)
	params := testGestureParams()
	res := run(t, g, params, snapPress(200, 140), snapDrag(250, 140))
	res = stepGesture(res.state, snapRelease(500, 400, true), g, params)
	wantNoIntents(t, res)

	// release near a same-side pin: not compatible, no intent.
	res = run(t, g, params, snapPress(200, 140), snapDrag(250, 140))
	res = stepGesture(res.state, snapRelease(398, 141, true), g, params)
	wantNoIntents(t, res)
}

func TestNodeCanvas_PinClickEmitsNothing(t *testing.T) {
	g := testGraph(nil, nil)
	params := testGestureParams()
	res := run(t, g, params, snapPress(200, 140))
	if res.state.kind != gesturePressedPin {
		t.Fatalf("expected pressedPin, got %v", res.state.kind)
	}
	res = stepGesture(res.state, snapRelease(200, 140, false), g, params)
	wantNoIntents(t, res)
}

func TestNodeCanvas_LockedMode(t *testing.T) {
	params := testGestureParams()
	params.locked = true

	// plain press on an already-selected node: the collapse applies on
	// press — the release delay exists only to keep a multi-drag startable.
	g := testGraph([]string{"a", "b"}, nil)
	res := run(t, g, params, snapPress(150, 140))
	wantSelection(t, res, []string{"a"}, nil)
	if res.state.kind != gestureIdle {
		t.Fatalf("no pending state survives a locked press, got %v", res.state.kind)
	}

	// dragging after a locked press does nothing.
	res = stepGesture(res.state, snapDrag(180, 170), g, params)
	res = stepGesture(res.state, snapRelease(200, 190, true), g, params)
	wantNoIntents(t, res)

	// box select stays live.
	g = testGraph(nil, nil)
	res = run(t, g, params, snapPress(50, 50), snapDrag(150, 100))
	if res.state.kind != gestureBoxSelect {
		t.Fatalf("expected boxSelect under locked mode, got %v", res.state.kind)
	}
	res = stepGesture(res.state, snapRelease(250, 150, true), g, params)
	wantSelection(t, res, []string{"a"}, []string{"l1"})

	// link creation stays live.
	res = run(t, g, params, snapPress(200, 140), snapDrag(250, 140))
	res = stepGesture(res.state, snapRelease(295, 138, true), g, params)
	if res.intents.LinkCreated == nil {
		t.Fatal("expected link creation under locked mode")
	}
}

func TestNodeCanvas_MiddleDragPans(t *testing.T) {
	g := testGraph(nil, nil)
	params := testGestureParams()

	press := inputSnapshot{mouse: imgui.Vec2{X: 400, Y: 300}, middlePressed: true, middleDown: true, canvasHovered: true}
	res := stepGesture(gestureState[string]{}, press, g, params)
	if res.state.kind != gesturePan {
		t.Fatalf("expected pan, got %v", res.state.kind)
	}

	move := inputSnapshot{mouse: imgui.Vec2{X: 450, Y: 320}, middleDown: true}
	res = stepGesture(res.state, move, g, params)
	if !res.viewChanged || !approxVec2(res.view.Pan, imgui.Vec2{X: 50, Y: 20}) {
		t.Fatalf("expected pan (50,20), got %+v changed=%v", res.view.Pan, res.viewChanged)
	}
	g.view = res.view

	end := inputSnapshot{mouse: imgui.Vec2{X: 450, Y: 320}, middleReleased: true}
	res = stepGesture(res.state, end, g, params)
	if res.state.kind != gestureIdle || !approxVec2(res.view.Pan, imgui.Vec2{X: 50, Y: 20}) {
		t.Fatalf("expected idle with pan kept, got %v %+v", res.state.kind, res.view.Pan)
	}
}

func TestNodeCanvas_PanScalesWithZoom(t *testing.T) {
	g := testGraph(nil, nil)
	g.view = View{Zoom: 0.5}
	params := testGestureParams()

	press := inputSnapshot{mouse: imgui.Vec2{X: 400, Y: 300}, middlePressed: true, middleDown: true, canvasHovered: true}
	res := stepGesture(gestureState[string]{}, press, g, params)
	g.view = res.view

	move := inputSnapshot{mouse: imgui.Vec2{X: 450, Y: 320}, middleDown: true}
	res = stepGesture(res.state, move, g, params)
	if !approxVec2(res.view.Pan, imgui.Vec2{X: 100, Y: 40}) {
		t.Fatalf("expected screen delta divided by zoom, got %+v", res.view.Pan)
	}
}

func TestNodeCanvas_WheelStepsDetentTowardCursor(t *testing.T) {
	g := testGraph(nil, nil)
	params := testGestureParams()
	mouse := imgui.Vec2{X: 400, Y: 300}

	before := canvasFromScreen(mouse, g.view, g.origin)
	in := inputSnapshot{mouse: mouse, wheel: -1, canvasHovered: true}
	res := stepGesture(gestureState[string]{}, in, g, params)
	if !res.viewChanged || res.view.Zoom != 0.75 {
		t.Fatalf("expected step down to 0.75, got %+v changed=%v", res.view, res.viewChanged)
	}
	if after := canvasFromScreen(mouse, res.view, g.origin); !approxVec2(before, after) {
		t.Fatalf("canvas point under cursor moved: %v -> %v", before, after)
	}

	// wheel up at the maximum detent: no change.
	in = inputSnapshot{mouse: mouse, wheel: 1, canvasHovered: true}
	res = stepGesture(gestureState[string]{}, in, g, params)
	if res.viewChanged {
		t.Fatalf("expected no change at the maximum detent, got %+v", res.view)
	}
}

func TestNodeCanvas_WheelIgnoredDuringGesture(t *testing.T) {
	g := testGraph(nil, nil)
	params := testGestureParams()
	res := run(t, g, params, snapPress(50, 50), snapDrag(150, 100))

	in := snapDrag(150, 100)
	in.wheel = -1
	res = stepGesture(res.state, in, g, params)
	if res.viewChanged || res.view.Zoom != 1.0 {
		t.Fatalf("wheel stepped the detent mid-gesture: %+v", res.view)
	}
	if res.state.kind != gestureBoxSelect {
		t.Fatalf("gesture lost: %v", res.state.kind)
	}
}

func TestNodeCanvas_MiddlePressIgnoredDuringGesture(t *testing.T) {
	g := testGraph(nil, nil)
	params := testGestureParams()
	res := run(t, g, params, snapPress(150, 140))

	in := snapDrag(160, 150)
	in.middlePressed = true
	in.middleDown = true
	res = stepGesture(res.state, in, g, params)
	if res.state.kind != gestureDragNodes {
		t.Fatalf("expected the left gesture to continue, got %v", res.state.kind)
	}
}
