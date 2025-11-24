package dfx

import "github.com/AllenDang/cimgui-go/imgui"

const (
	// default padding and spacing
	DefaultWindowPadding = 4
	DefaultFramePadding  = 4
	DefaultItemSpacing   = 4

	// scrollbar and border sizes
	DefaultScrollbarSize = 12
	DefaultWindowBorder  = 0
	DefaultChildBorder   = 0
	DefaultPopupBorder   = 1
	DefaultFrameBorder   = 0

	// rounding values
	DefaultWindowRounding    = 3
	DefaultChildRounding     = 0
	DefaultFrameRounding     = 3
	DefaultPopupRounding     = 3
	DefaultScrollbarRounding = 2
	DefaultGrabRounding      = 2
)

// DefaultStyle sets up the default ImGui style parameters
// this should be called after font setup but before theme application
func DefaultStyle() {
	style := imgui.CurrentStyle()

	// spacing and padding
	style.SetWindowPadding(imgui.Vec2{X: DefaultWindowPadding, Y: DefaultWindowPadding})
	style.SetFramePadding(imgui.Vec2{X: DefaultFramePadding, Y: DefaultFramePadding})
	style.SetItemSpacing(imgui.Vec2{X: DefaultItemSpacing, Y: DefaultItemSpacing})

	// sizes
	style.SetScrollbarSize(DefaultScrollbarSize)

	// borders
	style.SetWindowBorderSize(DefaultWindowBorder)
	style.SetChildBorderSize(DefaultChildBorder)
	style.SetPopupBorderSize(DefaultPopupBorder)
	style.SetFrameBorderSize(DefaultFrameBorder)

	// rounding
	style.SetWindowRounding(DefaultWindowRounding)
	style.SetChildRounding(DefaultChildRounding)
	style.SetFrameRounding(DefaultFrameRounding)
	style.SetPopupRounding(DefaultPopupRounding)
	style.SetScrollbarRounding(DefaultScrollbarRounding)
	style.SetGrabRounding(DefaultGrabRounding)
}
