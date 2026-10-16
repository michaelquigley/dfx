package main

import (
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
	"github.com/michaelquigley/dfx/internal/nodecanvasstress"
)

const sampleWindow = 120

type timings struct {
	values [sampleWindow]time.Duration
	count  int
	next   int
	warm   int
}

func (t *timings) reset() { *t = timings{warm: 30} }

func (t *timings) add(d time.Duration) {
	if t.warm > 0 {
		t.warm--
		return
	}
	t.values[t.next] = d
	t.next = (t.next + 1) % sampleWindow
	if t.count < sampleWindow {
		t.count++
	}
}

func (t *timings) stats() (time.Duration, time.Duration) {
	if t.count == 0 {
		return 0, 0
	}
	v := append([]time.Duration(nil), t.values[:t.count]...)
	sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
	return v[t.count/2], v[(95*t.count+99)/100-1]
}

type model struct {
	settings     nodecanvasstress.Settings
	scene        nodecanvasstress.Scene
	values       []float32
	selected     map[int]bool
	links        map[int]bool
	userLinkID   int
	canvas       *dfx.NodeCanvas[int]
	cpu, cadence timings
	lastTick     time.Time
}

func newModel() *model {
	m := &model{
		settings: nodecanvasstress.Settings{Count: 1000, Layout: nodecanvasstress.Grid, Spacing: 1, Fanout: 1, Seed: 0xD1F},
		selected: make(map[int]bool), links: make(map[int]bool),
		canvas: dfx.NewNodeCanvas[int](dfx.NodeCanvasConfig{}),
	}
	m.rebuild()
	return m
}

func (m *model) resetTimings() { m.cpu.reset(); m.cadence.reset(); m.lastTick = time.Time{} }

func (m *model) rebuild() {
	m.scene = nodecanvasstress.Generate(m.settings)
	m.values = make([]float32, len(m.scene.Nodes))
	for i := range m.values {
		m.values[i] = .5
	}
	clear(m.selected)
	clear(m.links)
	m.resetTimings()
}

func (m *model) tick() {
	now := time.Now()
	if !m.lastTick.IsZero() {
		m.cadence.add(now.Sub(m.lastTick))
	}
	m.lastTick = now
}

func (m *model) controls(size imgui.Vec2) {
	imgui.BeginChildStrV("controls", imgui.Vec2{X: 300, Y: size.Y}, imgui.ChildFlagsBorders, 0)
	imgui.TextUnformattedV("NodeCanvas stress baseline")
	imgui.TextUnformattedV("F: fit  C: center  1/2/3: view presets")
	imgui.TextUnformattedV("middle drag: pan  Ctrl+wheel: zoom")
	if m.canvas.MiniMapVisible() {
		if imgui.Button("hide mini-map") {
			m.canvas.SetMiniMapVisible(false)
			m.resetTimings()
		}
	} else if imgui.Button("show mini-map") {
		m.canvas.SetMiniMapVisible(true)
		m.resetTimings()
	}
	corner := m.canvas.MiniMapConfig().Corner
	cornerNames := [...]string{"lower-left", "lower-right", "upper-right", "upper-left"}
	if imgui.Button("map corner: " + cornerNames[corner] + " (next)") {
		cfg := m.canvas.MiniMapConfig()
		cfg.Corner = dfx.NodeCanvasCorner((int(cfg.Corner) + 1) % len(cornerNames))
		m.canvas.SetMiniMapConfig(cfg)
		m.resetTimings()
	}
	mapSize := m.canvas.MiniMapConfig().Size
	if imgui.SliderFloatV("map width", &mapSize.X, 120, 400, "%.0f px", imgui.SliderFlagsNone) {
		cfg := m.canvas.MiniMapConfig()
		cfg.Size.X = mapSize.X
		m.canvas.SetMiniMapConfig(cfg)
		m.resetTimings()
	}
	if imgui.SliderFloatV("map height", &mapSize.Y, 80, 280, "%.0f px", imgui.SliderFlagsNone) {
		cfg := m.canvas.MiniMapConfig()
		cfg.Size.Y = mapSize.Y
		m.canvas.SetMiniMapConfig(cfg)
		m.resetTimings()
	}
	imgui.Separator()
	if imgui.Button("100") {
		m.settings.Count = 100
		m.rebuild()
	}
	imgui.SameLine()
	if imgui.Button("1,000") {
		m.settings.Count = 1000
		m.rebuild()
	}
	imgui.SameLine()
	if imgui.Button("5,000") {
		m.settings.Count = 5000
		m.rebuild()
	}
	if imgui.Button("10,000") {
		m.settings.Count = 10000
		m.rebuild()
	}
	imgui.SameLine()
	if imgui.Button("20,000") {
		m.settings.Count = 20000
		m.rebuild()
	}
	count := int32(m.settings.Count)
	if imgui.SliderInt("nodes", &count, 0, nodecanvasstress.MaxNodes) {
		m.settings.Count = int(count)
		m.rebuild()
	}
	if imgui.Button("compact grid") {
		m.settings.Layout = nodecanvasstress.Grid
		m.rebuild()
	}
	imgui.SameLine()
	if imgui.Button("clusters") {
		m.settings.Layout = nodecanvasstress.Clusters
		m.rebuild()
	}
	spacing := m.settings.Spacing
	if imgui.SliderFloatV("spacing", &spacing, .5, 3, "%.2f", imgui.SliderFlagsNone) {
		m.settings.Spacing = spacing
		m.rebuild()
	}
	if imgui.Button("0 links/node") {
		m.settings.Fanout = 0
		m.rebuild()
	}
	imgui.SameLine()
	if imgui.Button("1 link/node") {
		m.settings.Fanout = 1
		m.rebuild()
	}
	imgui.SameLine()
	if imgui.Button("4 links/node") {
		m.settings.Fanout = 4
		m.rebuild()
	}
	imgui.TextUnformattedV(fmt.Sprintf("nodes %d  links %d  layout %v  fanout %d", len(m.scene.Nodes), len(m.scene.Links), m.settings.Layout, m.settings.Fanout))
	imgui.TextUnformattedV(fmt.Sprintf("viewport %.0f x %.0f  zoom %.2f", size.X-308, size.Y, m.canvas.View().Zoom))
	imgui.TextUnformattedV(fmt.Sprintf("spacing %.2f  OS %s/%s", m.settings.Spacing, runtime.GOOS, runtime.GOARCH))
	imgui.TextUnformattedV("GLFW/OpenGL; VSync: enabled (swap interval 1)")
	imgui.Separator()
	cm, cp := m.cpu.stats()
	fm, fp := m.cadence.stats()
	imgui.TextUnformattedV(fmt.Sprintf("canvas CPU (%d samples): median %s / p95 %s", m.cpu.count, cm.Round(time.Microsecond), cp.Round(time.Microsecond)))
	imgui.TextUnformattedV(fmt.Sprintf("displayed frame cadence (%d): median %s / p95 %s", m.cadence.count, fm.Round(time.Microsecond), fp.Round(time.Microsecond)))
	if m.cpu.warm > 0 {
		imgui.TextUnformattedV(fmt.Sprintf("warming: %d frames", m.cpu.warm))
	}
	imgui.TextUnformattedV("cadence includes presentation pacing; no GPU timing.")
	imgui.EndChild()
}

func (m *model) draw(state *dfx.State) {
	m.controls(state.Size)
	imgui.SameLine()
	canvasSize := imgui.Vec2{X: state.Size.X - 308, Y: state.Size.Y}
	if canvasSize.X < 1 {
		canvasSize.X = 1
	}
	start := time.Now()
	m.canvas.Begin(&dfx.State{Size: canvasSize, IO: state.IO, App: state.App})
	for i := range m.scene.Nodes {
		node := &m.scene.Nodes[i]
		value := &m.values[node.ID-1]
		m.canvas.Node(node.ID, imgui.Vec2{X: node.Pos.X, Y: node.Pos.Y}, dfx.NodeFlags{
			Selected: m.selected[node.ID],
			Accent:   imgui.Vec4{X: node.Accent.R, Y: node.Accent.G, Z: node.Accent.B, W: node.Accent.A},
		}, func(n *dfx.NodeContext[int]) {
			n.TitleBar(func() { n.Label(fmt.Sprintf("node %d", node.ID)) })
			n.Label("processor")
			if n.Detent() < 1 {
				n.Label(fmt.Sprintf("value %.2f", *value))
			} else {
				imgui.PushItemWidth(120)
				*value, _ = dfx.WheelSlider("value", *value, 0, 1, 100, "%.2f", imgui.SliderFlagsNone)
				imgui.PopItemWidth()
			}
			n.Input(node.Inputs[0], "in A")
			n.Input(node.Inputs[1], "in B")
			n.Output(node.Output, "out")
		})
	}
	for _, link := range m.scene.Links {
		m.canvas.Link(link.ID, link.From, link.To, dfx.LinkFlags{Selected: m.links[link.ID]})
	}
	intents := m.canvas.End()
	m.cpu.add(time.Since(start))
	m.apply(intents)
}

func (m *model) apply(in dfx.Intents[int]) {
	if in.NodeRaised != nil {
		for i := range m.scene.Nodes {
			if m.scene.Nodes[i].ID == *in.NodeRaised {
				raised := m.scene.Nodes[i]
				copy(m.scene.Nodes[i:], m.scene.Nodes[i+1:])
				m.scene.Nodes[len(m.scene.Nodes)-1] = raised
				break
			}
		}
	}
	if in.SelectionChanged != nil {
		clear(m.selected)
		clear(m.links)
		for _, id := range in.SelectionChanged.Nodes {
			m.selected[id] = true
		}
		for _, id := range in.SelectionChanged.Links {
			m.links[id] = true
		}
	}
	for _, move := range in.NodesMoved {
		for i := range m.scene.Nodes {
			if m.scene.Nodes[i].ID == move.ID {
				m.scene.Nodes[i].Pos = nodecanvasstress.Point{X: move.To.X, Y: move.To.Y}
				break
			}
		}
	}
	if in.LinkCreated != nil {
		for _, l := range m.scene.Links {
			if l.From == in.LinkCreated.FromPin && l.To == in.LinkCreated.ToPin {
				return
			}
		}
		m.userLinkID++
		m.scene.Links = append(m.scene.Links, nodecanvasstress.Link{
			ID: 300000 + m.userLinkID, From: in.LinkCreated.FromPin, To: in.LinkCreated.ToPin,
		})
		m.resetTimings()
	}
}

func main() {
	m := newModel()
	root := dfx.NewFunc(m.draw)
	root.Actions().MustRegister("fit", "F", func() { m.canvas.ZoomToFit(); m.resetTimings() })
	root.Actions().MustRegister("center", "C", func() { m.canvas.CenterOn(); m.resetTimings() })
	for i, key := range []string{"1", "2", "3"} {
		preset := i
		root.Actions().MustRegister("view "+key, key, func() {
			if len(m.scene.Nodes) == 0 {
				return
			}
			id := 1 + preset*(len(m.scene.Nodes)-1)/2
			m.canvas.CenterOn(id)
			m.resetTimings()
		})
	}
	app := dfx.New(root, dfx.Config{Title: "dfx NodeCanvas stress", Width: 1400, Height: 900,
		OnTick:     func(*dfx.App) { m.tick() },
		OnShutdown: func(*dfx.App) { m.canvas.Destroy() },
	})
	if err := app.Run(); err != nil {
		panic(err)
	}
}
