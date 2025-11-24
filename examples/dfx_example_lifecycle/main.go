package main

import (
	"fmt"

	"github.com/michaelquigley/dfx"
	"github.com/AllenDang/cimgui-go/imgui"
)

type state struct {
	windowWidth     int
	windowHeight    int
	windowX         int
	windowY         int
	resizeCount     int
	titleCounter    int
	showCloseDialog bool
	unsavedChanges  bool
}

func main() {
	s := &state{
		unsavedChanges: false,
	}

	root := dfx.NewFunc(func(state *dfx.State) {
		dfx.Text("Window Lifecycle Demo")
		dfx.Separator()
		dfx.Spacing()

		// window information display
		dfx.Text("Window Information:")
		dfx.Text(fmt.Sprintf("  Size: %d x %d", s.windowWidth, s.windowHeight))
		dfx.Text(fmt.Sprintf("  Position: (%d, %d)", s.windowX, s.windowY))
		dfx.Text(fmt.Sprintf("  Resize count: %d", s.resizeCount))

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		// dynamic title update
		dfx.Text("Dynamic Title:")
		if dfx.Button("Update Window Title") {
			s.titleCounter++
			newTitle := fmt.Sprintf("Lifecycle Demo - Count: %d", s.titleCounter)
			state.App.SetWindowTitle(newTitle)
		}

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		// unsaved changes simulation
		dfx.Text("Close Behavior:")
		if newValue, changed := dfx.Checkbox("Simulate unsaved changes", s.unsavedChanges); changed {
			s.unsavedChanges = newValue
		}
		dfx.Text("(Try closing the window with unsaved changes)")

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		// window position controls
		dfx.Text("Window Controls:")
		if dfx.Button("Get Current Position") {
			s.windowX, s.windowY = state.App.GetWindowPos()
		}
		dfx.SameLine()
		if dfx.Button("Get Current Size") {
			s.windowWidth, s.windowHeight = state.App.GetWindowSize()
		}

		// close confirmation dialog
		if s.showCloseDialog {
			imgui.OpenPopupStr("Confirm Close")
			s.showCloseDialog = false
		}

		// modal dialog
		if imgui.BeginPopupModalV("Confirm Close", nil, imgui.WindowFlagsAlwaysAutoResize) {
			dfx.Text("You have unsaved changes!")
			dfx.Text("Are you sure you want to close?")
			dfx.Spacing()

			if dfx.Button("Close Anyway") {
				s.unsavedChanges = false
				state.App.SetShouldClose(true)
				imgui.CloseCurrentPopup()
			}
			dfx.SameLine()
			if dfx.Button("Cancel") {
				// window will not close
				imgui.CloseCurrentPopup()
			}

			imgui.EndPopup()
		}
	})

	app := dfx.New(root, dfx.Config{
		Title:  "Lifecycle Demo",
		Width:  600,
		Height: 400,
		OnSetup: func(app *dfx.App) {
			// initialize window info on startup
			s.windowWidth, s.windowHeight = app.GetWindowSize()
			s.windowX, s.windowY = app.GetWindowPos()

			// register quit action
			app.Actions().Register("quit", "Ctrl+Q", func() {
				fmt.Println("quitting via Ctrl+Q")
				app.Stop()
			})
		},
		OnClose: func(app *dfx.App) {
			fmt.Println("window close requested")
			if s.unsavedChanges {
				// cancel the close and show confirmation dialog
				app.SetShouldClose(false)
				s.showCloseDialog = true
			}
		},
		OnSizeChange: func(width, height int) {
			fmt.Printf("window resized to %d x %d\n", width, height)
			s.windowWidth = width
			s.windowHeight = height
			s.resizeCount++
		},
	})

	app.Run()
}
