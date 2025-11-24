package main

import (
	"fmt"

	"github.com/michaelquigley/dfx"
	"github.com/michaelquigley/dfx/fonts"
	"github.com/AllenDang/cimgui-go/imgui"
)

type state struct {
	// combo state
	selectedOption int
	options        []string

	// toggle states
	playEnabled   bool
	mouseTracking bool
	loopEnabled   bool

	// wheel slider states
	zoom        float32
	volume      float32
	speed       float32
	temperature float32
}

func main() {
	s := &state{
		selectedOption: 0,
		options:        []string{"Option 1", "Option 2", "Option 3", "Option 4"},
		playEnabled:    false,
		mouseTracking:  true,
		loopEnabled:    false,
		zoom:           1.0,
		volume:         0.5,
		speed:          1.0,
		temperature:    0.7,
	}

	root := dfx.NewFunc(func(state *dfx.State) {
		dfx.Text("dfx Control Wrappers Demo")
		dfx.Separator()
		dfx.Spacing()

		// Combo / IndexedCombo demo
		dfx.Text("Combo:")
		if newIndex, changed := dfx.Combo("Select Option", s.selectedOption, s.options); changed {
			s.selectedOption = newIndex
			fmt.Printf("combo changed to index %d: '%s'\n", newIndex, s.options[newIndex])
		}
		dfx.SameLine()
		dfx.Text(fmt.Sprintf("Current: %s", s.options[s.selectedOption]))

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		// Toggle demo
		dfx.Text("Toggle - Boolean buttons with visual feedback:")

		// Play toggle with icon
		if newValue, changed := dfx.Toggle(fonts.ICON_PLAY_ARROW, s.playEnabled); changed {
			s.playEnabled = newValue
			fmt.Printf("play enabled: %v\n", s.playEnabled)
		}
		dfx.SameLine()
		dfx.Text(fmt.Sprintf("Play: %v", s.playEnabled))

		// Mouse tracking toggle with icon
		if newValue, changed := dfx.Toggle(fonts.ICON_MOUSE, s.mouseTracking); changed {
			s.mouseTracking = newValue
			fmt.Printf("mouse tracking: %v\n", s.mouseTracking)
		}
		dfx.SameLine()
		dfx.Text(fmt.Sprintf("Mouse Tracking: %v", s.mouseTracking))

		// Loop toggle with icon
		if newValue, changed := dfx.Toggle(fonts.ICON_LOOP, s.loopEnabled); changed {
			s.loopEnabled = newValue
			fmt.Printf("loop enabled: %v\n", s.loopEnabled)
		}
		dfx.SameLine()
		dfx.Text(fmt.Sprintf("Loop: %v", s.loopEnabled))

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		// WheelSlider demo
		dfx.Text("WheelSlider - Hover and use mouse wheel to adjust")
		dfx.Text("(Ctrl = 10x faster, Alt = 10x slower)")
		dfx.Spacing()

		// Zoom slider
		if newValue, changed := dfx.WheelSlider("Zoom", s.zoom, 0.25, 5.0, 100, "%.2f", imgui.SliderFlagsNone); changed {
			s.zoom = newValue
			fmt.Printf("zoom: %.2f\n", s.zoom)
		}

		// Volume slider
		if newValue, changed := dfx.WheelSlider("Volume", s.volume, 0.0, 1.0, 100, "%.2f", imgui.SliderFlagsNone); changed {
			s.volume = newValue
			fmt.Printf("volume: %.2f\n", s.volume)
		}

		// Speed slider
		if newValue, changed := dfx.WheelSlider("Speed", s.speed, 0.1, 3.0, 100, "%.2fx", imgui.SliderFlagsNone); changed {
			s.speed = newValue
			fmt.Printf("speed: %.2f\n", s.speed)
		}

		// Temperature slider (for AI/ML demos)
		if newValue, changed := dfx.WheelSlider("Temperature", s.temperature, 0.0, 2.0, 100, "%.2f", imgui.SliderFlagsNone); changed {
			s.temperature = newValue
			fmt.Printf("temperature: %.2f\n", s.temperature)
		}

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		// Standard controls for comparison
		dfx.Text("Standard Controls (for comparison):")

		// Standard button
		if dfx.Button("Standard Button") {
			fmt.Println("standard button clicked")
		}
		dfx.SameLine()

		// Standard checkbox
		if newValue, changed := dfx.Checkbox("Checkbox", s.loopEnabled); changed {
			s.loopEnabled = newValue
		}
		dfx.SameLine()

		// Standard slider
		_, _ = dfx.Slider("Standard Slider", s.volume, 0.0, 1.0)
	})

	app := dfx.New(root, dfx.Config{
		Title:  "Control Wrappers Demo",
		Width:  700,
		Height: 600,
		OnSetup: func(app *dfx.App) {
			app.Actions().Register("quit", "Ctrl+Q", func() {
				fmt.Println("quitting via Ctrl+Q")
				app.Stop()
			})
		},
	})

	app.Run()
}
