package dfx

import (
	"fmt"
	"math"
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
)

func TestNodeCanvas_PanInputLossKeepsLastValidView(t *testing.T) {
	for _, zoom := range []float32{0.5, 1, 1.5} {
		for _, loss := range []string{"focus_invalid_mouse", "focus_valid_mouse", "invalid_mouse_held", "invalid_mouse_release", "lost_button"} {
			t.Run(fmt.Sprintf("zoom_%g/%s", zoom, loss), func(t *testing.T) {
				h := newCanvasWidgetTest(t)
				h.nc.SetView(View{Pan: imgui.Vec2{X: 12, Y: -7}, Zoom: zoom})
				var drawn View
				draw := func() { drawn = h.nc.frameView }
				h.point(imgui.Vec2{X: 450, Y: 350})
				h.frame(draw)
				h.frame(draw)
				h.io.AddMouseButtonEvent(2, true)
				h.frame(draw)
				h.point(imgui.Vec2{X: 490, Y: 370})
				h.frame(draw)
				want := View{Pan: imgui.Vec2{X: 12 + 40/zoom, Y: -7 + 20/zoom}, Zoom: zoom}
				if h.nc.View() != want || drawn != want || h.nc.gesture.kind != gesturePan {
					t.Fatalf("pan setup: drawn=%+v committed=%+v gesture=%v", drawn, h.nc.View(), h.nc.gesture.kind)
				}

				invalid := imgui.Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32}
				switch loss {
				case "focus_invalid_mouse":
					h.io.AddFocusEvent(false)
					h.point(invalid)
				case "focus_valid_mouse":
					h.io.AddFocusEvent(false)
					h.point(imgui.Vec2{X: 600, Y: 450})
				case "invalid_mouse_held":
					h.point(invalid)
				case "invalid_mouse_release":
					h.point(invalid)
					h.io.AddMouseButtonEvent(2, false)
				case "lost_button":
					// no release event: model a backend losing its held-button state.
					h.io.ClearInputMouse()
					h.point(imgui.Vec2{X: 600, Y: 450})
				}
				h.frame(draw)
				if drawn != want || h.nc.View() != want || h.nc.gesture.kind != gestureIdle {
					t.Fatalf("input loss changed pan: drawn=%+v committed=%+v gesture=%v; want %+v idle", drawn, h.nc.View(), h.nc.gesture.kind, want)
				}

				// restored position/focus must not resume the canceled gesture,
				// even if a backend still reports the middle button held.
				h.io.AddFocusEvent(true)
				h.point(imgui.Vec2{X: 500, Y: 400})
				h.frame(draw)
				if drawn != want || h.nc.View() != want || h.nc.gesture.kind != gestureIdle {
					t.Fatal("canceled pan resumed when input returned")
				}
				h.io.AddMouseButtonEvent(2, false)
				h.frame(draw)
				h.io.AddMouseButtonEvent(2, true)
				h.frame(draw)
				h.point(imgui.Vec2{X: 510, Y: 400})
				h.frame(draw)
				want.Pan.X += 10 / zoom
				if !approxVec2(drawn.Pan, want.Pan) || !approxVec2(h.nc.View().Pan, want.Pan) || h.nc.gesture.kind != gesturePan {
					t.Fatal("a fresh pan failed after input recovery")
				}
			})
		}
	}
}

func TestNodeCanvas_PanReleaseOutsideCanvasCommitsFinalPosition(t *testing.T) {
	for _, zoom := range []float32{0.5, 1, 1.5} {
		t.Run(fmt.Sprintf("zoom_%g", zoom), func(t *testing.T) {
			h := newCanvasWidgetTest(t)
			h.nc.SetView(View{Zoom: zoom})
			var drawn View
			var hovered bool
			draw := func() {
				drawn = h.nc.frameView
				hovered = imgui.IsWindowHovered()
			}
			h.point(imgui.Vec2{X: 450, Y: 350})
			h.frame(draw)
			h.frame(draw)
			h.io.AddMouseButtonEvent(2, true)
			h.frame(draw)
			h.point(imgui.Vec2{X: 490, Y: 370})
			h.frame(draw)
			h.point(imgui.Vec2{X: 980, Y: 760})
			h.io.AddMouseButtonEvent(2, false)
			h.frame(draw)
			want := imgui.Vec2{X: 530 / zoom, Y: 410 / zoom}
			if hovered || !approxVec2(drawn.Pan, want) || !approxVec2(h.nc.View().Pan, want) || h.nc.gesture.kind != gestureIdle {
				t.Fatalf("outside release: drawn=%+v committed=%+v gesture=%v hovered=%v", drawn, h.nc.View(), h.nc.gesture.kind, hovered)
			}
		})
	}
}
