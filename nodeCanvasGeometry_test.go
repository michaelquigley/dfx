package dfx

import (
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
)

const geomEpsilon = 1e-3

func approx32(a, b float32) bool {
	return abs32(a-b) <= geomEpsilon
}

func approxVec2(a, b imgui.Vec2) bool {
	return approx32(a.X, b.X) && approx32(a.Y, b.Y)
}

func TestNodeCanvas_TransformKnownValues(t *testing.T) {
	v := View{Pan: imgui.Vec2{X: 10, Y: 20}, Zoom: 0.5}
	origin := imgui.Vec2{X: 100, Y: 200}

	s := screenFromCanvas(imgui.Vec2{X: 30, Y: 40}, v, origin)
	if !approxVec2(s, imgui.Vec2{X: (30+10)*0.5 + 100, Y: (40+20)*0.5 + 200}) {
		t.Fatalf("unexpected screen point %v", s)
	}

	c := canvasFromScreen(s, v, origin)
	if !approxVec2(c, imgui.Vec2{X: 30, Y: 40}) {
		t.Fatalf("unexpected canvas point %v", c)
	}
}

func TestNodeCanvas_TransformRoundTrip(t *testing.T) {
	views := []View{
		{Pan: imgui.Vec2{}, Zoom: 1},
		{Pan: imgui.Vec2{X: -250, Y: 75}, Zoom: 0.25},
		{Pan: imgui.Vec2{X: 3.5, Y: -8.25}, Zoom: 0.75},
	}
	origins := []imgui.Vec2{{}, {X: 42, Y: 17}}
	points := []imgui.Vec2{{}, {X: 100, Y: -300}, {X: -7.5, Y: 0.125}}

	for _, v := range views {
		for _, origin := range origins {
			for _, p := range points {
				if got := canvasFromScreen(screenFromCanvas(p, v, origin), v, origin); !approxVec2(got, p) {
					t.Fatalf("round trip failed for %v under %+v origin %v: got %v", p, v, origin, got)
				}
			}
		}
	}
}

func TestNodeCanvas_NearestDetent(t *testing.T) {
	detents := []float32{0.25, 0.5, 0.75, 1.0}
	cases := []struct{ zoom, want float32 }{
		{0.1, 0.25},
		{0.25, 0.25},
		{0.6, 0.5},
		{0.7, 0.75},
		{0.625, 0.75}, // exact midpoint tie snaps to the larger detent
		{2.0, 1.0},
	}
	for _, c := range cases {
		if got := nearestDetent(detents, c.zoom); got != c.want {
			t.Fatalf("nearestDetent(%v) = %v, want %v", c.zoom, got, c.want)
		}
	}
}

func TestNodeCanvas_StepDetent(t *testing.T) {
	detents := []float32{0.25, 0.5, 0.75, 1.0}
	cases := []struct {
		current float32
		dir     int
		want    float32
	}{
		{1.0, 1, 1.0}, // clamped at max
		{1.0, -1, 0.75},
		{0.25, -1, 0.25}, // clamped at min
		{0.25, 1, 0.5},
		{0.6, 1, 0.75}, // snaps to 0.5 first, then steps
		{0.6, -1, 0.25},
	}
	for _, c := range cases {
		if got := stepDetent(detents, c.current, c.dir); got != c.want {
			t.Fatalf("stepDetent(%v, %d) = %v, want %v", c.current, c.dir, got, c.want)
		}
	}
}

func TestNodeCanvas_ZoomTowardPointKeepsCursorFixed(t *testing.T) {
	v := View{Pan: imgui.Vec2{X: 5, Y: 5}, Zoom: 1.0}
	origin := imgui.Vec2{X: 10, Y: 20}
	mouse := imgui.Vec2{X: 200, Y: 300}

	before := canvasFromScreen(mouse, v, origin)
	zoomed := zoomTowardPoint(v, origin, mouse, 0.5)
	after := canvasFromScreen(mouse, zoomed, origin)

	if zoomed.Zoom != 0.5 {
		t.Fatalf("expected zoom 0.5, got %v", zoomed.Zoom)
	}
	if !approxVec2(before, after) {
		t.Fatalf("canvas point under cursor moved: %v -> %v", before, after)
	}
}

func TestNodeCanvas_RectOps(t *testing.T) {
	r := normalizedRect(imgui.Vec2{X: 50, Y: 80}, imgui.Vec2{X: 10, Y: 20})
	if r.Min.X != 10 || r.Min.Y != 20 || r.Max.X != 50 || r.Max.Y != 80 {
		t.Fatalf("normalizedRect got %+v", r)
	}

	if !r.contains(imgui.Vec2{X: 10, Y: 20}) || !r.contains(imgui.Vec2{X: 30, Y: 50}) {
		t.Fatal("expected containment")
	}
	if r.contains(imgui.Vec2{X: 9, Y: 50}) || r.contains(imgui.Vec2{X: 30, Y: 81}) {
		t.Fatal("unexpected containment")
	}

	o := canvasRect{Min: imgui.Vec2{X: 45, Y: 75}, Max: imgui.Vec2{X: 100, Y: 100}}
	if !r.overlaps(o) || !o.overlaps(r) {
		t.Fatal("expected overlap")
	}
	far := canvasRect{Min: imgui.Vec2{X: 51, Y: 81}, Max: imgui.Vec2{X: 60, Y: 90}}
	if r.overlaps(far) {
		t.Fatal("unexpected overlap")
	}

	u := rectUnion(r, o)
	if u.Min.X != 10 || u.Min.Y != 20 || u.Max.X != 100 || u.Max.Y != 100 {
		t.Fatalf("rectUnion got %+v", u)
	}

	if !approxVec2(r.center(), imgui.Vec2{X: 30, Y: 50}) {
		t.Fatalf("center got %v", r.center())
	}
}

func TestNodeCanvas_FitsAtZoom(t *testing.T) {
	bounds := canvasRect{Max: imgui.Vec2{X: 100, Y: 100}}
	viewport := imgui.Vec2{X: 300, Y: 300}

	if !fitsAtZoom(bounds, viewport, 2.8, 10) {
		t.Fatal("expected fit at 2.8 with margin 10")
	}
	if fitsAtZoom(bounds, viewport, 2.9, 10) {
		t.Fatal("unexpected fit at 2.9 with margin 10")
	}
}

func TestNodeCanvas_CenteredPan(t *testing.T) {
	bounds := canvasRect{Min: imgui.Vec2{X: 100, Y: 100}, Max: imgui.Vec2{X: 300, Y: 200}}
	viewport := imgui.Vec2{X: 800, Y: 600}
	zoom := float32(0.5)

	pan := centeredPan(bounds, viewport, zoom)
	v := View{Pan: pan, Zoom: zoom}
	center := screenFromCanvas(bounds.center(), v, imgui.Vec2{})
	if !approxVec2(center, imgui.Vec2{X: 400, Y: 300}) {
		t.Fatalf("bounds center maps to %v, want viewport center", center)
	}
}

func TestNodeCanvas_LinkCubic(t *testing.T) {
	c := linkCubic(imgui.Vec2{X: 0, Y: 10}, imgui.Vec2{X: 100, Y: 50}, 25)
	if !approxVec2(c[1], imgui.Vec2{X: 25, Y: 10}) || !approxVec2(c[2], imgui.Vec2{X: 75, Y: 50}) {
		t.Fatalf("expected horizontal tangents, got %v %v", c[1], c[2])
	}
}

func TestNodeCanvas_PointSegmentDistance(t *testing.T) {
	a, b := imgui.Vec2{X: 0, Y: 0}, imgui.Vec2{X: 100, Y: 0}
	if d := pointSegmentDistance(imgui.Vec2{X: 50, Y: 10}, a, b); !approx32(d, 10) {
		t.Fatalf("perpendicular distance got %v", d)
	}
	if d := pointSegmentDistance(imgui.Vec2{X: 150, Y: 0}, a, b); !approx32(d, 50) {
		t.Fatalf("endpoint-clamped distance got %v", d)
	}
	if d := pointSegmentDistance(imgui.Vec2{X: 3, Y: 4}, a, a); !approx32(d, 5) {
		t.Fatalf("degenerate segment distance got %v", d)
	}
}

func TestNodeCanvas_PointBezierDistance(t *testing.T) {
	// collinear control points make an exactly straight cubic, so the
	// polyline approximation introduces no error.
	c := linkCubic(imgui.Vec2{X: 0, Y: 0}, imgui.Vec2{X: 100, Y: 0}, 25)
	if d := pointBezierDistance(c, imgui.Vec2{X: 50, Y: 10}); !approx32(d, 10) {
		t.Fatalf("mid-span distance got %v", d)
	}
	if d := pointBezierDistance(c, imgui.Vec2{X: 150, Y: 0}); !approx32(d, 50) {
		t.Fatalf("beyond-endpoint distance got %v", d)
	}
}

func TestNodeCanvas_SegmentIntersectsRect(t *testing.T) {
	r := canvasRect{Max: imgui.Vec2{X: 10, Y: 10}}
	cases := []struct {
		a, b imgui.Vec2
		want bool
	}{
		{imgui.Vec2{X: -5, Y: 5}, imgui.Vec2{X: 15, Y: 5}, true}, // crossing
		{imgui.Vec2{X: 2, Y: 2}, imgui.Vec2{X: 8, Y: 8}, true},   // fully inside
		{imgui.Vec2{X: 20, Y: 20}, imgui.Vec2{X: 30, Y: 30}, false},
		{imgui.Vec2{X: -5, Y: -5}, imgui.Vec2{X: 5, Y: -1}, false},
		{imgui.Vec2{X: 10, Y: 2}, imgui.Vec2{X: 10, Y: 8}, true}, // touching an edge
		{imgui.Vec2{X: 5, Y: 5}, imgui.Vec2{X: 5, Y: 5}, true},   // degenerate point inside
	}
	for i, c := range cases {
		if got := segmentIntersectsRect(c.a, c.b, r); got != c.want {
			t.Fatalf("case %d: segmentIntersectsRect(%v, %v) = %v, want %v", i, c.a, c.b, got, c.want)
		}
	}
}

func TestNodeCanvas_BezierIntersectsRect(t *testing.T) {
	c := linkCubic(imgui.Vec2{X: 0, Y: 0}, imgui.Vec2{X: 100, Y: 0}, 25)
	if !bezierIntersectsRect(c, canvasRect{Min: imgui.Vec2{X: 40, Y: -5}, Max: imgui.Vec2{X: 60, Y: 5}}) {
		t.Fatal("expected mid-span intersection")
	}
	if bezierIntersectsRect(c, canvasRect{Min: imgui.Vec2{X: 40, Y: 10}, Max: imgui.Vec2{X: 60, Y: 20}}) {
		t.Fatal("unexpected intersection with offset rect")
	}
	if !bezierIntersectsRect(c, canvasRect{Min: imgui.Vec2{X: -5, Y: -5}, Max: imgui.Vec2{X: 5, Y: 5}}) {
		t.Fatal("expected endpoint-inside intersection")
	}
}

// testPin, testNode, and testGraph build the synthetic fixture shared with
// the input tests: node a (100,100)-(200,180) with output pin a.out at
// (200,140); node b (300,100)-(400,180) with input pin b.in at (300,140) and
// output pin b.out at (400,140); link l1 running a.out -> b.in.
func testPin(id string, x, y float32, side pinSide, decl int) pinGeometry[string] {
	return pinGeometry[string]{id: id, pos: imgui.Vec2{X: x, Y: y}, side: side, declIndex: decl}
}

func testNode(id string, minX, minY, maxX, maxY float32, decl int, pins ...pinGeometry[string]) nodeGeometry[string] {
	return nodeGeometry[string]{
		id:        id,
		rect:      canvasRect{Min: imgui.Vec2{X: minX, Y: minY}, Max: imgui.Vec2{X: maxX, Y: maxY}},
		declIndex: decl,
		pins:      pins,
	}
}

func testGraph(selNodes, selLinks []string) *canvasGeometry[string] {
	g := &canvasGeometry[string]{
		view:     View{Zoom: 1},
		viewport: imgui.Vec2{X: 800, Y: 600},
		nodes: []nodeGeometry[string]{
			testNode("a", 100, 100, 200, 180, 0, testPin("a.out", 200, 140, pinOutput, 0)),
			testNode("b", 300, 100, 400, 180, 1,
				testPin("b.in", 300, 140, pinInput, 1),
				testPin("b.out", 400, 140, pinOutput, 2)),
		},
		links: []linkGeometry[string]{
			{id: "l1", cubic: linkCubic(imgui.Vec2{X: 200, Y: 140}, imgui.Vec2{X: 300, Y: 140}, 50), declIndex: 0},
		},
	}
	for i := range g.nodes {
		g.nodes[i].selected = idInSlice(selNodes, g.nodes[i].id)
	}
	for i := range g.links {
		g.links[i].selected = idInSlice(selLinks, g.links[i].id)
	}
	return g
}

func testHitParams() hitParams {
	return hitParams{pinHitRadius: 10, linkHitDistance: 6, linkSnapRadius: 24}
}

func TestNodeCanvas_HitTestNodeBody(t *testing.T) {
	g := testGraph(nil, nil)
	hit := hitTest(g, imgui.Vec2{X: 150, Y: 140}, testHitParams())
	if hit.kind != hitNode || hit.node != "a" {
		t.Fatalf("expected node a, got %+v", hit)
	}
}

func TestNodeCanvas_HitTestPinOutranksOwnBody(t *testing.T) {
	g := testGraph(nil, nil)
	hit := hitTest(g, imgui.Vec2{X: 200, Y: 140}, testHitParams())
	if hit.kind != hitPin || hit.pin != "a.out" || hit.node != "a" || hit.side != pinOutput {
		t.Fatalf("expected pin a.out, got %+v", hit)
	}
}

func TestNodeCanvas_HitTestTopmostNodeWins(t *testing.T) {
	g := testGraph(nil, nil)
	// node c overlaps node a and covers a's pin; declared later, so it is
	// topmost and its body wins over both.
	g.nodes = append(g.nodes, testNode("c", 150, 120, 260, 200, 2))

	hit := hitTest(g, imgui.Vec2{X: 170, Y: 150}, testHitParams())
	if hit.kind != hitNode || hit.node != "c" {
		t.Fatalf("expected topmost node c, got %+v", hit)
	}

	// a's pin at (200,140) is covered by c: pins never reach through a node
	// covering them.
	hit = hitTest(g, imgui.Vec2{X: 200, Y: 140}, testHitParams())
	if hit.kind != hitNode || hit.node != "c" {
		t.Fatalf("expected covering node c over a's pin, got %+v", hit)
	}
}

func TestNodeCanvas_HitTestPinRadiusIsScreenSpace(t *testing.T) {
	g := testGraph(nil, nil)
	g.view = View{Pan: imgui.Vec2{X: 5, Y: 5}, Zoom: 0.25}
	g.origin = imgui.Vec2{X: 10, Y: 20}

	pinScreen := screenFromCanvas(imgui.Vec2{X: 200, Y: 140}, g.view, g.origin)
	if hit := hitTest(g, imgui.Vec2{X: pinScreen.X + 9, Y: pinScreen.Y}, testHitParams()); hit.kind != hitPin || hit.pin != "a.out" {
		t.Fatalf("expected pin hit 9 screen px away at zoom 0.25, got %+v", hit)
	}
	if hit := hitTest(g, imgui.Vec2{X: pinScreen.X + 11, Y: pinScreen.Y - 11}, testHitParams()); hit.kind == hitPin {
		t.Fatalf("unexpected pin hit outside the screen-space radius: %+v", hit)
	}
}

func TestNodeCanvas_HitTestLink(t *testing.T) {
	g := testGraph(nil, nil)
	hit := hitTest(g, imgui.Vec2{X: 250, Y: 143}, testHitParams())
	if hit.kind != hitLink || hit.link != "l1" {
		t.Fatalf("expected link l1, got %+v", hit)
	}
	if hit := hitTest(g, imgui.Vec2{X: 250, Y: 150}, testHitParams()); hit.kind != hitNone {
		t.Fatalf("expected miss beyond link tolerance, got %+v", hit)
	}
}

func TestNodeCanvas_HitTestNearestLinkWins(t *testing.T) {
	g := &canvasGeometry[string]{
		view:     View{Zoom: 1},
		viewport: imgui.Vec2{X: 800, Y: 600},
		links: []linkGeometry[string]{
			{id: "l1", cubic: linkCubic(imgui.Vec2{X: 0, Y: 0}, imgui.Vec2{X: 100, Y: 0}, 25), declIndex: 0},
			{id: "l2", cubic: linkCubic(imgui.Vec2{X: 0, Y: 8}, imgui.Vec2{X: 100, Y: 8}, 25), declIndex: 1},
		},
	}
	if hit := hitTest(g, imgui.Vec2{X: 50, Y: 3}, testHitParams()); hit.link != "l1" {
		t.Fatalf("expected nearest link l1, got %+v", hit)
	}
	if hit := hitTest(g, imgui.Vec2{X: 50, Y: 5}, testHitParams()); hit.link != "l2" {
		t.Fatalf("expected nearest link l2, got %+v", hit)
	}
	// an exact tie breaks to the later-declared (topmost) link.
	if hit := hitTest(g, imgui.Vec2{X: 50, Y: 4}, testHitParams()); hit.link != "l2" {
		t.Fatalf("expected tie to go to topmost link l2, got %+v", hit)
	}
}

func TestNodeCanvas_HitTestMiss(t *testing.T) {
	g := testGraph(nil, nil)
	if hit := hitTest(g, imgui.Vec2{X: 500, Y: 500}, testHitParams()); hit.kind != hitNone {
		t.Fatalf("expected miss, got %+v", hit)
	}
}

func TestNodeCanvas_SnapPin(t *testing.T) {
	g := testGraph(nil, nil)

	// dragging from an output: only input pins are compatible.
	pin, pos, ok := snapPin(g, imgui.Vec2{X: 295, Y: 138}, pinOutput, testHitParams())
	if !ok || pin != "b.in" || !approxVec2(pos, imgui.Vec2{X: 300, Y: 140}) {
		t.Fatalf("expected snap to b.in, got %v %v %v", pin, pos, ok)
	}

	// near b.out (same side as the source): no compatible pin in range.
	if _, _, ok := snapPin(g, imgui.Vec2{X: 400, Y: 140}, pinOutput, testHitParams()); ok {
		t.Fatal("unexpected snap to a same-side pin")
	}

	// dragging from an input: outputs are compatible; nearest wins.
	pin, _, ok = snapPin(g, imgui.Vec2{X: 203, Y: 142}, pinInput, testHitParams())
	if !ok || pin != "a.out" {
		t.Fatalf("expected snap to a.out, got %v %v", pin, ok)
	}

	// out of radius: no snap.
	if _, _, ok := snapPin(g, imgui.Vec2{X: 500, Y: 400}, pinOutput, testHitParams()); ok {
		t.Fatal("unexpected snap outside the radius")
	}
}

func TestNodeCanvas_SnapPinRadiusIsScreenSpace(t *testing.T) {
	g := testGraph(nil, nil)
	g.view = View{Zoom: 0.25}

	// at zoom 0.25 the canvas gap from b.in (300,140) to a point 80 canvas
	// units away is only 20 screen px — inside the 24 px snap radius.
	target := screenFromCanvas(imgui.Vec2{X: 300, Y: 140}, g.view, g.origin)
	pin, _, ok := snapPin(g, imgui.Vec2{X: target.X + 20, Y: target.Y}, pinOutput, testHitParams())
	if !ok || pin != "b.in" {
		t.Fatalf("expected screen-space snap to b.in, got %v %v", pin, ok)
	}
}
