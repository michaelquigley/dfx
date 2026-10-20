package dfx

import (
	"fmt"
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
)

// probe records what one draw was allocated and where it actually drew: its
// state size, the content region at its origin, and the outer rectangles of its
// window and every ancestor window, innermost first.
type probe struct {
	drawn   int
	size    imgui.Vec2
	avail   imgui.Vec2
	origin  imgui.Vec2
	windows []imgui.Rect
}

func (p *probe) Draw(state *State) {
	p.drawn++
	p.size, p.avail, p.origin = state.Size, imgui.ContentRegionAvail(), imgui.CursorScreenPos()
	p.windows = windowRects()
}

func (p *probe) Actions() *ActionRegistry { return nil }

func windowRects() []imgui.Rect {
	var rects []imgui.Rect
	for w := imgui.InternalCurrentWindow(); ; w = w.ParentWindow() {
		ptr, finish := w.Handle()
		finish()
		if ptr == nil {
			return rects
		}
		rects = append(rects, imgui.Rect{Min: w.Pos(), Max: w.Pos().Add(w.Size())})
	}
}

func within(inner, outer imgui.Rect) bool {
	const slack = 0.5
	return inner.Min.X >= outer.Min.X-slack && inner.Min.Y >= outer.Min.Y-slack &&
		inner.Max.X <= outer.Max.X+slack && inner.Max.Y <= outer.Max.Y+slack
}

func near(a, b imgui.Vec2) bool {
	return abs32(a.X-b.X) <= 0.5 && abs32(a.Y-b.Y) <= 0.5
}

// assertFits checks that a probe drew with a nonnegative allocation equal to the
// content region at its origin, and that each window lies inside its parent.
func assertFits(t *testing.T, name string, p *probe) {
	t.Helper()
	if p.drawn == 0 {
		t.Fatalf("%s: never drew", name)
	}
	if p.size.X < 0 || p.size.Y < 0 {
		t.Fatalf("%s: negative allocation '%v'", name, p.size)
	}
	usable := imgui.Vec2{X: max(0, p.avail.X), Y: max(0, p.avail.Y)}
	if !near(p.size, usable) {
		t.Fatalf("%s: allocation '%v' differs from its usable region '%v'", name, p.size, usable)
	}
	for i := 0; i+1 < len(p.windows); i++ {
		if !within(p.windows[i], p.windows[i+1]) {
			t.Fatalf("%s: window '%v' escapes its parent '%v'", name, p.windows[i], p.windows[i+1])
		}
	}
}

var presets = map[string]func() Style{"classic": ClassicStyle, "faceplate": FaceplateStyle}

func runRoot(t *testing.T, style Style, root Component, menu Component, size imgui.Vec2, frames int) {
	t.Helper()
	style.Name += " (test)"
	app := New(root, Config{Style: &style, MenuBar: menu})
	if err := runHeadless(t, app, &headlessBackend{size: size, frames: frames}); err != nil {
		t.Fatal(err)
	}
}

func testMenu() Component {
	return NewFunc(func(*State) {
		if imgui.BeginMenu("File") {
			imgui.EndMenu()
		}
	})
}

func TestAllocation_RootUsesOuterMarginOnce(t *testing.T) {
	for name, preset := range presets {
		for _, menu := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/menu=%v", name, menu), func(t *testing.T) {
				style := preset()
				p := &probe{}
				var bar Component
				if menu {
					bar = testMenu()
				}
				runRoot(t, style, p, bar, imgui.Vec2{X: 800, Y: 600}, 2)
				assertFits(t, "root", p)
				root := p.windows[0]
				if menu && root.Min.Y <= 0 {
					t.Fatal("the root must start below the menu bar")
				}
				if !menu && root.Min != (imgui.Vec2{}) {
					t.Fatalf("the root must fill the viewport, starts at '%v'", root.Min)
				}
				m := style.Spacing.OuterMargin
				if !near(p.origin, root.Min.Add(imgui.Vec2{X: m, Y: m})) {
					t.Fatalf("content origin '%v' is not the root inset by '%v'", p.origin, m)
				}
				expected := root.Max.Sub(root.Min).Sub(imgui.Vec2{X: 2 * m, Y: 2 * m})
				if !near(p.size, expected) {
					t.Fatalf("root allocation '%v', expected '%v'", p.size, expected)
				}
			})
		}
	}
}

func TestAllocation_TinyRootIsNeverNegative(t *testing.T) {
	for name, preset := range presets {
		t.Run(name, func(t *testing.T) {
			p := &probe{}
			runRoot(t, preset(), p, testMenu(), imgui.Vec2{X: 20, Y: 30}, 2)
			if p.size.X < 0 || p.size.Y < 0 {
				t.Fatalf("negative root allocation '%v'", p.size)
			}
		})
	}
}

func TestAllocation_WorkspaceSelector(t *testing.T) {
	for name, preset := range presets {
		for _, selector := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/selector=%v", name, selector), func(t *testing.T) {
				p := &probe{}
				ws := NewWorkspace()
				ws.ShowSelector = selector
				ws.Add("one", "One", p)
				ws.Add("two", "Two", &probe{})
				outer := &probe{}
				var frameHeight float32
				root := NewFunc(func(state *State) {
					frameHeight = imgui.FrameHeight()
					outer.Draw(state)
					ws.Draw(state)
				})
				runRoot(t, preset(), root, nil, imgui.Vec2{X: 640, Y: 480}, 2)
				assertFits(t, "workspace item", p)
				advance := p.origin.Y - outer.origin.Y
				if !selector && advance != 0 {
					t.Fatalf("without a selector the item must start at the workspace origin, moved '%v'", advance)
				}
				if selector && advance <= frameHeight {
					t.Fatalf("the item must start below the selector, moved '%v'", advance)
				}
			})
		}
	}
}

func TestAllocation_DashManagerAtNestedOrigin(t *testing.T) {
	for name, preset := range presets {
		t.Run(name, func(t *testing.T) {
			left, top, right, bottom, inner := &probe{}, &probe{}, &probe{}, &probe{}, &probe{}
			dm := NewDashManager()
			dm.Margin = 6 // explicit public margin survives
			dm.Left, dm.Top, dm.Right, dm.Bottom = NewDash("left", left), NewDash("top", top), NewDash("right", right), NewDash("bottom", bottom)
			dm.Left.TargetSize, dm.Left.CurrentSize = 150, 150
			dm.Right.TargetSize, dm.Right.CurrentSize = 140, 140
			dm.Top.TargetSize, dm.Top.CurrentSize = 90, 90
			dm.Bottom.TargetSize, dm.Bottom.CurrentSize = 80, 80
			dm.Right.Resizable = false
			dm.Inner = inner
			var origin imgui.Vec2
			var rootRect imgui.Rect
			root := NewFunc(func(state *State) {
				imgui.Text("above the manager") // a nonzero content origin
				origin = imgui.CursorScreenPos()
				rootRect = windowRects()[0]
				dm.Draw(&State{Size: availableSize(), IO: state.IO, App: state.App, Parent: nil})
			})
			runRoot(t, preset(), root, nil, imgui.Vec2{X: 900, Y: 700}, 2)
			for name, p := range map[string]*probe{"left": left, "top": top, "right": right, "bottom": bottom, "inner": inner} {
				assertFits(t, name, p)
				if !within(p.windows[len(p.windows)-1], rootRect) {
					t.Fatalf("%s escapes the root", name)
				}
			}
			// HorizontalPrecedence: top spans the width at the origin
			if panel := top.windows[1]; !near(panel.Min, origin) {
				t.Fatalf("top dash at '%v', expected the content origin '%v'", panel.Min, origin)
			}
			expected := origin.Add(imgui.Vec2{X: 150 + dm.Margin, Y: 90 + dm.Margin})
			if got := inner.windows[0].Min; !near(got, expected) {
				t.Fatalf("inner panel at '%v', expected '%v'", got, expected)
			}
		})
	}
}

func TestAllocation_DashManagerExhaustedInner(t *testing.T) {
	inner := &probe{}
	dm := NewDashManager()
	dm.Left = NewDash("left", &probe{})
	dm.Inner = inner
	root := NewFunc(func(state *State) {
		dm.Draw(&State{Size: imgui.Vec2{X: 300, Y: 200}, IO: state.IO, App: state.App})
	})
	runRoot(t, FaceplateStyle(), root, nil, imgui.Vec2{X: 800, Y: 600}, 1)
	if inner.drawn != 0 {
		t.Fatalf("an exhausted inner panel must not draw, got size '%v'", inner.size)
	}
}

func TestAllocation_HCollapse(t *testing.T) {
	for name, preset := range presets {
		for _, anchor := range []HCollapseAnchor{AnchorLeft, AnchorRight} {
			t.Run(fmt.Sprintf("%s/anchor=%v", name, anchor), func(t *testing.T) {
				content := &probe{}
				expanded := NewHCollapse(content, HCollapseConfig{Title: "Expanded", ExpandedWidth: 240, Expanded: true, Resizable: true, Anchor: anchor})
				collapsed := NewHCollapse(&probe{}, HCollapseConfig{Title: "Collapsed", ExpandedWidth: 240, Anchor: anchor})
				fixed := NewHCollapse(&probe{}, HCollapseConfig{Title: "Fixed", ExpandedWidth: 120, Expanded: true, Height: 140, Anchor: anchor})
				var strip, fixedPanel imgui.Rect
				var stripReach, frameHeight float32
				root := NewFunc(func(state *State) {
					frameHeight = imgui.FrameHeight()
					expanded.Draw(state)
					imgui.SameLine()
					collapsed.Draw(state)
					strip, stripReach = childExtent(collapsed.imguiID())
					imgui.SameLine()
					fixed.Draw(state)
					fixedPanel, _ = childExtent(fixed.imguiID())
				})
				runRoot(t, preset(), root, nil, imgui.Vec2{X: 800, Y: 600}, 2)
				assertFits(t, "content", content)
				panel := content.windows[1]
				if !near(panel.Max.Sub(panel.Min), imgui.Vec2{X: 240, Y: panel.Max.Y - panel.Min.Y}) {
					t.Fatalf("panel width '%v', expected 240", panel.Max.X-panel.Min.X)
				}
				if content.origin.Y-panel.Min.Y < frameHeight {
					t.Fatal("content must start below the header")
				}
				if width := strip.Max.X - strip.Min.X; width != collapsed.MinWidth {
					t.Fatalf("collapsed strip width '%v', expected '%v'", width, collapsed.MinWidth)
				}
				if stripReach > strip.Max.X+0.5 {
					t.Fatalf("collapsed toggle reaches '%v', past its strip's edge '%v'", stripReach, strip.Max.X)
				}
				if height := fixedPanel.Max.Y - fixedPanel.Min.Y; height != 140 {
					t.Fatalf("explicit height '%v', expected 140", height)
				}
			})
		}
	}
}

// childExtent finds a child window drawn in the current window under label, and
// returns its rectangle and the furthest x its items reached this frame.
func childExtent(label string) (imgui.Rect, float32) {
	parent := imgui.InternalCurrentWindow()
	name := fmt.Sprintf("%s/%s_%08X", parent.Name(), label, uint32(imgui.IDStr(label)))
	w := imgui.InternalFindWindowByName(name)
	dc := w.DC()
	return imgui.Rect{Min: w.Pos(), Max: w.Pos().Add(w.Size())}, dc.CursorMaxPos().X
}

func TestAllocation_HCollapseClassicHeader(t *testing.T) {
	styleTestContext(t)
	classic := ClassicStyle()
	New(nil, Config{Style: &classic}).setupFontsAndTheme()
	imgui.NewFrame()
	defer imgui.EndFrame()
	h := NewHCollapse(nil, HCollapseConfig{Expanded: true, ExpandedWidth: 200})
	if got := h.headerHeight(); got != HCollapseHeaderHeight {
		t.Fatalf("Classic's measured header '%v' must match HCollapseHeaderHeight '%v'", got, HCollapseHeaderHeight)
	}
	if got := dashHandleRow(); got != DashTitleBarHeight {
		t.Fatalf("Classic's measured dash handle row '%v' must match DashTitleBarHeight '%v'", got, DashTitleBarHeight)
	}
}

func TestAllocation_Layouts(t *testing.T) {
	for name, preset := range presets {
		t.Run(name, func(t *testing.T) {
			probes := map[string]*probe{}
			flex, grid := NewMultiGrid(), NewMultiGrid()
			for _, id := range []string{"a", "b", "c", "d"} {
				probes["flex "+id], probes["grid "+id] = &probe{}, &probe{}
				flex.AddComponent(id, probes["flex "+id])
				grid.AddComponent(id, probes["grid "+id])
			}
			flex.SetLayout(NewFlexLayout([][]string{{"a", "b"}, {"c", "d"}}))
			layout := NewGridLayout(3, 2)
			layout.SetCell("a", 0, 0, 1, 2)
			layout.SetCell("b", 0, 2, 2, 1)
			layout.SetCell("c", 1, 0, 1, 1)
			layout.SetCell("d", 1, 1, 1, 1)
			grid.SetLayout(layout)
			var flexArea, gridArea imgui.Rect
			root := NewFunc(func(state *State) {
				imgui.Text("above the layouts")
				half := imgui.Vec2{X: availableSize().X, Y: availableSize().Y/2 - imgui.CurrentStyle().ItemSpacing().Y}
				for _, pair := range []struct {
					mg   *MultiGrid
					area *imgui.Rect
				}{{flex, &flexArea}, {grid, &gridArea}} {
					imgui.BeginGroup()
					*pair.area = imgui.Rect{Min: imgui.CursorScreenPos(), Max: imgui.CursorScreenPos().Add(half)}
					pair.mg.Draw(&State{Size: half, IO: state.IO, App: state.App})
					imgui.EndGroup()
					imgui.SetCursorScreenPos(imgui.Vec2{X: pair.area.Min.X, Y: pair.area.Max.Y + imgui.CurrentStyle().ItemSpacing().Y})
				}
				imgui.Dummy(imgui.Vec2{})
			})
			runRoot(t, preset(), root, nil, imgui.Vec2{X: 900, Y: 700}, 2)
			for id, p := range probes {
				assertFits(t, id, p)
				area := flexArea
				if id[0] == 'g' {
					area = gridArea
				}
				if !within(p.windows[0], area) {
					t.Fatalf("%s cell '%v' escapes its layout '%v'", id, p.windows[0], area)
				}
			}
		})
	}
}

func TestAllocation_ExhaustedLayoutCells(t *testing.T) {
	for name, preset := range presets {
		t.Run(name, func(t *testing.T) {
			probes := []*probe{{}, {}, {}, {}}
			flex, grid := NewMultiGrid(), NewMultiGrid()
			flex.AddComponent("a", probes[0])
			flex.AddComponent("b", probes[1])
			grid.AddComponent("a", probes[2])
			grid.AddComponent("b", probes[3])
			flex.SetLayout(NewFlexLayout([][]string{{"a", "b"}}))
			layout := NewGridLayout(2, 1)
			layout.SetCell("a", 0, 0, 1, 1)
			layout.SetCell("b", 0, 1, 1, 1)
			grid.SetLayout(layout)
			root := NewFunc(func(state *State) {
				tiny := imgui.Vec2{X: 6, Y: 6}
				flex.Draw(&State{Size: tiny, IO: state.IO, App: state.App})
				grid.Draw(&State{Size: tiny, IO: state.IO, App: state.App})
			})
			runRoot(t, preset(), root, nil, imgui.Vec2{X: 400, Y: 300}, 1)
			for i, p := range probes {
				if p.drawn > 0 && (p.size.X < 0 || p.size.Y < 0) {
					t.Fatalf("cell %d drew with negative size '%v'", i, p.size)
				}
			}
			if probes[1].drawn != 0 {
				t.Fatal("a flex cell exhausted by its splitter must not draw")
			}
		})
	}
}

func TestAllocation_ToolbarRegistersItsExtent(t *testing.T) {
	for name, preset := range presets {
		t.Run(name, func(t *testing.T) {
			var start, after imgui.Vec2
			var lineHeight, spacing float32
			root := NewFunc(func(*State) {
				start = imgui.CursorPos()
				lineHeight = max(imgui.CalcTextSize("Tools").Y, imgui.FrameHeight()) + imgui.CurrentStyle().FramePadding().Y*2.5
				spacing = imgui.CurrentStyle().ItemSpacing().Y
				ToolbarEx("Tools", func() { imgui.Button("go") })
				after = imgui.CursorPos()
			})
			runRoot(t, preset(), root, nil, imgui.Vec2{X: 600, Y: 400}, 1)
			if !near(after, imgui.Vec2{X: start.X, Y: start.Y + lineHeight + spacing}) {
				t.Fatalf("after the toolbar the cursor is at '%v', expected '%v'", after, imgui.Vec2{X: start.X, Y: start.Y + lineHeight + spacing})
			}
		})
	}
}

// a positive cell whose padding leaves no usable area stays empty rather than
// drawing its component at zero size. Classic's unpadded cell is the control.
func TestAllocation_PaddingExhaustsContent(t *testing.T) {
	for name, preset := range presets {
		t.Run(name, func(t *testing.T) {
			grid, flex, collapse := &probe{}, &probe{}, &probe{}
			gmg, fmg := NewMultiGrid(), NewMultiGrid()
			gmg.AddComponent("one", grid)
			layout := NewGridLayout(1, 1)
			layout.SetCell("one", 0, 0, 1, 1)
			gmg.SetLayout(layout)
			fmg.AddComponent("one", flex)
			fmg.SetLayout(NewFlexLayout([][]string{{"one"}}))
			h := NewHCollapse(collapse, HCollapseConfig{Title: "short", ExpandedWidth: 120, Expanded: true})
			var headerHeight float32
			root := NewFunc(func(state *State) {
				cell := imgui.Vec2{X: 20, Y: 20}
				gmg.Draw(&State{Size: cell, IO: state.IO, App: state.App})
				fmg.Draw(&State{Size: cell, IO: state.IO, App: state.App})
				headerHeight = h.headerHeight()
				h.Height = headerHeight + 10 // ten pixels of content, less than Faceplate's padding
				h.Draw(state)
			})
			runRoot(t, preset(), root, nil, imgui.Vec2{X: 600, Y: 400}, 2)
			if name == "faceplate" {
				for label, p := range map[string]*probe{"grid": grid, "flex": flex, "hcollapse": collapse} {
					if p.drawn != 0 {
						t.Fatalf("%s content drew with exhausted allocation '%v'", label, p.size)
					}
				}
				return
			}
			assertFits(t, "grid", grid)
			if grid.size != (imgui.Vec2{X: 20, Y: 20}) {
				t.Fatalf("Classic's unpadded cell must keep its whole allocation, got '%v'", grid.size)
			}
			assertFits(t, "hcollapse", collapse)
		})
	}
}

// a side dash left no height by its neighbors is skipped before any window state
// is queued: it neither draws nor keeps focus, the next window is not resized by
// leftover geometry, and its show/hide animation still advances.
func TestAllocation_DashExhaustedSide(t *testing.T) {
	for _, visible := range []bool{true, false} {
		t.Run(fmt.Sprintf("visible=%v", visible), func(t *testing.T) {
			p := &probe{}
			dm := NewDashManager()
			dm.Top, dm.Bottom, dm.Left = NewDash("top", &probe{}), NewDash("bottom", &probe{}), NewDash("left", p)
			dm.Top.TargetSize, dm.Top.CurrentSize = 90, 90
			dm.Bottom.TargetSize, dm.Bottom.CurrentSize = 90, 90
			dm.Left.TargetSize, dm.Left.CurrentSize = 120, 60 // growing while shown
			dm.Left.Visible = visible
			dm.Left.Focused = true // stale from an earlier frame
			var after imgui.Vec2
			panelCreated := false
			root := NewFunc(func(state *State) {
				dm.Draw(&State{Size: imgui.Vec2{X: 300, Y: 100}, IO: state.IO, App: state.App})
				name := fmt.Sprintf("%s/%s_%08X", imgui.InternalCurrentWindow().Name(), "left", uint32(imgui.IDStr("left")))
				ptr, finish := imgui.InternalFindWindowByName(name).Handle()
				panelCreated = panelCreated || ptr != nil
				finish()
				imgui.BeginChildStrV("after", imgui.Vec2{X: 50, Y: 40}, imgui.ChildFlagsNone, imgui.WindowFlagsNone)
				after = imgui.WindowSize()
				imgui.EndChild()
			})
			runRoot(t, FaceplateStyle(), root, nil, imgui.Vec2{X: 600, Y: 400}, 3)
			if p.drawn != 0 {
				t.Fatalf("exhausted side dash drew its content at '%v'", p.size)
			}
			if panelCreated {
				t.Fatal("an exhausted side dash must not create its panel window")
			}
			if dm.Left.Focused || dm.Focused == dm.Left {
				t.Fatal("an undrawn dash must not hold focus")
			}
			if after != (imgui.Vec2{X: 50, Y: 40}) {
				t.Fatalf("the next window took leftover dash geometry: '%v'", after)
			}
			if visible && dm.Left.CurrentSize <= 60 {
				t.Fatalf("a shown dash must keep growing while exhausted, at '%d'", dm.Left.CurrentSize)
			}
			if !visible && dm.Left.CurrentSize >= 60 {
				t.Fatalf("a hidden dash must keep shrinking while exhausted, at '%d'", dm.Left.CurrentSize)
			}
		})
	}
}

// TestAllocation_DashManagerTopMargin checks that TopMargin reserves space above
// the whole manager layout once, under either precedence: every panel and the
// inner surface lie within the remainder below the inset.
func TestAllocation_DashManagerTopMargin(t *testing.T) {
	precedences := map[string]DashPrecedence{"vertical": VerticalPrecedence, "horizontal": HorizontalPrecedence}
	for styleName, preset := range presets {
		for precedenceName, precedence := range precedences {
			for _, topMargin := range []float32{0, 40} {
				t.Run(fmt.Sprintf("%s/%s/top=%v", styleName, precedenceName, topMargin), func(t *testing.T) {
					left, top, right, bottom, inner := &probe{}, &probe{}, &probe{}, &probe{}, &probe{}
					dm := NewDashManager()
					dm.Precedence, dm.TopMargin, dm.Margin = precedence, topMargin, 6
					dm.Left, dm.Top, dm.Right, dm.Bottom = NewDash("left", left), NewDash("top", top), NewDash("right", right), NewDash("bottom", bottom)
					dm.Left.TargetSize, dm.Left.CurrentSize = 150, 150
					dm.Right.TargetSize, dm.Right.CurrentSize = 140, 140
					dm.Top.TargetSize, dm.Top.CurrentSize = 90, 90
					dm.Bottom.TargetSize, dm.Bottom.CurrentSize = 80, 80
					dm.Inner = inner
					var manager imgui.Rect
					root := NewFunc(func(state *State) {
						imgui.Text("above the manager") // a nonzero content origin
						size := availableSize()
						manager = imgui.Rect{Min: imgui.CursorScreenPos(), Max: imgui.CursorScreenPos().Add(size)}
						dm.Draw(&State{Size: size, IO: state.IO, App: state.App})
					})
					runRoot(t, preset(), root, nil, imgui.Vec2{X: 900, Y: 700}, 2)

					remainder := imgui.Rect{Min: manager.Min.Add(imgui.Vec2{Y: topMargin}), Max: manager.Max}
					panels := map[string]imgui.Rect{"inner": inner.windows[0]}
					for name, p := range map[string]*probe{"left": left, "top": top, "right": right, "bottom": bottom} {
						panels[name] = p.windows[1] // the panel around the dash surface
					}
					for name, p := range map[string]*probe{"left": left, "top": top, "right": right, "bottom": bottom, "inner": inner} {
						assertFits(t, name, p)
						if !within(panels[name], remainder) {
							t.Fatalf("%s panel '%v' escapes the area below the top margin '%v'", name, panels[name], remainder)
						}
					}

					// the inset is applied once: the top dash starts at it, the bottom
					// dash still ends at the manager's bottom edge, and the inner panel
					// loses it from its height exactly once.
					if got := panels["top"].Min.Y; abs32(got-remainder.Min.Y) > 0.5 {
						t.Fatalf("top dash at y '%v', expected '%v'", got, remainder.Min.Y)
					}
					if got := panels["bottom"].Max.Y; abs32(got-manager.Max.Y) > 0.5 {
						t.Fatalf("bottom dash ends at y '%v', expected '%v'", got, manager.Max.Y)
					}
					sideY := remainder.Min.Y
					if precedence == HorizontalPrecedence {
						sideY += 90 + dm.Margin
					}
					for _, side := range []string{"left", "right"} {
						if got := panels[side].Min.Y; abs32(got-sideY) > 0.5 {
							t.Fatalf("%s dash at y '%v', expected '%v'", side, got, sideY)
						}
					}
					expected := imgui.Rect{
						Min: remainder.Min.Add(imgui.Vec2{X: 150 + dm.Margin, Y: 90 + dm.Margin}),
						Max: remainder.Max.Sub(imgui.Vec2{X: 140 + dm.Margin, Y: 80 + dm.Margin}),
					}
					if got := panels["inner"]; !near(got.Min, expected.Min) || !near(got.Max, expected.Max) {
						t.Fatalf("inner panel '%v', expected '%v'", got, expected)
					}
				})
			}
		}
	}
}

// TestAllocation_DashManagerTopMarginExhausts checks that a TopMargin consuming
// the manager's height skips every panel and the inner surface, clears focus,
// leaks no window geometry, and keeps the dash animations running.
func TestAllocation_DashManagerTopMarginExhausts(t *testing.T) {
	precedences := map[string]DashPrecedence{"vertical": VerticalPrecedence, "horizontal": HorizontalPrecedence}
	for styleName, preset := range presets {
		for precedenceName, precedence := range precedences {
			for _, topMargin := range []float32{200, 260} { // exactly the height, and beyond it
				t.Run(fmt.Sprintf("%s/%s/top=%v", styleName, precedenceName, topMargin), func(t *testing.T) {
					probes := map[string]*probe{"left": {}, "top": {}, "right": {}, "bottom": {}}
					inner := &probe{}
					dm := NewDashManager()
					dm.Precedence, dm.TopMargin = precedence, topMargin
					dm.Left, dm.Top, dm.Right, dm.Bottom = NewDash("left", probes["left"]), NewDash("top", probes["top"]), NewDash("right", probes["right"]), NewDash("bottom", probes["bottom"])
					for _, dash := range []*Dash{dm.Left, dm.Top, dm.Right, dm.Bottom} {
						dash.TargetSize, dash.CurrentSize = 60, 30 // growing while shown
						dash.Focused = true                        // stale from an earlier frame
					}
					dm.Right.Visible = false // and one shrinking while hidden
					dm.Inner = inner
					created := map[string]bool{}
					var after imgui.Vec2
					root := NewFunc(func(state *State) {
						dm.Draw(&State{Size: imgui.Vec2{X: 600, Y: 200}, IO: state.IO, App: state.App})
						for name := range probes {
							window := fmt.Sprintf("%s/%s_%08X", imgui.InternalCurrentWindow().Name(), name, uint32(imgui.IDStr(name)))
							ptr, finish := imgui.InternalFindWindowByName(window).Handle()
							created[name] = created[name] || ptr != nil
							finish()
						}
						imgui.BeginChildStrV("after", imgui.Vec2{X: 50, Y: 40}, imgui.ChildFlagsNone, imgui.WindowFlagsNone)
						after = imgui.WindowSize()
						imgui.EndChild()
					})
					runRoot(t, preset(), root, nil, imgui.Vec2{X: 900, Y: 700}, 3)
					for name, p := range probes {
						if p.drawn != 0 {
							t.Fatalf("%s dash drew under an exhausting top margin at '%v'", name, p.size)
						}
						if created[name] {
							t.Fatalf("%s dash created its panel window under an exhausting top margin", name)
						}
					}
					if inner.drawn != 0 {
						t.Fatalf("inner drew under an exhausting top margin at '%v'", inner.size)
					}
					for _, dash := range []*Dash{dm.Left, dm.Top, dm.Right, dm.Bottom} {
						if dash.Focused {
							t.Fatalf("undrawn dash '%s' must not hold focus", dash.Name)
						}
					}
					if dm.Focused != nil {
						t.Fatalf("manager routes focus to undrawn dash '%s'", dm.Focused.Name)
					}
					if after != (imgui.Vec2{X: 50, Y: 40}) {
						t.Fatalf("the next window took leftover dash geometry: '%v'", after)
					}
					if dm.Left.CurrentSize <= 30 || dm.Top.CurrentSize <= 30 || dm.Bottom.CurrentSize <= 30 {
						t.Fatalf("shown dashes must keep growing while exhausted: left '%d', top '%d', bottom '%d'", dm.Left.CurrentSize, dm.Top.CurrentSize, dm.Bottom.CurrentSize)
					}
					if dm.Right.CurrentSize >= 30 {
						t.Fatalf("a hidden dash must keep shrinking while exhausted, at '%d'", dm.Right.CurrentSize)
					}
				})
			}
		}
	}
}

// TestAllocation_DashManagerTopMarginShortRemainder checks that dashes taller than
// the height left below TopMargin are clamped to it, never reaching into the inset.
func TestAllocation_DashManagerTopMarginShortRemainder(t *testing.T) {
	precedences := map[string]DashPrecedence{"vertical": VerticalPrecedence, "horizontal": HorizontalPrecedence}
	for styleName, preset := range presets {
		for precedenceName, precedence := range precedences {
			t.Run(styleName+"/"+precedenceName, func(t *testing.T) {
				dm := NewDashManager()
				dm.Precedence, dm.TopMargin = precedence, 140
				dm.Top, dm.Bottom = NewDash("top", &probe{}), NewDash("bottom", &probe{})
				dm.Top.TargetSize, dm.Top.CurrentSize = 90, 90
				dm.Bottom.TargetSize, dm.Bottom.CurrentSize = 80, 80
				var remainder imgui.Rect
				panels := map[string]imgui.Rect{}
				root := NewFunc(func(state *State) {
					size := imgui.Vec2{X: 600, Y: 200}
					remainder = imgui.Rect{Min: imgui.CursorScreenPos().Add(imgui.Vec2{Y: 140}), Max: imgui.CursorScreenPos().Add(size)}
					dm.Draw(&State{Size: size, IO: state.IO, App: state.App})
					for _, name := range []string{"top", "bottom"} {
						window := imgui.InternalFindWindowByName(fmt.Sprintf("%s/%s_%08X", imgui.InternalCurrentWindow().Name(), name, uint32(imgui.IDStr(name))))
						if ptr, finish := window.Handle(); ptr != nil {
							finish()
							panels[name] = imgui.Rect{Min: window.Pos(), Max: window.Pos().Add(window.Size())}
						} else {
							finish()
						}
					}
				})
				runRoot(t, preset(), root, nil, imgui.Vec2{X: 900, Y: 700}, 2)
				for _, name := range []string{"top", "bottom"} {
					panel, ok := panels[name]
					if !ok {
						t.Fatalf("%s dash skipped with 60 px still available", name)
					}
					if !within(panel, remainder) {
						t.Fatalf("%s panel '%v' escapes the area below the top margin '%v'", name, panel, remainder)
					}
				}
			})
		}
	}
}
