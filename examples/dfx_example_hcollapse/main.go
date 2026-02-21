package main

import (
	"fmt"
	"math"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
)

// MixerChannel represents a single mixer channel with fader and meter
type MixerChannel struct {
	name       string
	normalized float32
	meter      *dfx.VUMeter
}

func newMixerChannel(name string, initial float32) *MixerChannel {
	meter := dfx.NewVUMeter(1)
	meter.Height = 180
	meter.ChannelWidth = 10
	meter.SetLabel(0, "")
	return &MixerChannel{
		name:       name,
		normalized: initial,
		meter:      meter,
	}
}

func main() {
	drumChannels := []*MixerChannel{
		newMixerChannel("Kick", 0.75),
		newMixerChannel("Snare", 0.70),
		newMixerChannel("HiHat", 0.55),
		newMixerChannel("Toms", 0.60),
	}

	synthChannels := []*MixerChannel{
		newMixerChannel("Bass", 0.80),
		newMixerChannel("Lead", 0.65),
		newMixerChannel("Pad", 0.50),
		newMixerChannel("FX", 0.40),
	}

	// create content components for each collapsible panel
	drumsContent := createMixerPanel("drums", drumChannels)
	synthsContent := createMixerPanel("synths", synthChannels)

	// create the HCollapse panels
	// 4 channels * 60px = 240px content + padding + resize handle
	drumsCollapse := dfx.NewHCollapse(drumsContent, dfx.HCollapseConfig{
		Title:         "Drums",
		ExpandedWidth: 300,
		TransitionMs:  150,
		Resizable:     true,
		Expanded:      true,
	})

	synthsCollapse := dfx.NewHCollapse(synthsContent, dfx.HCollapseConfig{
		Title:         "Synths",
		ExpandedWidth: 300,
		TransitionMs:  150,
		Resizable:     true,
		Expanded:      true,
	})

	// simulation state
	startTime := time.Now()
	paused := false

	// root component
	root := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("HCollapse Demo - Collapsible Mixer Panels")
		imgui.Separator()

		// controls row
		if newValue, changed := dfx.Checkbox("Pause Animation", paused); changed {
			paused = newValue
		}
		imgui.SameLine()
		if imgui.Button("Toggle Drums") {
			drumsCollapse.Toggle()
		}
		imgui.SameLine()
		if imgui.Button("Toggle Synths") {
			synthsCollapse.Toggle()
		}

		imgui.Spacing()
		imgui.Separator()
		imgui.Spacing()

		// simulate meter levels if not paused
		if !paused {
			t := time.Since(startTime).Seconds()
			simulateLevels(drumChannels, t, 0)
			simulateLevels(synthChannels, t, 4)
		}

		// draw the collapsible panels side by side
		// calculate available height for panels
		panelHeight := state.Size.Y - 120 // leave room for header and controls

		// create a state with adjusted height for the panels
		panelState := &dfx.State{
			Size:     imgui.Vec2{X: state.Size.X, Y: panelHeight},
			Position: state.Position,
			IO:       state.IO,
			App:      state.App,
			Parent:   nil,
		}

		// first collapsible panel (draws directly, no wrapper)
		drumsCollapse.Draw(panelState)

		imgui.SameLine()

		// second collapsible panel
		synthsCollapse.Draw(panelState)

		imgui.SameLine()

		// main content area (fills remaining space)
		remaining := state.Size.X - drumsCollapse.CurrentWidth - synthsCollapse.CurrentWidth - 20
		if remaining > 50 {
			imgui.BeginChildStrV("main", imgui.Vec2{X: remaining, Y: panelHeight}, imgui.ChildFlagsBorders, 0)
			imgui.Text("Main Content Area")
			imgui.Separator()
			imgui.Spacing()
			imgui.Text("This area expands as panels collapse.")
			imgui.Spacing()
			imgui.Text(fmt.Sprintf("Drums panel: %.0fpx", drumsCollapse.CurrentWidth))
			imgui.Text(fmt.Sprintf("Synths panel: %.0fpx", synthsCollapse.CurrentWidth))
			imgui.Text(fmt.Sprintf("Main area: %.0fpx", remaining))
			imgui.Spacing()
			imgui.Separator()
			imgui.Spacing()
			imgui.Text("Tips:")
			imgui.BulletText("Click chevron icons to toggle panels")
			imgui.BulletText("Drag the resize handle to adjust width")
			imgui.BulletText("Use keyboard shortcuts [ and ] to toggle")
			imgui.EndChild()
		}
	})

	app := dfx.New(root, dfx.Config{
		Title:  "HCollapse Demo",
		Width:  900,
		Height: 600,
		OnSetup: func(app *dfx.App) {
			// keyboard shortcuts for toggling panels
			app.Actions().Register("toggle-drums", "[", func() {
				drumsCollapse.Toggle()
			})
			app.Actions().Register("toggle-synths", "]", func() {
				synthsCollapse.Toggle()
			})
			app.Actions().Register("quit", "Ctrl+Q", func() {
				app.Stop()
			})
		},
	})

	app.Run()
}

// createMixerPanel creates the content for a mixer panel with faders and meters
func createMixerPanel(id string, channels []*MixerChannel) dfx.Component {
	return dfx.NewFunc(func(state *dfx.State) {
		// use a table for clean layout - NoClip prevents cell content clipping
		channelWidth := float32(60)
		contentWidth := float32(len(channels)) * channelWidth
		imgui.Dummy(imgui.Vec2{X: 20, Y: 20})

		imgui.Dummy(imgui.Vec2{X: 20, Y: 5})
		imgui.SameLine()

		tableFlags := imgui.TableFlagsNoClip
		if imgui.BeginTableV(id+"_table", int32(len(channels)), tableFlags, imgui.Vec2{X: contentWidth, Y: 0}, 0) {
			// setup columns
			for i := 0; i < len(channels); i++ {
				imgui.TableSetupColumnV(fmt.Sprintf("##col%d", i), imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoClip, channelWidth, 0)
			}

			// row 1: channel names
			imgui.TableNextRow()
			for _, ch := range channels {
				imgui.TableNextColumn()
				imgui.Text(ch.name)
			}

			// row 2: meters
			imgui.TableNextRow()
			for _, ch := range channels {
				imgui.TableNextColumn()
				// center the meter in the column
				meterWidth := ch.meter.Width()
				padding := (channelWidth - meterWidth) / 2
				if padding > 0 {
					imgui.Dummy(imgui.Vec2{X: padding, Y: 0})
					imgui.SameLine()
				}
				ch.meter.Draw(state)
			}

			// row 3: faders
			imgui.TableNextRow()
			for i, ch := range channels {
				imgui.TableNextColumn()

				params := dfx.DefaultFaderParams()
				params.Taper = dfx.AudioTaper()
				params.Height = 150
				params.Width = 20
				params.Format = func(norm float32) string {
					db := norm*72.0 - 60.0
					if db <= -59.9 {
						return "-inf dB"
					}
					return fmt.Sprintf("%.1f dB", db)
				}

				if newValue, changed := dfx.FaderN(fmt.Sprintf("##%s_fader%d", id, i), ch.normalized, params); changed {
					ch.normalized = newValue
				}
			}

			// row 4: values
			imgui.TableNextRow()
			for _, ch := range channels {
				imgui.TableNextColumn()
				db := ch.normalized*72.0 - 60.0
				if db <= -59.9 {
					imgui.Text("-inf")
				} else {
					imgui.Text(fmt.Sprintf("%.0f", db))
				}
			}

			imgui.EndTable()
		}
	})
}

// simulateLevels generates animated meter levels based on fader positions
func simulateLevels(channels []*MixerChannel, t float64, offset int) {
	for i, ch := range channels {
		// base level from fader position
		baseLevel := ch.normalized * 0.7

		// add some variation based on time and channel index
		freq := 1.5 + float64(i+offset)*0.3
		phase := float64(i+offset) * 0.5
		variation := float32(0.15 * math.Sin(t*freq+phase))
		variation += float32(0.1 * math.Sin(t*freq*2.3+phase*1.7))

		level := baseLevel + variation
		if level < 0 {
			level = 0
		}
		if level > 1 {
			level = 1
		}

		// occasional peaks based on fader level
		if ch.normalized > 0.6 && math.Sin(t*0.5+float64(i+offset)*0.8) > 0.97 {
			level = 1.0
		}

		ch.meter.SetLevel(0, level)
	}
}
