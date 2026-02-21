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
		imgui.Text("Window Lifecycle Demo")
		imgui.Separator()
		imgui.Spacing()

		// window information display
		imgui.Text("Window Information:")
		imgui.Text(fmt.Sprintf("  Size: %d x %d", s.windowWidth, s.windowHeight))
		imgui.Text(fmt.Sprintf("  Position: (%d, %d)", s.windowX, s.windowY))
		imgui.Text(fmt.Sprintf("  Resize count: %d", s.resizeCount))

		imgui.Spacing()
		imgui.Separator()
		imgui.Spacing()

		// dynamic title update
		imgui.Text("Dynamic Title:")
		if imgui.Button("Update Window Title") {
			s.titleCounter++
			newTitle := fmt.Sprintf("Lifecycle Demo - Count: %d", s.titleCounter)
			state.App.SetWindowTitle(newTitle)
		}

		imgui.Spacing()
		imgui.Separator()
		imgui.Spacing()

		// unsaved changes simulation
		imgui.Text("Close Behavior:")
		if newValue, changed := dfx.Checkbox("Simulate unsaved changes", s.unsavedChanges); changed {
			s.unsavedChanges = newValue
		}
		imgui.Text("(Try closing the window with unsaved changes)")

		imgui.Spacing()
		imgui.Separator()
		imgui.Spacing()

		// window position controls
		imgui.Text("Window Controls:")
		if imgui.Button("Get Current Position") {
			s.windowX, s.windowY = state.App.GetWindowPos()
		}
		imgui.SameLine()
		if imgui.Button("Get Current Size") {
			s.windowWidth, s.windowHeight = state.App.GetWindowSize()
		}

		// close confirmation dialog
		if s.showCloseDialog {
			imgui.OpenPopupStr("Confirm Close")
			s.showCloseDialog = false
		}

		// modal dialog
		if imgui.BeginPopupModalV("Confirm Close", nil, imgui.WindowFlagsAlwaysAutoResize) {
			imgui.Text("You have unsaved changes!")
			imgui.Text("Are you sure you want to close?")
			imgui.Spacing()

			if imgui.Button("Close Anyway") {
				s.unsavedChanges = false
				state.App.SetShouldClose(true)
				imgui.CloseCurrentPopup()
			}
			imgui.SameLine()
			if imgui.Button("Cancel") {
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
