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

func (d *Dash) DrawDash(state *State, bounds Bounds, attachment DashAttachment) {
	if d.CurrentSize > 0 {
		imgui.SetNextWindowBgAlpha(DashBackgroundAlpha)
		imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, DashWindowRounding)
		sfSize := d.boundsAndSize(bounds, attachment)

		windowFlags := imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoTitleBar |
			imgui.WindowFlagsNoResize | imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoScrollWithMouse

		imgui.BeginChildStrV(d.Name, imgui.Vec2{X: bounds.W, Y: bounds.H}, imgui.ChildFlagsNone, windowFlags)

		if d.CurrentSize == d.TargetSize {
			if d.Resizable {
				dhp := d.dragHandlePos(bounds, attachment)
				imgui.SetCursorPos(dhp)
				imgui.PushStyleColorVec4(imgui.ColText, imgui.CurrentStyle().Colors()[imgui.ColHeaderActive])
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

			childSize := imgui.Vec2{X: 0, Y: 0}
			if attachment != TopDash {
				windowPadding := imgui.CurrentStyle().WindowPadding()
				if d.Resizable {
					imgui.SetCursorPos(imgui.Vec2{X: windowPadding.X, Y: DashTitleBarHeight})
				} else {
					imgui.SetCursorPos(windowPadding)
				}
			} else {
				windowPadding := imgui.CurrentStyle().WindowPadding()
				imgui.SetCursorPos(windowPadding)
				if d.Resizable {
					childSize = imgui.Vec2{X: bounds.W - (windowPadding.X * 2), Y: bounds.H - (windowPadding.Y * 2) - DashTitleBarOffset}
				}
			}
			imgui.PushStyleVarFloat(imgui.StyleVarScrollbarSize, DashScrollbarSize)
			imgui.BeginChildStrV("##dashSurface", childSize, 0, 0)
			if d.Visible && d.Component != nil {
				windowPadding := imgui.CurrentStyle().WindowPadding()
				sfSize = sfSize.Sub(imgui.Vec2{X: windowPadding.X * 2, Y: windowPadding.Y * 2})
				if d.Resizable {
					sfSize = sfSize.Sub(imgui.Vec2{X: 0, Y: DashSurfacePadding})
				}

				// create state for the child component
				childState := &State{
					Size:     sfSize,
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

			if !d.Focused {
				d.Focused = d.Visible && imgui.IsWindowFocused() // outer window may be focused, dragging
			}
		}
		imgui.EndChild()
		imgui.PopStyleVar()
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

func (d *Dash) boundsAndSize(bounds Bounds, attachment DashAttachment) imgui.Vec2 {
	winPos := imgui.WindowPos()
	switch attachment {
	case LeftDash:
		imgui.SetNextWindowPos(winPos.Add(imgui.Vec2{X: bounds.X, Y: bounds.Y}))
		size := imgui.Vec2{X: float32(d.CurrentSize), Y: bounds.H}
		imgui.SetNextWindowSize(size)
		return size

	case TopDash:
		imgui.SetNextWindowPos(winPos.Add(imgui.Vec2{X: bounds.X, Y: bounds.Y}))
		size := imgui.Vec2{X: bounds.W, Y: float32(d.CurrentSize)}
		imgui.SetNextWindowSize(size)
		return size

	case RightDash:
		imgui.SetNextWindowPos(winPos.Add(imgui.Vec2{X: bounds.X, Y: bounds.Y}))
		size := imgui.Vec2{X: float32(d.CurrentSize), Y: bounds.H}
		imgui.SetNextWindowSize(size)
		return size

	default: // BottomDash
		imgui.SetNextWindowPos(winPos.Add(imgui.Vec2{X: bounds.X, Y: bounds.Y + bounds.H - float32(d.CurrentSize)}))
		size := imgui.Vec2{X: bounds.W, Y: float32(d.CurrentSize)}
		imgui.SetNextWindowSize(size)
		return size
	}
}

func (d *Dash) dragHandlePos(bounds Bounds, attachment DashAttachment) imgui.Vec2 {
	switch attachment {
	case LeftDash:
		return imgui.Vec2{X: float32(d.CurrentSize) - DashDragHandleOffset, Y: DefaultItemSpacing + 1}

	case TopDash:
		return imgui.Vec2{X: bounds.W - DashDragHandleOffset, Y: bounds.H - DashDragHandleOffset}

	case RightDash:
		return imgui.Vec2{X: DefaultWindowPadding, Y: DefaultItemSpacing + 1}

	default: // BottomDash
		return imgui.Vec2{X: bounds.W - DashDragHandleOffset, Y: DefaultItemSpacing + 1}
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
