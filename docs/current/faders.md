# Faders

`FaderN`, `FaderF`, and `FaderI` draw vertical gain-style controls with a narrow slot, a continuous accent fill below a rounded handle, and a horizontal position marker. The APIs return the new value and whether input changed it. `FaderN` works in normalized 0–1 space; `FaderF` and `FaderI` convert arbitrary float and integer ranges to that space.

The interaction remains an ImGui vertical slider: click the track to set a position, drag the handle without a jump at the grab point, and use ImGui keyboard/gamepad navigation when enabled by the application. Wheel input adjusts in tapered visual space, with Ctrl moving ten times faster and Alt ten times slower. Right-click resets to `ResetValue`. Range stops and reset values are normalized, including when using the float/integer wrappers.

## Appearance and sizing

Start with `DefaultFaderParams()`. `Width` and `Height` describe the complete slider hit area and layout allocation, defaulting to 30×300 pixels. `HandleWidth` defaults to `Width - 4`, `HandleHeight` to 18, and `TrackWidth` to 4. Handle dimensions are capped to the hit area minus a 2-pixel margin on each side; track width is capped to handle width. Positive hit-area dimensions smaller than 6 pixels are raised to 6.

All default colors come from the active ImGui style and are read each frame. The fill uses `ColButtonHovered` (the default Modern Dark theme's yellow). The slot and idle handle body use `ColFrameBg`; hovered and active bodies use `ColFrameBgHovered` and `ColFrameBgActive`. The handle border uses `ColBorder`, then `ColButtonHovered` on hover and `ColButtonActive` when active. The marker uses `ColText`, shadows use `ColBorderShadow`, and the scale uses `ColTextDisabled`.

`AccentColor` overrides the fill and the hovered/active handle border. `TrackColor` overrides the unfilled slot. Both are optional `*imgui.Vec4` values. Color alpha is preserved and drawing respects ImGui's style alpha, including disabled controls. Theme changes take effect on the next draw without rebuilding the fader parameters.

```go
params := dfx.DefaultFaderParams()
params.Width, params.Height = 60, 340
params.HandleWidth, params.HandleHeight, params.TrackWidth = 56, 28, 7
params.ResetValue = 0.8 // 0 dB within -40 to +10
params.Format = func(normalized float32) string {
    return fmt.Sprintf("%.1f dB", normalized*50-40)
}
value, changed := dfx.FaderF("##gain", gainDB, -40, 10, params)
```

## Tapers and scales

`Taper` maps normalized values to visual positions and back; nil selects linear. Existing linear, logarithmic, audio, decibel, and custom taper implementations are supported. Values already expressed in dB need a linear taper for equal travel per dB; `DecibelTaper` is intended for values proportional to linear amplitude.

`FaderWithScaleN/F/I` add `ScaleConfig` marks and labels. Mark positions are normalized values before tapering. Scale marks use the same inset travel as the handle center, so the scale and handle agree at both endpoints. Labels and ticks are drawn outside the hit area on the configured left or right side; callers must leave room for them in their layout. `DefaultScaleConfig()` enables `GuideLines`, extending each tick across the hit area behind the slot and handle. Set it false for short external ticks only.

The mixer example (`go run ./examples/dfx_example_mixer`) opens on a five-channel reference bank with larger handles and a -40 to +10 dB scale. Its second tab demonstrates the original taper, normalized/integer range, range-stop, accent-color, and reset cases.
