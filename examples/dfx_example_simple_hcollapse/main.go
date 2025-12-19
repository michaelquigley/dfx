package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
)

func main() {
	content := dfx.NewFunc(func(state *dfx.State) {
		imgui.BeginChildStrV("##content", imgui.Vec2{}, 0, imgui.WindowFlagsAlwaysVerticalScrollbar)
		for i := 0; i < 24; i++ {
			dfx.Text(fmt.Sprintf("oh yes: %d", i))
		}
		imgui.EndChild()
	})
	collapse := dfx.NewHCollapse(content, dfx.HCollapseConfig{
		Title:         "Main",
		ExpandedWidth: 200,
		Expanded:      true,
		Resizable:     true,
	})
	root := dfx.NewFunc(func(state *dfx.State) {
		collapse.Draw(state)
	})
	app := dfx.New(root, dfx.Config{
		Title:  "Simple HCollapse",
		Width:  400,
		Height: 300,
	})
	if err := app.Run(); err != nil {
		panic(err)
	}
}
