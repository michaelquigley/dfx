package dfx

import (
	"fmt"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
)

// Command is the core interface for undoable operations.
// simple commands only need to implement these three methods.
type Command interface {
	Description() string
	Run()
	Undo()
}

// MergeableCommand extends Command with merge capability.
// implement this interface when commands can be merged together.
type MergeableCommand interface {
	Command
	Merge(other Command) bool
}

// StampedCommand extends Command with timestamp capability.
// implement this interface when commands need timestamp tracking.
type StampedCommand interface {
	Command
	Stamp() time.Time
}

// FullCommand combines both mergeable and stamped capabilities.
type FullCommand interface {
	MergeableCommand
	StampedCommand
}

// UndoSystem manages command history and undo/redo operations.
type UndoSystem struct {
	// RunF is called whenever a command is executed, useful for tracking modifications
	RunF func(Command)
	undo []Command
	redo []Command
}

// NewUndoSystem creates a new undo system.
func NewUndoSystem() *UndoSystem {
	return &UndoSystem{}
}

// Run executes a command and adds it to the undo stack.
// if the command is mergeable and can merge with the previous command,
// they will be merged instead of creating a new stack entry.
func (us *UndoSystem) Run(command Command) {
	if us.RunF != nil {
		us.RunF(command)
	}
	command.Run()

	// attempt to merge with previous command if both support it
	if len(us.undo) > 0 {
		if mergeableCmd, ok := command.(MergeableCommand); ok {
			top := us.undo[len(us.undo)-1]
			if mergeableCmd.Merge(top) {
				us.undo = us.undo[:len(us.undo)-1]
			}
		}
		us.undo = append(us.undo, command)
		us.redo = nil
	} else {
		us.undo = append(us.undo, command)
		us.redo = nil
	}
}

// Undo reverses the last command and moves it to the redo stack.
func (us *UndoSystem) Undo() {
	if len(us.undo) > 0 {
		top := us.undo[len(us.undo)-1]
		top.Undo()
		us.undo = us.undo[:len(us.undo)-1]
		us.redo = append(us.redo, top)
	}
}

// Redo re-executes the last undone command and moves it back to the undo stack.
func (us *UndoSystem) Redo() {
	if len(us.redo) > 0 {
		top := us.redo[len(us.redo)-1]
		top.Run()
		us.redo = us.redo[:len(us.redo)-1]
		us.undo = append(us.undo, top)
	}
}

// Clear removes all commands from both undo and redo stacks.
func (us *UndoSystem) Clear() {
	us.undo = nil
	us.redo = nil
}

// CanUndo returns true if there are commands that can be undone.
func (us *UndoSystem) CanUndo() bool {
	return len(us.undo) > 0
}

// CanRedo returns true if there are commands that can be redone.
func (us *UndoSystem) CanRedo() bool {
	return len(us.redo) > 0
}

// HistoryComponent returns a component that displays the undo/redo history.
// this replaces the original Draw method with dfx's component architecture.
func (us *UndoSystem) HistoryComponent() Component {
	return NewFunc(func(state *State) {
		imgui.BeginChildStrV(fmt.Sprintf("##%p", us), imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsNone, imgui.WindowFlagsHorizontalScrollbar)

		// draw redo commands in muted color
		imgui.PushStyleColorVec4(imgui.ColText, imgui.CurrentStyle().Colors()[imgui.ColPlotLinesHovered])
		for i := range us.redo {
			imgui.TextUnformatted(us.redo[i].Description())
		}
		imgui.PopStyleColor()

		// draw undo commands in normal color (most recent first)
		for i := len(us.undo) - 1; i >= 0; i-- {
			imgui.TextUnformatted(us.undo[i].Description())
		}

		imgui.EndChild()
	})
}

// BaseCommand provides default implementations for common command functionality.
// embed this struct to get sensible defaults for timestamp tracking.
type BaseCommand struct {
	stamp time.Time
}

// Stamp implements StampedCommand with automatic timestamp on first call.
func (bc *BaseCommand) Stamp() time.Time {
	if bc.stamp.IsZero() {
		bc.stamp = time.Now()
	}
	return bc.stamp
}

// SetStamp allows manual setting of the timestamp.
func (bc *BaseCommand) SetStamp(t time.Time) {
	bc.stamp = t
}
