package fonts

import (
	"testing"

	"golang.org/x/image/font/sfnt"
)

func TestMaterialSymbolsSupplementPreservesLegacyCodepoints(t *testing.T) {
	legacy, err := sfnt.Parse(MaterialIconsRegular)
	if err != nil {
		t.Fatal(err)
	}
	symbols, err := sfnt.Parse(MaterialSymbolsSupplement)
	if err != nil {
		t.Fatal(err)
	}
	for code := rune(0); code <= 0xffff; code++ {
		oldGlyph, err := legacy.GlyphIndex(nil, code)
		if err != nil {
			t.Fatal(err)
		}
		newGlyph, err := symbols.GlyphIndex(nil, code)
		if err != nil {
			t.Fatal(err)
		}
		if oldGlyph != 0 && newGlyph != 0 {
			t.Fatalf("supplement shadows existing codepoint 'U+%04X'", code)
		}
	}
	code := []rune(ICON_SYMBOL_VIEW_REAL_SIZE)[0]
	glyph, err := symbols.GlyphIndex(nil, code)
	if err != nil || glyph == 0 {
		t.Fatalf("view real size glyph missing: '%v'", err)
	}
}
