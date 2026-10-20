package dfx

import (
	"math"
	"runtime"
	"slices"
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
)

// styleTestContext gives a test a fresh imgui context able to build fonts, and
// restores the style metadata and font globals that setup replaces.
func styleTestContext(t *testing.T) {
	t.Helper()
	runtime.LockOSThread()
	ctx := imgui.CreateContext()
	previousStyle, previousFonts, previousSizes := activeStyle, Fonts, fontSizes
	io := imgui.CurrentIO()
	io.SetIniFilename("")
	io.SetDisplaySize(imgui.Vec2{X: 800, Y: 600})
	io.SetDeltaTime(1.0 / 60)
	io.SetBackendFlags(imgui.BackendFlagsRendererHasTextures)
	t.Cleanup(func() {
		activeStyle, Fonts, fontSizes = previousStyle, previousFonts, previousSizes
		imgui.DestroyContextV(ctx)
		runtime.UnlockOSThread()
	})
}

// presentation fingerprints what startup applied: the selection, the metrics,
// every color slot, and each font slot's size and glyph advances.
type presentation struct {
	name                     string
	spacing                  StyleSpacing
	roles                    StyleRoles
	chrome                   StyleChrome
	controls                 StyleControls
	padding, frame, spacing2 imgui.Vec2
	colors                   [imgui.ColCOUNT]imgui.Vec4
	sizes                    []float32
	advances                 []float32
}

// startupPresentation runs startup in its own subtest: imgui's CreateContext keeps
// an existing context current, so contexts created within one test would share it.
func startupPresentation(t *testing.T, name string, config Config) presentation {
	t.Helper()
	var p presentation
	t.Run(name, func(t *testing.T) {
		styleTestContext(t)
		New(nil, config).setupFontsAndTheme()
		current, style := CurrentStyle(), imgui.CurrentStyle()
		p = presentation{
			name: current.Name, spacing: current.Spacing, roles: current.Roles, chrome: current.Chrome, controls: current.Controls,
			padding: style.WindowPadding(), frame: style.FramePadding(), spacing2: style.ItemSpacing(),
			colors: style.Colors(), sizes: append([]float32(nil), fontSizes...),
		}
		imgui.NewFrame()
		defer imgui.EndFrame()
		for slot, font := range Fonts {
			baked := font.FontBakedV(fontSizes[slot], 1)
			p.advances = append(p.advances, baked.FindGlyphNoFallback('W').AdvanceX(), baked.FindGlyphNoFallback('i').AdvanceX())
		}
	})
	return p
}

func samePresentation(a, b presentation) bool {
	return a.name == b.name && a.spacing == b.spacing && a.roles == b.roles && a.chrome == b.chrome && a.controls == b.controls &&
		a.padding == b.padding && a.frame == b.frame && a.spacing2 == b.spacing2 && a.colors == b.colors &&
		slices.Equal(a.sizes, b.sizes) && slices.Equal(a.advances, b.advances)
}

// with no style, startup selects Faceplate: the same presentation as selecting it.
func TestDefaultStartup_IsFaceplate(t *testing.T) {
	faceplate := FaceplateStyle()
	implicit := startupPresentation(t, "implicit", Config{})
	explicit := startupPresentation(t, "explicit", Config{Style: &faceplate})
	if implicit.name != "Faceplate" || !samePresentation(implicit, explicit) {
		t.Fatalf("nil Config.Style must select Faceplate, got '%s'", implicit.name)
	}
	if len(implicit.sizes) != 4 {
		t.Fatalf("the default must load Faceplate's four font slots, got '%d'", len(implicit.sizes))
	}
}

// a color theme overrides only the colors of whichever style is selected: Modern
// Dark alone does not select Classic, and Classic with another theme keeps its
// typography and metrics.
func TestStartup_ColorThemeOverridesKeepTheSelectedStyle(t *testing.T) {
	classic, faceplate := ClassicStyle(), FaceplateStyle()
	for _, c := range []struct {
		name     string
		config   Config
		base     Config
		colorsOf Config
	}{
		{"default with ModernDark", Config{Theme: ModernDark}, Config{Style: &faceplate}, Config{Style: &classic}},
		{"Classic with BlueTheme", Config{Style: &classic, Theme: BlueTheme}, Config{Style: &classic}, Config{Theme: BlueTheme}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, base, colors := startupPresentation(t, "selected", c.config), startupPresentation(t, "style", c.base), startupPresentation(t, "colors", c.colorsOf)
			if got.name != base.name || got.frame != base.frame || got.padding != base.padding || got.chrome != base.chrome ||
				!slices.Equal(got.sizes, base.sizes) || !slices.Equal(got.advances, base.advances) {
				t.Fatalf("the theme override changed the selected style: '%s' instead of '%s'", got.name, base.name)
			}
			if got.colors != colors.colors {
				t.Fatal("the theme override did not supply the colors")
			}
		})
	}
}

func TestClassicStartupBaseline_Explicit(t *testing.T) {
	styleTestContext(t)
	classic := ClassicStyle()
	New(nil, Config{Style: &classic}).setupFontsAndTheme()
	assertClassicBaseline(t)
}

func TestClassicStartupBaseline_LegacyHelpers(t *testing.T) {
	styleTestContext(t)
	SetupFonts()
	DefaultStyle()
	SetTheme(ModernDark)
	assertClassicBaseline(t)
}

// these literals were captured from the pre-complete-style startup on a fresh
// context; Classic must keep reproducing them.
func assertClassicBaseline(t *testing.T) {
	t.Helper()
	style := imgui.CurrentStyle()

	vec2s := map[string][2]imgui.Vec2{
		"window padding": {style.WindowPadding(), {X: 4, Y: 4}},
		"frame padding":  {style.FramePadding(), {X: 4, Y: 4}},
		"item spacing":   {style.ItemSpacing(), {X: 4, Y: 4}},
		"cell padding":   {style.CellPadding(), {X: 4, Y: 2}},
	}
	for name, pair := range vec2s {
		if pair[0] != pair[1] {
			t.Fatalf("%s: expected '%v', got '%v'", name, pair[1], pair[0])
		}
	}
	floats := map[string][2]float32{
		"scrollbar size":     {style.ScrollbarSize(), 12},
		"window border":      {style.WindowBorderSize(), 0},
		"child border":       {style.ChildBorderSize(), 0},
		"popup border":       {style.PopupBorderSize(), 1},
		"frame border":       {style.FrameBorderSize(), 0},
		"window rounding":    {style.WindowRounding(), 3},
		"child rounding":     {style.ChildRounding(), 0},
		"frame rounding":     {style.FrameRounding(), 3},
		"popup rounding":     {style.PopupRounding(), 3},
		"scrollbar rounding": {style.ScrollbarRounding(), 2},
		"grab rounding":      {style.GrabRounding(), 2},
	}
	for name, pair := range floats {
		if pair[0] != pair[1] {
			t.Fatalf("%s: expected '%v', got '%v'", name, pair[1], pair[0])
		}
	}

	colors := style.Colors()
	slots := map[imgui.Col]imgui.Vec4{
		imgui.ColText:             {X: 1, Y: 1, Z: 1, W: 1},
		imgui.ColWindowBg:         {X: 0.039, Y: 0.039, Z: 0.039, W: 0.94},
		imgui.ColFrameBg:          {X: 0, Y: 0, Z: 0, W: 0.421},
		imgui.ColCheckMark:        {X: 0.98, Y: 0.259, Z: 0.259, W: 1},
		imgui.ColButtonHovered:    {X: 1, Y: 0.733, Z: 0.124, W: 0.828},
		imgui.ColButtonActive:     {X: 1, Y: 0.231, Z: 0.231, W: 1},
		imgui.ColTableHeaderBg:    {X: 1, Y: 0.235, Z: 0.235, W: 1},
		imgui.ColModalWindowDimBg: {X: 0, Y: 0, Z: 0, W: 0.798},
		// slots Modern Dark leaves at imgui's dark defaults
		imgui.ColTabSelectedOverline: {X: 0.26, Y: 0.59, Z: 0.98, W: 1},
		imgui.ColNavCursor:           {X: 0.26, Y: 0.59, Z: 0.98, W: 1},
		imgui.ColTreeLines:           {X: 0.43, Y: 0.43, Z: 0.5, W: 0.5},
	}
	for col, expected := range slots {
		if colors[col] != expected {
			t.Fatalf("color '%s': expected '%v', got '%v'", imgui.StyleColorName(col), expected, colors[col])
		}
	}

	if len(Fonts) != 3 {
		t.Fatalf("expected three font slots, got '%d'", len(Fonts))
	}
	// glyph advances fingerprint the family and size in each slot: proportional
	// Gidole at 20 and 16, monospace JetBrains Mono at 16.
	imgui.NewFrame()
	defer imgui.EndFrame()
	fingerprints := []struct {
		slot  int
		size  float32
		w, i  float32
		label string
	}{
		{MainFont, 20, 11.904762, 3.730159, "main"},
		{MonospaceFont, 16, 7.272727, 7.272727, "monospace"},
		{SmallFont, 16, 9.523809, 2.984127, "small"},
	}
	for _, f := range fingerprints {
		if fontSizes[f.slot] != f.size {
			t.Fatalf("%s font size: expected '%v', got '%v'", f.label, f.size, fontSizes[f.slot])
		}
		baked := Fonts[f.slot].FontBakedV(f.size, 1)
		w, i := baked.FindGlyphNoFallback('W').AdvanceX(), baked.FindGlyphNoFallback('i').AdvanceX()
		if math.Abs(float64(w-f.w)) > 0.001 || math.Abs(float64(i-f.i)) > 0.001 {
			t.Fatalf("%s font fingerprint: expected W/i '%v/%v', got '%v/%v'", f.label, f.w, f.i, w, i)
		}
	}
}

func TestClassicStyle_ReturnsIndependentValues(t *testing.T) {
	first := ClassicStyle()
	first.Name = "Mine"
	first.Theme = GreenTheme
	first.Typography.Main.Size = 30
	first.Metrics.FramePadding = imgui.Vec2{X: 10, Y: 10}
	first.Spacing.OuterMargin = 40
	first.Roles.FaderFill = imgui.ColCheckMark

	second := ClassicStyle()
	if second.Name != "Classic" || second.Typography.Main.Size != 20 || second.Metrics.FramePadding != (imgui.Vec2{X: 4, Y: 4}) ||
		second.Spacing.OuterMargin != DefaultWindowPadding || second.Roles.FaderFill != imgui.ColButtonHovered {
		t.Fatal("customizing one Classic value changed another")
	}
	if _, ok := second.Theme.(*ModernTheme); !ok {
		t.Fatal("Classic's default theme must be Modern Dark")
	}
	if CurrentStyle().Name != "Classic" {
		t.Fatal("outside a run, CurrentStyle must describe Classic")
	}
}

func TestModernDark_IndependentOfPreviousTheme(t *testing.T) {
	styleTestContext(t)
	fresh := imgui.CurrentStyle().Colors()
	SetTheme(ModernDark)
	expected := imgui.CurrentStyle().Colors()
	for _, theme := range []Theme{GreenTheme, &spyTheme{accent: imgui.Vec4{X: 0.5, W: 1}}} {
		SetTheme(theme)
		// scribble on a slot Modern Dark leaves unset, as a custom theme may
		colors := imgui.CurrentStyle().Colors()
		colors[imgui.ColNavCursor] = imgui.Vec4{X: 0.3, Y: 0.3, Z: 0.3, W: 0.3}
		imgui.CurrentStyle().SetColors(&colors)
		SetTheme(ModernDark)
		if got := imgui.CurrentStyle().Colors(); got != expected {
			for i := range got {
				if got[i] != expected[i] {
					t.Fatalf("Modern Dark after '%s' differs at '%s': '%v' vs '%v'", theme.Name(), imgui.StyleColorName(imgui.Col(i)), got[i], expected[i])
				}
			}
		}
	}
	// the baseline is imgui's dark defaults, which a fresh context already holds
	if expected[imgui.ColNavCursor] != fresh[imgui.ColNavCursor] || expected[imgui.ColTextLink] != fresh[imgui.ColTextLink] {
		t.Fatal("Modern Dark recolored a slot it leaves unset")
	}
}

func TestPushHeadingFont_OutsideRun(t *testing.T) {
	styleTestContext(t)
	frame := func(check func()) {
		imgui.NewFrame()
		imgui.Begin("heading")
		check()
		imgui.End()
		imgui.EndFrame()
	}

	// no dfx fonts: the current font, balanced with PopFont
	Fonts = nil
	frame(func() {
		before := imgui.CurrentFont()
		PushHeadingFont()
		if !sameFont(imgui.CurrentFont(), before) {
			t.Fatal("expected the current font without dfx fonts")
		}
		PopFont()
	})

	// classic fonts: the main font
	SetupFonts()
	frame(func() {
		PushHeadingFont()
		if !sameFont(imgui.CurrentFont(), Fonts[MainFont]) || imgui.FontSize() != 20 {
			t.Fatal("expected Classic headings in the main font")
		}
		PopFont()
	})

	// a heading face: its own slot and size, after the three fixed slots
	typography := ClassicStyle().Typography
	typography.Heading = FontFace{Data: typography.Main.Data, Size: 26, Icons: true, IconOffset: 6}
	setupFonts(typography)
	if len(Fonts) != 4 || fontSizes[headingFont] != 26 {
		t.Fatalf("expected a fourth heading slot, got '%d' sizes '%v'", len(Fonts), fontSizes)
	}
	frame(func() {
		PushHeadingFont()
		if !sameFont(imgui.CurrentFont(), Fonts[headingFont]) || imgui.FontSize() != 26 {
			t.Fatal("expected the heading face")
		}
		PopFont()
	})
}
