package main

import (
	"fmt"
	"sort"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/df/dl"
	"github.com/michaelquigley/dfx"
	"github.com/michaelquigley/dfx/fonts"
)

// the NodeCanvas acceptance example: a synthetic graph over a trivial
// in-memory model. the app owns all graph truth and declares it every frame;
// completed gestures come back as intents, logged and applied here — node
// moves run through a dfx UndoSystem command (one gesture, one intent, one
// undo command; Ctrl+Z / Ctrl+Shift+Z).
//
// interaction: click selects (ctrl toggles, shift adds), drag moves the
// selection, drag on empty canvas box-selects, drag from a pin creates a
// link (snaps near a compatible pin), middle-drag pans, wheel zooms through
// the detents toward the cursor.

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

	undo := dfx.NewUndoSystem()
	nc := dfx.NewNodeCanvas[string](dfx.NodeCanvasConfig{})

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
			// node content owns its widget widths: the canvas window's
			// default item width is meaningless inside a node.
			imgui.PushItemWidth(140 * n.Detent())
			imgui.SliderFloat("cutoff", &cutoff, 20, 20000)
			imgui.SliderFloat("res", &resonance, 0, 1)
			imgui.PopItemWidth()
			n.Input("filter.in", "in")
			n.Output("filter.out", "out")
		})

		nc.Node("gain", nodes["gain"].pos, dfx.NodeFlags{Selected: nodes["gain"].selected}, func(n *dfx.NodeContext[string]) {
			n.TitleBar(func() { n.Label(nodes["gain"].title) })
			imgui.PushItemWidth(140 * n.Detent())
			imgui.SliderFloat("level", &level, 0, 1)
			imgui.PopItemWidth()
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

	app := dfx.New(root, dfx.Config{
		Title:  "dfx NodeCanvas",
		Width:  1280,
		Height: 800,
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
