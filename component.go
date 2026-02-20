package dfx

import "github.com/AllenDang/cimgui-go/imgui"

// Component is the core abstraction - a drawable, interactive UI element.
type Component interface {
	// Draw renders the component. Unlike Surface.DrawF, we pass a State
	// that contains more than just size - it has everything needed to draw.
	Draw(state *State)

	// Actions returns the component's action registry.
	// this provides a consistent API for registering keyboard shortcuts.
	Actions() *ActionRegistry
}

// ChildActionProvider exposes child components for action traversal.
// components that compose other components can implement this to participate
// in hierarchical action lookup.
type ChildActionProvider interface {
	ChildActions() []Component
}

// LocalActionProvider exposes a component's local action registry.
// this allows traversal code to use local actions directly even when Actions()
// is used for legacy delegation behavior.
type LocalActionProvider interface {
	LocalActions() *ActionRegistry
}

// State provides everything a component needs to draw.
// this consolidates what Surface scattered across multiple parameters.
type State struct {
	// Size available for this component to draw in
	Size imgui.Vec2

	// Position where this component should draw (for absolute positioning)
	Position imgui.Vec2

	// IO provides access to imgui's input/output system
	IO *imgui.IO

	// App provides access to the application instance
	App *App

	// Parent component (nil for root)
	Parent Component
}

// Container is a basic component implementation that others can embed.
// provides default implementations and common fields.
type Container struct {
	Visible  bool
	Children []Component
	OnDraw   func(*State)
	actions  *ActionRegistry
}

// Draw implements Component with a simple delegation pattern
func (c *Container) Draw(state *State) {
	if !c.Visible {
		return
	}
	if c.OnDraw != nil {
		c.OnDraw(state)
	}
	// draw children if any
	for _, child := range c.Children {
		child.Draw(state)
	}
}

// ChildActions returns action-traversable children.
func (c *Container) ChildActions() []Component {
	return c.Children
}

// Actions implements Component
func (c *Container) Actions() *ActionRegistry {
	if c.actions == nil {
		c.actions = NewActionRegistry()
	}
	return c.actions
}

// LocalActions returns this container's local action registry.
func (c *Container) LocalActions() *ActionRegistry {
	return c.Actions()
}

// Func is a function component that can have keyboard actions.
// use this when you need a simple component with keyboard shortcuts.
type Func struct {
	drawFunc func(*State)
	actions  *ActionRegistry
}

func NewFunc(draw func(*State)) *Func {
	return &Func{
		drawFunc: draw,
		actions:  NewActionRegistry(),
	}
}

func (f *Func) Draw(state *State) {
	if f.drawFunc != nil {
		f.drawFunc(state)
	}
}

func (f *Func) Actions() *ActionRegistry {
	if f.actions == nil {
		f.actions = NewActionRegistry()
	}
	return f.actions
}

// LocalActions returns this component's local action registry.
func (f *Func) LocalActions() *ActionRegistry {
	return f.Actions()
}
