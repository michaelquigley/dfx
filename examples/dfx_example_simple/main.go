package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
)

func main() {
	clickCount := 0

	// create a simple component using FuncWithActions for keyboard support
	root := dfx.NewFunc(func(state *dfx.State) {
		imgui.Text("Hello from dfx!")

		if imgui.Button("Click Me") {
			clickCount++
			fmt.Printf("button clicked %d times!\n", clickCount)
		}

		imgui.Separator()
		imgui.Text("This demonstrates FuncWithActions for simple components with keyboard shortcuts.")
		imgui.Text(fmt.Sprintf("Click count: %d", clickCount))
		imgui.Text("Press Space to increment counter")
	})

	// add keyboard action to the function component
	root.Actions().Register("increment", "Space", func() {
		clickCount++
		fmt.Printf("counter incremented to %d via Space key\n", clickCount)
	})

	app := dfx.New(root, dfx.Config{
		Title:  "Simple Example with Actions",
		Width:  500,
		Height: 300,
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
