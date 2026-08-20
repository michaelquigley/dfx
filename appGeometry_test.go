package dfx

import (
	"testing"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
)

// geometryBackend records the geometry calls the app makes. it embeds the backend
// interface so only the methods under test need implementing -- reaching any other
// method nil-panics, which is intended: these tests must never touch a real window.
type geometryBackend struct {
	backend.Backend[glfwbackend.GLFWWindowFlags]
	sizes [][2]int
	poss  [][2]int
}

func (b *geometryBackend) SetWindowSize(w, h int) { b.sizes = append(b.sizes, [2]int{w, h}) }
func (b *geometryBackend) SetWindowPos(x, y int)  { b.poss = append(b.poss, [2]int{x, y}) }

func TestWindowGeometryDeferredToFrameBoundary(t *testing.T) {
	be := &geometryBackend{}
	app := New(nil, Config{})
	app.backend = be

	// requesting geometry from inside a frame must not reach the backend. on wayland
	// the glfw call dispatches a surface configure synchronously, which re-enters the
	// render loop and opens a second imgui frame inside the current one.
	app.SetWindowSize(1280, 1024)
	app.SetWindowPos(40, 50)
	if len(be.sizes) != 0 || len(be.poss) != 0 {
		t.Fatalf("expected no backend calls before the frame boundary, got sizes '%v' pos '%v'", be.sizes, be.poss)
	}

	// the afterRender hook flushes them, size before position.
	app.applyPendingGeometry()
	if len(be.sizes) != 1 || be.sizes[0] != [2]int{1280, 1024} {
		t.Fatalf("expected one size call '(1280,1024)', got '%v'", be.sizes)
	}
	if len(be.poss) != 1 || be.poss[0] != [2]int{40, 50} {
		t.Fatalf("expected one position call '(40,50)', got '%v'", be.poss)
	}

	// a boundary with nothing pending is a no-op. reasserting geometry every frame
	// would fight the user's own resizing of the window.
	app.applyPendingGeometry()
	if len(be.sizes) != 1 || len(be.poss) != 1 {
		t.Fatalf("expected no repeat calls, got sizes '%v' pos '%v'", be.sizes, be.poss)
	}
}

func TestWindowGeometryLastRequestWins(t *testing.T) {
	be := &geometryBackend{}
	app := New(nil, Config{})
	app.backend = be

	app.SetWindowSize(800, 600)
	app.SetWindowSize(1024, 768)
	app.applyPendingGeometry()

	if len(be.sizes) != 1 || be.sizes[0] != [2]int{1024, 768} {
		t.Fatalf("expected only the last size '(1024,768)', got '%v'", be.sizes)
	}
}

func TestWindowGeometryWithoutBackendIsInert(t *testing.T) {
	// geometry may be requested from OnSetup, before Run() creates the backend.
	app := New(nil, Config{})
	app.SetWindowSize(800, 600)
	app.SetWindowPos(10, 20)
	app.applyPendingGeometry()
}
