package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
)

func main() {
	// state for our component
	showExtra := false
	counter := 0

	// create a component with the Container type for more control
	root := &dfx.Container{
		Visible: true,
		OnDraw: func(state *dfx.State) {
			imgui.Text("Action Example - Component Actions")
			imgui.Separator()

			if imgui.Button("Main Button") {
				fmt.Println("main button clicked")
			}

			if showExtra {
				imgui.SameLine()
				if imgui.Button("Extra Button") {
					fmt.Println("extra button clicked")
				}
			}

			imgui.Spacing()
			imgui.Text(fmt.Sprintf("Counter: %d", counter))

			imgui.Spacing()
			imgui.Text("Keyboard shortcuts:")
			imgui.Text("  Ctrl+E - Toggle extra button (component action)")
			imgui.Text("  Ctrl+= - Increment counter (component action)")
			imgui.Text("  Ctrl+- - Decrement counter (component action)")
			imgui.Text("  Ctrl+Q - Quit application (global action)")
		},
	}

	// add component-local actions using the consistent API
	root.Actions().MustRegister("toggle-extra", "Ctrl+E", func() {
		showExtra = !showExtra
		fmt.Printf("toggled extra button: %v\n", showExtra)
	})
	root.Actions().MustRegister("increment", "Ctrl+=", func() {
		counter++
		fmt.Printf("incremented counter to %d\n", counter)
	})
	root.Actions().MustRegister("decrement", "Ctrl+-", func() {
		counter--
		fmt.Printf("decremented counter to %d\n", counter)
	})

	app := dfx.New(root, dfx.Config{
		Title:  "Actions Example",
		Width:  500,
		Height: 400,
	})
	app.Actions().MustRegister("quit", "Ctrl+Q", func() {
		fmt.Println("quitting application")
		app.Stop()
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
