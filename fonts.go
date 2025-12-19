package dfx

import (
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx/fonts"
)

// font indices for easy access
// note: icon fonts are merged into their base fonts, not separate entries
const (
	MainFont      = 0 // default font (Gidole Regular, 20px) with Material Icons merged
	MonospaceFont = 1 // monospace font (JetBrains Mono, 18px)
	SmallFont     = 2 // small font (Gidole Regular, 16px) with Material Icons merged
)

var Fonts []*imgui.Font

// SetupFonts initializes and loads all fonts
// this should be called during app initialization
func SetupFonts() {
	// clear any existing fonts
	imgui.CurrentIO().Fonts().Clear()
	Fonts = Fonts[:0] // clear slice but keep capacity

	// add Gidole Regular as the main font
	gidoleConfig := imgui.NewFontConfig()
	gidoleConfig.SetFontData(uintptr(unsafe.Pointer(&fonts.GidoleRegular[0])))
	gidoleConfig.SetFontDataSize(int32(len(fonts.GidoleRegular)))
	gidoleConfig.SetFontDataOwnedByAtlas(false)
	gidoleConfig.SetSizePixels(20.0)
	Fonts = append(Fonts, imgui.CurrentIO().Fonts().AddFont(gidoleConfig))

	// build glyph ranges for material icons (used for both main and small fonts)
	builder := imgui.NewFontGlyphRangesBuilder()
	ranges := []imgui.Wchar{'\ue000', '\uff01', 0}
	builder.AddRanges(&ranges[0])
	glyphRanges := imgui.NewGlyphRange()
	builder.BuildRanges(glyphRanges)

	// add Material Icons merged with main font (don't append - it merges into MainFont)
	materialConfig := imgui.NewFontConfig()
	materialConfig.SetFontData(uintptr(unsafe.Pointer(&fonts.MaterialIconsRegular[0])))
	materialConfig.SetFontDataSize(int32(len(fonts.MaterialIconsRegular)))
	materialConfig.SetFontDataOwnedByAtlas(false)
	materialConfig.SetSizePixels(20.0)
	materialConfig.SetGlyphOffset(imgui.Vec2{X: 0, Y: 4})
	materialConfig.SetGlyphRanges(glyphRanges.Data())
	materialConfig.SetMergeMode(true) // merge with previous font
	imgui.CurrentIO().Fonts().AddFont(materialConfig)

	// add JetBrains Mono as monospace font
	monoConfig := imgui.NewFontConfig()
	monoConfig.SetFontData(uintptr(unsafe.Pointer(&fonts.JetBrainsMonoMedium[0])))
	monoConfig.SetFontDataSize(int32(len(fonts.JetBrainsMonoMedium)))
	monoConfig.SetFontDataOwnedByAtlas(false)
	monoConfig.SetSizePixels(18.0)
	Fonts = append(Fonts, imgui.CurrentIO().Fonts().AddFont(monoConfig))

	// add small font (Gidole for small labels/indicators)
	smallConfig := imgui.NewFontConfig()
	smallConfig.SetFontData(uintptr(unsafe.Pointer(&fonts.GidoleRegular[0])))
	smallConfig.SetFontDataSize(int32(len(fonts.GidoleRegular)))
	smallConfig.SetFontDataOwnedByAtlas(false)
	smallConfig.SetSizePixels(16.0)
	Fonts = append(Fonts, imgui.CurrentIO().Fonts().AddFont(smallConfig))

	// add small Material Icons merged with small font
	smallMaterialConfig := imgui.NewFontConfig()
	smallMaterialConfig.SetFontData(uintptr(unsafe.Pointer(&fonts.MaterialIconsRegular[0])))
	smallMaterialConfig.SetFontDataSize(int32(len(fonts.MaterialIconsRegular)))
	smallMaterialConfig.SetFontDataOwnedByAtlas(false)
	smallMaterialConfig.SetSizePixels(16.0)
	smallMaterialConfig.SetGlyphOffset(imgui.Vec2{X: 0, Y: 3}) // scaled offset (4px at 20px -> 3px at 16px)
	smallMaterialConfig.SetGlyphRanges(glyphRanges.Data())
	smallMaterialConfig.SetMergeMode(true) // merge with previous font (small font)
	imgui.CurrentIO().Fonts().AddFont(smallMaterialConfig)
}

// font sizes corresponding to each font index
var fontSizes = []float32{20.0, 18.0, 16.0}

// PushFont convenience function for temporarily switching fonts
func PushFont(fontIndex int) {
	if fontIndex >= 0 && fontIndex < len(Fonts) {
		size := float32(0)
		if fontIndex < len(fontSizes) {
			size = fontSizes[fontIndex]
		}
		imgui.PushFont(Fonts[fontIndex], size)
	}
}

// PopFont convenience function - matches PushFont
func PopFont() {
	imgui.PopFont()
}
