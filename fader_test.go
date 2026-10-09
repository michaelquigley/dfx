package dfx

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
)

func faderTestFrame(t *testing.T) (*imgui.IO, func(func())) {
	t.Helper()
	runtime.LockOSThread()
	ctx := imgui.CreateContext()
	io := imgui.CurrentIO()
	io.SetIniFilename("")
	io.SetDisplaySize(imgui.Vec2{X: 600, Y: 500})
	io.SetDeltaTime(1.0 / 60)
	io.SetBackendFlags(imgui.BackendFlagsRendererHasTextures)
	t.Cleanup(func() {
		imgui.DestroyContextV(ctx)
		runtime.UnlockOSThread()
	})
	return io, func(draw func()) {
		imgui.NewFrame()
		imgui.SetNextWindowPos(imgui.Vec2{})
		imgui.SetNextWindowSize(imgui.Vec2{X: 600, Y: 500})
		imgui.BeginV("fader test", nil, imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove)
		draw()
		imgui.End()
		imgui.Render()
	}
}

func TestFader_HandleAndScaleMatchNativeSliderTravel(t *testing.T) {
	for _, tc := range []struct {
		name  string
		taper Taper
	}{{"linear", LinearTaper()}, {"audio", AudioTaper()}} {
		t.Run(tc.name, func(t *testing.T) {
			taper := tc.taper
			io, frame := faderTestFrame(t)
			params := FaderParams{Width: 60, Height: 300, HandleHeight: 28, Taper: taper}
			value := float32(0.5)
			var g faderGeometry
			draw := func() {
				value, _ = FaderWithScaleN("##gain", value, params, DefaultScaleConfig())
				g = newFaderGeometry(imgui.ItemRectMin(), imgui.ItemRectMax(), resolveFaderParams(params))
			}
			frame(draw)
			// track clicks at a scale mark must put the handle at that mark,
			// including endpoints and a nonlinear taper.
			for _, mark := range []float32{0, 1, 0.25, 0.75} {
				value = 0.5
				frame(draw)
				y := g.y(taper.Apply(mark))
				io.AddMousePosEvent(g.centerX, y)
				frame(draw)
				io.AddMouseButtonEvent(0, true)
				frame(draw)
				if math.Abs(float64(g.y(taper.Apply(value))-y)) > 1.25 {
					t.Fatalf("mark %v: native value %v", mark, value)
				}
				io.AddMouseButtonEvent(0, false)
				frame(draw)
			}
			// clicking off-center on the handle must preserve its grab offset.
			before := value
			io.AddMousePosEvent(g.centerX, g.y(taper.Apply(value))+5)
			frame(draw)
			io.AddMouseButtonEvent(0, true)
			frame(draw)
			if math.Abs(float64(value-before)) > 0.003 {
				t.Fatalf("handle jumped on grab: %v -> %v", before, value)
			}
			io.AddMousePosEvent(g.centerX, g.y(taper.Apply(before))+5-30)
			frame(draw)
			if value <= before {
				t.Fatalf("upward drag failed: %v -> %v", before, value)
			}
			io.AddMouseButtonEvent(0, false)
			frame(draw)
		})
	}
}

func TestFader_WheelResetStopsAndDisabled(t *testing.T) {
	io, frame := faderTestFrame(t)
	params := FaderParams{MinStop: 0.2, MaxStop: 0.8, ResetValue: 0.6, HandleHeight: 28}
	value := float32(0.5)
	var changed, disabled bool
	var g faderGeometry
	draw := func() {
		imgui.BeginDisabledV(disabled)
		value, changed = FaderN("##gain", value, params)
		g = newFaderGeometry(imgui.ItemRectMin(), imgui.ItemRectMax(), resolveFaderParams(params))
		imgui.EndDisabled()
	}
	frame(draw)
	io.AddMousePosEvent(g.centerX, g.y(0.5))
	frame(draw)
	for _, modifier := range []struct {
		key  imgui.Key
		step float32
	}{{imgui.KeyNone, 0.01}, {imgui.ModCtrl, 0.1}, {imgui.ModAlt, 0.001}} {
		if modifier.key != imgui.KeyNone {
			io.AddKeyEvent(modifier.key, true)
		}
		before := value
		io.AddMouseWheelEvent(0, 1)
		frame(draw)
		if math.Abs(float64(value-before-modifier.step)) > 0.00001 || !changed {
			t.Fatalf("wheel %v: %v -> %v, changed=%v", modifier.key, before, value, changed)
		}
		if modifier.key != imgui.KeyNone {
			io.AddKeyEvent(modifier.key, false)
		}
	}
	io.AddMouseWheelEvent(0, 100)
	frame(draw)
	if value != 0.8 {
		t.Fatalf("upper stop: %v", value)
	}
	io.AddMouseButtonEvent(1, true)
	frame(draw)
	if value != 0.6 || !changed {
		t.Fatalf("reset: %v, changed=%v", value, changed)
	}
	io.AddMouseButtonEvent(1, false)
	frame(draw)
	disabled = true
	frame(draw)
	io.AddMouseWheelEvent(0, -100)
	io.AddMouseButtonEvent(1, true)
	frame(draw)
	if value != 0.6 || changed {
		t.Fatalf("disabled fader changed: %v, %v", value, changed)
	}
}

func TestFader_DefaultScaleAndStyleRestoration(t *testing.T) {
	_, frame := faderTestFrame(t)
	frame(func() {
		style := imgui.CurrentStyle()
		colors, grab, border := style.Colors(), style.GrabMinSize(), style.FrameBorderSize()
		value, changed := FaderWithScaleF("gain", -10, -40, 10, FaderParams{}, DefaultScaleConfig())
		if math.Abs(float64(value+10)) > 0.00001 || changed {
			t.Fatalf("idle float range changed: %v, %v", value, changed)
		}
		if style.Colors() != colors || style.GrabMinSize() != grab || style.FrameBorderSize() != border {
			t.Fatal("fader leaked its temporary style")
		}
		imgui.SameLine()
		integer, changed := FaderWithScaleI("##integer", 64, 0, 127, FaderParams{}, DefaultScaleConfig())
		if integer != 64 || changed {
			t.Fatalf("idle integer range changed: %v, %v", integer, changed)
		}
	})
}

func TestFader_ThemeChangesAndColorOverrides(t *testing.T) {
	_, frame := faderTestFrame(t)
	params := DefaultFaderParams() // reusable across theme changes
	var previous uint32
	for _, theme := range []Theme{ModernDark, GreenTheme} {
		frame(func() {
			SetTheme(theme)
			expected := imgui.ColorU32Col(imgui.ColButtonHovered)
			if expected == previous {
				t.Fatal("test themes must have different accents")
			}
			previous = expected
			dl := imgui.WindowDrawList()
			start := dl.VtxBuffer().Size
			FaderN("##gain", 0.5, params)
			found := false
			for _, color := range faderVertexColors(dl, start) {
				if color == expected {
					found = true
				}
			}
			if !found {
				t.Fatalf("fader did not render %s's accent", theme.Name())
			}
		})
	}
	// explicit colors still win, including their alpha and disabled opacity.
	accent := imgui.Vec4{X: 0.85, Y: 0.1, Z: 0.65, W: 0.7}
	track := imgui.Vec4{X: 0.2, Y: 0.4, Z: 0.6, W: 0.8}
	params.AccentColor, params.TrackColor = &accent, &track
	frame(func() {
		imgui.BeginDisabled()
		expected := map[uint32]bool{imgui.ColorU32Vec4(accent): false, imgui.ColorU32Vec4(track): false}
		dl := imgui.WindowDrawList()
		start := dl.VtxBuffer().Size
		FaderN("##gain", 0.5, params)
		for _, color := range faderVertexColors(dl, start) {
			if _, ok := expected[color]; ok {
				expected[color] = true
			}
		}
		imgui.EndDisabled()
		for color, found := range expected {
			if !found {
				t.Fatalf("missing disabled override color %#x", color)
			}
		}
	})
}

// cimgui-go's vector Data is a wrapper around the first C vertex, rather than
// a contiguous array of Go wrappers. walk the C array using its actual stride.
func faderVertexColors(dl *imgui.DrawList, start int) []uint32 {
	buffer := dl.VtxBuffer()
	ptr, done := buffer.Data.Handle()
	defer done()
	colors := make([]uint32, 0, buffer.Size-start)
	for i := start; i < buffer.Size; i++ {
		vertex := imgui.NewDrawVertFromC(unsafe.Add(unsafe.Pointer(ptr), uintptr(i)*unsafe.Sizeof(*ptr)))
		colors = append(colors, vertex.Col())
	}
	return colors
}
