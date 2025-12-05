package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
)

// FaderChannel represents a single mixer channel with multiple value representations
type FaderChannel struct {
	name       string
	normalized float32 // 0.0 - 1.0 (master value)
	hardware   int     // 0 - 32767 (hardware representation)
	decibels   float32 // -60.0 to +12.0 dB (user-facing)
}

// updateFromNormalized keeps all representations in sync from normalized value
func (fc *FaderChannel) updateFromNormalized(norm float32) {
	fc.normalized = norm
	fc.hardware = int(norm * 32767)
	// Map 0-1 to -60dB to +12dB
	fc.decibels = norm*72.0 - 60.0
}

func main() {
	// Hue counter for cycling track color demo (0-360)
	hue := float32(0.0)

	// Create multiple fader channels with different initial values
	channels := []FaderChannel{
		{name: "Rainbow", normalized: 0.75},
		{name: "Log", normalized: 0.5},
		{name: "Audio", normalized: 0.8},
		{name: "Limited", normalized: 0.5},
		{name: "Hardware", normalized: 0.6},
		{name: "dB Scale", normalized: 0.7},
		{name: "Custom", normalized: 0.4},
		{name: "Reset", normalized: 0.5},
	}

	// Initialize all representations
	for i := range channels {
		channels[i].updateFromNormalized(channels[i].normalized)
	}

	// Create the root component with horizontally scrollable fader bank
	root := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("Advanced Fader Demo - Horizontal Scrollable Mixer")
		imgui.Separator()

		// Instructions
		imgui.TextWrapped("• Drag faders to adjust values")
		imgui.TextWrapped("• Right-click to reset to default")
		imgui.TextWrapped("• Mouse wheel to fine-tune (Ctrl = 10x faster, Alt = 10x slower)")
		imgui.TextWrapped("• Resize window to see horizontal scrolling")
		imgui.Separator()

		// Begin child window with horizontal scrollbar
		childHeight := float32(450.0)
		channelWidth := float32(120.0) // fixed width per channel (wider for scale labels)
		contentWidth := float32(len(channels)) * channelWidth

		// Create child window with horizontal scrollbar
		childSize := imgui.Vec2{X: 0, Y: childHeight} // X=0 means fill available width
		if imgui.BeginChildStrV("FaderBank", childSize, imgui.ChildFlagsNone, imgui.WindowFlagsHorizontalScrollbar) {
			// let it breathe to the left
			imgui.Dummy(imgui.Vec2{X: 50, Y: 0})
			imgui.SameLine()

			// Use table layout with fixed column widths to prevent jiggling
			// TableFlagsScrollX makes the table honor its width for scrolling
			if imgui.BeginTableV("mixer_table", int32(len(channels)), imgui.TableFlagsNone, imgui.Vec2{X: contentWidth + 50, Y: 0}, 0.0) {
				// Setup columns with fixed width
				for i := 0; i < len(channels); i++ {
					imgui.TableSetupColumnV(fmt.Sprintf("##col%d", i), imgui.TableColumnFlagsWidthFixed, channelWidth, 0)
				}

				// Row 1: Channel labels
				imgui.TableNextRow()
				for i := range channels {
					imgui.TableNextColumn()
					imgui.Text(channels[i].name)
				}

				// Row 2: Faders
				imgui.TableNextRow()
				for i := range channels {
					ch := &channels[i]
					imgui.TableNextColumn()

					// Select appropriate fader based on channel type
					switch i {
					case 0: // Rainbow track color demo (cycles hue every frame)
						// Increment hue each frame
						hue += 0.5
						if hue >= 360.0 {
							hue = 0.0
						}

						// Convert HSV to RGB for track color
						var r, g, b float32
						imgui.ColorConvertHSVtoRGB(hue/360.0, 1.0, 0.5, &r, &g, &b)
						trackColor := imgui.Vec4{X: r, Y: g, Z: b, W: 1.0}

						params := dfx.DefaultFaderParams()
						params.Taper = dfx.LinearTaper()
						params.TrackColor = &trackColor
						params.Format = func(norm float32) string {
							return fmt.Sprintf("%.3f", norm)
						}

						// Percentage scale
						scale := dfx.DefaultScaleConfig()
						scale.Marks = []float32{0.0, 0.25, 0.5, 0.75, 1.0}
						scale.Labels = map[float32]string{
							0.0:  "0%",
							0.25: "25%",
							0.5:  "50%",
							0.75: "75%",
							1.0:  "100%",
						}

						if newValue, changed := dfx.FaderWithScaleN(fmt.Sprintf("##fader%d", i), ch.normalized, params, scale); changed {
							ch.updateFromNormalized(newValue)
						}

					case 1: // Log taper
						params := dfx.DefaultFaderParams()
						params.Taper = dfx.LogTaper(3.0) // moderate curve
						params.Format = func(norm float32) string {
							return fmt.Sprintf("Log: %.3f", norm)
						}

						if newValue, changed := dfx.FaderN(fmt.Sprintf("##fader%d", i), ch.normalized, params); changed {
							ch.updateFromNormalized(newValue)
						}

					case 2: // Audio taper with scale
						params := dfx.DefaultFaderParams()
						params.Taper = dfx.AudioTaper()
						params.Format = func(norm float32) string {
							return fmt.Sprintf("Audio: %.3f", norm)
						}

						// Scale showing taper-aware positioning
						scale := dfx.DefaultScaleConfig()
						scale.Marks = []float32{0.0, 0.25, 0.5, 0.75, 1.0}
						scale.Labels = map[float32]string{
							0.0: "0",
							0.5: ".5",
							1.0: "1",
						}

						if newValue, changed := dfx.FaderWithScaleN(fmt.Sprintf("##fader%d", i), ch.normalized, params, scale); changed {
							ch.updateFromNormalized(newValue)
						}

					case 3: // Range-limited (20% - 80%)
						params := dfx.DefaultFaderParams()
						params.MinStop = 0.2
						params.MaxStop = 0.8
						params.ResetValue = 0.5
						params.Format = func(norm float32) string {
							return fmt.Sprintf("Limited: %.3f", norm)
						}

						if newValue, changed := dfx.FaderN(fmt.Sprintf("##fader%d", i), ch.normalized, params); changed {
							ch.updateFromNormalized(newValue)
						}

					case 4: // Hardware int (0-32767)
						params := dfx.DefaultFaderParams()
						params.Taper = dfx.LinearTaper()
						params.Format = func(norm float32) string {
							hw := int(norm * 32767)
							return fmt.Sprintf("HW: %d", hw)
						}

						if newValue, changed := dfx.FaderI(fmt.Sprintf("##fader%d", i), ch.hardware, 0, 32767, params); changed {
							ch.hardware = newValue
							ch.normalized = float32(newValue) / 32767.0
							ch.decibels = ch.normalized*72.0 - 60.0
						}

					case 5: // dB scale with audio taper and dB labels
						params := dfx.DefaultFaderParams()
						params.Taper = dfx.AudioTaper()
						params.Format = func(norm float32) string {
							db := norm*72.0 - 60.0
							if db <= -59.9 {
								return "-∞ dB"
							}
							return fmt.Sprintf("%.1f dB", db)
						}

						// dB scale (normalized positions for -60dB to +12dB range)
						scale := dfx.DefaultScaleConfig()
						scale.Marks = []float32{0.0, 0.417, 0.667, 0.833, 1.0}
						scale.Labels = map[float32]string{
							0.0:   "-60",
							0.417: "-30",
							0.667: "-12",
							0.833: "0",
							1.0:   "+12",
						}
						scale.TickLength = 6.0

						if newValue, changed := dfx.FaderWithScaleF(fmt.Sprintf("##fader%d", i), ch.decibels, -60.0, 12.0, params, scale); changed {
							ch.decibels = newValue
							ch.normalized = (newValue + 60.0) / 72.0
							ch.hardware = int(ch.normalized * 32767)
						}

					case 6: // Custom steep log taper
						params := dfx.DefaultFaderParams()
						params.Taper = dfx.LogTaper(10.0) // steep curve
						params.Format = func(norm float32) string {
							return fmt.Sprintf("Steep: %.3f", norm)
						}

						if newValue, changed := dfx.FaderN(fmt.Sprintf("##fader%d", i), ch.normalized, params); changed {
							ch.updateFromNormalized(newValue)
						}

					case 7: // Reset demo (right-click to reset to 0.75)
						params := dfx.DefaultFaderParams()
						params.ResetValue = 0.75
						params.Format = func(norm float32) string {
							return fmt.Sprintf("Reset→%.2f", 0.75)
						}

						if newValue, changed := dfx.FaderN(fmt.Sprintf("##fader%d", i), ch.normalized, params); changed {
							ch.updateFromNormalized(newValue)
						}
					}
				}

				// Row 3: Value displays
				imgui.TableNextRow()
				for i := range channels {
					ch := &channels[i]
					imgui.TableNextColumn()

					// Display values (fixed width text to prevent jiggling)
					imgui.Text(fmt.Sprintf("N: %.3f", ch.normalized))
					imgui.Text(fmt.Sprintf("HW: %5d", ch.hardware))
					if ch.decibels <= -59.9 {
						imgui.Text("dB: -∞    ")
					} else {
						imgui.Text(fmt.Sprintf("dB: %5.1f", ch.decibels))
					}
				}

				imgui.EndTable()
			}

			imgui.EndChild()
		}

		imgui.Separator()
		imgui.Text("Fader Types:")
		imgui.BulletText("Rainbow: Cycling track color (hue changes every frame)")
		imgui.BulletText("Log: Logarithmic taper (moderate)")
		imgui.BulletText("Audio: Audio fader curve with taper-aware scale marks")
		imgui.BulletText("Limited: Range stops at 20%%-80%%")
		imgui.BulletText("Hardware: Integer range (0-32767)")
		imgui.BulletText("dB Scale: Float range (-60dB to +12dB) with dB scale labels")
		imgui.BulletText("Custom: Steep logarithmic curve")
		imgui.BulletText("Reset: Right-click resets to 75%%")
		imgui.Separator()
		imgui.TextWrapped("Note: Faders with scales (Rainbow, Audio, dB) demonstrate the FaderWithScale API with tick marks and labels.")
	})

	app := dfx.New(root, dfx.Config{
		Title:  "Advanced Fader Demo",
		Width:  1100,
		Height: 850,
	})

	app.Run()
}
