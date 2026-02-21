package main

import (
	"fmt"

	"github.com/michaelquigley/dfx"
	"github.com/AllenDang/cimgui-go/imgui"
)

func main() {
	// application state
	counter := 0
	text := "Hello from dfx menu system!"
	showDialog := false

	// create menu actions
	fileNew := dfx.NewMenuAction("New", "Ctrl+N", func() {
		counter = 0
		text = "New file created"
		fmt.Println("File > New")
	})

	fileOpen := dfx.NewMenuAction("Open", "Ctrl+O", func() {
		text = "Open file dialog would appear here"
		fmt.Println("File > Open")
	})

	fileSave := dfx.NewMenuAction("Save", "Ctrl+S", func() {
		text = fmt.Sprintf("Saved! Counter value: %d", counter)
		fmt.Println("File > Save")
	})

	fileSaveAs := dfx.NewMenuAction("Save As...", "Ctrl+Shift+S", func() {
		text = "Save As dialog would appear here"
		fmt.Println("File > Save As")
	})

	editIncrement := dfx.NewMenuAction("Increment Counter", "Ctrl+=", func() {
		counter++
		text = fmt.Sprintf("Counter: %d", counter)
		fmt.Println("Edit > Increment")
	})

	editDecrement := dfx.NewMenuAction("Decrement Counter", "Ctrl+-", func() {
		counter--
		text = fmt.Sprintf("Counter: %d", counter)
		fmt.Println("Edit > Decrement")
	})

	editReset := dfx.NewMenuAction("Reset Counter", "Ctrl+R", func() {
		counter = 0
		text = "Counter reset to 0"
		fmt.Println("Edit > Reset")
	})

	viewShowDialog := dfx.NewMenuAction("Show Dialog", "Ctrl+D", func() {
		showDialog = !showDialog
		fmt.Printf("View > Show Dialog (now %v)\n", showDialog)
	})

	helpAbout := dfx.NewMenuAction("About", "F1", func() {
		text = "dfx Menu System Example - Demonstrates menu-compatible actions"
		fmt.Println("Help > About")
	})

	// create menu bar component
	// NOTE: dfx.Config.MenuBar already wraps this in BeginMainMenuBar/EndMainMenuBar
	menuBar := dfx.NewFunc(func(state *dfx.State) {
		if imgui.BeginMenu("File") {
			fileNew.DrawMenuItem()
			imgui.Separator()
			fileOpen.DrawMenuItem()
			imgui.Separator()
			fileSave.DrawMenuItem()
			fileSaveAs.DrawMenuItem()
			imgui.EndMenu()
		}

		if imgui.BeginMenu("Edit") {
			editIncrement.DrawMenuItem()
			editDecrement.DrawMenuItem()
			imgui.Separator()
			editReset.DrawMenuItem()
			imgui.EndMenu()
		}

		if imgui.BeginMenu("View") {
			viewShowDialog.DrawMenuItem()
			imgui.EndMenu()
		}

		if imgui.BeginMenu("Help") {
			helpAbout.DrawMenuItem()
			imgui.EndMenu()
		}
	})

	// create main content component
	root := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("Menu System Example")
		imgui.Separator()
		imgui.Spacing()

		imgui.Text(fmt.Sprintf("Counter: %d", counter))
		imgui.Text(fmt.Sprintf("Status: %s", text))

		imgui.Spacing()
		imgui.Separator()
		imgui.Spacing()

		if imgui.Button("Increment (or use Ctrl+=)") {
			editIncrement.Handler()
		}

		imgui.SameLine()
		if imgui.Button("Decrement (or use Ctrl+-)") {
			editDecrement.Handler()
		}

		imgui.SameLine()
		if imgui.Button("Reset (or use Ctrl+R)") {
			editReset.Handler()
		}

		imgui.Spacing()
		imgui.Text("Try using the keyboard shortcuts!")
		imgui.Text("All menu actions work via keyboard and menu clicks.")

		// optional dialog
		if showDialog {
			imgui.Spacing()
			imgui.Separator()
			imgui.Spacing()
			imgui.Text("Dialog is visible!")
			imgui.Text("Press Ctrl+D to toggle this dialog")
		}
	})

	// register all menu actions for keyboard shortcuts
	actions := root.Actions()
	actions.MustRegisterAction(fileNew)
	actions.MustRegisterAction(fileOpen)
	actions.MustRegisterAction(fileSave)
	actions.MustRegisterAction(fileSaveAs)
	actions.MustRegisterAction(editIncrement)
	actions.MustRegisterAction(editDecrement)
	actions.MustRegisterAction(editReset)
	actions.MustRegisterAction(viewShowDialog)
	actions.MustRegisterAction(helpAbout)

	// run the application
	app := dfx.New(root, dfx.Config{
		Title:   "dfx Menu Example",
		Width:   800,
		Height:  600,
		MenuBar: menuBar,
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
