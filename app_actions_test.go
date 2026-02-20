package dfx

import "testing"

type embeddedContainerComponent struct {
	Container
}

func newEmbeddedContainerComponent(children ...Component) *embeddedContainerComponent {
	return &embeddedContainerComponent{
		Container: Container{
			Visible:  true,
			Children: children,
		},
	}
}

func actionIDs(registries []*ActionRegistry) []string {
	var ids []string
	for _, registry := range registries {
		for _, action := range registry.actions {
			ids = append(ids, action.Id)
		}
	}
	return ids
}

func TestGatherComponentActions_TraversesEmbeddedContainerChildren(t *testing.T) {
	child := NewFunc(func(*State) {})
	child.Actions().MustRegister("child", "Ctrl+1", func() {})
	parent := newEmbeddedContainerComponent(child)
	parent.Actions().MustRegister("parent", "Ctrl+2", func() {})

	app := New(parent, Config{})
	got := actionIDs(app.gatherComponentActions(parent))

	if len(got) != 2 || got[0] != "child" || got[1] != "parent" {
		t.Fatalf("expected traversal order ['child', 'parent'], got %v", got)
	}
}

func TestWorkspace_IncludesChildAndLocalActions(t *testing.T) {
	child := NewFunc(func(*State) {})
	child.Actions().MustRegister("child", "Ctrl+1", func() {})

	ws := NewWorkspace()
	ws.Add("one", "One", child)
	ws.Actions().MustRegister("local", "Ctrl+2", func() {})

	app := New(ws, Config{})
	got := actionIDs(app.gatherComponentActions(ws))

	if len(got) != 2 || got[0] != "child" || got[1] != "local" {
		t.Fatalf("expected workspace order ['child', 'local'], got %v", got)
	}
}

func TestHCollapse_IncludesContentAndLocalActions(t *testing.T) {
	content := NewFunc(func(*State) {})
	content.Actions().MustRegister("content", "Ctrl+1", func() {})

	panel := NewHCollapse(content, HCollapseConfig{Title: "panel", ExpandedWidth: 120, Expanded: true})
	panel.Container.Actions().MustRegister("local", "Ctrl+2", func() {})

	app := New(panel, Config{})
	got := actionIDs(app.gatherComponentActions(panel))

	if len(got) != 2 || got[0] != "content" || got[1] != "local" {
		t.Fatalf("expected hcollapse order ['content', 'local'], got %v", got)
	}
}

func TestDash_IncludesComponentAndLocalActions(t *testing.T) {
	content := NewFunc(func(*State) {})
	content.Actions().MustRegister("content", "Ctrl+1", func() {})

	dash := NewDash("test-dash", content)
	dash.Container.Actions().MustRegister("local", "Ctrl+2", func() {})

	app := New(dash, Config{})
	got := actionIDs(app.gatherComponentActions(dash))

	if len(got) != 2 || got[0] != "content" || got[1] != "local" {
		t.Fatalf("expected dash order ['content', 'local'], got %v", got)
	}
}
