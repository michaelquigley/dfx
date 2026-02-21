package main

import (
	"fmt"

	"github.com/michaelquigley/dfx"
	"github.com/michaelquigley/dfx/fonts"
	"github.com/AllenDang/cimgui-go/imgui"
)

func main() {
	// shared application state
	var editorText = "// edit your code here\npackage main\n\nfunc main() {\n    fmt.Println(\"oh, wow!\")\n}"
	var viewerScale float32 = 1.0
	var settingsUsername = "user"
	var settingsDarkMode = true
	var settingsVolume float32 = 50.0

	// create editor workspace
	editor := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("Code Editor")
		imgui.Separator()
		imgui.Spacing()

		// multiline text editor
		newText, changed := dfx.InputMultiline("##editor", editorText, state.Size.X-20, state.Size.Y-80)
		if changed {
			editorText = newText
		}

		// status bar
		imgui.Separator()
		imgui.Text(fmt.Sprintf("lines: %d | chars: %d", countLines(editorText), len(editorText)))
	})

	// create viewer workspace
	viewer := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("Content Viewer")
		imgui.Separator()
		imgui.Spacing()

		// zoom control
		newScale, changed := dfx.Slider("Zoom", viewerScale, 0.5, 3.0)
		if changed {
			viewerScale = newScale
		}

		imgui.Spacing()
		imgui.Separator()

		// simulate viewing content with colored rectangles
		if imgui.BeginChildStrV("viewer_content", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsNone, imgui.WindowFlagsNone) {
			drawList := imgui.WindowDrawList()
			pos := imgui.CursorScreenPos()

			// draw scaled rectangles to simulate content
			baseSize := float32(50) * viewerScale
			spacing := float32(10)
			colors := []imgui.Vec4{
				{X: 0.8, Y: 0.2, Z: 0.2, W: 1.0}, // red
				{X: 0.2, Y: 0.8, Z: 0.2, W: 1.0}, // green
				{X: 0.2, Y: 0.2, Z: 0.8, W: 1.0}, // blue
				{X: 0.8, Y: 0.8, Z: 0.2, W: 1.0}, // yellow
			}

			for i, col := range colors {
				offset := float32(i) * (baseSize + spacing)
				rectPos := pos.Add(imgui.Vec2{X: offset, Y: 0})
				rectEnd := rectPos.Add(imgui.Vec2{X: baseSize, Y: baseSize})
				drawList.AddRectFilled(rectPos, rectEnd, imgui.ColorConvertFloat4ToU32(col))
			}

			// advance cursor
			imgui.Dummy(imgui.Vec2{X: 0, Y: baseSize + spacing})
			imgui.Text(fmt.Sprintf("viewing at %.1fx zoom", viewerScale))

			imgui.EndChild()
		}
	})

	// create settings workspace
	settings := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("Application Settings")
		imgui.Separator()
		imgui.Spacing()

		// username input
		newUsername, changed := dfx.Input("Username", settingsUsername)
		if changed {
			settingsUsername = newUsername
		}

		// dark mode toggle
		newDarkMode, changed := dfx.Checkbox("Dark Mode", settingsDarkMode)
		if changed {
			settingsDarkMode = newDarkMode
		}

		// volume slider
		newVolume, changed := dfx.Slider("Volume", settingsVolume, 0, 100)
		if changed {
			settingsVolume = newVolume
		}

		imgui.Spacing()
		imgui.Separator()
		imgui.Spacing()

		// display current settings
		imgui.Text("Current Configuration:")
		imgui.Text(fmt.Sprintf("  username: '%s'", settingsUsername))
		imgui.Text(fmt.Sprintf("  dark mode: %v", settingsDarkMode))
		imgui.Text(fmt.Sprintf("  volume: %.0f%%", settingsVolume))

		imgui.Spacing()
		if imgui.Button("Reset to Defaults") {
			settingsUsername = "user"
			settingsDarkMode = true
			settingsVolume = 50.0
		}
	})

	// create workspace manager with stable IDs and display names (with icons)
	ws := dfx.NewWorkspace()
	ws.Add("editor", fonts.ICON_ADDCHART+" Editor", editor)
	ws.Add("viewer", fonts.ICON_PANORAMA_FISH_EYE+" Viewer", viewer)
	ws.Add("settings", fonts.ICON_SETTINGS+" Settings", settings)
	ws.ShowSelector = true
	ws.SelectorLabel = "Workspace"
	ws.SelectorWidth = 200

	// add switch callback (receives stable IDs)
	ws.OnSwitch = func(oldID, newID string) {
		fmt.Printf("switched from '%s' to '%s'\n", oldID, newID)
		// display names can be retrieved if needed
		fmt.Printf("  (display: '%s' -> '%s')\n", ws.GetName(oldID), ws.GetName(newID))
	}

	// add keyboard shortcuts for switching workspaces (use stable IDs)
	ws.Actions().MustRegister("Switch to Editor", "Ctrl+1", func() {
		ws.Switch("editor") // stable Id, won't break if display name changes
	})
	ws.Actions().MustRegister("Switch to Viewer", "Ctrl+2", func() {
		ws.Switch("viewer")
	})
	ws.Actions().MustRegister("Switch to Settings", "Ctrl+3", func() {
		ws.Switch("settings")
	})

	// demonstrate changing a display name (Id stays the same)
	// ws.SetName("editor", "✏️ Code Editor") // would update the display name

	// create menu bar
	menuBar := dfx.NewFunc(func(state *dfx.State) {
		if imgui.BeginMenu("File") {
			if imgui.MenuItemBoolV("Exit", "", false, true) {
				state.App.Stop()
			}
			imgui.EndMenu()
		}

		if imgui.BeginMenu("Workspace") {
			// use display names in menu but switch by Id
			if imgui.MenuItemBoolV(fonts.ICON_ADDCHART+" Editor", "Ctrl+1", false, true) {
				ws.Switch("editor")
			}
			if imgui.MenuItemBoolV(fonts.ICON_PANORAMA_FISH_EYE+" Viewer", "Ctrl+2", false, true) {
				ws.Switch("viewer")
			}
			if imgui.MenuItemBoolV(fonts.ICON_SETTINGS+" Settings", "Ctrl+3", false, true) {
				ws.Switch("settings")
			}
			imgui.EndMenu()
		}

		if imgui.BeginMenu("Help") {
			if imgui.MenuItemBoolV("About", "", false, true) {
				fmt.Println("workspace example - demonstrates workspace switching")
				fmt.Printf("current workspace: '%s' (%s)\n", ws.Current(), ws.CurrentName())
			}
			imgui.EndMenu()
		}
	})

	// create and run application
	app := dfx.New(ws, dfx.Config{
		Title:   "Workspace Example",
		Width:   800,
		Height:  600,
		MenuBar: menuBar,
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}

// countLines counts newlines in a string
func countLines(s string) int {
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}
