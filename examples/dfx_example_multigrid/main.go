package main

import (
	"fmt"
	"strconv"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
)

func main() {
	// shared application state
	var counter1, counter2, counter3 int
	var textValue string
	var usingFlexLayout = true

	// create sample components
	editorComponent := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("Text Editor")
		imgui.Separator()
		newText, changed := dfx.Input("Edit Text", textValue)
		if changed {
			textValue = newText
		}
		if textValue != "" {
			imgui.Spacing()
			imgui.Text("Content: " + textValue)
		}
		imgui.Spacing()
		if imgui.Button("Clear") {
			textValue = ""
		}
	})

	sidebarComponent := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("Tools")
		imgui.Separator()
		if imgui.Button("Tool 1") {
			counter1++
		}
		imgui.Text("Used: " + strconv.Itoa(counter1) + " times")
		imgui.Spacing()
		if imgui.Button("Tool 2") {
			counter2++
		}
		imgui.Text("Used: " + strconv.Itoa(counter2) + " times")
	})

	terminalComponent := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("Terminal")
		imgui.Separator()
		imgui.Text("> Commands executed:")
		for i := 0; i < counter3; i++ {
			imgui.Text(fmt.Sprintf("  command_%d completed", i+1))
		}
		if imgui.Button("Run Command") {
			counter3++
		}
		imgui.SameLine()
		if imgui.Button("Clear Log") {
			counter3 = 0
		}
	})

	propertiesComponent := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("Properties")
		imgui.Separator()
		imgui.Text("Layout: " + map[bool]string{true: "Flexible", false: "Fixed Grid"}[usingFlexLayout])
		imgui.Text("Window Size:")
		imgui.Text(fmt.Sprintf("  %.0fx%.0f", state.Size.X, state.Size.Y))
		if len(textValue) > 0 {
			imgui.Text("Text Length: " + strconv.Itoa(len(textValue)))
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
		if imgui.BeginMenu("Layout") {
			if imgui.MenuItemBoolV("Flexible Layout", "F1", false, true) {
				if !usingFlexLayout {
					mg.SetLayout(flexLayout)
					usingFlexLayout = true
				}
			}
			if imgui.MenuItemBoolV("Fixed Grid Layout", "F1", false, true) {
				if usingFlexLayout {
					mg.SetLayout(gridLayout)
					usingFlexLayout = false
				}
			}
			imgui.EndMenu()
		}
		if imgui.BeginMenu("Actions") {
			if imgui.MenuItemBoolV("Reset Counters", "", false, true) {
				counter1, counter2, counter3 = 0, 0, 0
			}
			if imgui.MenuItemBoolV("Clear Text", "", false, true) {
				textValue = ""
			}
			imgui.EndMenu()
		}
		if imgui.BeginMenu("Help") {
			if imgui.MenuItemBoolV("About", "", false, true) {
				// could show about dialog
			}
			imgui.EndMenu()
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
