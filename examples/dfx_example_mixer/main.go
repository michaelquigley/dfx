package main

import (
	"fmt"

	"github.com/michaelquigley/dfx"
	"github.com/AllenDang/cimgui-go/imgui"
)

func main() {
	// fader dimensions
	const faderWidth = 30.0
	const faderHeight = 300.0
	const labelWidth = 80.0

	// fader values (persists between frames)
	faders := []float32{0.0, 1.0, 0.5, 0.75, 0.25}

	root := dfx.NewFunc(func(state *dfx.State) {
		// use table layout for proper alignment of labels and faders
		if imgui.BeginTable("mixer_table", int32(len(faders))) {
			// setup columns with fixed width
			for i := 0; i < len(faders); i++ {
				imgui.TableSetupColumnV(fmt.Sprintf("##col%d", i), imgui.TableColumnFlagsWidthFixed, labelWidth, 0)
			}

			// row 1: channel labels (centered)
			imgui.TableNextRow()
			for i := 0; i < len(faders); i++ {
				imgui.TableNextColumn()
				channelLabel := fmt.Sprintf("Channel %d", i+1)
				labelSize := imgui.CalcTextSize(channelLabel)
				padding := float32((labelWidth - labelSize.X) / 2)
				if padding > 0 {
					imgui.Dummy(imgui.Vec2{X: padding, Y: 1})
					imgui.SameLine()
				}
				dfx.Text(channelLabel)
			}

			// row 2: faders (centered)
			imgui.TableNextRow()
			for i := 0; i < len(faders); i++ {
				imgui.TableNextColumn()
				// center fader in column
				faderPadding := float32((labelWidth - faderWidth) / 2)
				if faderPadding > 0 {
					imgui.Dummy(imgui.Vec2{X: faderPadding, Y: 1})
					imgui.SameLine()
				}
				faderLabel := fmt.Sprintf("##fader%d", i)
				if newValue, changed := dfx.Fader(faderLabel, faders[i], 0.0, faderWidth, faderHeight); changed {
					faders[i] = newValue
					fmt.Printf("channel %d: %.2f\n", i+1, faders[i])
				}
			}

			imgui.EndTable()
		}
	})

	app := dfx.New(root, dfx.Config{
		Title:  "Audio Mixer Example",
		Width:  400,
		Height: 500,
		OnSetup: func(app *dfx.App) {
			app.Actions().Register("quit", "Ctrl+Q", func() {
				fmt.Println("quitting via Ctrl+Q")
				app.Stop()
			})
		},
	})

	app.Run()
}
