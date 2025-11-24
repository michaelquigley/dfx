package dfx

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
)

type SizeDebugger struct {
	Margin         float32
	disabled       bool
	actionRegistry *ActionRegistry
}

func NewSizeDebugger() *SizeDebugger {
	sd := &SizeDebugger{
		Margin:         4,
		actionRegistry: NewActionRegistry(),
	}
	sd.actionRegistry.MustRegister("disable", "Shift+Alt+D", func() { sd.disabled = !sd.disabled })
	return sd
}

func (sd *SizeDebugger) Draw(state *State) {
	cursor := imgui.WindowPos()
	dl := imgui.WindowDrawList()
	color := imgui.ColorConvertFloat4ToU32(imgui.CurrentStyle().Colors()[imgui.ColHeaderHovered])

	ul := cursor.Add(imgui.Vec2{X: sd.Margin, Y: sd.Margin})
	ll := cursor.Add(imgui.Vec2{X: sd.Margin, Y: state.Size.Y - sd.Margin})
	ur := cursor.Add(imgui.Vec2{X: state.Size.X - sd.Margin, Y: sd.Margin})
	lr := cursor.Add(imgui.Vec2{X: state.Size.X - sd.Margin, Y: state.Size.Y - sd.Margin})

	// crosses
	dl.AddLine(ul, lr, color)
	dl.AddLine(ur, ll, color)

	// outer box
	dl.AddLine(ul, ur, color)
	dl.AddLine(ur, lr, color)
	dl.AddLine(lr, ll, color)
	dl.AddLine(ll, ul, color)

	if !sd.disabled {
		// label
		label := fmt.Sprintf("( x: %v, y: %v )", state.Size.X, state.Size.Y)
		labelSize := imgui.CalcTextSize(label)
		imgui.SetCursorPos(imgui.Vec2{X: (state.Size.X / 2) - (labelSize.X / 2), Y: (state.Size.Y / 2) - (labelSize.Y / 2)})
		imgui.TextUnformatted(label)
	}
}

func (sd *SizeDebugger) Actions() *ActionRegistry {
	return sd.actionRegistry
}
