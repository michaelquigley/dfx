package dfx

import "github.com/AllenDang/cimgui-go/imgui"

// CenterText draws text centered horizontally and vertically in the available content region.
func CenterText(text string) {
	avail := imgui.ContentRegionAvail()
	textSize := imgui.CalcTextSize(text)
	cursorPos := imgui.CursorPos()
	imgui.SetCursorPos(imgui.Vec2{
		X: cursorPos.X + (avail.X-textSize.X)/2,
		Y: cursorPos.Y + (avail.Y-textSize.Y)/2,
	})
	imgui.Text(text)
}

// CenterTextDisabled draws disabled text centered horizontally and vertically in the available content region.
func CenterTextDisabled(text string) {
	avail := imgui.ContentRegionAvail()
	textSize := imgui.CalcTextSize(text)
	cursorPos := imgui.CursorPos()
	imgui.SetCursorPos(imgui.Vec2{
		X: cursorPos.X + (avail.X-textSize.X)/2,
		Y: cursorPos.Y + (avail.Y-textSize.Y)/2,
	})
	imgui.TextDisabled(text)
}
