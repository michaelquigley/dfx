package dfx

import (
	"errors"
	"testing"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
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
