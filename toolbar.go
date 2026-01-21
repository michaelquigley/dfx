package dfx

import (
	"github.com/AllenDang/cimgui-go/imgui"
)

// Toolbar draws a full-width header bar with the given label.
func Toolbar(label string) {
	ToolbarEx(label, nil)
}

// ToolbarEx draws a full-width header bar with the given label,
// and optionally calls extra to draw additional controls on the same line.
func ToolbarEx(label string, extra func()) {
	availWidth := imgui.ContentRegionAvail().X
	framePadding := imgui.CurrentStyle().FramePadding().Y
	textSize := imgui.CalcTextSize(label)
	lineHeight := textSize.Y + framePadding*2.5

	// save starting position for absolute positioning
	startY := imgui.CursorPosY()

	// draw background rectangle
	cursorPos := imgui.CursorScreenPos()
	drawList := imgui.WindowDrawList()
	headerColor := imgui.ColorConvertFloat4ToU32(imgui.CurrentStyle().Colors()[imgui.ColHeader])
	drawList.AddRectFilled(
		cursorPos,
		imgui.Vec2{X: cursorPos.X + availWidth, Y: cursorPos.Y + lineHeight},
		headerColor,
	)

	// draw text with padding
	imgui.SetCursorPosY(imgui.CursorPosY() + imgui.CurrentStyle().FramePadding().Y)
	Text(label)

	// draw extra controls if provided
	if extra != nil {
		SameLine()
		extra()
	}

	// set cursor to end of toolbar using absolute position
	imgui.SetCursorPosY(startY + lineHeight)

	// imgui requires an item after SetCursorPos to validate window boundaries
	imgui.Dummy(imgui.Vec2{})
}
