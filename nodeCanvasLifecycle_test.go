package dfx

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
)

func canvasNativeAllocations() (allocated, freed int32) {
	info := imgui.CurrentContext().DebugAllocInfo()
	return info.TotalAllocCount(), info.TotalFreeCount()
}

func canvasPanics(f func()) (panicked bool) {
	defer func() { panicked = recover() != nil }()
	f()
	return false
}

func TestNodeCanvas_DestroyBeforeBegin(t *testing.T) {
	nc := NewNodeCanvas[string](NodeCanvasConfig{})
	nc.Destroy()
	nc.Destroy()
	if !canvasPanics(func() { nc.Begin(nil) }) {
		t.Fatal("destroyed canvas accepted Begin")
	}
}

func TestNodeCanvas_DestroyDuringFrame(t *testing.T) {
	h := newCanvasWidgetTest(t)
	var rejected bool
	h.frame(func() {
		rejected = canvasPanics(h.nc.Destroy)
		h.nc.Node("node", imgui.Vec2{}, NodeFlags{}, func(n *NodeContext[string]) {
			n.Label("still drawable")
		})
	})
	if !rejected {
		t.Fatal("Destroy accepted an unmerged frame")
	}
	h.nc.Destroy()
	if h.nc.splitter != nil || h.nc.retained != nil || h.nc.frameNodes != nil || h.nc.framePins != nil {
		t.Fatal("Destroy retained native or frame state")
	}
	allocated, freed := canvasNativeAllocations()
	h.nc.Destroy()
	afterAllocated, afterFreed := canvasNativeAllocations()
	if allocated != afterAllocated || freed != afterFreed {
		t.Fatal("repeated Destroy touched native allocations")
	}
}

func TestNodeCanvas_NativeAllocationLifecycle(t *testing.T) {
	h := newCanvasWidgetTest(t)
	// warm imgui's delayed settings serialization along with the drawing
	// buffers, so its first save does not fall inside the measured frames.
	h.io.SetIniSavingRate(0.01)
	draw := func() {
		for i := 0; i < 12; i++ {
			h.nc.Node(fmt.Sprint(i), imgui.Vec2{X: float32(i * 30), Y: float32(i * 20)}, NodeFlags{}, func(n *NodeContext[string]) {
				n.Label("native buffers")
				n.Input(fmt.Sprintf("in%d", i), "in")
				n.Output(fmt.Sprintf("out%d", i), "out")
			})
			if i > 0 {
				h.nc.Link(fmt.Sprintf("link%d", i), fmt.Sprintf("out%d", i-1), fmt.Sprintf("in%d", i), LinkFlags{})
			}
		}
	}
	for i := 0; i < 20; i++ {
		h.frame(draw)
	}
	allocated, freed := canvasNativeAllocations()
	for i := 0; i < 1000; i++ {
		h.frame(draw)
	}
	afterAllocated, afterFreed := canvasNativeAllocations()
	if afterAllocated-afterFreed != allocated-freed {
		t.Fatalf("stable graph grew live native allocations: before=%d after=%d", allocated-freed, afterAllocated-afterFreed)
	}
	h.nc.Destroy()
	baselineAllocated, baselineFreed := canvasNativeAllocations()
	if baselineFreed <= afterFreed+1 {
		t.Fatal("Destroy did not free the splitter's channel buffers")
	}
	for cycle := 0; cycle < 25; cycle++ {
		// reuse the Go address so imgui reuses the child window's own caches;
		// this isolates canvas-owned resources from context-owned windows.
		*h.nc = *NewNodeCanvas[string](NodeCanvasConfig{})
		for i := 0; i < 20; i++ {
			h.frame(draw)
		}
		h.nc.Destroy()
		a, f := canvasNativeAllocations()
		if a-f != baselineAllocated-baselineFreed {
			t.Fatalf("cycle %d leaked native allocations: baseline=%d current=%d", cycle, baselineAllocated-baselineFreed, a-f)
		}
	}
}

func TestNodeCanvas_DestroyAfterContextShutdown(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	nc := NewNodeCanvas[string](NodeCanvasConfig{})
	// dfx calls OnShutdown after the backend destroys the imgui context.
	// keep the native splitter alive through that same order of teardown.
	func() {
		ctx := imgui.CreateContext()
		defer imgui.DestroyContextV(ctx)
		io := imgui.CurrentIO()
		io.SetIniFilename("")
		io.SetDisplaySize(imgui.Vec2{X: 1000, Y: 800})
		io.SetDeltaTime(1.0 / 60)
		io.SetBackendFlags(imgui.BackendFlagsRendererHasTextures)
		h := &canvasWidgetTest{nc: nc, io: io}
		for i := 0; i < 3; i++ {
			h.frame(func() {
				nc.Node("node", imgui.Vec2{}, NodeFlags{}, func(n *NodeContext[string]) {
					n.Label("teardown")
				})
			})
		}
	}()
	nc.Destroy()
	nc.Destroy()
}
