package dfx

import (
	"github.com/AllenDang/cimgui-go/imgui"
)

// ToolbarLayout provides vertical centering helpers for items inside a toolbar.
type ToolbarLayout struct {
	startY     float32
	lineHeight float32
}

// CenterFrame sets the cursor Y to vertically center a frame-height item
// (combo, button, input) within the toolbar.
func (t *ToolbarLayout) CenterFrame() {
	imgui.SetCursorPosY(t.startY + (t.lineHeight-imgui.FrameHeight())/2)
}

// CenterText sets the cursor Y to align standalone text with the text baseline
// of frame-height items (combos, buttons) in the toolbar.
func (t *ToolbarLayout) CenterText() {
	t.CenterFrame()
	imgui.AlignTextToFramePadding()
}

// Toolbar draws a full-width header bar with the given label.
func Toolbar(label string) {
	ToolbarEx(label, nil)
}

// ToolbarEx draws a full-width header bar with the given label,
// and optionally calls extra to draw additional controls on the same line.
func ToolbarEx(label string, extra func()) {
	ToolbarExLayout(label, func(_ *ToolbarLayout) {
		if extra != nil {
			extra()
		}
	})
}

// ToolbarExLayout draws a full-width header bar with the given label,
// and calls extra with a ToolbarLayout for precise vertical centering of
// mixed item types (combos, text, buttons).
func ToolbarExLayout(label string, extra func(*ToolbarLayout)) {
	availWidth := imgui.ContentRegionAvail().X
	framePadding := imgui.CurrentStyle().FramePadding().Y
	textSize := imgui.CalcTextSize(label)
	contentHeight := textSize.Y
	if fh := imgui.FrameHeight(); fh > contentHeight {
		contentHeight = fh
	}
	lineHeight := contentHeight + framePadding*2.5

	// save starting position for absolute positioning
	startY := imgui.CursorPosY()

	layout := &ToolbarLayout{startY: startY, lineHeight: lineHeight}

	// draw background rectangle
	cursorPos := imgui.CursorScreenPos()
	drawList := imgui.WindowDrawList()
	headerColor := imgui.ColorConvertFloat4ToU32(imgui.CurrentStyle().Colors()[imgui.ColHeader])
	drawList.AddRectFilled(
		cursorPos,
		imgui.Vec2{X: cursorPos.X + availWidth, Y: cursorPos.Y + lineHeight},
		headerColor,
	)

	imgui.Dummy(imgui.Vec2{X: 3, Y: 3})
	imgui.SameLine()

	// draw text centered vertically
	layout.CenterText()
	Text(label)

	// draw extra controls
	if extra != nil {
		SameLine()
		extra(layout)
	}

	// set cursor to end of toolbar using absolute position
	imgui.SetCursorPosY(startY + lineHeight)

	// imgui requires an item after SetCursorPos to validate window boundaries
	imgui.Dummy(imgui.Vec2{})
}
