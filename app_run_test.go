package dfx

import (
	"errors"
	"testing"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
)

func TestRun_ReturnsBackendInitializationError(t *testing.T) {
	expectedErr := errors.New("backend init failed")

	originalCreateBackend := createBackend
	createBackend = func() (backend.Backend[glfwbackend.GLFWWindowFlags], error) {
		return nil, expectedErr
	}
	defer func() {
		createBackend = originalCreateBackend
	}()

	app := New(nil, Config{})
	err := app.Run()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error '%v', got '%v'", expectedErr, err)
	}

	waitErr := app.Wait()
	if !errors.Is(waitErr, expectedErr) {
		t.Fatalf("expected Wait to return '%v', got '%v'", expectedErr, waitErr)
	}
}

func TestRootWindowRect_NoMenuBarUsesFullViewport(t *testing.T) {
	viewport := imgui.Vec2{X: 800, Y: 600}
	pos, size := rootWindowRect(viewport, 0, false)

	if pos.X != 0 || pos.Y != 0 {
		t.Fatalf("expected root position '(0,0)', got '(%.2f,%.2f)'", pos.X, pos.Y)
	}
	if size.X != 800 || size.Y != 600 {
		t.Fatalf("expected root size '(800,600)', got '(%.2f,%.2f)'", size.X, size.Y)
	}
}

func TestRootWindowRect_MenuBarOffsetsAndShrinks(t *testing.T) {
	viewport := imgui.Vec2{X: 1024, Y: 768}
	pos, size := rootWindowRect(viewport, 31, true)

	if pos.X != 0 || pos.Y != 31 {
		t.Fatalf("expected root position '(0,31)', got '(%.2f,%.2f)'", pos.X, pos.Y)
	}
	if size.X != 1024 || size.Y != 737 {
		t.Fatalf("expected root size '(1024,737)', got '(%.2f,%.2f)'", size.X, size.Y)
	}
}

func TestRootWindowRect_ClampsSmallViewportHeight(t *testing.T) {
	viewport := imgui.Vec2{X: 400, Y: 10}
	pos, size := rootWindowRect(viewport, 25, true)

	if pos.Y != 25 {
		t.Fatalf("expected root position y '25', got '%.2f'", pos.Y)
	}
	if size.Y != 0 {
		t.Fatalf("expected clamped root height '0', got '%.2f'", size.Y)
	}
}
