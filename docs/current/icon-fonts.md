# icon fonts

`SetupFonts` merges the existing Material Icons font and an additive Material Symbols font into both the main and small UI fonts. the font slots and existing `fonts.ICON_*` constants are unchanged. new symbols use `fonts.ICON_SYMBOL_*`, including `ICON_SYMBOL_VIEW_REAL_SIZE` for the 1:1 zoom reset icon.

the supplement contains only private-use BMP codepoints absent from the bundled Material Icons font. codepoint collisions keep their existing Material Icons artwork; the supplement does not replace them. symbols outside the BMP are omitted because the pinned cimgui-go build uses 16-bit `ImWchar`. applications can use the exported supplemental constants directly as widget labels, with the same font scaling and baseline alignment as existing icons.

## source and regeneration

the source is Google's [Material Symbols Outlined font and codepoints](https://github.com/google/material-design-icons/tree/bd8cb85bd4bad964fe6918f79665bb40c3a8efef/variablefont), pinned at commit `bd8cb85bd4bad964fe6918f79665bb40c3a8efef` (2026-09-25), under the bundled Apache 2.0 license in `fonts/material-symbols-license.txt`. `generateMaterialSymbols.py` freezes FILL=1, GRAD=0, opsz=24 and wght=400, then subsets to the supported non-overlapping codepoints. the resulting font is a fixed filled style; runtime variable-font support is not required.

the generator normalizes line metrics to the square em used by Material Icons: ImGui scales by ascent minus descent, so Symbols' original extra leading would make its icon cells smaller at the same pixel size. the compact `view_real_size` lettering is enlarged 1.5 times and centered within its unchanged square advance, so the `1:1` button reads alongside the fit icon. existing Material Icons glyphs and button padding are unchanged.

download `MaterialSymbolsOutlined[FILL,GRAD,opsz,wght].ttf` and the matching `.codepoints` from that commit, then run from the repository root using Python with `fonttools==4.66.1` installed:

```bash
python3 fonts/generateMaterialSymbols.py /path/to/source.ttf /path/to/source.codepoints
gofmt -w fonts/materialSymbols.go
make test
```

source SHA-256 checksums:

- font: `0128da5981791d2fe09918c337ea6657b0ef2e4e5f8f4af3e1dc5f31889b2311`
- codepoints: `225bd09137103cb7746bc93dc08d08764c9f0c3bd04f4b958d4a3c3c19432dd6`

the generator and regression tests keep the legacy and supplemental cmaps disjoint. the font setup test verifies that the real-size glyph and an existing icon resolve without fallback in both font slots at multiple scales. visual alignment still needs a live display check.
