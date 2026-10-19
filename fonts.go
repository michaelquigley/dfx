package dfx

import (
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx/fonts"
)

// font indices for easy access
// note: icon fonts are merged into their base fonts, not separate entries
const (
	MainFont      = 0 // default font (Gidole Regular, 20px) with icons and supplemental symbols merged
	MonospaceFont = 1 // monospace font (JetBrains Mono, 16px)
	SmallFont     = 2 // small font (Gidole Regular, 16px) with icons and supplemental symbols merged
)

// headingFont is the slot a style's optional heading face is appended to; reach it
// through PushHeadingFont.
const headingFont = SmallFont + 1

var Fonts []*imgui.Font

// SetupFonts loads Classic's fonts into the current atlas, replacing any existing
// fonts. App startup loads the selected style's typography itself.
func SetupFonts() {
	setupFonts(ClassicStyle().Typography)
}

// setupFonts loads a typography into the font slots, merging icons into each face
// that asks for them, and records the slot sizes PushFont uses.
func setupFonts(t Typography) {
	// clear any existing fonts
	imgui.CurrentIO().Fonts().Clear()
	Fonts = Fonts[:0] // clear slice but keep capacity

	// build glyph ranges for material icons (shared by every face that merges them)
	builder := imgui.NewFontGlyphRangesBuilder()
	ranges := []imgui.Wchar{'', '！', 0}
	builder.AddRanges(&ranges[0])
	glyphRanges := imgui.NewGlyphRange()
	builder.BuildRanges(glyphRanges)

	sizes := make([]float32, 0, headingFont+1)
	faces := []FontFace{t.Main, t.Monospace, t.Small}
	if len(t.Heading.Data) > 0 {
		faces = append(faces, t.Heading)
	}
	for _, face := range faces {
		Fonts = append(Fonts, imgui.CurrentIO().Fonts().AddFont(fontConfig(face.Data, face.Size, 0, nil)))
		if face.Icons {
			mergeMaterialIcons(face.Size, face.IconOffset, glyphRanges.Data())
		}
		sizes = append(sizes, face.Size)
	}
	fontSizes = sizes
	if activeStyle != nil {
		activeStyle.fontsLoaded = true
	}
}

func fontConfig(data []byte, size, offset float32, ranges *imgui.Wchar) *imgui.FontConfig {
	config := imgui.NewFontConfig()
	config.SetFontData(uintptr(unsafe.Pointer(&data[0])))
	config.SetFontDataSize(int32(len(data)))
	config.SetFontDataOwnedByAtlas(false)
	config.SetSizePixels(size)
	if ranges != nil {
		config.SetGlyphOffset(imgui.Vec2{Y: offset})
		config.SetGlyphRanges(ranges)
		config.SetMergeMode(true) // merge with the preceding font
	}
	return config
}

// mergeMaterialIcons merges Material Icons into the preceding font, then the
// fixed-style supplemental symbols. the supplement's cmap excludes all codepoints in
// Material Icons, so existing icons retain their original artwork.
func mergeMaterialIcons(size, offset float32, ranges *imgui.Wchar) {
	imgui.CurrentIO().Fonts().AddFont(fontConfig(fonts.MaterialIconsRegular, size, offset, ranges))
	imgui.CurrentIO().Fonts().AddFont(fontConfig(fonts.MaterialSymbolsSupplement, size, offset, ranges))
}

// font sizes corresponding to each font index
var fontSizes = []float32{20.0, 16.0, 16.0}

// PushFont convenience function for temporarily switching fonts.
// passes the font's configured size to ensure correct rendering.
func PushFont(fontIndex int) {
	if fontIndex >= 0 && fontIndex < len(Fonts) {
		size := float32(0)
		if fontIndex < len(fontSizes) {
			size = fontSizes[fontIndex]
		}
		imgui.PushFont(Fonts[fontIndex], size)
	}
}

// PushHeadingFont switches to the style's heading face, or the main font when the
// style has none. when dfx fonts were not loaded it pushes the current font, so it
// always pairs with PopFont.
func PushHeadingFont() {
	loaded := len(Fonts) > MainFont
	if activeStyle != nil {
		loaded = activeStyle.fontsLoaded
	}
	switch {
	case !loaded:
		imgui.PushFont(imgui.CurrentFont(), 0)
	case len(Fonts) > headingFont:
		PushFont(headingFont)
	default:
		PushFont(MainFont)
	}
}

// PopFont convenience function - matches PushFont
func PopFont() {
	imgui.PopFont()
}
