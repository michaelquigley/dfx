package main

import (
	"fmt"
	"math"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
)

func main() {
	// create meters with different channel counts
	singleMeter := dfx.NewVUMeter(1)
	singleMeter.SetLabel(0, "M")

	stereoMeter := dfx.NewVUMeter(2)
	stereoMeter.SetLabels([]string{"L", "R"})

	multiMeter := dfx.NewVUMeter(8)
	multiMeter.SetLabels([]string{"1", "2", "3", "4", "5", "6", "7", "8"})
	multiMeter.Height = 250

	// simulation state
	startTime := time.Now()
	paused := false

	root := dfx.NewFunc(func(state *dfx.State) {
		dfx.Text("VU Meter Demo")
		dfx.Separator()
		dfx.Spacing()

		// pause control
		if newValue, changed := dfx.Checkbox("Pause Animation", paused); changed {
			paused = newValue
		}
		dfx.Spacing()

		// simulate audio levels if not paused
		if !paused {
			t := time.Since(startTime).Seconds()
			updateSimulatedLevels(singleMeter, stereoMeter, multiMeter, t)
		}

		// layout the meters horizontally
		dfx.Text("Single Channel:")
		singleMeter.Draw(state)

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		dfx.Text("Stereo (2 channels):")
		stereoMeter.Draw(state)

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		dfx.Text("Multi-channel (8 channels):")
		multiMeter.Draw(state)

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		// configuration section
		dfx.Text("Configuration:")
		if imgui.TreeNodeStr("Meter Settings") {
			dfx.Text("Single Meter:")
			if newValue, changed := dfx.WheelSlider("Height##single", singleMeter.Height, 100, 400, 50, "%.0f", imgui.SliderFlagsNone); changed {
				singleMeter.Height = newValue
			}
			if newValue, changed := dfx.WheelSlider("Channel Width##single", singleMeter.ChannelWidth, 8, 32, 50, "%.0f", imgui.SliderFlagsNone); changed {
				singleMeter.ChannelWidth = newValue
			}
			if newValue, changed := dfx.WheelSlider("Segment Count##single", float32(singleMeter.SegmentCount), 10, 40, 30, "%.0f", imgui.SliderFlagsNone); changed {
				singleMeter.SegmentCount = int(newValue)
			}

			dfx.Spacing()
			dfx.Text("Peak Hold:")
			if newValue, changed := dfx.WheelSlider("Peak Hold (ms)", float32(singleMeter.PeakHoldMs), 0, 3000, 50, "%.0f", imgui.SliderFlagsNone); changed {
				singleMeter.PeakHoldMs = int(newValue)
				stereoMeter.PeakHoldMs = int(newValue)
				multiMeter.PeakHoldMs = int(newValue)
			}
			if newValue, changed := dfx.WheelSlider("Peak Decay Rate", singleMeter.PeakDecayRate, 0.1, 2.0, 50, "%.2f", imgui.SliderFlagsNone); changed {
				singleMeter.PeakDecayRate = newValue
				stereoMeter.PeakDecayRate = newValue
				multiMeter.PeakDecayRate = newValue
			}

			dfx.Spacing()
			dfx.Text("Clip Indicator:")
			if newValue, changed := dfx.WheelSlider("Clip Hold (ms)", float32(singleMeter.ClipHoldMs), 500, 5000, 50, "%.0f", imgui.SliderFlagsNone); changed {
				singleMeter.ClipHoldMs = int(newValue)
				stereoMeter.ClipHoldMs = int(newValue)
				multiMeter.ClipHoldMs = int(newValue)
			}

			imgui.TreePop()
		}

		dfx.Spacing()
		dfx.Text("Tip: Animation occasionally clips to demonstrate clip indicator")
	})

	app := dfx.New(root, dfx.Config{
		Title:  "VU Meter Demo",
		Width:  400,
		Height: 800,
		OnSetup: func(app *dfx.App) {
			app.Actions().Register("quit", "Ctrl+Q", func() {
				fmt.Println("quitting via Ctrl+Q")
				app.Stop()
			})
		},
	})

	app.Run()
}

// updateSimulatedLevels generates animated audio levels for demonstration.
func updateSimulatedLevels(single, stereo, multi *dfx.VUMeter, t float64) {
	// single channel: smooth sine wave with occasional peaks
	singleLevel := float32(0.3 + 0.3*math.Sin(t*2.0) + 0.2*math.Sin(t*5.0))
	if math.Sin(t*0.5) > 0.95 {
		singleLevel = 1.0 // occasional clip
	}
	single.SetLevel(0, singleLevel)

	// stereo: slightly different phase for L and R
	leftLevel := float32(0.3 + 0.35*math.Sin(t*2.5) + 0.15*math.Sin(t*7.0))
	rightLevel := float32(0.3 + 0.35*math.Sin(t*2.5+0.5) + 0.15*math.Sin(t*6.0))
	if math.Sin(t*0.7) > 0.97 {
		leftLevel = 1.0 // occasional clip on left
	}
	if math.Sin(t*0.8) > 0.97 {
		rightLevel = 1.0 // occasional clip on right
	}
	stereo.SetLevels([]float32{leftLevel, rightLevel})

	// multi-channel: different frequencies for each channel
	levels := make([]float32, 8)
	for i := 0; i < 8; i++ {
		freq := 1.5 + float64(i)*0.3
		phase := float64(i) * 0.4
		levels[i] = float32(0.2 + 0.4*math.Sin(t*freq+phase) + 0.2*math.Sin(t*freq*2.0+phase))
		// occasional clips on different channels
		if math.Sin(t*0.3+float64(i)*0.7) > 0.98 {
			levels[i] = 1.0
		}
	}
	multi.SetLevels(levels)
}
