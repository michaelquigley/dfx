package dfx

import (
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx/fonts"
)

func TestFaceplateStyle_IndependentValuesAndAccents(t *testing.T) {
	first, second := FaceplateStyle(), FaceplateStyle()
	first.Theme.(*FaceplateTheme).Accent = rgb(0x4da3ff)
	first.Spacing.SectionGap = 40
	first.Metrics.ItemSpacing = imgui.Vec2{X: 2, Y: 2}
	if second.Theme.(*FaceplateTheme).Accent != faceplateAccent || second.Spacing.SectionGap != 16 || second.Metrics.ItemSpacing != (imgui.Vec2{X: 8, Y: 8}) {
		t.Fatal("customizing one Faceplate value changed another")
	}
	if FaceplateStyle().Theme.(*FaceplateTheme).Accent != faceplateAccent {
		t.Fatal("customizing a Faceplate value changed later presets")
	}

	// two accents share every neutral; only accent-derived slots differ
	styleTestContext(t)
	SetTheme(first.Theme)
	blue := imgui.CurrentStyle().Colors()
	SetTheme(second.Theme)
	yellow := imgui.CurrentStyle().Colors()
	for _, col := range []imgui.Col{imgui.ColText, imgui.ColTextDisabled, imgui.ColWindowBg, imgui.ColChildBg, imgui.ColButton, imgui.ColButtonHovered, imgui.ColFrameBg, imgui.ColBorder} {
		if blue[col] != yellow[col] {
			t.Fatalf("neutral '%s' depends on the accent", imgui.StyleColorName(col))
		}
	}
	if blue[imgui.ColCheckMark] == yellow[imgui.ColCheckMark] {
		t.Fatal("the accent did not reach the check mark")
	}
	// hover stays neutral rather than becoming the accent
	if yellow[imgui.ColButtonHovered] == yellow[imgui.ColCheckMark] || yellow[imgui.ColHeaderHovered] == yellow[imgui.ColCheckMark] {
		t.Fatal("hover must not be a solid accent")
	}
}

func TestFaceplate_GlyphsAtUISizesAndCanvasDetents(t *testing.T) {
	styleTestContext(t)
	style := FaceplateStyle()
	New(nil, Config{Style: &style}).setupFontsAndTheme()
	if len(Fonts) != 4 {
		t.Fatalf("expected four font slots, got '%d'", len(Fonts))
	}
	imgui.NewFrame()
	defer imgui.EndFrame()
	detents := []float32{0.25, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1, 1.2, 1.3, 1.4, 1.5}
	runes := []rune{'A', 'g', '0', []rune(fonts.ICON_PLAY_ARROW)[0], []rune(fonts.ICON_DRAG_INDICATOR)[0], []rune(fonts.ICON_SYMBOL_VIEW_REAL_SIZE)[0]}
	for _, slot := range []int{MainFont, SmallFont, headingFont} {
		for _, scale := range detents {
			baked := Fonts[slot].FontBakedV(fontSizes[slot]*scale, 1)
			for _, r := range runes {
				glyph := baked.FindGlyphNoFallback(imgui.Wchar(r))
				ptr, finish := glyph.Handle()
				missing := ptr == nil
				finish()
				if missing {
					t.Fatalf("glyph 'U+%04X' missing in slot '%d' at scale '%v'", r, slot, scale)
				}
			}
		}
	}
	// the heading is the real bold face: wider strokes than regular at one size
	regular := Fonts[MainFont].FontBakedV(20, 1).FindGlyphNoFallback('W').AdvanceX()
	bold := Fonts[headingFont].FontBakedV(20, 1).FindGlyphNoFallback('W').AdvanceX()
	if bold <= regular {
		t.Fatalf("heading 'W' advance '%v' is not wider than regular '%v'", bold, regular)
	}
}

func TestFaceplate_ThemeChangesKeepMetricsFontsAndExplicitColors(t *testing.T) {
	styleTestContext(t)
	style := FaceplateStyle()
	New(nil, Config{Style: &style}).setupFontsAndTheme()
	fontsBefore := append([]*imgui.Font(nil), Fonts...)
	padding, spacing := imgui.CurrentStyle().FramePadding(), imgui.CurrentStyle().ItemSpacing()

	imgui.NewFrame()
	explicit := DefaultNodeCanvasStyle()
	explicit.LinkColor = imgui.Vec4{X: 0.1, Y: 0.9, Z: 0.3, W: 1}
	canvas := NewNodeCanvas[int](NodeCanvasConfig{Style: explicit})
	imgui.EndFrame()

	for _, theme := range []Theme{BlueTheme, ModernDark, &FaceplateTheme{Accent: rgb(0xff5c8a)}} {
		SetTheme(theme)
		if imgui.CurrentStyle().FramePadding() != padding || imgui.CurrentStyle().ItemSpacing() != spacing {
			t.Fatalf("%s changed metrics", theme.Name())
		}
		for i := range Fonts {
			if !sameFont(Fonts[i], fontsBefore[i]) {
				t.Fatalf("%s changed font slot '%d'", theme.Name(), i)
			}
		}
		if canvas.style.LinkColor != explicit.LinkColor {
			t.Fatalf("%s changed an explicit canvas style", theme.Name())
		}
		accent := imgui.Vec4{X: 0.85, Y: 0.1, Z: 0.65, W: 0.7}
		params := DefaultFaderParams()
		params.AccentColor = &accent
		if palette := resolveFaderPalette(params, true, true); palette.fill != accent || palette.border != accent {
			t.Fatalf("%s overrode an explicit fader accent", theme.Name())
		}
	}
}

func TestFaceplate_FaderAndCanvasFollowAccent(t *testing.T) {
	styleTestContext(t)
	style := FaceplateStyle()
	New(nil, Config{Style: &style}).setupFontsAndTheme()
	imgui.NewFrame()
	defer imgui.EndFrame()
	accent := styleColor(imgui.ColCheckMark)
	palette := resolveFaderPalette(DefaultFaderParams(), false, true)
	if palette.fill != accent || palette.border != accent {
		t.Fatalf("Faceplate faders must fill and press in the accent: '%v'", palette)
	}
	if canvas := DefaultNodeCanvasStyle(); canvas.NodeBorderColorSelected != accent {
		t.Fatalf("Faceplate canvas selection must use the accent: '%v'", canvas.NodeBorderColorSelected)
	}
}

// toggleColors draws an on toggle in a fresh frame, with the mouse over it when
// hovered and held down when pressed, and returns its frame and label colors.
func toggleColors(t *testing.T, opts ToggleOptions, hovered, pressed bool) (imgui.Vec4, imgui.Vec4, imgui.Vec4) {
	t.Helper()
	io := imgui.CurrentIO()
	var bg, text, window imgui.Vec4
	var rect imgui.Rect
	frame := func(capture bool) {
		imgui.NewFrame()
		imgui.SetNextWindowPos(imgui.Vec2{})
		imgui.SetNextWindowSize(imgui.Vec2{X: 300, Y: 200})
		imgui.BeginV("toggle", nil, imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)
		dl := imgui.WindowDrawList()
		start := dl.VtxBuffer().Size
		ToggleV("On##toggle", true, opts)
		rect = imgui.Rect{Min: imgui.ItemRectMin(), Max: imgui.ItemRectMax()}
		if capture {
			colors := faderVertexColors(dl, start)
			bg, text = imgui.ColorConvertU32ToFloat4(colors[0]), imgui.ColorConvertU32ToFloat4(colors[len(colors)-1])
			window = styleColor(imgui.ColWindowBg)
		}
		imgui.End()
		imgui.Render()
	}
	io.AddMousePosEvent(-100, -100)
	frame(false)
	if hovered || pressed {
		io.AddMousePosEvent((rect.Min.X+rect.Max.X)/2, (rect.Min.Y+rect.Max.Y)/2)
		frame(false)
	}
	if pressed {
		io.AddMouseButtonEvent(0, true)
		frame(false)
	}
	frame(true)
	if pressed {
		io.AddMouseButtonEvent(0, false)
		frame(false)
	}
	return bg, text, window
}

// contrastOver is the WCAG ratio of text over a possibly translucent fill,
// composited over the window background.
func contrastOver(fill, text, window imgui.Vec4) float64 {
	solid := mix(window, imgui.Vec4{X: fill.X, Y: fill.Y, Z: fill.Z, W: 1}, fill.W)
	a, b := relativeLuminance(solid), relativeLuminance(text)
	return (max(a, b) + 0.05) / (min(a, b) + 0.05)
}

// under Faceplate an on toggle keeps a readable label at rest, hovered, and
// pressed: at least 4.5:1 on Faceplate accents, and never worse than at rest
// after a live switch to a legacy theme, whose own text-on-main contrast is lower.
func TestFaceplate_ToggleForegroundContrast(t *testing.T) {
	styleTestContext(t)
	style := FaceplateStyle()
	New(nil, Config{Style: &style}).setupFontsAndTheme()
	states := []struct {
		name             string
		hovered, pressed bool
	}{{"rest", false, false}, {"hovered", true, false}, {"pressed", true, true}}
	themes := []Theme{style.Theme, &FaceplateTheme{Accent: rgb(0x4da3ff)}, BlueTheme, GreenTheme, RedTheme, PurpleTheme, ModernDark}
	for i, theme := range themes {
		SetTheme(theme)
		var rest float64
		for _, state := range states {
			bg, text, window := toggleColors(t, ToggleOptions{}, state.hovered, state.pressed)
			ratio := contrastOver(bg, text, window)
			if state.name == "rest" {
				rest = ratio
			}
			floor := rest - 0.01
			if i < 2 {
				floor = 4.5 // Faceplate accents
			} else if state.name == "rest" {
				floor = 3
			}
			if ratio < floor {
				t.Fatalf("%s %s toggle contrast '%.2f' below '%.2f'", theme.Name(), state.name, ratio, floor)
			}
		}
	}

	// an explicit foreground wins in every state
	SetTheme(style.Theme)
	override := imgui.Vec4{X: 0.2, Y: 0.1, Z: 0.6, W: 1}
	for _, state := range states {
		if _, text, _ := toggleColors(t, ToggleOptions{OnText: &override}, state.hovered, state.pressed); !nearColor(text, override) {
			t.Fatalf("%s: explicit foreground '%v' drew as '%v'", state.name, override, text)
		}
	}
}

// Classic keeps its toggle treatment: the check mark fill at rest, the theme's
// button hover and press colors, and the theme's text.
func TestClassic_ToggleTreatmentUnchanged(t *testing.T) {
	styleTestContext(t)
	New(nil, Config{}).setupFontsAndTheme()
	expected := map[string]imgui.Col{"rest": imgui.ColCheckMark, "hovered": imgui.ColButtonHovered, "pressed": imgui.ColButtonActive}
	for name, state := range map[string][2]bool{"rest": {false, false}, "hovered": {true, false}, "pressed": {true, true}} {
		bg, text, _ := toggleColors(t, ToggleOptions{}, state[0], state[1])
		if !nearColor(bg, styleColor(expected[name])) || !nearColor(text, styleColor(imgui.ColText)) {
			t.Fatalf("%s: Classic toggle drew '%v' on '%v'", name, text, bg)
		}
	}
}

// nearColor compares colors at the 8-bit precision they are drawn with.
func nearColor(a, b imgui.Vec4) bool {
	return imgui.ColorConvertFloat4ToU32(a) == imgui.ColorConvertFloat4ToU32(b)
}

// title labels on a node's band: under Faceplate they contrast with the band,
// selected yellow, normal raised, or a declared accent; Classic keeps the
// theme's text color.
func TestNodeCanvas_TitleInkFollowsStyle(t *testing.T) {
	for name, preset := range presets {
		t.Run(name, func(t *testing.T) {
			styleTestContext(t)
			style := preset()
			New(nil, Config{Style: &style}).setupFontsAndTheme()
			nc := NewNodeCanvas[string](NodeCanvasConfig{})
			defer nc.Destroy()
			pale := imgui.Vec4{X: 0.85, Y: 0.95, Z: 0.7, W: 1}
			flags := map[string]NodeFlags{"selected": {Selected: true}, "normal": {}, "accent": {Accent: pale}}
			inks := map[string]uint32{}
			imgui.NewFrame()
			imgui.SetNextWindowPos(imgui.Vec2{})
			imgui.SetNextWindowSize(imgui.Vec2{X: 800, Y: 600})
			imgui.BeginV("canvas", nil, imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)
			nc.Begin(&State{Size: imgui.Vec2{X: 700, Y: 500}})
			for i, id := range []string{"selected", "normal", "accent"} {
				nc.Node(id, imgui.Vec2{X: float32(40 + 200*i), Y: 60}, flags[id], func(n *NodeContext[string]) {
					n.TitleBar(func() {
						inks[id] = n.labelColor()
						n.Label(id)
					})
					n.Label("body")
				})
			}
			nc.End()
			text := imgui.ColorU32Col(imgui.ColText)
			canvas := nc.style
			imgui.End()
			imgui.EndFrame()

			if name == "classic" {
				for id, ink := range inks {
					if ink != text {
						t.Fatalf("Classic %s title must use the theme's text color", id)
					}
				}
				return
			}
			for id, band := range map[string]imgui.Vec4{"selected": canvas.TitleBandColorSelected, "normal": canvas.TitleBandColor, "accent": pale} {
				ink := imgui.ColorConvertU32ToFloat4(inks[id])
				if ratio := contrastOver(band, ink, canvas.NodeBodyColor); ratio < 4.5 {
					t.Fatalf("Faceplate %s title contrast '%.2f' below 4.5", id, ratio)
				}
			}
			if inks["selected"] == text || inks["normal"] != text {
				t.Fatal("Faceplate titles: dark ink on the accent band, the theme's text on the raised band")
			}
		})
	}
}

// TestFaceplateTheme_StartupOwnsItsCopy checks that the running selection owns a
// copy of a FaceplateTheme from either startup source, so editing the source after
// startup changes neither CurrentStyle nor the next snapshot, while font bytes
// stay shared.
func TestFaceplateTheme_StartupOwnsItsCopy(t *testing.T) {
	for _, source := range []string{"style", "config theme"} {
		t.Run(source, func(t *testing.T) {
			accent := rgb(0x4da3ff)
			style := FaceplateStyle()
			theme := &FaceplateTheme{Accent: accent}
			config := Config{Style: &style}
			if source == "style" {
				style.Theme = theme
			} else {
				config.Theme = theme
			}
			var applied, current imgui.Vec4
			var owned Theme
			sharedFont := false
			config.OnSetup = func(*App) {
				applied = styleColor(imgui.ColCheckMark)
				theme.Accent = rgb(0xff5c8a) // the caller edits its source after startup
				snapshot := CurrentStyle()
				owned = snapshot.Theme
				current = snapshot.Theme.(*FaceplateTheme).Accent
				sharedFont = &snapshot.Typography.Main.Data[0] == &fonts.ChakraPetchRegular[0]
			}
			if err := runHeadless(t, New(nil, config), &headlessBackend{}); err != nil {
				t.Fatal(err)
			}
			if applied != accent {
				t.Fatalf("startup applied accent '%v', expected '%v'", applied, accent)
			}
			if owned == Theme(theme) {
				t.Fatal("the running selection shares the caller's FaceplateTheme")
			}
			if current != accent {
				t.Fatalf("editing the source theme changed the running selection's accent to '%v'", current)
			}
			if !sharedFont {
				t.Fatal("CurrentStyle cloned immutable font bytes")
			}
		})
	}
}

// TestFaceplateTheme_SnapshotsAreIndependent checks that each CurrentStyle snapshot
// owns its FaceplateTheme, and that custom themes keep their existing ownership.
func TestFaceplateTheme_SnapshotsAreIndependent(t *testing.T) {
	style := FaceplateStyle()
	app := New(nil, Config{Style: &style, OnSetup: func(*App) {
		first, second := CurrentStyle(), CurrentStyle()
		if first.Theme == second.Theme {
			t.Fatal("two snapshots share one FaceplateTheme")
		}
		first.Theme.(*FaceplateTheme).Accent = rgb(0x4da3ff)
		if got := second.Theme.(*FaceplateTheme).Accent; got != faceplateAccent {
			t.Errorf("editing one snapshot changed another to '%v'", got)
		}
		if got := CurrentStyle().Theme.(*FaceplateTheme).Accent; got != faceplateAccent {
			t.Errorf("editing a snapshot changed the running selection to '%v'", got)
		}
	}})
	if err := runHeadless(t, app, &headlessBackend{}); err != nil {
		t.Fatal(err)
	}

	custom := &spyTheme{name: "custom", accent: imgui.Vec4{X: 0.1, W: 1}}
	app = New(nil, Config{Theme: custom, OnSetup: func(*App) {
		if CurrentStyle().Theme != Theme(custom) {
			t.Error("a custom theme must keep its existing ownership, not be copied")
		}
	}})
	if err := runHeadless(t, app, &headlessBackend{}); err != nil {
		t.Fatal(err)
	}
}
