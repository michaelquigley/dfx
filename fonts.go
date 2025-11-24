package dfx

import (
	"unsafe"

	"github.com/michaelquigley/dfx/fonts"
	"github.com/AllenDang/cimgui-go/imgui"
)

// font indices for easy access
const (
	MainFont      = 0 // default font (Gidole Regular)
	IconFont      = 1 // material icons (merged with main font)
	MonospaceFont = 2 // monospace font (JetBrains Mono)
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

	// add Material Icons merged with main font
	// build glyph ranges for material icons
	builder := imgui.NewFontGlyphRangesBuilder()
	ranges := []imgui.Wchar{'\ue000', '\uff01', 0}
	builder.AddRanges(&ranges[0])
	glyphRanges := imgui.NewGlyphRange()
	builder.BuildRanges(glyphRanges)

	materialConfig := imgui.NewFontConfig()
	materialConfig.SetFontData(uintptr(unsafe.Pointer(&fonts.MaterialIconsRegular[0])))
	materialConfig.SetFontDataSize(int32(len(fonts.MaterialIconsRegular)))
	materialConfig.SetFontDataOwnedByAtlas(false)
	materialConfig.SetSizePixels(20.0)
	materialConfig.SetGlyphOffset(imgui.Vec2{X: 0, Y: 4})
	materialConfig.SetGlyphRanges(glyphRanges.Data())
	materialConfig.SetMergeMode(true) // merge with previous font
	Fonts = append(Fonts, imgui.CurrentIO().Fonts().AddFont(materialConfig))

	// add JetBrains Mono as monospace font
	monoConfig := imgui.NewFontConfig()
	monoConfig.SetFontData(uintptr(unsafe.Pointer(&fonts.JetBrainsMonoMedium[0])))
	monoConfig.SetFontDataSize(int32(len(fonts.JetBrainsMonoMedium)))
	monoConfig.SetFontDataOwnedByAtlas(false)
	monoConfig.SetSizePixels(18.0)
	Fonts = append(Fonts, imgui.CurrentIO().Fonts().AddFont(monoConfig))
}

// PushFont convenience function for temporarily switching fonts
func PushFont(fontIndex int) {
	if fontIndex >= 0 && fontIndex < len(Fonts) {
		imgui.PushFont(Fonts[fontIndex], 0)
	}
}

// PopFont convenience function - matches PushFont
func PopFont() {
	imgui.PopFont()
}
