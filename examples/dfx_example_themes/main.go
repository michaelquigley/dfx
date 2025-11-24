package main

import (
	"fmt"

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
		dfx.Text("dfx Theming Example")
		dfx.Separator()

		// theme selector
		dfx.Text("Current Theme: " + themes[currentTheme].Name())

		if dfx.Button("Previous Theme") {
			currentTheme--
			if currentTheme < 0 {
				currentTheme = len(themes) - 1
			}
			dfx.SetTheme(themes[currentTheme])
		}

		dfx.SameLine()
		if dfx.Button("Next Theme") {
			currentTheme++
			if currentTheme >= len(themes) {
				currentTheme = 0
			}
			dfx.SetTheme(themes[currentTheme])
		}

		dfx.Separator()

		// font demonstration
		dfx.Text("Font Examples:")

		// default font with icon
		dfx.Text("Default Font " + string(fonts.ICON_FAVORITE) + " with Material Icon")

		// monospace font toggle
		var changed bool
		showMonospace, changed = dfx.Checkbox("Show Monospace Text", showMonospace)
		if changed {
			// checkbox was toggled
		}

		if showMonospace {
			dfx.PushFont(dfx.MonospaceFont)
			dfx.Text("This is monospace text: 1234567890")
			dfx.Text("function hello() { return 'world'; }")
			dfx.PopFont()
		}

		dfx.Separator()

		// ui controls demonstration
		dfx.Text("UI Controls:")

		sliderValue, _ = dfx.Slider("Slider", sliderValue, 0.0, 1.0)
		inputText, _ = dfx.Input("Input", inputText)

		if dfx.Button("Sample Button") {
			fmt.Println("button clicked with theme:", themes[currentTheme].Name())
		}

		dfx.SameLine()
		if dfx.Button("Another Button") {
			fmt.Println("another button clicked")
		}

		dfx.Separator()

		// colored elements
		dfx.Text("This demonstrates how the theme affects all UI elements:")
		dfx.Text("- Window backgrounds")
		dfx.Text("- Button states (hover, active)")
		dfx.Text("- Text colors")
		dfx.Text("- Border colors")
		dfx.Text("- Input field styling")
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
