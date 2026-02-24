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
		if imgui.BeginChildStrV("sidebar", imgui.Vec2{X: 150, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone) {
			imgui.Text("Select Item:")
			imgui.Separator()

			for i, item := range items {
				if imgui.Button(item) {
					selectedItem = i
				}
			}

			imgui.EndChild()
		}
	})

	// main content component
	content := dfx.NewFunc(func(state *dfx.State) {
		imgui.SameLine()
		if imgui.BeginChildStrV("content", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsNone, imgui.WindowFlagsNone) {
			imgui.Text(fmt.Sprintf("Selected: %s", items[selectedItem]))
			imgui.Separator()

			// text input
			newText, changed := dfx.Input("Text Input", textValue)
			if changed {
				textValue = newText
			}

			if textValue != "" {
				imgui.Text("You typed: " + textValue)
			}

			// checkbox
			newCheck, changed := dfx.Checkbox("Enable Feature", checkValue)
			if changed {
				checkValue = newCheck
			}

			if checkValue {
				imgui.TextColored(imgui.Vec4{X: 0.2, Y: 1.0, Z: 0.2, W: 1.0}, "Feature is enabled!")
			}

			// combo box
			comboOptions := []string{"Option A", "Option B", "Option C"}
			var comboIndex int
			comboIndex, _ = dfx.Combo("Select Option", comboIndex, comboOptions)

			// color picker
			imgui.Spacing()
			imgui.Text("Pick a color:")
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

			imgui.EndChild()
		}
	})

	// create menu bar component
	menuBar := dfx.NewFunc(func(state *dfx.State) {
		if imgui.BeginMenu("File") {
			if imgui.MenuItemBoolV("New", "Ctrl+N", false, true) {
				fmt.Println("new file")
			}
			if imgui.MenuItemBoolV("Open", "Ctrl+O", false, true) {
				fmt.Println("open file")
			}
			imgui.Separator()
			if imgui.MenuItemBoolV("Exit", "", false, true) {
				state.App.Stop()
			}
			imgui.EndMenu()
		}
		if imgui.BeginMenu("Edit") {
			if imgui.MenuItemBoolV("Reset", "", false, true) {
				selectedItem = 0
				textValue = ""
				checkValue = false
				colorR, colorG, colorB = 0.5, 0.5, 0.5
			}
			imgui.EndMenu()
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
