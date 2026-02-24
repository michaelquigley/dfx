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

	size := state.Size
	d.Focused = nil
	leftWidth := float32(0)
	topHeight := float32(0)
	rightWidth := float32(0)
	bottomHeight := float32(0)

	if d.Precedence == VerticalPrecedence {
		if d.Left != nil {
			leftWidth = float32(d.Left.CurrentSize)
			d.Left.DrawDash(state, Bounds{X: 0, Y: d.TopMargin, W: leftWidth, H: size.Y}, LeftDash)
			if d.Left.Focused {
				d.Focused = d.Left
			}
		}
		if d.Right != nil {
			rightWidth = float32(d.Right.CurrentSize)
			d.Right.DrawDash(state, Bounds{X: size.X - rightWidth, Y: d.TopMargin, W: rightWidth, H: size.Y}, RightDash)
			if d.Right.Focused {
				d.Focused = d.Right
			}
		}
		if d.Top != nil {
			topHeight = float32(d.Top.CurrentSize)
			availW := size.X - (leftWidth + d.Margin*2 + rightWidth)
			d.Top.DrawDash(state, Bounds{X: leftWidth + d.Margin, Y: d.TopMargin, W: availW, H: topHeight}, TopDash)
			if d.Top.Focused {
				d.Focused = d.Top
			}
		}
		if d.Bottom != nil {
			bottomHeight = float32(d.Bottom.CurrentSize)
			availW := size.X - (leftWidth + d.Margin*2 + rightWidth)
			d.Bottom.DrawDash(state, Bounds{X: leftWidth + d.Margin, Y: 0, W: availW, H: size.Y}, BottomDash)
			if d.Bottom.Focused {
				d.Focused = d.Bottom
			}
		}
	} else if d.Precedence == HorizontalPrecedence {
		if d.Top != nil {
			topHeight = d.TopMargin + float32(d.Top.CurrentSize)
			panelH := float32(d.Top.CurrentSize)
			d.Top.DrawDash(state, Bounds{X: 0, Y: d.TopMargin, W: size.X, H: panelH}, TopDash)
			if d.Top.Focused {
				d.Focused = d.Top
			}
		}
		if d.Bottom != nil {
			bottomHeight = float32(d.Bottom.CurrentSize)
			d.Bottom.DrawDash(state, Bounds{X: 0, Y: 0, W: size.X, H: size.Y}, BottomDash)
			if d.Bottom.Focused {
				d.Focused = d.Bottom
			}
		}
		if d.Left != nil {
			leftWidth = float32(d.Left.CurrentSize)
			availH := size.Y - (bottomHeight + d.Margin*2 + topHeight)
			d.Left.DrawDash(state, Bounds{X: 0, Y: topHeight + d.Margin, W: leftWidth, H: availH}, LeftDash)
			if d.Left.Focused {
				d.Focused = d.Left
			}
		}
		if d.Right != nil {
			rightWidth = float32(d.Right.CurrentSize)
			availH := size.Y - (bottomHeight + d.Margin*2 + topHeight)
			d.Right.DrawDash(state, Bounds{X: size.X - rightWidth, Y: topHeight + d.Margin, W: rightWidth, H: availH}, RightDash)
			if d.Right.Focused {
				d.Focused = d.Right
			}
		}
	}

	if d.Inner != nil {
		pos := imgui.WindowPos().Add(imgui.Vec2{X: leftWidth + d.Margin, Y: topHeight + d.Margin})
		imgui.SetNextWindowPos(pos)
		innerSize := imgui.Vec2{X: size.X - leftWidth - rightWidth - (d.Margin * 2.0), Y: size.Y - topHeight - bottomHeight - (d.Margin * 2.0)}
		imgui.SetNextWindowSize(innerSize)

		windowFlags := imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoScrollWithMouse

		imgui.BeginChildStrV("##dashManagerInner", innerSize, imgui.ChildFlagsNone, windowFlags)

		// create state for the inner component
		innerState := &State{
			Size:     innerSize,
			Position: imgui.Vec2{}, // position is relative to the child window
			IO:       state.IO,
			App:      state.App,
			Parent:   d,
		}
		d.Inner.Draw(innerState)
		imgui.EndChild()
	}

	drawContainerExtensions(&d.Container, state)
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
