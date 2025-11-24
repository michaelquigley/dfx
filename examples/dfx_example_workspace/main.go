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
		dfx.Text("Code Editor")
		dfx.Separator()
		dfx.Spacing()

		// multiline text editor
		newText, changed := dfx.InputMultiline("##editor", editorText, state.Size.X-20, state.Size.Y-80)
		if changed {
			editorText = newText
		}

		// status bar
		dfx.Separator()
		dfx.Text(fmt.Sprintf("lines: %d | chars: %d", countLines(editorText), len(editorText)))
	})

	// create viewer workspace
	viewer := dfx.NewFunc(func(state *dfx.State) {
		dfx.Text("Content Viewer")
		dfx.Separator()
		dfx.Spacing()

		// zoom control
		newScale, changed := dfx.Slider("Zoom", viewerScale, 0.5, 3.0)
		if changed {
			viewerScale = newScale
		}

		dfx.Spacing()
		dfx.Separator()

		// simulate viewing content with colored rectangles
		if dfx.BeginChild("viewer_content", 0, 0, false) {
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
			dfx.Text(fmt.Sprintf("viewing at %.1fx zoom", viewerScale))

			dfx.EndChild()
		}
	})

	// create settings workspace
	settings := dfx.NewFunc(func(state *dfx.State) {
		dfx.Text("Application Settings")
		dfx.Separator()
		dfx.Spacing()

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

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		// display current settings
		dfx.Text("Current Configuration:")
		dfx.Text(fmt.Sprintf("  username: '%s'", settingsUsername))
		dfx.Text(fmt.Sprintf("  dark mode: %v", settingsDarkMode))
		dfx.Text(fmt.Sprintf("  volume: %.0f%%", settingsVolume))

		dfx.Spacing()
		if dfx.Button("Reset to Defaults") {
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
		if dfx.BeginMenu("File") {
			if dfx.MenuItem("Exit", "") {
				state.App.Stop()
			}
			dfx.EndMenu()
		}

		if dfx.BeginMenu("Workspace") {
			// use display names in menu but switch by Id
			if dfx.MenuItem(fonts.ICON_ADDCHART+" Editor", "Ctrl+1") {
				ws.Switch("editor")
			}
			if dfx.MenuItem(fonts.ICON_PANORAMA_FISH_EYE+" Viewer", "Ctrl+2") {
				ws.Switch("viewer")
			}
			if dfx.MenuItem(fonts.ICON_SETTINGS+" Settings", "Ctrl+3") {
				ws.Switch("settings")
			}
			dfx.EndMenu()
		}

		if dfx.BeginMenu("Help") {
			if dfx.MenuItem("About", "") {
				fmt.Println("workspace example - demonstrates workspace switching")
				fmt.Printf("current workspace: '%s' (%s)\n", ws.Current(), ws.CurrentName())
			}
			dfx.EndMenu()
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
