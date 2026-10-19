# Styles

A `Style` is a complete presentation configuration: a display name, a default color `Theme`, the `Typography` loaded into the font slots, the ImGui `Metrics` applied at startup, dfx `Spacing` values ImGui has no metric for, and `Roles` binding dfx's custom chrome to ImGui color slots. An application selects one with `Config.Style`; with no style, dfx uses `ClassicStyle()`, its original compact presentation.

## Selecting and customizing

```go
style := dfx.ClassicStyle()
style.Metrics.FramePadding = imgui.Vec2{X: 6, Y: 4}
style.Typography.Small.Size = 15
app := dfx.New(root, dfx.Config{Style: &style})
```

`ClassicStyle()` returns independent configuration on every call: changing metrics, sizes, roles, or a face's `Data` slice assignment leaves other presets unchanged. Font bytes are shared, immutable assets; constructing or copying a style does not clone them. A style is complete rather than a sparse overlay: every field is meaningful, including zeros, so customization starts from a constructor. `App.Run` copies the configuration when it starts, and rejects one whose main, monospace, or small face has no font data, or any face with a non-positive size, before opening the window.

Style selection happens at startup only. There is no API for switching complete styles while an App runs; select another on a later launch.

`CurrentStyle()` returns a configuration copy of the running App's selection, with `Theme` set to the effective color theme. Font bytes remain shared and immutable. It describes the selected defaults — for example, `Spacing.SectionGap` for deliberate separation between groups. Live colors and metrics, including local overrides and `SetTheme` calls, come from the current ImGui context. Outside a running App, `CurrentStyle()` returns Classic.

## Startup order

`App.Run` sets up presentation in this order, then calls `OnSetup`, so application setup has the final word:

1. Select the style: `Config.Style`, or Classic. A non-nil `Config.Theme` overrides the style's default theme; a style whose theme is nil uses Modern Dark.
2. Load the style's typography, unless `DisableFonts` is set. Disabled fonts leave the atlas and the `Fonts` globals untouched.
3. Apply the style's metrics. This happens even when `DisableTheming` is set.
4. Apply the effective color theme, unless `DisableTheming` is set.

The selection and the font globals belong to that run. `App.Run` restores their previous values on every exit, including a panic, so a later App or a headless test never inherits a finished run's style or font pointers into a destroyed context. Like `Fonts`, this follows dfx's single-active-UI model.

## Color themes

`Theme` remains the color-only extension point. `SetTheme` changes colors in the current ImGui context at any time, without rebuilding fonts or changing metrics. Modern Dark starts from ImGui's dark colors before applying its table, so it produces the same colors whichever theme preceded it; the hue schemes and custom themes keep their own `Apply` semantics. Selecting a color theme does not select a complete style: Classic is chosen through `Config.Style` or by leaving it nil.

## Typography

`Typography` fills `MainFont`, `MonospaceFont`, and `SmallFont` from its `Main`, `Monospace`, and `Small` faces. A face with `Icons` set has Material Icons and the supplemental Material Symbols merged in at its size, shifted down by `IconOffset` (see [icon-fonts.md](icon-fonts.md)). Every `FontFace.Data` is borrowed, immutable font data, read in place by the atlas. Never edit these bytes through a preset or `CurrentStyle()`. To customize a face, assign another immutable asset before startup. If preparing your own bytes, finish editing them before supplying the slice as `Data` and keep them immutable while referenced by a style or font atlas.

`Heading` is optional. When it has data, it is loaded into a fourth slot after the three fixed ones. `PushHeadingFont()` pushes that face, or the main font when the style has none, or the current font when dfx fonts were not loaded; it always pairs with `PopFont()`. Classic has no heading face.

## Roles

`Roles` name the ImGui color slots dfx's custom chrome reads, resolved from the current context on each draw so theme changes and local `PushStyleColor` overrides reach it. Explicit per-widget colors still win.

| Role | Classic slot | Used by |
|---|---|---|
| `Accent` | `ColCheckMark` | — |
| `Panel` | `ColChildBg` | — |
| `SecondaryText` | `ColTextDisabled` | — |
| `FaderFill` | `ColButtonHovered` | fader fill and hovered handle border |
| `FaderActive` | `ColButtonActive` | active fader handle border |

## Classic

Classic is Gidole Regular at 20 px (main) and 16 px (small), both with icons merged, and JetBrains Mono Medium at 16 px; 4 px window padding, frame padding, and item spacing; ImGui's default 4×2 table cell padding; the `Default*` scrollbar, border, and rounding constants in `style.go`; and Modern Dark colors. Its `Spacing` sets a 4 px outer margin, matching the root window's padding, and an 8 px section gap.

## Legacy helpers

The original helpers keep their explicit meanings. `SetupFonts()` loads Classic's three fonts into the current atlas. `DefaultStyle()` applies Classic's metrics, including its cell padding. `ModernDark`, `ModernTheme`, the hue schemes, and `SetTheme` apply colors only. App startup uses the selected style rather than calling these helpers, so a direct call does what it names regardless of the selected style.
