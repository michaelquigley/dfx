package dfx

import "github.com/AllenDang/cimgui-go/imgui"

// Theme interface allows for extensible theming system
type Theme interface {
	Apply() // applies the theme to ImGui style
	Name() string
}

// HueColorScheme creates themes based on HSV color space
// this allows dynamic theme generation with consistent relationships
type HueColorScheme struct {
	name           string
	Hue            int
	TextSaturation float32
	TextValue      float32
	MainSaturation float32
	MainValue      float32
	AreaSaturation float32
	AreaValue      float32
	BgSaturation   float32
	BgValue        float32
}

// NewHueColorScheme creates a new HSV-based color scheme.
// hue is expected in the 0-255 range (matching the ImGui style editor convention),
// not degrees. sat and value are also in 0-255 range.
func NewHueColorScheme(name string, hue int, sat, value float32) *HueColorScheme {
	return &HueColorScheme{
		name:           name,
		Hue:            hue,
		TextSaturation: 0.0 / 255.0,
		TextValue:      value / 255.0,
		MainSaturation: sat / 255.0,
		MainValue:      (value / 2) / 255.0,
		AreaSaturation: (sat / 2) / 255.0,
		AreaValue:      (value / 2) / 255.0,
		BgSaturation:   (sat / 3) / 255.0,
		BgValue:        (value / 3) / 255.0,
	}
}

func (s *HueColorScheme) Name() string {
	return s.name
}

func (s *HueColorScheme) Apply() {
	// create color values from HSV
	text := imgui.Color{}
	text.SetHSV(float32(s.Hue)/255.0, s.TextSaturation, s.TextValue)
	main := imgui.Color{}
	main.SetHSV(float32(s.Hue)/255.0, s.MainSaturation, s.MainValue)
	area := imgui.Color{}
	area.SetHSV(float32(s.Hue)/255.0, s.AreaSaturation, s.AreaValue)
	bg := imgui.Color{}
	bg.SetHSV(float32(s.Hue)/255.0, s.BgSaturation, s.BgValue)

	// apply to ImGui style
	colors := imgui.CurrentStyle().Colors()
	colors[imgui.ColText] = imgui.Vec4{X: text.FieldValue.X, Y: text.FieldValue.Y, Z: text.FieldValue.Z, W: 1}
	colors[imgui.ColTextDisabled] = imgui.Vec4{X: text.FieldValue.X, Y: text.FieldValue.Y, Z: text.FieldValue.Z, W: 0.58}
	colors[imgui.ColWindowBg] = imgui.Vec4{X: bg.FieldValue.X, Y: bg.FieldValue.Y, Z: bg.FieldValue.Z, W: 1}
	colors[imgui.ColChildBg] = imgui.Vec4{X: area.FieldValue.X, Y: area.FieldValue.Y, Z: area.FieldValue.Z, W: 0}
	colors[imgui.ColPopupBg] = imgui.Vec4{X: area.FieldValue.X, Y: area.FieldValue.Y, Z: area.FieldValue.Z, W: 1}
	colors[imgui.ColBorder] = imgui.Vec4{X: text.FieldValue.X, Y: text.FieldValue.Y, Z: text.FieldValue.Z, W: 0.3}
	colors[imgui.ColBorderShadow] = imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0}
	colors[imgui.ColFrameBg] = imgui.Vec4{X: area.FieldValue.X, Y: area.FieldValue.Y, Z: area.FieldValue.Z, W: 0.0}
	colors[imgui.ColFrameBgHovered] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.68}
	colors[imgui.ColFrameBgActive] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 1}
	colors[imgui.ColTitleBg] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.45}
	colors[imgui.ColTitleBgCollapsed] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.35}
	colors[imgui.ColTitleBgActive] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.78}
	colors[imgui.ColMenuBarBg] = imgui.Vec4{X: area.FieldValue.X, Y: area.FieldValue.Y, Z: area.FieldValue.Z, W: 0.57}
	colors[imgui.ColScrollbarBg] = imgui.Vec4{X: area.FieldValue.X, Y: area.FieldValue.Y, Z: area.FieldValue.Z, W: 0.25}
	colors[imgui.ColScrollbarGrab] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 1}
	colors[imgui.ColScrollbarGrabHovered] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.78}
	colors[imgui.ColScrollbarGrabActive] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 1}
	colors[imgui.ColCheckMark] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.8}
	colors[imgui.ColSliderGrab] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 1}
	colors[imgui.ColSliderGrabActive] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 1}
	colors[imgui.ColButton] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.44}
	colors[imgui.ColButtonHovered] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.86}
	colors[imgui.ColButtonActive] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 1}
	colors[imgui.ColHeader] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.76}
	colors[imgui.ColHeaderHovered] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.86}
	colors[imgui.ColHeaderActive] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 1}
	colors[imgui.ColSeparator] = imgui.Vec4{X: text.FieldValue.X, Y: text.FieldValue.Y, Z: text.FieldValue.Z, W: 0.32}
	colors[imgui.ColSeparatorHovered] = imgui.Vec4{X: text.FieldValue.X, Y: text.FieldValue.Y, Z: text.FieldValue.Z, W: 0.78}
	colors[imgui.ColSeparatorActive] = imgui.Vec4{X: text.FieldValue.X, Y: text.FieldValue.Y, Z: text.FieldValue.Z, W: 1}
	colors[imgui.ColResizeGrip] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.2}
	colors[imgui.ColResizeGripHovered] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.78}
	colors[imgui.ColResizeGripActive] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 1}
	colors[imgui.ColTab] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.44}
	colors[imgui.ColTabHovered] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.86}
	colors[imgui.ColTabSelected] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 1}
	colors[imgui.ColPlotLines] = imgui.Vec4{X: text.FieldValue.X, Y: text.FieldValue.Y, Z: text.FieldValue.Z, W: 0.63}
	colors[imgui.ColPlotLinesHovered] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 1}
	colors[imgui.ColPlotHistogram] = imgui.Vec4{X: text.FieldValue.X, Y: text.FieldValue.Y, Z: text.FieldValue.Z, W: 0.63}
	colors[imgui.ColPlotHistogramHovered] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 1}
	colors[imgui.ColTextSelectedBg] = imgui.Vec4{X: main.FieldValue.X, Y: main.FieldValue.Y, Z: main.FieldValue.Z, W: 0.43}
	colors[imgui.ColModalWindowDimBg] = imgui.Vec4{X: 0.2, Y: 0.2, Z: 0.2, W: 0.35}
	imgui.CurrentStyle().SetColors(&colors)
}

// ModernTheme implements a predefined dark theme
type ModernTheme struct{}

func (m *ModernTheme) Name() string {
	return "Modern Dark"
}

func (m *ModernTheme) Apply() {
	colors := imgui.CurrentStyle().Colors()
	colors[imgui.ColText] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 1.0}
	colors[imgui.ColTextDisabled] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 0.399}
	colors[imgui.ColWindowBg] = imgui.Vec4{X: 0.039, Y: 0.039, Z: 0.039, W: 0.94}
	colors[imgui.ColChildBg] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.0}
	colors[imgui.ColPopupBg] = imgui.Vec4{X: 0.051, Y: 0.051, Z: 0.051, W: 0.94}
	colors[imgui.ColBorder] = imgui.Vec4{X: 0.427, Y: 0.427, Z: 0.498, W: 0.5}
	colors[imgui.ColBorderShadow] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.0}
	colors[imgui.ColFrameBg] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.421}
	colors[imgui.ColFrameBgHovered] = imgui.Vec4{X: 0.141, Y: 0.141, Z: 0.141, W: 0.4}
	colors[imgui.ColFrameBgActive] = imgui.Vec4{X: 0.231, Y: 0.231, Z: 0.231, W: 0.863}
	colors[imgui.ColTitleBg] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0}
	colors[imgui.ColTitleBgActive] = imgui.Vec4{X: 0.094, Y: 0.094, Z: 0.094, W: 1.0}
	colors[imgui.ColTitleBgCollapsed] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.292}
	colors[imgui.ColMenuBarBg] = imgui.Vec4{X: 0.137, Y: 0.137, Z: 0.137, W: 1.0}
	colors[imgui.ColScrollbarBg] = imgui.Vec4{X: 0.02, Y: 0.02, Z: 0.02, W: 0.53}
	colors[imgui.ColScrollbarGrab] = imgui.Vec4{X: 0.31, Y: 0.31, Z: 0.31, W: 1.0}
	colors[imgui.ColScrollbarGrabHovered] = imgui.Vec4{X: 0.408, Y: 0.408, Z: 0.408, W: 1.0}
	colors[imgui.ColScrollbarGrabActive] = imgui.Vec4{X: 0.51, Y: 0.51, Z: 0.51, W: 1.0}
	colors[imgui.ColCheckMark] = imgui.Vec4{X: 0.98, Y: 0.259, Z: 0.259, W: 1.0}
	colors[imgui.ColSliderGrab] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 1.0}
	colors[imgui.ColSliderGrabActive] = imgui.Vec4{X: 0.98, Y: 0.259, Z: 0.259, W: 1.0}
	colors[imgui.ColButton] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.579}
	colors[imgui.ColButtonHovered] = imgui.Vec4{X: 1.0, Y: 0.733, Z: 0.124, W: 0.828}
	colors[imgui.ColButtonActive] = imgui.Vec4{X: 1.0, Y: 0.231, Z: 0.231, W: 1.0}
	colors[imgui.ColHeader] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.455}
	colors[imgui.ColHeaderHovered] = imgui.Vec4{X: 0.18, Y: 0.18, Z: 0.18, W: 0.8}
	colors[imgui.ColHeaderActive] = imgui.Vec4{X: 0.976, Y: 0.259, Z: 0.259, W: 1.0}
	colors[imgui.ColSeparator] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.5}
	colors[imgui.ColSeparatorHovered] = imgui.Vec4{X: 0.098, Y: 0.4, Z: 0.749, W: 0.78}
	colors[imgui.ColSeparatorActive] = imgui.Vec4{X: 0.098, Y: 0.4, Z: 0.749, W: 1.0}
	colors[imgui.ColResizeGrip] = imgui.Vec4{X: 0.259, Y: 0.588, Z: 0.976, W: 0.2}
	colors[imgui.ColResizeGripHovered] = imgui.Vec4{X: 0.259, Y: 0.588, Z: 0.976, W: 0.67}
	colors[imgui.ColResizeGripActive] = imgui.Vec4{X: 0.259, Y: 0.588, Z: 0.976, W: 0.95}
	colors[imgui.ColTab] = imgui.Vec4{X: 0.106, Y: 0.106, Z: 0.106, W: 1.0}
	colors[imgui.ColTabHovered] = imgui.Vec4{X: 1.0, Y: 0.733, Z: 0.124, W: 0.828}
	colors[imgui.ColTabSelected] = imgui.Vec4{X: 1.0, Y: 0.224, Z: 0.224, W: 1.0}
	colors[imgui.ColPlotLines] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 1.0}
	colors[imgui.ColPlotLinesHovered] = imgui.Vec4{X: 1.0, Y: 0.427, Z: 0.349, W: 1.0}
	colors[imgui.ColPlotHistogram] = imgui.Vec4{X: 1.0, Y: 0.216, Z: 0.216, W: 1.0}
	colors[imgui.ColPlotHistogramHovered] = imgui.Vec4{X: 1.0, Y: 0.733, Z: 0.124, W: 0.828}
	colors[imgui.ColTableHeaderBg] = imgui.Vec4{X: 1.0, Y: 0.235, Z: 0.235, W: 1.0}
	colors[imgui.ColTableBorderStrong] = imgui.Vec4{X: 1.0, Y: 0.318, Z: 0.318, W: 1.0}
	colors[imgui.ColTableBorderLight] = imgui.Vec4{X: 1.0, Y: 0.565, Z: 0.565, W: 0.369}
	colors[imgui.ColTableRowBg] = imgui.Vec4{X: 0.725, Y: 0.337, Z: 1.0, W: 0.0}
	colors[imgui.ColTableRowBgAlt] = imgui.Vec4{X: 1.0, Y: 0.275, Z: 0.275, W: 0.112}
	colors[imgui.ColTextSelectedBg] = imgui.Vec4{X: 0.976, Y: 0.259, Z: 0.259, W: 1.0}
	colors[imgui.ColDragDropTarget] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 0.0, W: 0.9}
	colors[imgui.ColNavWindowingHighlight] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 0.468}
	colors[imgui.ColNavWindowingDimBg] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.734}
	colors[imgui.ColModalWindowDimBg] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.798}
	imgui.CurrentStyle().SetColors(&colors)
}

// predefined themes for convenience
var (
	BlueTheme   = NewHueColorScheme("Blue", 240, 50, 180)
	GreenTheme  = NewHueColorScheme("Green", 120, 40, 170)
	RedTheme    = NewHueColorScheme("Red", 0, 45, 175)
	PurpleTheme = NewHueColorScheme("Purple", 270, 35, 165)
	ModernDark  = &ModernTheme{}
)

// SetTheme applies a theme to the current ImGui style
func SetTheme(theme Theme) {
	theme.Apply()
}
