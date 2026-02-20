package dfx

import (
	"github.com/AllenDang/cimgui-go/imgui"
)

type DashPrecedence int

const (
	VerticalPrecedence = DashPrecedence(iota)
	HorizontalPrecedence
)

type DashManager struct {
	Container
	Precedence DashPrecedence
	TopMargin  float32
	Margin     float32
	Left       *Dash
	Top        *Dash
	Right      *Dash
	Bottom     *Dash
	Focused    *Dash
	Inner      Component
}

func NewDashManager() *DashManager {
	return &DashManager{
		Container: Container{
			Visible: true,
		},
		Precedence: HorizontalPrecedence,
		TopMargin:  0.0,
		Margin:     5.0,
	}
}

func (d *DashManager) Draw(state *State) {
	if !d.Visible {
		return
	}

	bounds := state.Size
	d.Focused = nil
	leftWidth := float32(0)
	topHeight := float32(0)
	rightWidth := float32(0)
	bottomHeight := float32(0)

	if d.Precedence == VerticalPrecedence {
		if d.Left != nil {
			leftWidth = float32(d.Left.CurrentSize)
			d.Left.DrawDash(state, imgui.Vec4{X: 0, Y: d.TopMargin, Z: leftWidth, W: bounds.Y}, LeftDash)
			if d.Left.Focused {
				d.Focused = d.Left
			}
		}
		if d.Right != nil {
			rightWidth = float32(d.Right.CurrentSize)
			bounds := imgui.Vec4{X: bounds.X - rightWidth, Y: d.TopMargin, Z: bounds.X, W: bounds.Y}
			d.Right.DrawDash(state, bounds, RightDash)
			if d.Right.Focused {
				d.Focused = d.Right
			}
		}
		if d.Top != nil {
			topHeight = float32(d.Top.CurrentSize)
			d.Top.DrawDash(state, imgui.Vec4{X: leftWidth + d.Margin, Y: d.TopMargin, Z: bounds.X - (leftWidth + d.Margin + d.Margin + rightWidth), W: d.TopMargin + topHeight}, TopDash)
			if d.Top.Focused {
				d.Focused = d.Top
			}
		}
		if d.Bottom != nil {
			bottomHeight = float32(d.Bottom.CurrentSize)
			d.Bottom.DrawDash(state, imgui.Vec4{X: leftWidth + d.Margin, Y: bottomHeight, Z: bounds.X - (leftWidth + d.Margin + d.Margin + rightWidth), W: bounds.Y}, BottomDash)
			if d.Bottom.Focused {
				d.Focused = d.Bottom
			}
		}
	} else if d.Precedence == HorizontalPrecedence {
		if d.Top != nil {
			topHeight = d.TopMargin + float32(d.Top.CurrentSize)
			d.Top.DrawDash(state, imgui.Vec4{X: 0, Y: d.TopMargin, Z: bounds.X, W: topHeight}, TopDash)
			if d.Top.Focused {
				d.Focused = d.Top
			}
		}
		if d.Bottom != nil {
			bottomHeight = float32(d.Bottom.CurrentSize)
			d.Bottom.DrawDash(state, imgui.Vec4{X: 0, Y: bounds.Y - bottomHeight, Z: bounds.X, W: bounds.Y}, BottomDash)
			if d.Bottom.Focused {
				d.Focused = d.Bottom
			}
		}
		if d.Left != nil {
			leftWidth = float32(d.Left.CurrentSize)
			d.Left.DrawDash(state, imgui.Vec4{X: 0, Y: topHeight + d.Margin, Z: leftWidth, W: bounds.Y - (bottomHeight + d.Margin + d.Margin + topHeight)}, LeftDash)
			if d.Left.Focused {
				d.Focused = d.Left
			}
		}
		if d.Right != nil {
			rightWidth = float32(d.Right.CurrentSize)
			d.Right.DrawDash(state, imgui.Vec4{X: bounds.X - rightWidth, Y: topHeight + d.Margin, Z: rightWidth, W: bounds.Y - (bottomHeight + d.Margin + d.Margin + topHeight)}, RightDash)
			if d.Right.Focused {
				d.Focused = d.Right
			}
		}
	}

	if d.Inner != nil {
		pos := imgui.WindowPos().Add(imgui.Vec2{X: leftWidth + d.Margin, Y: topHeight + d.Margin})
		imgui.SetNextWindowPos(pos)
		size := imgui.Vec2{X: bounds.X - leftWidth - rightWidth - (d.Margin * 2.0), Y: bounds.Y - topHeight - bottomHeight - (d.Margin * 2.0)}
		imgui.SetNextWindowSize(size)

		windowFlags := imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoScrollWithMouse

		imgui.BeginChildStrV("##dashManagerInner", size, imgui.ChildFlagsNone, windowFlags)

		// create state for the inner component
		innerState := &State{
			Size:     size,
			Position: imgui.Vec2{}, // position is relative to the child window
			IO:       state.IO,
			App:      state.App,
			Parent:   d,
		}
		d.Inner.Draw(innerState)
		imgui.EndChild()
	}

	// base container drawing support
	if d.OnDraw != nil {
		d.OnDraw(state)
	}
	for _, child := range d.Children {
		child.Draw(state)
	}
}

// Actions implements Component by prioritizing focused dash actions
func (d *DashManager) Actions() *ActionRegistry {
	// if there's a focused dash, prioritize its actions
	if d.Focused != nil && d.Focused.Component != nil {
		return d.Focused.Component.Actions()
	}
	// if there's an inner component, use its actions
	if d.Inner != nil {
		return d.Inner.Actions()
	}
	// fall back to container actions
	return d.Container.Actions()
}

// LocalActions returns dash manager-local actions without delegation.
func (d *DashManager) LocalActions() *ActionRegistry {
	return d.Container.Actions()
}

// ChildActions returns the active child component for action traversal.
func (d *DashManager) ChildActions() []Component {
	if d.Focused != nil && d.Focused.Component != nil {
		return []Component{d.Focused.Component}
	}
	if d.Inner != nil {
		return []Component{d.Inner}
	}
	return nil
}
