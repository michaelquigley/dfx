package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
)

// Counter is a custom component type that embeds Container
type Counter struct {
	dfx.Container // embed Container for default behavior
	count         int
	min           int
	max           int
	label         string
}

// NewCounter creates a new counter component
func NewCounter(label string, min, max int) *Counter {
	c := &Counter{
		label: label,
		min:   min,
		max:   max,
		count: 0,
	}

	// set up the draw function
	c.Visible = true
	c.OnDraw = func(state *dfx.State) {
		imgui.Text(fmt.Sprintf("%s: %d", c.label, c.count))

		if imgui.Button("Increment") && c.count < c.max {
			c.count++
		}

		imgui.SameLine()
		if imgui.Button("Decrement") && c.count > c.min {
			c.count--
		}

		imgui.SameLine()
		if imgui.Button("Reset") {
			c.count = 0
		}

		// demonstrate wrapped controls
		newCount, changed := dfx.SliderInt("Slider", c.count, c.min, c.max)
		if changed {
			c.count = newCount
		}

		// color based on value
		if c.count > 75 {
			imgui.TextColored(imgui.Vec4{X: 1.0, Y: 0.2, Z: 0.2, W: 1.0}, "High value!")
		} else if c.count < 25 {
			imgui.TextColored(imgui.Vec4{X: 0.2, Y: 0.2, Z: 1.0, W: 1.0}, "Low value!")
		} else {
			imgui.TextColored(imgui.Vec4{X: 0.2, Y: 1.0, Z: 0.2, W: 1.0}, "Normal value")
		}

		// show keyboard shortcuts
		imgui.Spacing()
		imgui.Text("Keys: Up/Down arrows")
	}

	// add component-specific keyboard actions using consistent API
	c.Actions().Register("increment", "Up", func() {
		if c.count < c.max {
			c.count++
			fmt.Printf("%s incremented to %d\n", c.label, c.count)
		}
	})

	c.Actions().Register("decrement", "Down", func() {
		if c.count > c.min {
			c.count--
			fmt.Printf("%s decremented to %d\n", c.label, c.count)
		}
	})

	return c
}

func main() {
	// create multiple custom components
	mainCounter := NewCounter("Main Counter", 0, 100)

	// create a container with multiple counters
	container := &dfx.Container{
		Visible: true,
		OnDraw: func(state *dfx.State) {
			imgui.Text("Custom Component Example")
			imgui.Text("This demonstrates creating reusable custom components")
			imgui.Separator()
		},
		Children: []dfx.Component{
			mainCounter,
			dfx.NewFunc(func(state *dfx.State) {
				imgui.Separator()
			}),
			NewCounter("Secondary Counter", -50, 50),
		},
	}

	app := dfx.New(container, dfx.Config{
		Title:  "Custom Component Example",
		Width:  600,
		Height: 500,
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
