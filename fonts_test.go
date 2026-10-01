package dfx

import (
	"math"
	"runtime"
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx/fonts"
)

func TestSetupFontsMergesMaterialSymbolsAtBothSizes(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ctx := imgui.CreateContext()
	defer imgui.DestroyContextV(ctx)
	previous := Fonts
	defer func() { Fonts = previous }()
	io := imgui.CurrentIO()
	io.SetIniFilename("")
	io.SetDisplaySize(imgui.Vec2{X: 800, Y: 600})
	io.SetDeltaTime(1.0 / 60)
	io.SetBackendFlags(imgui.BackendFlagsRendererHasTextures)
	SetupFonts()
	if len(Fonts) != 3 {
		t.Fatalf("font slots changed: '%d'", len(Fonts))
	}
	imgui.NewFrame()
	for _, slot := range []int{MainFont, SmallFont} {
		for _, scale := range []float32{0.5, 1, 1.5} {
			baked := Fonts[slot].FontBakedV(fontSizes[slot]*scale, 1)
			for _, icon := range []string{fonts.ICON_SYMBOL_VIEW_REAL_SIZE, fonts.ICON_PLAY_ARROW} {
				code := imgui.Wchar([]rune(icon)[0])
				glyph := baked.FindGlyphNoFallback(code)
				ptr, finish := glyph.Handle()
				missing := ptr == nil
				finish()
				if missing {
					t.Fatalf("glyph 'U+%04X' missing in font '%d' at scale '%v'", code, slot, scale)
				}
			}
			real := baked.FindGlyphNoFallback(imgui.Wchar([]rune(fonts.ICON_SYMBOL_VIEW_REAL_SIZE)[0]))
			fit := baked.FindGlyphNoFallback(imgui.Wchar([]rune(fonts.ICON_FIT_SCREEN)[0]))
			if math.Abs(float64(real.AdvanceX()-fit.AdvanceX())) > 1 {
				t.Fatalf("real-size and fit icon cells differ in font '%d' at scale '%v': '%v' vs '%v'", slot, scale, real.AdvanceX(), fit.AdvanceX())
			}
			if width := real.X1() - real.X0(); width < 0.9*(fit.X1()-fit.X0()) || width > 1.2*(fit.X1()-fit.X0()) {
				t.Fatalf("real-size lettering width '%v' does not match fit width '%v' in font '%d' at scale '%v'", width, fit.X1()-fit.X0(), slot, scale)
			}
			if height := real.Y1() - real.Y0(); height < 0.7*(fit.Y1()-fit.Y0()) {
				t.Fatalf("real-size lettering height '%v' is too small alongside fit height '%v' in font '%d' at scale '%v'", height, fit.Y1()-fit.Y0(), slot, scale)
			}
		}
	}
	imgui.EndFrame()
}
