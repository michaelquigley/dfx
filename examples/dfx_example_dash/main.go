package main

import (
	"github.com/michaelquigley/dfx"
)

func main() {
	// create dashes
	leftDash := dfx.NewDash("LeftDash", dfx.NewSizeDebugger())
	leftDash.TargetSize = 300
	leftDash.CurrentSize = 300

	rightDash := dfx.NewDash("RightDash", dfx.NewSizeDebugger())
	rightDash.TargetSize = 250
	rightDash.CurrentSize = 250

	topDash := dfx.NewDash("TopDash", dfx.NewSizeDebugger())
	topDash.TargetSize = 200
	topDash.CurrentSize = 200

	bottomDash := dfx.NewDash("BottomDash", dfx.NewSizeDebugger())
	bottomDash.TargetSize = 150
	bottomDash.CurrentSize = 150

	// create dash manager
	dashManager := dfx.NewDashManager()
	dashManager.Left = leftDash
	dashManager.Right = rightDash
	dashManager.Top = topDash
	dashManager.Bottom = bottomDash
	dashManager.Inner = dfx.NewSizeDebugger()
	dashManager.Precedence = dfx.HorizontalPrecedence

	// create app
	app := dfx.New(dashManager, dfx.Config{
		Title:  "dfx Dash Example",
		Width:  1200,
		Height: 800,
	})

	// add keyboard shortcuts to toggle dashes
	app.Actions().MustRegister("toggle-left", "Alt+L", func() {
		leftDash.Visible = !leftDash.Visible
	})
	app.Actions().MustRegister("toggle-right", "Alt+R", func() {
		rightDash.Visible = !rightDash.Visible
	})
	app.Actions().MustRegister("toggle-top", "Alt+T", func() {
		topDash.Visible = !topDash.Visible
	})
	app.Actions().MustRegister("toggle-bottom", "Alt+B", func() {
		bottomDash.Visible = !bottomDash.Visible
	})
	app.Actions().MustRegister("toggle-precedence", "Alt+P", func() {
		if dashManager.Precedence == dfx.HorizontalPrecedence {
			dashManager.Precedence = dfx.VerticalPrecedence
		} else {
			dashManager.Precedence = dfx.HorizontalPrecedence
		}
	})

	// run the app
	if err := app.Run(); err != nil {
		panic(err)
	}
}
