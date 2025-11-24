package main

import (
	"fmt"

	"github.com/michaelquigley/dfx"
	"github.com/AllenDang/cimgui-go/imgui"
)

func main() {
	// shared state
	var selectedItem int
	var textValue string
	var checkValue bool
	var colorR, colorG, colorB float32 = 0.5, 0.5, 0.5

	items := []string{"First Item", "Second Item", "Third Item", "Fourth Item"}

	// sidebar component
	sidebar := dfx.NewFunc(func(state *dfx.State) {
		if dfx.BeginChild("sidebar", 150, 0, true) {
			dfx.Text("Select Item:")
			dfx.Separator()

			for i, item := range items {
				if dfx.Button(item) {
					selectedItem = i
				}
			}

			dfx.EndChild()
		}
	})

	// main content component
	content := dfx.NewFunc(func(state *dfx.State) {
		dfx.SameLine()
		if dfx.BeginChild("content", 0, 0, false) {
			dfx.Text(fmt.Sprintf("Selected: %s", items[selectedItem]))
			dfx.Separator()

			// text input
			newText, changed := dfx.Input("Text Input", textValue)
			if changed {
				textValue = newText
			}

			if textValue != "" {
				dfx.Text("You typed: " + textValue)
			}

			// checkbox
			newCheck, changed := dfx.Checkbox("Enable Feature", checkValue)
			if changed {
				checkValue = newCheck
			}

			if checkValue {
				dfx.TextColored("Feature is enabled!", 0.2, 1.0, 0.2, 1.0)
			}

			// combo box
			comboOptions := []string{"Option A", "Option B", "Option C"}
			var comboIndex int
			comboIndex, _ = dfx.Combo("Select Option", comboIndex, comboOptions)

			// color picker
			dfx.Spacing()
			dfx.Text("Pick a color:")
			newR, newG, newB, changed := dfx.ColorEdit3("Color", colorR, colorG, colorB)
			if changed {
				colorR, colorG, colorB = newR, newG, newB
			}

			// show colored rectangle
			drawList := imgui.WindowDrawList()
			pos := imgui.CursorScreenPos()
			col := imgui.ColorConvertFloat4ToU32(imgui.Vec4{X: colorR, Y: colorG, Z: colorB, W: 1.0})
			drawList.AddRectFilled(pos, pos.Add(imgui.Vec2{X: 100, Y: 50}), col)
			imgui.Dummy(imgui.Vec2{X: 100, Y: 50})

			dfx.EndChild()
		}
	})

	// create menu bar component
	menuBar := dfx.NewFunc(func(state *dfx.State) {
		if dfx.BeginMenu("File") {
			if dfx.MenuItem("New", "Ctrl+N") {
				fmt.Println("new file")
			}
			if dfx.MenuItem("Open", "Ctrl+O") {
				fmt.Println("open file")
			}
			dfx.Separator()
			if dfx.MenuItem("Exit", "") {
				state.App.Stop()
			}
			dfx.EndMenu()
		}
		if dfx.BeginMenu("Edit") {
			if dfx.MenuItem("Reset", "") {
				selectedItem = 0
				textValue = ""
				checkValue = false
				colorR, colorG, colorB = 0.5, 0.5, 0.5
			}
			dfx.EndMenu()
		}
	})

	// compose main content
	root := &dfx.Container{
		Visible:  true,
		Children: []dfx.Component{sidebar, content},
	}

	app := dfx.New(root, dfx.Config{
		Title:   "Composition Example",
		Width:   700,
		Height:  500,
		MenuBar: menuBar,
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
