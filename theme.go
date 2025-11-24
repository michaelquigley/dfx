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

// NewHueColorScheme creates a new HSV-based color scheme
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
	colors[imgui.ColTextDisabled] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 0.3991416096687317}
	colors[imgui.ColWindowBg] = imgui.Vec4{X: 0.03921568766236305, Y: 0.03921568766236305, Z: 0.03921568766236305, W: 0.939999997615814}
	colors[imgui.ColChildBg] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.0}
	colors[imgui.ColPopupBg] = imgui.Vec4{X: 0.05098039284348488, Y: 0.05098039284348488, Z: 0.05098039284348488, W: 0.9399999976158142}
	colors[imgui.ColBorder] = imgui.Vec4{X: 0.4274509847164154, Y: 0.4274509847164154, Z: 0.4980392158031464, W: 0.5}
	colors[imgui.ColBorderShadow] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.0}
	colors[imgui.ColFrameBg] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.4206008315086365}
	colors[imgui.ColFrameBgHovered] = imgui.Vec4{X: 0.1411764770746231, Y: 0.1411764770746231, Z: 0.1411764770746231, W: 0.4000000059604645}
	colors[imgui.ColFrameBgActive] = imgui.Vec4{X: 0.2313725501298904, Y: 0.2313725501298904, Z: 0.2313725501298904, W: 0.8626609444618225}
	colors[imgui.ColTitleBg] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0}
	colors[imgui.ColTitleBgActive] = imgui.Vec4{X: 0.09411764889955521, Y: 0.09411764889955521, Z: 0.09411764889955521, W: 1.0}
	colors[imgui.ColTitleBgCollapsed] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.2918455004692078}
	colors[imgui.ColMenuBarBg] = imgui.Vec4{X: 0.1372549086809158, Y: 0.1372549086809158, Z: 0.1372549086809158, W: 1.0}
	colors[imgui.ColScrollbarBg] = imgui.Vec4{X: 0.01960784383118153, Y: 0.01960784383118153, Z: 0.01960784383118153, W: 0.5299999713897705}
	colors[imgui.ColScrollbarGrab] = imgui.Vec4{X: 0.3098039329051971, Y: 0.3098039329051971, Z: 0.3098039329051971, W: 1.0}
	colors[imgui.ColScrollbarGrabHovered] = imgui.Vec4{X: 0.407843142747879, Y: 0.407843142747879, Z: 0.407843142747879, W: 1.0}
	colors[imgui.ColScrollbarGrabActive] = imgui.Vec4{X: 0.5098039507865906, Y: 0.5098039507865906, Z: 0.5098039507865906, W: 1.0}
	colors[imgui.ColCheckMark] = imgui.Vec4{X: 0.9803921580314636, Y: 0.2588235437870026, Z: 0.2588235437870026, W: 1.0}
	colors[imgui.ColSliderGrab] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 1.0}
	colors[imgui.ColSliderGrabActive] = imgui.Vec4{X: 0.9803921580314636, Y: 0.2588235437870026, Z: 0.2588235437870026, W: 1.0}
	colors[imgui.ColButton] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.5793991088867188}
	colors[imgui.ColButtonHovered] = imgui.Vec4{X: 1.0, Y: 0.733, Z: 0.124, W: 0.828}
	colors[imgui.ColButtonActive] = imgui.Vec4{X: 1.0, Y: 0.2313725501298904, Z: 0.2313725501298904, W: 1.0}
	colors[imgui.ColHeader] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.454935610294342}
	colors[imgui.ColHeaderHovered] = imgui.Vec4{X: 0.1803921610116959, Y: 0.1803921610116959, Z: 0.1803921610116959, W: 0.800000011920929}
	colors[imgui.ColHeaderActive] = imgui.Vec4{X: 0.9764705896377563, Y: 0.2588235437870026, Z: 0.2588235437870026, W: 1.0}
	colors[imgui.ColSeparator] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.5}
	colors[imgui.ColSeparatorHovered] = imgui.Vec4{X: 0.09803921729326248, Y: 0.4000000059604645, Z: 0.7490196228027344, W: 0.7799999713897705}
	colors[imgui.ColSeparatorActive] = imgui.Vec4{X: 0.09803921729326248, Y: 0.4000000059604645, Z: 0.7490196228027344, W: 1.0}
	colors[imgui.ColResizeGrip] = imgui.Vec4{X: 0.2588235437870026, Y: 0.5882353186607361, Z: 0.9764705896377563, W: 0.2000000029802322}
	colors[imgui.ColResizeGripHovered] = imgui.Vec4{X: 0.2588235437870026, Y: 0.5882353186607361, Z: 0.9764705896377563, W: 0.6700000166893005}
	colors[imgui.ColResizeGripActive] = imgui.Vec4{X: 0.2588235437870026, Y: 0.5882353186607361, Z: 0.9764705896377563, W: 0.949999988079071}
	colors[imgui.ColTab] = imgui.Vec4{X: 0.105882354080677, Y: 0.105882354080677, Z: 0.105882354080677, W: 1.0}
	colors[imgui.ColTabHovered] = imgui.Vec4{X: 1.0, Y: 0.733, Z: 0.124, W: 0.828}
	colors[imgui.ColTabSelected] = imgui.Vec4{X: 1.0, Y: 0.2235294133424759, Z: 0.2235294133424759, W: 1.0}
	colors[imgui.ColPlotLines] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 1.0}
	colors[imgui.ColPlotLinesHovered] = imgui.Vec4{X: 1.0, Y: 0.4274509847164154, Z: 0.3490196168422699, W: 1.0}
	colors[imgui.ColPlotHistogram] = imgui.Vec4{X: 1.0, Y: 0.2156862765550613, Z: 0.2156862765550613, W: 1.0}
	colors[imgui.ColPlotHistogramHovered] = imgui.Vec4{X: 1.0, Y: 0.733, Z: 0.124, W: 0.828}
	colors[imgui.ColTableHeaderBg] = imgui.Vec4{X: 1.0, Y: 0.2352941185235977, Z: 0.2352941185235977, W: 1.0}
	colors[imgui.ColTableBorderStrong] = imgui.Vec4{X: 1.0, Y: 0.3176470696926117, Z: 0.3176470696926117, W: 1.0}
	colors[imgui.ColTableBorderLight] = imgui.Vec4{X: 1.0, Y: 0.5647059082984924, Z: 0.5647059082984924, W: 0.3690987229347229}
	colors[imgui.ColTableRowBg] = imgui.Vec4{X: 0.7254902124404907, Y: 0.3372549116611481, Z: 1.0, W: 0.0}
	colors[imgui.ColTableRowBgAlt] = imgui.Vec4{X: 1.0, Y: 0.2745098173618317, Z: 0.2745098173618317, W: 0.1115880012512207}
	colors[imgui.ColTextSelectedBg] = imgui.Vec4{X: 0.9764705896377563, Y: 0.2588235437870026, Z: 0.2588235437870026, W: 1.0}
	colors[imgui.ColDragDropTarget] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 0.0, W: 0.8999999761581421}
	colors[imgui.ColNavWindowingHighlight] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 0.4678111672401428}
	colors[imgui.ColNavWindowingDimBg] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.733905553817749}
	colors[imgui.ColModalWindowDimBg] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.7982832789421082}
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
