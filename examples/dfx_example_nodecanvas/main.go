package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/df/dl"
	"github.com/michaelquigley/dfx"
	"github.com/michaelquigley/dfx/fonts"
)

// the NodeCanvas acceptance example: a synthetic graph over a trivial
// in-memory model. the app owns all graph truth and declares it every frame;
// completed gestures come back as intents, logged and applied here — node
// moves run through a dfx UndoSystem command (one gesture, one intent, one
// undo command).
//
// mouse: click selects (ctrl toggles, shift adds), drag moves the selection,
// drag on empty canvas box-selects, drag from a pin creates a link (snaps
// near a compatible pin), middle-drag pans, wheel zooms through the detents
// toward the cursor. below detent 1.0 the nodes declare simplified,
// non-interactive content — labels, values, pins — per the reduced-detent
// contract.
//
// keys: F zoom-to-fit all, C center on selection, L toggle locked mode,
// V save the view, Shift+V restore it (simulated persistence round-trip),
// Delete removes the selection (app-owned — the canvas has no delete
// intent), Ctrl+Z / Ctrl+Shift+Z undo/redo.

type node struct {
	title    string
	pos      imgui.Vec2
	selected bool
}

type link struct {
	from, to string
	selected bool
}

// moveNodesCommand applies one completed drag gesture as one undoable
// command.
type moveNodesCommand struct {
	nodes map[string]*node
	moves []dfx.NodeMove[string]
}

func (c *moveNodesCommand) Description() string {
	return fmt.Sprintf("move %d node(s)", len(c.moves))
}

func (c *moveNodesCommand) Run() {
	for _, m := range c.moves {
		c.nodes[m.ID].pos = m.To
	}
}

func (c *moveNodesCommand) Undo() {
	for _, m := range c.moves {
		c.nodes[m.ID].pos = m.From
	}
}

// deleteSelectionCommand removes the selected nodes and links — deletion is
// app-owned: the canvas has no delete intent by design, so the app deletes
// from its own selection through its own undo machinery. deleting a node
// also removes the links attached to its pins.
type deleteSelectionCommand struct {
	nodes        map[string]*node
	links        map[string]*link
	removedNodes map[string]*node
	removedLinks map[string]*link
}

func newDeleteSelectionCommand(nodes map[string]*node, links map[string]*link) *deleteSelectionCommand {
	c := &deleteSelectionCommand{
		nodes:        nodes,
		links:        links,
		removedNodes: make(map[string]*node),
		removedLinks: make(map[string]*link),
	}
	for id, n := range nodes {
		if n.selected {
			c.removedNodes[id] = n
		}
	}
	for id, l := range links {
		if l.selected {
			c.removedLinks[id] = l
			continue
		}
		// cascade: the example's pins are named "<node>.<pin>", so a link
		// touching a deleted node's pins goes with it.
		for nodeID := range c.removedNodes {
			if strings.HasPrefix(l.from, nodeID+".") || strings.HasPrefix(l.to, nodeID+".") {
				c.removedLinks[id] = l
				break
			}
		}
	}
	return c
}

func (c *deleteSelectionCommand) empty() bool {
	return len(c.removedNodes) == 0 && len(c.removedLinks) == 0
}

func (c *deleteSelectionCommand) Description() string {
	return fmt.Sprintf("delete %d node(s), %d link(s)", len(c.removedNodes), len(c.removedLinks))
}

func (c *deleteSelectionCommand) Run() {
	for id := range c.removedNodes {
		delete(c.nodes, id)
	}
	for id := range c.removedLinks {
		delete(c.links, id)
	}
}

func (c *deleteSelectionCommand) Undo() {
	for id, n := range c.removedNodes {
		c.nodes[id] = n
	}
	for id, l := range c.removedLinks {
		c.links[id] = l
	}
}

func main() {
	nodes := map[string]*node{
		"source": {title: fonts.ICON_MUSIC_NOTE + " source", pos: imgui.Vec2{X: 60, Y: 120}},
		"filter": {title: fonts.ICON_TUNE + " filter", pos: imgui.Vec2{X: 320, Y: 80}},
		"gain":   {title: fonts.ICON_GRAPHIC_EQ + " gain", pos: imgui.Vec2{X: 620, Y: 140}},
		"meter":  {title: fonts.ICON_SPEAKER + " meter", pos: imgui.Vec2{X: 880, Y: 180}},
		"notes":  {title: "", pos: imgui.Vec2{X: 320, Y: 340}},
	}
	links := map[string]*link{
		"l.source-filter": {from: "source.out", to: "filter.in"},
		"l.filter-gain":   {from: "filter.out", to: "gain.in"},
		"l.source-side":   {from: "source.out", to: "gain.side"},
		"l.gain-meter":    {from: "gain.out", to: "meter.in"},
	}
	linkSeq := 0

	cutoff := float32(1200)
	resonance := float32(0.3)
	level := float32(0.8)

	locked := false
	var savedView *dfx.View

	undo := dfx.NewUndoSystem()
	nc := dfx.NewNodeCanvas[string](dfx.NodeCanvasConfig{})

	selectedNodes := func() []string {
		var ids []string
		for id, n := range nodes {
			if n.selected {
				ids = append(ids, id)
			}
		}
		sort.Strings(ids)
		return ids
	}

	applyIntents := func(intents dfx.Intents[string]) {
		if sc := intents.SelectionChanged; sc != nil {
			for _, n := range nodes {
				n.selected = false
			}
			for _, l := range links {
				l.selected = false
			}
			for _, id := range sc.Nodes {
				nodes[id].selected = true
			}
			for _, id := range sc.Links {
				links[id].selected = true
			}
			dl.Infof("selection changed: nodes=%v links=%v", sc.Nodes, sc.Links)
		}

		if lc := intents.LinkCreated; lc != nil {
			// app-side validation: the canvas knows sides, not semantics —
			// reject duplicates here.
			duplicate := false
			for _, l := range links {
				if l.from == lc.FromPin && l.to == lc.ToPin {
					duplicate = true
					break
				}
			}
			if duplicate {
				dl.Infof("link create ignored (duplicate): %v -> %v", lc.FromPin, lc.ToPin)
			} else {
				linkSeq++
				id := fmt.Sprintf("l.user-%d", linkSeq)
				links[id] = &link{from: lc.FromPin, to: lc.ToPin}
				dl.Infof("link created: %v -> %v (%v)", lc.FromPin, lc.ToPin, id)
			}
		}

		if len(intents.NodesMoved) > 0 {
			undo.Run(&moveNodesCommand{nodes: nodes, moves: intents.NodesMoved})
			dl.Infof("nodes moved: %v", intents.NodesMoved)
		}
	}

	root := dfx.NewFunc(func(state *dfx.State) {
		nc.Begin(state)

		nc.Node("source", nodes["source"].pos, dfx.NodeFlags{Selected: nodes["source"].selected}, func(n *dfx.NodeContext[string]) {
			n.TitleBar(func() { n.Label(nodes["source"].title) })
			n.Label("sine 440hz")
			n.Output("source.out", "out")
		})

		nc.Node("filter", nodes["filter"].pos, dfx.NodeFlags{Selected: nodes["filter"].selected}, func(n *dfx.NodeContext[string]) {
			n.TitleBar(func() { n.Label(nodes["filter"].title) })
			if n.Detent() < 1.0 {
				// reduced-detent contract: labels, values, pins — nothing
				// interactive.
				n.Label(fmt.Sprintf("cutoff %.0f", cutoff))
				n.Label(fmt.Sprintf("res %.2f", resonance))
			} else {
				// node content owns its widget widths: the canvas window's
				// default item width is meaningless inside a node.
				imgui.PushItemWidth(140)
				imgui.SliderFloat("cutoff", &cutoff, 20, 20000)
				imgui.SliderFloat("res", &resonance, 0, 1)
				imgui.PopItemWidth()
			}
			n.Input("filter.in", "in")
			n.Output("filter.out", "out")
		})

		nc.Node("gain", nodes["gain"].pos, dfx.NodeFlags{Selected: nodes["gain"].selected}, func(n *dfx.NodeContext[string]) {
			n.TitleBar(func() { n.Label(nodes["gain"].title) })
			if n.Detent() < 1.0 {
				n.Label(fmt.Sprintf("level %.2f", level))
			} else {
				imgui.PushItemWidth(140)
				imgui.SliderFloat("level", &level, 0, 1)
				imgui.PopItemWidth()
			}
			n.Input("gain.in", "in")
			n.Input("gain.side", "sidechain")
			n.Output("gain.out", "out")
		})

		nc.Node("meter", nodes["meter"].pos, dfx.NodeFlags{Selected: nodes["meter"].selected}, func(n *dfx.NodeContext[string]) {
			n.TitleBar(func() { n.Label(nodes["meter"].title) })
			n.Label("-12.4 dB")
			n.Input("meter.in", "in")
		})

		// a bare, title-less, label-only card.
		nc.Node("notes", nodes["notes"].pos, dfx.NodeFlags{Selected: nodes["notes"].selected}, func(n *dfx.NodeContext[string]) {
			n.Label("patch: warm pad")
			n.Label("bpm: 96")
		})

		// declare links in a stable order — map iteration would shuffle
		// declaration order (and with it z-order tie-breaks) every frame.
		linkIDs := make([]string, 0, len(links))
		for id := range links {
			linkIDs = append(linkIDs, id)
		}
		sort.Strings(linkIDs)
		for _, id := range linkIDs {
			l := links[id]
			nc.Link(id, l.from, l.to, dfx.LinkFlags{Selected: l.selected})
		}

		applyIntents(nc.End())
	})

	root.Actions().MustRegister("undo", "Ctrl+Z", func() {
		undo.Undo()
	})
	root.Actions().MustRegister("redo", "Ctrl+Shift+Z", func() {
		undo.Redo()
	})
	root.Actions().MustRegister("delete selection", "Delete", func() {
		cmd := newDeleteSelectionCommand(nodes, links)
		if !cmd.empty() {
			undo.Run(cmd)
			dl.Infof("deleted: %s", cmd.Description())
		}
	})
	root.Actions().MustRegister("zoom to fit", "F", func() {
		nc.ZoomToFit()
	})
	root.Actions().MustRegister("center on selection", "C", func() {
		nc.CenterOn(selectedNodes()...)
	})
	root.Actions().MustRegister("toggle locked", "L", func() {
		locked = !locked
		nc.SetLocked(locked)
		dl.Infof("locked: %v", locked)
	})
	root.Actions().MustRegister("save view", "V", func() {
		v := nc.View()
		savedView = &v
		dl.Infof("view saved: pan=(%v, %v) zoom=%v", v.Pan.X, v.Pan.Y, v.Zoom)
	})
	root.Actions().MustRegister("restore view", "Shift+V", func() {
		if savedView != nil {
			nc.SetView(*savedView)
			dl.Infof("view restored: pan=(%v, %v) zoom=%v", savedView.Pan.X, savedView.Pan.Y, savedView.Zoom)
		}
	})

	app := dfx.New(root, dfx.Config{
		Title:  "dfx NodeCanvas",
		Width:  1280,
		Height: 800,
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
