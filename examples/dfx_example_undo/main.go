package main

import (
	"fmt"
	"time"

	"github.com/michaelquigley/dfx"
	"github.com/AllenDang/cimgui-go/imgui"
)

var value int

// simpleSetValueCommand demonstrates the minimal Command interface
type simpleSetValueCommand struct {
	oldValue int
	newValue int
}

func (c *simpleSetValueCommand) Description() string {
	return fmt.Sprintf("set value to %d", c.newValue)
}

func (c *simpleSetValueCommand) Run() {
	value = c.newValue
}

func (c *simpleSetValueCommand) Undo() {
	value = c.oldValue
}

// mergeableSetValueCommand demonstrates MergeableCommand interface
type mergeableSetValueCommand struct {
	dfx.BaseCommand
	oldValue int
	newValue int
}

func (c *mergeableSetValueCommand) Description() string {
	return fmt.Sprintf("set value to %d (mergeable)", c.newValue)
}

func (c *mergeableSetValueCommand) Run() {
	value = c.newValue
}

func (c *mergeableSetValueCommand) Undo() {
	value = c.oldValue
}

// Merge implements MergeableCommand - merges commands within 1 second
func (c *mergeableSetValueCommand) Merge(other dfx.Command) bool {
	if otherC, ok := other.(*mergeableSetValueCommand); ok {
		// check if other command implements StampedCommand
		if stamped, hasStamp := other.(dfx.StampedCommand); hasStamp {
			if time.Since(stamped.Stamp()).Milliseconds() <= 1000 {
				c.oldValue = otherC.oldValue
				return true
			}
		}
	}
	return false
}

func main() {
	undoSystem := dfx.NewUndoSystem()

	// track when commands are executed
	undoSystem.RunF = func(cmd dfx.Command) {
		fmt.Printf("command executed: '%s' - file modified\n", cmd.Description())
	}

	// create the main component
	mainComponent := &dfx.Container{
		Visible: true,
		OnDraw: func(state *dfx.State) {
			// controls column
			imgui.BeginTableV("layout", 2, imgui.TableFlagsResizable, imgui.Vec2{}, 0.0)

			imgui.TableNextColumn()

			// simple commands (no merge capability)
			imgui.Text("Simple Commands (no merge):")
			if imgui.Button("Simple Increment") {
				cmd := &simpleSetValueCommand{
					oldValue: value,
					newValue: value + 1,
				}
				undoSystem.Run(cmd)
			}
			if imgui.Button("Simple Decrement") {
				cmd := &simpleSetValueCommand{
					oldValue: value,
					newValue: value - 1,
				}
				undoSystem.Run(cmd)
			}

			imgui.Separator()

			// mergeable commands
			imgui.Text("Mergeable Commands:")
			if imgui.Button("Mergeable Increment") {
				cmd := &mergeableSetValueCommand{
					oldValue: value,
					newValue: value + 1,
				}
				cmd.SetStamp(time.Now()) // explicitly set timestamp
				undoSystem.Run(cmd)
			}
			if imgui.Button("Mergeable Decrement") {
				cmd := &mergeableSetValueCommand{
					oldValue: value,
					newValue: value - 1,
				}
				cmd.SetStamp(time.Now()) // explicitly set timestamp
				undoSystem.Run(cmd)
			}

			imgui.Separator()

			// undo/redo controls
			imgui.BeginDisabledV(!undoSystem.CanUndo())
			if imgui.Button("Undo (Ctrl+Z)") {
				undoSystem.Undo()
			}
			imgui.EndDisabled()

			imgui.SameLine()
			imgui.BeginDisabledV(!undoSystem.CanRedo())
			if imgui.Button("Redo (Ctrl+Shift+Z)") {
				undoSystem.Redo()
			}
			imgui.EndDisabled()

			if imgui.Button("Clear History") {
				undoSystem.Clear()
			}

			imgui.Separator()
			imgui.Text(fmt.Sprintf("Current value: %d", value))

			// history column
			imgui.TableNextColumn()
			imgui.Text("Command History:")

			// use the history component
			historyComponent := undoSystem.HistoryComponent()
			historyComponent.Draw(state)

			imgui.EndTable()
		},
	}

	// add keyboard shortcuts
	mainComponent.Actions().MustRegister("undo", "Ctrl+Z", func() {
		undoSystem.Undo()
	})
	mainComponent.Actions().MustRegister("redo", "Ctrl+Shift+Z", func() {
		undoSystem.Redo()
	})

	// configure and run the app
	config := dfx.Config{
		Title:  "dfx Undo System Example",
		Width:  800,
		Height: 600,
	}

	app := dfx.New(mainComponent, config)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
