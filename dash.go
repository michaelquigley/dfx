package dfx

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx/fonts"
)

// Bounds represents a rectangular region with explicit position and dimensions.
type Bounds struct {
	X, Y float32 // top-left position
	W, H float32 // width and height
}

type DashAttachment int

type Dash struct {
	Container
	Name         string
	Component    Component
	TargetSize   int
	CurrentSize  int
	MinSize      int
	MaxSize      int
	Resizable    bool
	TransitionMs int
	Focused      bool
}

func NewDash(name string, component Component) *Dash {
	return &Dash{
		Container: Container{
			Visible: true,
		},
		Name:         name,
		Component:    component,
		TargetSize:   DefaultDashSize,
		CurrentSize:  DefaultDashSize,
		MinSize:      DefaultDashMinSize,
		MaxSize:      DefaultDashMaxSize,
		Resizable:    true,
		TransitionMs: DefaultTransitionMs,
		Focused:      false,
	}
}

// DrawDash draws the dash within bounds, which are relative to the cursor when it
// is called: the content origin of the component laying the dash out.
func (d *Dash) DrawDash(state *State, bounds Bounds, attachment DashAttachment) {
	// an exhausted panel is skipped before any window state is queued or pushed:
	// imgui reads a zero child size as fill-remaining and a negative one as
	// fill-relative. the animation below still advances.
	pos, size := d.panelRect(imgui.CursorScreenPos(), bounds, attachment)
	if d.CurrentSize > 0 && size.X > 0 && size.Y > 0 && bounds.W > 0 && bounds.H > 0 {
		imgui.SetNextWindowBgAlpha(DashBackgroundAlpha)
		imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, currentStyleChrome().PanelRounding)
		imgui.SetNextWindowPos(pos)
		imgui.SetNextWindowSize(size)

		windowFlags := imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoTitleBar |
			imgui.WindowFlagsNoResize | imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoScrollWithMouse

		imgui.BeginChildStrV(d.Name, imgui.Vec2{X: bounds.W, Y: bounds.H}, imgui.ChildFlagsNone, windowFlags)
		panel := imgui.WindowSize()

		if d.CurrentSize == d.TargetSize {
			if d.Resizable {
				dhp := d.dragHandlePos(bounds, attachment)
				imgui.SetCursorPos(dhp)
				imgui.PushStyleColorVec4(imgui.ColText, styleColor(currentStyleRoles().Handle))
				imgui.TextUnformatted(fonts.ICON_DRAG_INDICATOR)
				imgui.PopStyleColor()

				imgui.SetCursorPos(dhp)
				imgui.InvisibleButton("##resize", imgui.Vec2{X: DragHandleSize, Y: DragHandleSize})
				if imgui.IsItemHovered() {
					if attachment == LeftDash || attachment == RightDash {
						imgui.SetMouseCursor(imgui.MouseCursorResizeEW)
					} else {
						imgui.SetMouseCursor(imgui.MouseCursorResizeNS)
					}
				}
				if imgui.IsItemActive() {
					delta := float32(0)
					if attachment == LeftDash || attachment == RightDash {
						delta = imgui.CurrentIO().MouseDelta().X
						if attachment == RightDash {
							delta *= -1
						}
					} else if attachment == TopDash || attachment == BottomDash {
						delta = imgui.CurrentIO().MouseDelta().Y
						if attachment == BottomDash {
							delta *= -1
						}
					}
					d.CurrentSize += int(delta)
					d.TargetSize += int(delta)
					if d.CurrentSize < DefaultDashMinSize {
						d.CurrentSize = DefaultDashMinSize
						d.TargetSize = DefaultDashMinSize
					}
					if d.MinSize > -1 && d.CurrentSize < d.MinSize {
						d.CurrentSize = d.MinSize
						d.TargetSize = d.MinSize
					}
					if d.MaxSize > -1 && d.CurrentSize > d.MaxSize {
						d.CurrentSize = d.MaxSize
						d.TargetSize = d.MaxSize
					}
				}
			}

			// the surface is inset by the panel padding on every side, and below
			// the handle row once: at the top, or for a top dash at the bottom.
			surfacePos, surfaceSize := d.surfaceRect(panel, attachment)
			if surfaceSize.X > 0 && surfaceSize.Y > 0 {
				imgui.SetCursorPos(surfacePos)
				imgui.PushStyleVarFloat(imgui.StyleVarScrollbarSize, DashScrollbarSize)
				imgui.BeginChildStrV("##dashSurface", surfaceSize, 0, 0)
				if usable, ok := usableSize(); ok && d.Visible && d.Component != nil {
					// create state for the child component
					childState := &State{
						Size:     usable,
						Position: imgui.Vec2{}, // position is relative to the child window
						IO:       state.IO,
						App:      state.App,
						Parent:   d,
					}
					d.Component.Draw(childState)
				}
				d.Focused = d.Visible && imgui.IsWindowFocused()
				imgui.EndChild()
				imgui.PopStyleVar()
			} else {
				d.Focused = false
			}

			if !d.Focused {
				d.Focused = d.Visible && imgui.IsWindowFocused() // outer window may be focused, dragging
			}
		}
		imgui.EndChild()
		imgui.PopStyleVar()
	} else {
		d.Focused = false // an undrawn panel cannot hold focus for action routing
	}

	if d.Visible {
		if d.CurrentSize < d.TargetSize {
			d.CurrentSize += int(d.dashPxPerFrame())
			if d.CurrentSize > d.TargetSize {
				d.CurrentSize = d.TargetSize
			}
		}
	} else {
		if d.CurrentSize > 0 {
			d.CurrentSize -= int(d.dashPxPerFrame())
			if d.CurrentSize < 0 {
				d.CurrentSize = 0
			}
		}
	}
}

// dashHandleRow is the height a resizable dash reserves for its drag handle: the
// handle's inset from the panel edge, the handle or the glyph drawn in it,
// whichever is taller, and a small gap. 27 px in Classic.
func dashHandleRow() float32 {
	return imgui.CurrentStyle().ItemSpacing().Y + 1 + max(DragHandleSize, imgui.FontSize()) + 2
}

// surfaceRect is the content surface of a panel, in panel coordinates, clamped so
// an exhausted panel yields an empty size rather than a negative one.
func (d *Dash) surfaceRect(panel imgui.Vec2, attachment DashAttachment) (imgui.Vec2, imgui.Vec2) {
	padding := imgui.CurrentStyle().WindowPadding()
	top, bottom := padding.Y, padding.Y
	if d.Resizable {
		if attachment == TopDash {
			bottom += max(DragHandleSize, imgui.FontSize()) + 2
		} else {
			top = dashHandleRow()
		}
	}
	size := imgui.Vec2{X: max(0, panel.X-padding.X*2), Y: max(0, panel.Y-top-bottom)}
	return imgui.Vec2{X: padding.X, Y: top}, size
}

// panelRect is the panel's screen position and size within bounds, relative to
// origin: its current size, clamped so the panel never leaves its bounds. it
// queues nothing, so an exhausted panel can be skipped cleanly.
func (d *Dash) panelRect(origin imgui.Vec2, bounds Bounds, attachment DashAttachment) (imgui.Vec2, imgui.Vec2) {
	switch attachment {
	case LeftDash, RightDash:
		return origin.Add(imgui.Vec2{X: bounds.X, Y: bounds.Y}), imgui.Vec2{X: min(float32(d.CurrentSize), bounds.W), Y: bounds.H}
	case TopDash:
		return origin.Add(imgui.Vec2{X: bounds.X, Y: bounds.Y}), imgui.Vec2{X: bounds.W, Y: min(float32(d.CurrentSize), bounds.H)}
	default: // BottomDash
		h := min(float32(d.CurrentSize), bounds.H)
		return origin.Add(imgui.Vec2{X: bounds.X, Y: bounds.Y + bounds.H - h}), imgui.Vec2{X: bounds.W, Y: h}
	}
}

func (d *Dash) dragHandlePos(bounds Bounds, attachment DashAttachment) imgui.Vec2 {
	inset := imgui.CurrentStyle().ItemSpacing().Y + 1
	switch attachment {
	case LeftDash:
		return imgui.Vec2{X: float32(d.CurrentSize) - DashDragHandleOffset, Y: inset}

	case TopDash:
		return imgui.Vec2{X: bounds.W - DashDragHandleOffset, Y: bounds.H - DashDragHandleOffset}

	case RightDash:
		return imgui.Vec2{X: imgui.CurrentStyle().WindowPadding().X, Y: inset}

	default: // BottomDash
		return imgui.Vec2{X: bounds.W - DashDragHandleOffset, Y: inset}
	}
}

func (d *Dash) dashPxPerFrame() float32 {
	return pxPerFrame(float32(d.TargetSize), d.TransitionMs)
}

// Draw implements Component interface - this is for when Dash is used as a standalone component
func (d *Dash) Draw(state *State) {
	// when used as a standalone component, we just draw our inner component
	if d.Visible && d.Component != nil {
		d.Component.Draw(state)
	}
}

// Actions implements Component by delegating to the child component
func (d *Dash) Actions() *ActionRegistry {
	if d.Component != nil {
		return d.Component.Actions()
	}
	return d.Container.Actions()
}

// LocalActions returns dash-local actions without delegation.
func (d *Dash) LocalActions() *ActionRegistry {
	return d.Container.Actions()
}

// ChildActions returns dash content for action traversal.
func (d *Dash) ChildActions() []Component {
	if d.Component != nil {
		return []Component{d.Component}
	}
	return nil
}

const (
	LeftDash DashAttachment = iota
	RightDash
	TopDash
	BottomDash
)

const (
	DefaultDashSize      = 400
	DefaultDashMinSize   = 40
	DefaultDashMaxSize   = 1000
	DefaultTransitionMs  = 100
	DashBackgroundAlpha  = 0.85
	DashWindowRounding   = 5
	DashScrollbarSize    = 5
	DragHandleSize       = 20
	DashTitleBarHeight   = 27
	DashTitleBarOffset   = 22
	DashDragHandleOffset = 22
	DashSurfacePadding   = 20
	FramerateToMs        = 1000
)
