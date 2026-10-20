# Styles

A `Style` is a complete presentation configuration: a display name, a default color `Theme`, the `Typography` loaded into the font slots, the ImGui `Metrics` applied at startup, dfx `Spacing` values ImGui has no metric for, `Roles` binding dfx's custom chrome to ImGui color slots, `Chrome` defaults for dfx's composite components, and `Controls` defaults for its own controls. An application selects one with `Config.Style`. With no style, dfx uses `FaceplateStyle()`, the faceplate family's presentation (see [Faceplate](#faceplate)). `ClassicStyle()` is dfx's original compact presentation, the default before Faceplate; select it explicitly to keep the old look (see [Migrating from Classic](#migrating-from-classic)).

## Selecting and customizing

```go
style := dfx.FaceplateStyle()
style.Theme = &dfx.FaceplateTheme{Accent: imgui.Vec4{X: 0.3, Y: 0.64, Z: 1, W: 1}} // another identity color
style.Spacing.SectionGap = 24
style.Metrics.ItemSpacing = imgui.Vec2{X: 10, Y: 10}
style.Typography.Small.Size = 15
app := dfx.New(root, dfx.Config{Style: &style})
```

Customization starts from a constructor, the default's or Classic's, and changes fields before startup.

`ClassicStyle()` and `FaceplateStyle()` return independent configuration on every call, each with its own theme value: changing metrics, sizes, roles, or a face's `Data` slice assignment leaves other presets unchanged. Font bytes are shared, immutable assets; constructing or copying a style does not clone them. A style is complete rather than a sparse overlay: every field is meaningful, including zeros, so customization starts from a constructor. `App.Run` copies the configuration when it starts, and rejects one whose main, monospace, or small face has no font data, or any face with a non-positive size, before opening the window.

Style selection happens at startup only. There is no API for switching complete styles while an App runs; select another on a later launch.

`CurrentStyle()` returns a configuration copy of the running App's selection, with `Theme` set to the effective color theme. Font bytes remain shared and immutable. It describes the selected defaults — for example, `Spacing.SectionGap` for deliberate separation between groups. Live colors and metrics, including local overrides and `SetTheme` calls, come from the current ImGui context. Outside a running App, `CurrentStyle()` returns Classic, as do the defaults dfx's components read there: an application that drives its own ImGui context without `App.Run` sets it up through the legacy helpers, which are Classic's (see [Legacy helpers](#legacy-helpers)).

`Theme` is an interface value. A constructor's theme is new on each call, so presets stay independent, but copying a style by assignment shares the theme it names. The built-in `FaceplateTheme` is copied at the two points where ownership matters. At startup, the running App copies the effective theme, whether it came from the style or `Config.Theme`, so editing the source afterwards does not change the running selection. Each `CurrentStyle()` call then returns its own copy, so editing one snapshot changes neither the running selection nor another snapshot. Live colors still change only through `SetTheme`. Other `Theme` implementations, including the application's own, are not copied: a theme the application supplies remains the application's. Font bytes stay shared in every case.

## Startup order

`App.Run` sets up presentation in this order, then calls `OnSetup`, so application setup has the final word:

1. Select the style: `Config.Style`, or Faceplate. A non-nil `Config.Theme` overrides the style's default theme; a style whose theme is nil uses Modern Dark.
2. Load the style's typography, unless `DisableFonts` is set. Disabled fonts leave the atlas and the `Fonts` globals untouched.
3. Apply the style's metrics. This happens even when `DisableTheming` is set.
4. Apply the effective color theme, unless `DisableTheming` is set.

The selection and the font globals belong to that run. `App.Run` restores their previous values on every exit, including a panic, so a later App or a headless test never inherits a finished run's style or font pointers into a destroyed context. Like `Fonts`, this follows dfx's single-active-UI model.

## Color themes

`Theme` remains the color-only extension point. `SetTheme` changes colors in the current ImGui context at any time, without rebuilding fonts or changing metrics. Modern Dark starts from ImGui's dark colors before applying its table, so it produces the same colors whichever theme preceded it; the hue schemes and custom themes keep their own `Apply` semantics. Selecting a color theme does not select a complete style. `Config.Theme: dfx.ModernDark` alone gives Faceplate's fonts and spacing in Modern Dark colors; Classic is chosen only through `Config.Style`.

## Typography

`Typography` fills `MainFont`, `MonospaceFont`, and `SmallFont` from its `Main`, `Monospace`, and `Small` faces. A face with `Icons` set has Material Icons and the supplemental Material Symbols merged in at its size, shifted down by `IconOffset` (see [icon-fonts.md](icon-fonts.md)). Every `FontFace.Data` is borrowed, immutable font data, read in place by the atlas. Never edit these bytes through a preset or `CurrentStyle()`. To customize a face, assign another immutable asset before startup. If preparing your own bytes, finish editing them before supplying the slice as `Data` and keep them immutable while referenced by a style or font atlas.

`Heading` is optional. When it has data, it is loaded into a fourth slot after the three fixed ones. `PushHeadingFont()` pushes that face, or the main font when the style has none, or the current font when dfx fonts were not loaded; it always pairs with `PopFont()`. Classic has no heading face.

## Roles

`Roles` name the ImGui color slots dfx's custom chrome reads, resolved from the current context on each draw so theme changes and local `PushStyleColor` overrides reach it. Explicit per-widget colors still win.

| Role | Classic slot | Faceplate slot | Used by |
|---|---|---|---|
| `Accent` | `ColHeaderActive` | `ColCheckMark` | NodeCanvas selection emphasis, derived by `DefaultNodeCanvasStyle` |
| `Panel` | `ColChildBg` | `ColChildBg` | — |
| `SecondaryText` | `ColTextDisabled` | `ColTextDisabled` | — |
| `FaderFill` | `ColButtonHovered` | `ColCheckMark` | fader fill and hovered handle border |
| `FaderActive` | `ColButtonActive` | `ColCheckMark` | active fader handle border |
| `Toolbar` | `ColHeader` | `ColTitleBgActive` | toolbar background |
| `Handle` | `ColHeaderActive` | `ColTextDisabled` | Dash and HCollapse drag-handle glyphs |

`Panel` and `SecondaryText` name the family's surfaces for application chrome; no dfx component reads them yet. Under Faceplate, as under Classic, the `Panel` slot carries its hue at zero alpha, so application chrome that paints with it supplies its own alpha, as Dash and HCollapse do.

## Chrome and controls

`Chrome` holds the presentation defaults of dfx's composite components; their public fields, such as `DashManager.Margin`, `HCollapse` widths, and fader dimensions, stay explicit and win.

| Field | Classic | Faceplate | Used by |
|---|---|---|---|
| `PanelRounding` | 5 | 5 | Dash and HCollapse panel corners |
| `ContentPadding` | 0×0 | 12×12 | inside HCollapse content and MultiGrid cells |
| `ToolbarInset` | 3 | 8 | space before a toolbar's label |
| `FlexMargin` | 2 | 0 | FlexLayout slack at its trailing edges |
| `FlexGap` | 4 | 0 | FlexLayout gap beside its 10 px splitter strips |
| `GridGap` | 0 | 8 | GridLayout gap, split across interior cell edges |

`Controls.AccentInk` governs labels drawn on accent fills. Off (Classic), an on `Toggle` fills with `ColCheckMark` at rest and takes the theme's button hover and press colors under the theme's text, and NodeCanvas title labels use the theme's text. On (Faceplate), those labels are drawn in whichever of the theme's background and text colors has the higher WCAG contrast with the fill; for the default yellow that is the charcoal background, at 12:1. An on toggle also keeps its accent fill through hover and press, nudged away from the label so neither state reads worse than rest. NodeCanvas applies the choice to `NodeContext.Label` text inside `TitleBar`, against the node's resolved band, whether the style's selected band, its normal band, or a declared node accent; widgets in a title bar keep their own colors. `ToggleV` takes `ToggleOptions.OnText`, an explicit on-state label color that wins over the style's choice. Disabled toggles keep ImGui's disabled alpha.

## Faceplate

`FaceplateStyle()` is Chakra Petch Regular at 20 px (main) and 16 px (small) and real Chakra Petch Bold at 20 px (heading), all three with icons merged at Classic's offsets, which match Chakra Petch's cap height within half a pixel; JetBrains Mono Medium at 16 px; 12 px window padding and outer margin; 8×5 frame padding; 8 px item spacing; 8×4 table cell padding; a 16 px section gap; Classic's scrollbar, borders, and radii; and a `FaceplateTheme`.

`FaceplateTheme{Accent}` colors every ImGui slot from imgui's dark baseline, so it does not depend on a previous theme. Its neutrals are fixed: background `#1a1c20`, panel `#24272e` (Dash and HCollapse panels), raised `#2e3239` (buttons, fields, popups, table headers), track `#3a3f48` (borders and neutral hover), press `#474d57`, text `#d6d6d6`, secondary text `#878b93`. The accent, `#ffd800` by default, marks the check mark, slider grabs, selected-tab overline, text links, keyboard focus, and drag targets; selection headers and selected text use it translucently. Hover stays neutral. `ColChildBg` holds the panel hue at zero alpha, as Modern Dark's does: a plain `BeginChild` region is a layout region and stays unpainted, because ImGui gives a child with no border no padding, and a painted one would show its content pressed against its edges. Dash and HCollapse paint the panel by setting their own background alpha. To change the identity color, replace the style's theme before startup:

```go
style := dfx.FaceplateStyle()
style.Theme = &dfx.FaceplateTheme{Accent: imgui.Vec4{X: 0.3, Y: 0.64, Z: 1, W: 1}}
```

Meter zones, log severity colors, and colors an application declares on graphs, faders, or canvases keep their meaning; the accent does not recolor them.

### Faceplate snapshot

dfx owns its Go/ImGui realization of the faceplate family. The palette values above and the Chakra Petch faces were taken from faceplate at revision `c65b59e8319b8f030b1bbc3ef01a688779dcb677` (`src/faceplate/faceplate.h`, `assets/fonts/`); later faceplate changes are adopted deliberately. The fonts were fetched by faceplate on 2026-07-06 from `github.com/google/fonts` (`ofl/chakrapetch/`) and are licensed under the SIL Open Font License 1.1, bundled as `fonts/chakra-petch-ofl.txt`.

| File | sha256 |
|---|---|
| `fonts/ChakraPetch-Regular.ttf` | `98fcd638baa5c81ff0316b7538ce330ee3b23b1302726de3526d5933a8ecf986` |
| `fonts/ChakraPetch-Bold.ttf` | `65fbf76d95651697275e19db4d717c0e95a789ddd3476478b05292104db278a0` |

## Comparing styles

`examples/dfx_example_themes` is the comparison surface: controls, typography with icons beside regular and bold text, tables (one with a deliberate local density and color override), FileTree, LogViewer, meters and faders, Workspace, DashManager, HCollapse on both anchors, a Toolbar, both MultiGrid layouts, and NodeCanvas with a mini-map beside an explicitly styled canvas. Launch it once per style:

```bash
go run ./examples/dfx_example_themes                  # Faceplate, default yellow
go run ./examples/dfx_example_themes -accent 4da3ff   # Faceplate, another accent
go run ./examples/dfx_example_themes -style classic   # Classic
```

Its Colors menu changes colors only and starts on the colors actually applied; fonts and spacing change by relaunching with another `-style`; it re-derives the implicit canvas style with `SetStyle` and leaves the explicit one alone.

## Classic

Classic is dfx's original presentation and was its default before Faceplate: Gidole Regular at 20 px (main) and 16 px (small), both with icons merged, and JetBrains Mono Medium at 16 px; 4 px window padding, frame padding, and item spacing; ImGui's default 4×2 table cell padding; the `Default*` scrollbar, border, and rounding constants in `style.go`; and Modern Dark colors. Its `Spacing` sets a 4 px outer margin, matching the root window's padding, and an 8 px section gap.

## Migrating from Classic

An application that sets no `Config.Style` now starts in Faceplate: Chakra Petch instead of Gidole, a fourth (bold heading) font slot, roomier metrics and margins, charcoal colors with a yellow accent, and Faceplate's `Chrome` around dfx's panels, cells, and toolbars. Nothing else about the API changes. To keep the previous presentation, select Classic:

```go
classic := dfx.ClassicStyle()
app := dfx.New(root, dfx.Config{Style: &classic})
```

Classic reproduces the old fonts, metrics, colors, and widget treatments, but not the old allocation defects. Containers now hand components the usable area at their drawing origin under either style (see the allocation notes in [layout-guide.md](layout-guide.md#how-dfx-allocates)), so a layout that compensated for the old sizes may shift by a few pixels.

Points to check when moving an application to the new default:

- **Colors only.** `Config.Theme` and `SetTheme` still change colors only. An application that set `Theme: dfx.ModernDark`, or a hue scheme, keeps those colors but gets Faceplate's fonts and spacing; add `Style: &classic` to keep Classic's as well.
- **Disabled setup.** `DisableFonts` leaves the font atlas alone, so dfx's font slots stay unset, but Faceplate's metrics and colors still apply. `DisableTheming` skips the color theme but still applies Faceplate's metrics. Select Classic to keep Classic's metrics under either flag.
- **Fixed geometry.** Faceplate's text is wider and its spacing larger. Sizes written for Classic, such as a fixed window size, a panel width that adds a hand-tuned padding, or a height that subtracts a fixed number of pixels for a header, may now clip. Prefer the measured remainder (`imgui.ContentRegionAvail()`, `state.Size`). Where an older public field holds a size, such as `HCollapse.ExpandedWidth` or `DashManager.Margin`, adopt the selected style in `OnSetup`, for example by adding `2 * dfx.CurrentStyle().Chrome.ContentPadding.X` to a panel width. dfx never reinterprets those fields itself.
- **NodeCanvas.** A canvas with a zero `Config.Style` derives `DefaultNodeCanvasStyle()` from the active colors once, at its first `Begin`. It does not follow later `SetTheme` calls by itself: after changing colors, refresh it explicitly with `canvas.SetStyle(dfx.DefaultNodeCanvasStyle())`. A canvas given an explicit style keeps it.
- **Fonts.** `MainFont`, `MonospaceFont`, and `SmallFont` keep their slots. `PushHeadingFont` uses the bold heading slot Faceplate adds and falls back to the main font under Classic.

## Legacy helpers

The original helpers keep their explicit meanings. `SetupFonts()` loads Classic's three fonts into the current atlas. `DefaultStyle()` applies Classic's metrics, including its cell padding. `ModernDark`, `ModernTheme`, the hue schemes, and `SetTheme` apply colors only. They did not move to Faceplate with the default: they still name Classic. App startup uses the selected style rather than calling these helpers, so a direct call does what it names regardless of the selected style.
