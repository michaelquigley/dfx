package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
	"github.com/michaelquigley/dfx/fonts"
)

func main() {
	var currentTheme = 0
	var showMonospace = false
	var sliderValue float32 = 0.5
	var inputText = "Type here..."

	themes := []dfx.Theme{
		dfx.BlueTheme,
		dfx.GreenTheme,
		dfx.RedTheme,
		dfx.PurpleTheme,
		dfx.ModernDark,
	}

	root := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("dfx Theming Example")
		imgui.Separator()

		// theme selector
		imgui.Text("Current Theme: " + themes[currentTheme].Name())

		if imgui.Button("Previous Theme") {
			currentTheme--
			if currentTheme < 0 {
				currentTheme = len(themes) - 1
			}
			dfx.SetTheme(themes[currentTheme])
		}

		imgui.SameLine()
		if imgui.Button("Next Theme") {
			currentTheme++
			if currentTheme >= len(themes) {
				currentTheme = 0
			}
			dfx.SetTheme(themes[currentTheme])
		}

		imgui.Separator()

		// font demonstration
		imgui.Text("Font Examples:")

		// default font with icon
		imgui.Text("Default Font " + string(fonts.ICON_FAVORITE) + " with Material Icon")

		// monospace font toggle
		var changed bool
		showMonospace, changed = dfx.Checkbox("Show Monospace Text", showMonospace)
		if changed {
			// checkbox was toggled
		}

		if showMonospace {
			dfx.PushFont(dfx.MonospaceFont)
			imgui.Text("This is monospace text: 1234567890")
			imgui.Text("function hello() { return 'world'; }")
			dfx.PopFont()
		}

		imgui.Separator()

		// ui controls demonstration
		imgui.Text("UI Controls:")

		sliderValue, _ = dfx.Slider("Slider", sliderValue, 0.0, 1.0)
		inputText, _ = dfx.Input("Input", inputText)

		if imgui.Button("Sample Button") {
			fmt.Println("button clicked with theme:", themes[currentTheme].Name())
		}

		imgui.SameLine()
		if imgui.Button("Another Button") {
			fmt.Println("another button clicked")
		}

		imgui.Separator()

		// colored elements
		imgui.Text("This demonstrates how the theme affects all UI elements:")
		imgui.Text("- Window backgrounds")
		imgui.Text("- Button states (hover, active)")
		imgui.Text("- Text colors")
		imgui.Text("- Border colors")
		imgui.Text("- Input field styling")
	})

	app := dfx.New(root, dfx.Config{
		Title:  "Theming Example",
		Width:  500,
		Height: 600,
		Theme:  &dfx.ModernTheme{}, // start with default theme
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
