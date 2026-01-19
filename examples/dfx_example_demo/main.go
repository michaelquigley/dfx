package main

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
)

func main() {
	// create a simple component that shows the imgui demo window
	root := dfx.NewFunc(func(state *dfx.State) {
		imgui.ShowDemoWindow()
	})

	app := dfx.New(root, dfx.Config{
		Title:  "ImGui Demo Window",
		Width:  1280,
		Height: 800,
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
