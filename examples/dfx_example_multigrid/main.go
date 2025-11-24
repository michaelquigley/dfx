package main

import (
	"fmt"
	"strconv"

	"github.com/michaelquigley/dfx"
)

func main() {
	// shared application state
	var counter1, counter2, counter3 int
	var textValue string
	var usingFlexLayout = true

	// create sample components
	editorComponent := dfx.NewFunc(func(state *dfx.State) {
		dfx.Text("📝 Text Editor")
		dfx.Separator()
		newText, changed := dfx.Input("Edit Text", textValue)
		if changed {
			textValue = newText
		}
		if textValue != "" {
			dfx.Spacing()
			dfx.Text("Content: " + textValue)
		}
		dfx.Spacing()
		if dfx.Button("Clear") {
			textValue = ""
		}
	})

	sidebarComponent := dfx.NewFunc(func(state *dfx.State) {
		dfx.Text("🔧 Tools")
		dfx.Separator()
		if dfx.Button("Tool 1") {
			counter1++
		}
		dfx.Text("Used: " + strconv.Itoa(counter1) + " times")
		dfx.Spacing()
		if dfx.Button("Tool 2") {
			counter2++
		}
		dfx.Text("Used: " + strconv.Itoa(counter2) + " times")
	})

	terminalComponent := dfx.NewFunc(func(state *dfx.State) {
		dfx.Text("💻 Terminal")
		dfx.Separator()
		dfx.Text("> Commands executed:")
		for i := 0; i < counter3; i++ {
			dfx.Text(fmt.Sprintf("  command_%d completed", i+1))
		}
		if dfx.Button("Run Command") {
			counter3++
		}
		dfx.SameLine()
		if dfx.Button("Clear Log") {
			counter3 = 0
		}
	})

	propertiesComponent := dfx.NewFunc(func(state *dfx.State) {
		dfx.Text("⚙️ Properties")
		dfx.Separator()
		dfx.Text("Layout: " + map[bool]string{true: "Flexible", false: "Fixed Grid"}[usingFlexLayout])
		dfx.Text("Window Size:")
		dfx.Text(fmt.Sprintf("  %.0fx%.0f", state.Size.X, state.Size.Y))
		if len(textValue) > 0 {
			dfx.Text("Text Length: " + strconv.Itoa(len(textValue)))
		}
	})

	// create MultiGrid
	mg := dfx.NewMultiGrid()
	mg.AddComponent("editor", editorComponent)
	mg.AddComponent("sidebar", sidebarComponent)
	mg.AddComponent("terminal", terminalComponent)
	mg.AddComponent("properties", propertiesComponent)

	// create flexible layout (resizable)
	flexLayout := dfx.NewFlexLayout([][]string{
		{"editor", "sidebar"},
		{"terminal", "properties"},
	})

	// create fixed grid layout
	gridLayout := dfx.NewGridLayout(3, 2)
	gridLayout.SetCell("editor", 0, 0, 1, 2)     // row 0, col 0, 1 row, 2 cols
	gridLayout.SetCell("sidebar", 0, 2, 1, 1)    // row 0, col 2, 1 row, 1 col
	gridLayout.SetCell("terminal", 1, 0, 1, 2)   // row 1, col 0, 1 row, 2 cols
	gridLayout.SetCell("properties", 1, 2, 1, 1) // row 1, col 2, 1 row, 1 col

	// start with flexible layout
	mg.SetLayout(flexLayout)

	// add keyboard shortcut to switch layouts
	mg.Actions().Register("toggle_layout", "F1", func() {
		if usingFlexLayout {
			mg.SetLayout(gridLayout)
			usingFlexLayout = false
		} else {
			mg.SetLayout(flexLayout)
			usingFlexLayout = true
		}
	})

	// create menu bar
	menuBar := dfx.NewFunc(func(state *dfx.State) {
		if dfx.BeginMenu("Layout") {
			if dfx.MenuItem("Flexible Layout", "F1") {
				if !usingFlexLayout {
					mg.SetLayout(flexLayout)
					usingFlexLayout = true
				}
			}
			if dfx.MenuItem("Fixed Grid Layout", "F1") {
				if usingFlexLayout {
					mg.SetLayout(gridLayout)
					usingFlexLayout = false
				}
			}
			dfx.EndMenu()
		}
		if dfx.BeginMenu("Actions") {
			if dfx.MenuItem("Reset Counters", "") {
				counter1, counter2, counter3 = 0, 0, 0
			}
			if dfx.MenuItem("Clear Text", "") {
				textValue = ""
			}
			dfx.EndMenu()
		}
		if dfx.BeginMenu("Help") {
			if dfx.MenuItem("About", "") {
				// could show about dialog
			}
			dfx.EndMenu()
		}
	})

	app := dfx.New(mg, dfx.Config{
		Title:   "MultiGrid Example",
		Width:   900,
		Height:  600,
		MenuBar: menuBar,
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
