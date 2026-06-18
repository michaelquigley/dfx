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

func TestDispatchAction_NotifiesOnActionObserver(t *testing.T) {
	var got []ActionEvent
	app := New(nil, Config{
		OnAction: func(e ActionEvent) { got = append(got, e) },
	})

	var handlerRan bool
	action := &Action{Id: "save", Keys: "Ctrl+S", Handler: func() { handlerRan = true }}

	app.dispatchAction(action, ActionSourceKeyboard)

	if !handlerRan {
		t.Fatal("expected handler to run")
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 observed event, got %d", len(got))
	}
	if got[0].Action != action {
		t.Fatal("expected observed event to carry the dispatched action")
	}
	if got[0].Source != ActionSourceKeyboard {
		t.Fatalf("expected source Keyboard, got %v", got[0].Source)
	}
	if got[0].Time.IsZero() {
		t.Fatal("expected event time to be set")
	}
}

func TestDispatchAction_NoObserverStillRunsHandler(t *testing.T) {
	app := New(nil, Config{})

	var ran bool
	app.dispatchAction(&Action{Id: "x", Handler: func() { ran = true }}, ActionSourceKeyboard)

	if !ran {
		t.Fatal("expected handler to run when no observer is configured")
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
