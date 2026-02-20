package dfx

import "github.com/AllenDang/cimgui-go/imgui"

// Workspace manages multiple named components and allows switching between them.
// provides a high-level component for building applications with multiple views/modes.
// separates stable IDs from display names for flexibility.
type Workspace struct {
	Container

	// workspace storage
	items        []*workspaceItem          // ordered list for iteration
	itemsById    map[string]*workspaceItem // fast lookup by id
	currentIndex int

	// configuration
	ShowSelector  bool    // if true, shows a combo selector at the top
	SelectorLabel string  // label for the combo selector
	SelectorWidth float32 // width of selector (-1 for auto-width)

	// callbacks
	OnSwitch func(oldId, newId string) // called when workspace changes (passes IDs)
}

// NewWorkspace creates a new workspace manager.
func NewWorkspace() *Workspace {
	ws := &Workspace{
		items:         []*workspaceItem{},
		itemsById:     make(map[string]*workspaceItem),
		currentIndex:  0,
		ShowSelector:  true,
		SelectorLabel: "Workspace",
		SelectorWidth: 200,
	}

	ws.Visible = true
	ws.OnDraw = ws.draw

	return ws
}

// Add adds or replaces a workspace with the given id, display name, and component.
// if this is the first workspace added, it becomes current.
// if a workspace with the same id exists, it is replaced.
func (ws *Workspace) Add(id, name string, component Component) {
	// check if already exists
	existing, exists := ws.itemsById[id]

	if exists {
		// update existing item
		existing.Name = name
		existing.Component = component
	} else {
		// create new item
		item := &workspaceItem{
			Id:        id,
			Name:      name,
			Component: component,
		}

		// add to ordered list
		ws.items = append(ws.items, item)

		// add to map
		ws.itemsById[id] = item

		// if this is first workspace, make it current
		if len(ws.items) == 1 {
			ws.currentIndex = 0
		}
	}
}

// Remove removes a workspace by id.
// if the current workspace is removed, switches to the first available workspace.
func (ws *Workspace) Remove(id string) {
	// find item
	item, exists := ws.itemsById[id]
	if !exists {
		return // not found
	}

	// find index in ordered list
	idx := -1
	for i, it := range ws.items {
		if it == item {
			idx = i
			break
		}
	}

	if idx == -1 {
		return // should not happen
	}

	// remove from map
	delete(ws.itemsById, id)

	// remove from ordered list
	ws.items = append(ws.items[:idx], ws.items[idx+1:]...)

	// adjust current index if needed
	if len(ws.items) == 0 {
		ws.currentIndex = 0
	} else if ws.currentIndex >= len(ws.items) {
		ws.currentIndex = len(ws.items) - 1
	}
}

// Switch changes to the workspace with the given id.
// returns true if the switch was successful.
func (ws *Workspace) Switch(id string) bool {
	// find item
	item, exists := ws.itemsById[id]
	if !exists {
		return false // not found
	}

	// find index
	idx := -1
	for i, it := range ws.items {
		if it == item {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false // should not happen
	}

	// switch
	oldID := ws.Current()
	ws.currentIndex = idx
	newID := ws.Current()

	// trigger callback if changed
	if oldID != newID && ws.OnSwitch != nil {
		ws.OnSwitch(oldID, newID)
	}

	return true
}

// SwitchByIndex changes to the workspace at the given index.
// returns true if the switch was successful.
func (ws *Workspace) SwitchByIndex(index int) bool {
	if index < 0 || index >= len(ws.items) {
		return false
	}

	oldID := ws.Current()
	ws.currentIndex = index
	newID := ws.Current()

	// trigger callback if changed
	if oldID != newID && ws.OnSwitch != nil {
		ws.OnSwitch(oldID, newID)
	}

	return true
}

// Current returns the id of the current workspace.
// returns empty string if no workspaces exist.
func (ws *Workspace) Current() string {
	if len(ws.items) == 0 {
		return ""
	}
	if ws.currentIndex < 0 || ws.currentIndex >= len(ws.items) {
		return ""
	}
	return ws.items[ws.currentIndex].Id
}

// CurrentName returns the display name of the current workspace.
// returns empty string if no workspaces exist.
func (ws *Workspace) CurrentName() string {
	if len(ws.items) == 0 {
		return ""
	}
	if ws.currentIndex < 0 || ws.currentIndex >= len(ws.items) {
		return ""
	}
	return ws.items[ws.currentIndex].Name
}

// CurrentComponent returns the current workspace component.
// returns nil if no workspaces exist.
func (ws *Workspace) CurrentComponent() Component {
	if len(ws.items) == 0 {
		return nil
	}
	if ws.currentIndex < 0 || ws.currentIndex >= len(ws.items) {
		return nil
	}
	return ws.items[ws.currentIndex].Component
}

// SetName changes the display name of a workspace without affecting its Id.
// returns true if the workspace was found and updated.
func (ws *Workspace) SetName(id, name string) bool {
	item, exists := ws.itemsById[id]
	if !exists {
		return false
	}
	item.Name = name
	return true
}

// GetName returns the display name for the given workspace Id.
// returns empty string if the workspace doesn't exist.
func (ws *Workspace) GetName(id string) string {
	item, exists := ws.itemsById[id]
	if !exists {
		return ""
	}
	return item.Name
}

// WorkspaceIds returns a copy of the workspace IDs in order.
func (ws *Workspace) WorkspaceIds() []string {
	result := make([]string, len(ws.items))
	for i, item := range ws.items {
		result[i] = item.Id
	}
	return result
}

// WorkspaceNames returns a copy of the workspace display names in order.
func (ws *Workspace) WorkspaceNames() []string {
	result := make([]string, len(ws.items))
	for i, item := range ws.items {
		result[i] = item.Name
	}
	return result
}

// draw renders the workspace selector and current component.
func (ws *Workspace) draw(state *State) {
	if len(ws.items) == 0 {
		Text("no workspaces configured")
		return
	}

	// calculate available size
	availableSize := state.Size
	selectorHeight := float32(0)

	// draw selector if enabled
	if ws.ShowSelector {
		selectorHeight = 30 // approximate height for combo

		// set width if specified
		if ws.SelectorWidth > 0 {
			imgui.PushItemWidth(ws.SelectorWidth)
			defer imgui.PopItemWidth()
		}

		// get display names for combo
		names := ws.WorkspaceNames()

		// draw combo with display names
		newIndex, changed := Combo(ws.SelectorLabel, ws.currentIndex, names)
		if changed {
			ws.SwitchByIndex(newIndex)
		}

		// add spacing
		Spacing()
	}

	// draw current component
	current := ws.CurrentComponent()
	if current != nil {
		// create state for current component with adjusted size
		componentState := &State{
			Size:     imgui.Vec2{X: availableSize.X, Y: availableSize.Y - selectorHeight},
			Position: state.Position,
			IO:       state.IO,
			App:      state.App,
			Parent:   ws,
		}
		current.Draw(componentState)
	}
}

// Actions returns the action registry of the current workspace component,
// enabling action propagation through the workspace to the active component.
func (ws *Workspace) Actions() *ActionRegistry {
	if current := ws.CurrentComponent(); current != nil {
		if actions := current.Actions(); actions != nil {
			return actions
		}
	}
	return ws.Container.Actions()
}

// LocalActions returns workspace-local actions without delegation.
func (ws *Workspace) LocalActions() *ActionRegistry {
	return ws.Container.Actions()
}

// ChildActions returns the current active workspace component for action traversal.
func (ws *Workspace) ChildActions() []Component {
	if current := ws.CurrentComponent(); current != nil {
		return []Component{current}
	}
	return nil
}

type workspaceItem struct {
	Id        string    // stable identifier used in code
	Name      string    // human-facing display name (can include icons, formatting)
	Component Component // the component to display
}
