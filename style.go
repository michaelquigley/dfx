package dfx

import (
	"math"
	"slices"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx/fonts"
	"github.com/pkg/errors"
)

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

// Style is a complete presentation configuration: typography, ImGui metrics, the
// default color theme, and the defaults dfx's own chrome consumes. it is a plain
// value; start from a built-in constructor such as ClassicStyle and change fields
// before passing it to Config.Style. every field is meaningful, including zeros.
// Configuration fields copy by value; FontFace.Data refers to shared, immutable assets.
type Style struct {
	Name       string
	Theme      Theme // default color theme; Config.Theme overrides it, and nil falls back to Modern Dark
	Typography Typography
	Metrics    StyleMetrics
	Spacing    StyleSpacing
	Roles      StyleRoles
	Chrome     StyleChrome
	Controls   StyleControls
}

// Typography names the faces loaded into the font slots. Main, Monospace, and Small
// fill MainFont, MonospaceFont, and SmallFont. Heading is optional: with no Data,
// PushHeadingFont uses the main face.
type Typography struct {
	Main      FontFace
	Monospace FontFace
	Small     FontFace
	Heading   FontFace
}

// FontFace is one loaded font. Data is borrowed, immutable TrueType data shared
// across style copies and read in place by the font atlas. Do not modify its bytes.
// To customize a face, replace Data before startup with another immutable asset;
// finish any edits to caller-owned bytes before supplying them as Data.
type FontFace struct {
	Data       []byte
	Size       float32 // pixel size
	Icons      bool    // merge Material Icons and the supplemental Material Symbols
	IconOffset float32 // vertical offset of merged icon glyphs
}

// StyleMetrics are the ImGui style metrics a style applies at startup.
// WindowPadding is the padding inside panels and ItemSpacing the gap between
// related items.
type StyleMetrics struct {
	WindowPadding imgui.Vec2
	FramePadding  imgui.Vec2
	ItemSpacing   imgui.Vec2
	CellPadding   imgui.Vec2

	ScrollbarSize float32

	WindowBorder float32
	ChildBorder  float32
	PopupBorder  float32
	FrameBorder  float32

	WindowRounding    float32
	ChildRounding     float32
	FrameRounding     float32
	PopupRounding     float32
	ScrollbarRounding float32
	GrabRounding      float32
}

// StyleSpacing is the part of the spacing vocabulary ImGui has no metric for.
type StyleSpacing struct {
	OuterMargin float32 // inset of application content from the window edge
	SectionGap  float32 // separation applications place deliberately between groups
}

// StyleRoles bind dfx's custom chrome to ImGui color slots. colors are read from
// the current ImGui context when drawing, so theme changes and local color pushes
// reach the chrome without rebuilding anything.
type StyleRoles struct {
	Accent        imgui.Col // active values, selection, focus, enabled state; NodeCanvas emphasis
	Panel         imgui.Col // raised panel surfaces
	SecondaryText imgui.Col // quieter labels and annotations
	FaderFill     imgui.Col // fader fill and hovered handle border
	FaderActive   imgui.Col // active fader handle border
	Toolbar       imgui.Col // toolbar background
	Handle        imgui.Col // drag-handle glyphs on Dash and HCollapse panels
}

// StyleChrome holds the presentation defaults of dfx's composite components. their
// public dimensions and margins, such as DashManager.Margin, remain explicit fields.
type StyleChrome struct {
	PanelRounding  float32    // corner radius of Dash and HCollapse panels
	ContentPadding imgui.Vec2 // padding inside HCollapse content and MultiGrid cells
	ToolbarInset   float32    // space before a toolbar's label
	FlexMargin     float32    // FlexLayout slack left at its trailing edges
	FlexGap        float32    // FlexLayout gap between cells, alongside its splitters
	GridGap        float32    // GridLayout gap between cells
}

// StyleControls holds defaults for dfx's own controls.
type StyleControls struct {
	// AccentInk draws labels on accent fills, an on Toggle and NodeCanvas title
	// labels, in whichever of the theme's background and text colors contrasts
	// more with the fill, and keeps an on Toggle's fill through hover and press.
	// off, those labels use the theme's text color and an on Toggle takes the
	// theme's button hover and press colors.
	AccentInk bool
}

// ClassicStyle returns dfx's original compact presentation: Gidole and JetBrains
// Mono, 4 px padding and spacing, and Modern Dark colors. each call returns an
// independent configuration sharing immutable font assets.
func ClassicStyle() Style {
	return Style{
		Name:  "Classic",
		Theme: &ModernTheme{},
		Typography: Typography{
			Main:      FontFace{Data: fonts.GidoleRegular, Size: 20, Icons: true, IconOffset: 5},
			Monospace: FontFace{Data: fonts.JetBrainsMonoMedium, Size: 16},
			Small:     FontFace{Data: fonts.GidoleRegular, Size: 16, Icons: true, IconOffset: 3}, // 4px at 20px -> 3px at 16px
		},
		Metrics: StyleMetrics{
			WindowPadding:     imgui.Vec2{X: DefaultWindowPadding, Y: DefaultWindowPadding},
			FramePadding:      imgui.Vec2{X: DefaultFramePadding, Y: DefaultFramePadding},
			ItemSpacing:       imgui.Vec2{X: DefaultItemSpacing, Y: DefaultItemSpacing},
			CellPadding:       imgui.Vec2{X: 4, Y: 2}, // imgui's default, which classic never changed
			ScrollbarSize:     DefaultScrollbarSize,
			WindowBorder:      DefaultWindowBorder,
			ChildBorder:       DefaultChildBorder,
			PopupBorder:       DefaultPopupBorder,
			FrameBorder:       DefaultFrameBorder,
			WindowRounding:    DefaultWindowRounding,
			ChildRounding:     DefaultChildRounding,
			FrameRounding:     DefaultFrameRounding,
			PopupRounding:     DefaultPopupRounding,
			ScrollbarRounding: DefaultScrollbarRounding,
			GrabRounding:      DefaultGrabRounding,
		},
		Spacing: StyleSpacing{
			OuterMargin: DefaultWindowPadding,
			SectionGap:  2 * DefaultItemSpacing,
		},
		Roles: classicStyleRoles(),
		Chrome: StyleChrome{
			PanelRounding: DashWindowRounding,
			ToolbarInset:  3,
			FlexMargin:    2,
			FlexGap:       4,
		},
	}
}

func classicStyleRoles() StyleRoles {
	return StyleRoles{
		Accent:        imgui.ColHeaderActive,
		Panel:         imgui.ColChildBg,
		SecondaryText: imgui.ColTextDisabled,
		FaderFill:     imgui.ColButtonHovered,
		FaderActive:   imgui.ColButtonActive,
		Toolbar:       imgui.ColHeader,
		Handle:        imgui.ColHeaderActive,
	}
}

// FaceplateStyle returns the presentation shared with the faceplate plugins:
// Chakra Petch with a real bold heading face, charcoal surfaces, one identity
// accent, and roomier spacing. replace Theme with a FaceplateTheme of another
// accent to change the identity color. each call returns an independent
// configuration, with its own theme, sharing immutable font assets.
func FaceplateStyle() Style {
	return Style{
		Name:  "Faceplate",
		Theme: &FaceplateTheme{Accent: faceplateAccent},
		Typography: Typography{
			Main:      FontFace{Data: fonts.ChakraPetchRegular, Size: 20, Icons: true, IconOffset: 5},
			Monospace: FontFace{Data: fonts.JetBrainsMonoMedium, Size: 16},
			Small:     FontFace{Data: fonts.ChakraPetchRegular, Size: 16, Icons: true, IconOffset: 3},
			Heading:   FontFace{Data: fonts.ChakraPetchBold, Size: 20, Icons: true, IconOffset: 5},
		},
		Metrics: StyleMetrics{
			WindowPadding:     imgui.Vec2{X: 12, Y: 12},
			FramePadding:      imgui.Vec2{X: 8, Y: 5},
			ItemSpacing:       imgui.Vec2{X: 8, Y: 8},
			CellPadding:       imgui.Vec2{X: 8, Y: 4},
			ScrollbarSize:     DefaultScrollbarSize,
			WindowBorder:      DefaultWindowBorder,
			ChildBorder:       DefaultChildBorder,
			PopupBorder:       DefaultPopupBorder,
			FrameBorder:       DefaultFrameBorder,
			WindowRounding:    DefaultWindowRounding,
			ChildRounding:     DefaultChildRounding,
			FrameRounding:     DefaultFrameRounding,
			PopupRounding:     DefaultPopupRounding,
			ScrollbarRounding: DefaultScrollbarRounding,
			GrabRounding:      DefaultGrabRounding,
		},
		Spacing: StyleSpacing{
			OuterMargin: 12,
			SectionGap:  16,
		},
		Roles: StyleRoles{
			Accent:        imgui.ColCheckMark,
			Panel:         imgui.ColChildBg,
			SecondaryText: imgui.ColTextDisabled,
			FaderFill:     imgui.ColCheckMark,
			FaderActive:   imgui.ColCheckMark,
			Toolbar:       imgui.ColTitleBgActive,
			Handle:        imgui.ColTextDisabled,
		},
		Chrome: StyleChrome{
			PanelRounding:  DashWindowRounding,
			ContentPadding: imgui.Vec2{X: 12, Y: 12},
			ToolbarInset:   8,
			GridGap:        8, // FlexLayout's 10 px splitter strips already separate its cells
		},
		Controls: StyleControls{AccentInk: true},
	}
}

func currentStyleSpacing() StyleSpacing {
	if activeStyle == nil {
		return ClassicStyle().Spacing
	}
	return activeStyle.style.Spacing
}

// availableSize is the content region left at the cursor, never negative: a
// negative size passed on would read as imgui's fill-relative size syntax.
func availableSize() imgui.Vec2 {
	avail := imgui.ContentRegionAvail()
	return imgui.Vec2{X: max(0, avail.X), Y: max(0, avail.Y)}
}

// usableSize is availableSize, and whether it leaves anything to draw in: after a
// container's padding, a positive child can still have no usable area.
func usableSize() (imgui.Vec2, bool) {
	size := availableSize()
	return size, size.X > 0 && size.Y > 0
}

func currentStyleChrome() StyleChrome {
	if activeStyle == nil {
		return ClassicStyle().Chrome
	}
	return activeStyle.style.Chrome
}

func currentStyleControls() StyleControls {
	if activeStyle == nil {
		return StyleControls{}
	}
	return activeStyle.style.Controls
}

// validate rejects a style whose required font slots could not load.
func (s Style) validate() error {
	faces := []struct {
		name     string
		face     FontFace
		optional bool
	}{
		{"main", s.Typography.Main, false},
		{"monospace", s.Typography.Monospace, false},
		{"small", s.Typography.Small, false},
		{"heading", s.Typography.Heading, true},
	}
	for _, f := range faces {
		if len(f.face.Data) == 0 {
			if f.optional {
				continue
			}
			return errors.Errorf("style '%s': %s font has no data", s.Name, f.name)
		}
		if f.face.Size <= 0 {
			return errors.Errorf("style '%s': %s font size '%v' is not positive", s.Name, f.name, f.face.Size)
		}
	}
	return nil
}

// activeStyleState describes the style selected by the running App. it holds no
// ImGui objects, only the selection.
type activeStyleState struct {
	style       Style
	fontsLoaded bool // dfx font slots belong to this App's context
}

// activeStyle follows the single-active-UI model, like Fonts: App.Run installs it
// during setup and restores the previous value when it exits. nil outside a run.
var activeStyle *activeStyleState

// CurrentStyle returns a copy of the style selected at startup, with Theme set to
// the effective color theme. it describes the selected defaults; live colors and
// metrics, including local overrides, come from the current ImGui context. outside
// a running App it returns ClassicStyle. Font data remains shared and immutable.
// a built-in FaceplateTheme is copied, so editing a snapshot changes nothing else.
func CurrentStyle() Style {
	if activeStyle == nil {
		return ClassicStyle()
	}
	style := activeStyle.style
	style.Theme = ownTheme(style.Theme)
	return style
}

// ownTheme copies a built-in FaceplateTheme, whose accent is configuration held
// behind a pointer: sharing it would let an edit to the source, or to one
// CurrentStyle snapshot, reach the running selection and every other snapshot.
// other Theme implementations keep their existing ownership.
func ownTheme(theme Theme) Theme {
	if faceplate, ok := theme.(*FaceplateTheme); ok && faceplate != nil {
		owned := *faceplate
		return &owned
	}
	return theme
}

func currentStyleRoles() StyleRoles {
	if activeStyle == nil {
		return classicStyleRoles()
	}
	return activeStyle.style.Roles
}

// styleColor reads one color from the current ImGui style without copying the table.
func styleColor(col imgui.Col) imgui.Vec4 {
	return *imgui.StyleColorVec4(col)
}

// saveStyleGlobals captures the active style and font globals, returning a function
// that restores them. App.Run defers it so a later App, or a test, never inherits a
// finished run's metadata or font pointers into a destroyed context.
func saveStyleGlobals() func() {
	style, fonts, sizes := activeStyle, slices.Clone(Fonts), fontSizes
	return func() {
		activeStyle, Fonts, fontSizes = style, fonts, sizes
	}
}

// pxPerFrame calculates pixels to animate per frame for smooth transitions.
func pxPerFrame(targetSize float32, transitionMs int) float32 {
	msFrame := FramerateToMs / imgui.CurrentIO().Framerate()
	frames := float32(transitionMs) / msFrame
	return float32(math.Ceil(float64(targetSize) / float64(frames)))
}

// DefaultStyle applies Classic's metrics to the current ImGui style. App startup
// applies the selected style's metrics itself; this remains for direct callers.
func DefaultStyle() {
	applyStyleMetrics(ClassicStyle().Metrics)
}

func applyStyleMetrics(m StyleMetrics) {
	style := imgui.CurrentStyle()

	// spacing and padding
	style.SetWindowPadding(m.WindowPadding)
	style.SetFramePadding(m.FramePadding)
	style.SetItemSpacing(m.ItemSpacing)
	style.SetCellPadding(m.CellPadding)

	// sizes
	style.SetScrollbarSize(m.ScrollbarSize)

	// borders
	style.SetWindowBorderSize(m.WindowBorder)
	style.SetChildBorderSize(m.ChildBorder)
	style.SetPopupBorderSize(m.PopupBorder)
	style.SetFrameBorderSize(m.FrameBorder)

	// rounding
	style.SetWindowRounding(m.WindowRounding)
	style.SetChildRounding(m.ChildRounding)
	style.SetFrameRounding(m.FrameRounding)
	style.SetPopupRounding(m.PopupRounding)
	style.SetScrollbarRounding(m.ScrollbarRounding)
	style.SetGrabRounding(m.GrabRounding)
}
