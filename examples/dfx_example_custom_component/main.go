package main

import (
	"fmt"

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
		dfx.Text(fmt.Sprintf("%s: %d", c.label, c.count))

		if dfx.Button("Increment") && c.count < c.max {
			c.count++
		}

		dfx.SameLine()
		if dfx.Button("Decrement") && c.count > c.min {
			c.count--
		}

		dfx.SameLine()
		if dfx.Button("Reset") {
			c.count = 0
		}

		// demonstrate wrapped controls
		newCount, changed := dfx.SliderInt("Slider", c.count, c.min, c.max)
		if changed {
			c.count = newCount
		}

		// color based on value
		if c.count > 75 {
			dfx.TextColored("High value!", 1.0, 0.2, 0.2, 1.0)
		} else if c.count < 25 {
			dfx.TextColored("Low value!", 0.2, 0.2, 1.0, 1.0)
		} else {
			dfx.TextColored("Normal value", 0.2, 1.0, 0.2, 1.0)
		}

		// show keyboard shortcuts
		dfx.Spacing()
		dfx.Text("Keys: Up/Down arrows")
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
			dfx.Text("Custom Component Example")
			dfx.Text("This demonstrates creating reusable custom components")
			dfx.Separator()
		},
		Children: []dfx.Component{
			mainCounter,
			dfx.NewFunc(func(state *dfx.State) {
				dfx.Separator()
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
