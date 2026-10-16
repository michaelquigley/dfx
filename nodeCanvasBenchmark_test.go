package dfx

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx/internal/nodecanvasstress"
)

var benchmarkBounds canvasRect
var benchmarkHit hitResult[int]

func benchmarkGeometry(scene nodecanvasstress.Scene) *canvasGeometry[int] {
	g := &canvasGeometry[int]{view: View{Zoom: 1}, viewport: imgui.Vec2{X: 1200, Y: 800},
		nodes: make([]nodeGeometry[int], len(scene.Nodes)), links: make([]linkGeometry[int], 0, len(scene.Links)+1)}
	pinPos := make(map[int]imgui.Vec2, len(scene.Nodes)*3)
	for i, n := range scene.Nodes {
		p := imgui.Vec2{X: n.Pos.X, Y: n.Pos.Y}
		pinPos[n.Inputs[0]] = imgui.Vec2{X: p.X, Y: p.Y + 80}
		pinPos[n.Inputs[1]] = imgui.Vec2{X: p.X, Y: p.Y + 110}
		pinPos[n.Output] = imgui.Vec2{X: p.X + 180, Y: p.Y + 80}
		g.nodes[i] = nodeGeometry[int]{id: n.ID, rect: canvasRect{Min: p, Max: imgui.Vec2{X: p.X + 180, Y: p.Y + 140}},
			pins: []pinGeometry[int]{{id: n.Inputs[0], pos: pinPos[n.Inputs[0]]},
				{id: n.Inputs[1], pos: pinPos[n.Inputs[1]]},
				{id: n.Output, pos: pinPos[n.Output]}}}
	}
	for _, l := range scene.Links {
		g.links = append(g.links, linkGeometry[int]{id: l.ID, cubic: linkCubic(pinPos[l.From], pinPos[l.To], 60)})
	}
	// a known isolated link ensures the link-hit case traverses both full
	// populations regardless of scene layout and density.
	g.links = append(g.links, linkGeometry[int]{id: 999999, cubic: linkCubic(imgui.Vec2{X: -1000, Y: -1000}, imgui.Vec2{X: -800, Y: -1000}, 60)})
	return g
}

func BenchmarkNodeCanvasGeometry(b *testing.B) {
	for _, count := range []int{100, 1000, 5000, 10000, 20000} {
		b.Run(fmt.Sprintf("nodes_%d", count), func(b *testing.B) {
			scene := nodecanvasstress.Generate(nodecanvasstress.Settings{Count: count, Fanout: 4, Seed: 42})
			g := benchmarkGeometry(scene)
			params := hitParams{pinHitRadius: 8, linkHitDistance: 6}
			last := g.nodes[len(g.nodes)-1].rect.center()
			cases := []struct {
				name string
				pos  imgui.Vec2
			}{
				{"node", last}, {"link", imgui.Vec2{X: -900, Y: -1000}}, {"miss", imgui.Vec2{X: -50000, Y: -50000}},
			}
			b.Run("bounds", func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					benchmarkBounds, _ = nodeBounds(g.nodes, nil)
				}
			})
			for _, tc := range cases {
				b.Run(tc.name, func(b *testing.B) {
					if got := hitTest(g, tc.pos, params); (tc.name == "node" && got.kind != hitNode) || (tc.name == "link" && got.kind != hitLink) || (tc.name == "miss" && got.kind != hitNone) {
						b.Fatalf("wrong hit for %s: %+v", tc.name, got)
					}
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						benchmarkHit = hitTest(g, tc.pos, params)
					}
				})
			}
		})
	}
}

// this measures CPU submission with a real ImGui context, never GPU drawing.
func BenchmarkNodeCanvasCPUFrame(b *testing.B) {
	for _, count := range []int{100, 1000, 5000, 10000, 20000} {
		b.Run(fmt.Sprintf("nodes_%d", count), func(b *testing.B) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			ctx := imgui.CreateContext()
			defer imgui.DestroyContextV(ctx)
			io := imgui.CurrentIO()
			io.SetIniFilename("")
			io.SetDisplaySize(imgui.Vec2{X: 1200, Y: 800})
			io.SetDeltaTime(1.0 / 60)
			// the real GLFW/OpenGL renderer advertises vertex offsets, which
			// are required once the scene exceeds 64K vertices per draw list.
			io.SetBackendFlags(imgui.BackendFlagsRendererHasTextures | imgui.BackendFlagsRendererHasVtxOffset)
			scene := nodecanvasstress.Generate(nodecanvasstress.Settings{Count: count, Fanout: 1, Seed: 42})
			values := make([]float32, count)
			for i := range values {
				values[i] = .5
			}
			nc := NewNodeCanvas[int](NodeCanvasConfig{})
			defer nc.Destroy()
			frame := func() {
				imgui.NewFrame()
				imgui.SetNextWindowPos(imgui.Vec2{})
				imgui.SetNextWindowSize(imgui.Vec2{X: 1200, Y: 800})
				imgui.BeginV("stress", nil, imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)
				nc.Begin(&State{Size: imgui.Vec2{X: 1100, Y: 700}})
				for i := range scene.Nodes {
					node := &scene.Nodes[i]
					nc.Node(node.ID, imgui.Vec2{X: node.Pos.X, Y: node.Pos.Y}, NodeFlags{Accent: imgui.Vec4{X: node.Accent.R, Y: node.Accent.G, Z: node.Accent.B, W: 1}}, func(n *NodeContext[int]) {
					n.TitleBar(func() { n.Label(fmt.Sprintf("node %d", node.ID)) })
						n.Label("processor")
						imgui.PushItemWidth(120)
					values[node.ID-1], _ = WheelSlider("value", values[node.ID-1], 0, 1, 100, "%.2f", imgui.SliderFlagsNone)
						imgui.PopItemWidth()
						n.Input(node.Inputs[0], "in A")
						n.Input(node.Inputs[1], "in B")
						n.Output(node.Output, "out")
					})
				}
				for _, l := range scene.Links {
					nc.Link(l.ID, l.From, l.To, LinkFlags{})
				}
				nc.End()
				imgui.End()
				imgui.Render()
			}
			frame()
			frame() // warm font/layout and splitter capacity
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				frame()
			}
		})
	}
}
