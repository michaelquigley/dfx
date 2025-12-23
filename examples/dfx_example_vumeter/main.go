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

	// create waterfall displays
	stereoWaterfall := dfx.NewVUWaterfall(2)
	stereoWaterfall.Height = 200
	stereoWaterfall.ChannelWidth = 30
	stereoWaterfall.RowHeight = 2
	stereoWaterfall.HistorySize = 100

	multiWaterfall := dfx.NewVUWaterfall(8)
	multiWaterfall.Height = 200
	multiWaterfall.ChannelWidth = 12
	multiWaterfall.ChannelGap = 2
	multiWaterfall.RowHeight = 2
	multiWaterfall.HistorySize = 100

	// wrap each meter in an HCollapse
	padding := float32(20) // account for window padding
	singleCollapse := dfx.NewHCollapse(singleMeter, dfx.HCollapseConfig{
		Title:         "Mono",
		ExpandedWidth: singleMeter.Width() + padding,
		Expanded:      true,
		Height:        -1,
	})
	stereoCollapse := dfx.NewHCollapse(stereoMeter, dfx.HCollapseConfig{
		Title:         "Stereo",
		ExpandedWidth: stereoMeter.Width() + padding,
		Expanded:      true,
		Height:        -1,
	})
	multiCollapse := dfx.NewHCollapse(multiMeter, dfx.HCollapseConfig{
		Title:         "Multi (8ch)",
		ExpandedWidth: multiMeter.Width() + padding,
		Expanded:      true,
		Height:        -1,
	})
	stereoWaterfallCollapse := dfx.NewHCollapse(stereoWaterfall, dfx.HCollapseConfig{
		Title:         "Waterfall (2ch)",
		ExpandedWidth: stereoWaterfall.Width() + padding,
		Expanded:      true,
		Height:        -1,
	})
	multiWaterfallCollapse := dfx.NewHCollapse(multiWaterfall, dfx.HCollapseConfig{
		Title:         "Waterfall (8ch)",
		ExpandedWidth: multiWaterfall.Width() + padding,
		Expanded:      true,
		Height:        -1,
	})

	// simulation state
	startTime := time.Now()
	paused := false
	modeIndex := 0
	modeNames := []string{"Solid", "Highres", "Segmented"}

	controlsCollapse := dfx.NewHCollapse(dfx.NewFunc(func(state *dfx.State) {
		if newValue, changed := dfx.Checkbox("Pause Animation", paused); changed {
			paused = newValue
		}

		dfx.Spacing()
		dfx.Text("Display Mode:")
		if newIndex, changed := dfx.Combo("##mode", modeIndex, modeNames); changed {
			modeIndex = newIndex
			mode := dfx.VUMeterMode(newIndex)
			singleMeter.Mode = mode
			stereoMeter.Mode = mode
			multiMeter.Mode = mode
		}

		dfx.Spacing()
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
	}), dfx.HCollapseConfig{
		Title:         "Controls",
		ExpandedWidth: 400,
		Expanded:      true,
		Height:        -1,
	})

	root := dfx.NewFunc(func(state *dfx.State) {
		imgui.BeginChildStr("##dfx_example_vumeter")

		if !paused {
			t := time.Since(startTime).Seconds()
			updateSimulatedLevels(singleMeter, stereoMeter, multiMeter, stereoWaterfall, multiWaterfall, t)
		}

		// layout the meters horizontally in a scrollable child window
		singleCollapse.Draw(state)
		imgui.SameLine()
		stereoCollapse.Draw(state)
		imgui.SameLine()
		multiCollapse.Draw(state)
		imgui.SameLine()
		stereoWaterfallCollapse.Draw(state)
		imgui.SameLine()
		multiWaterfallCollapse.Draw(state)
		imgui.SameLine()
		controlsCollapse.Draw(state)
		imgui.EndChild()
	})

	app := dfx.New(root, dfx.Config{
		Title:  "VU Meter and Waterfall Demo",
		Width:  1200,
		Height: 360,
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
func updateSimulatedLevels(single, stereo, multi *dfx.VUMeter, stereoWf, multiWf *dfx.VUWaterfall, t float64) {
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
	stereoLevels := []float32{leftLevel, rightLevel}
	stereo.SetLevels(stereoLevels)
	stereoWf.SetLevels(stereoLevels)

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
	multiWf.SetLevels(levels)
}
