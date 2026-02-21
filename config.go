package dfx

import (
	"os"
	"path/filepath"

	"github.com/michaelquigley/df/dd"
	"github.com/pkg/errors"
)

// DashConfig holds configuration for a single dashboard panel
type DashConfig struct {
	Visible bool
	Size    int
}

// WindowConfig holds window position and size configuration
type WindowConfig struct {
	X         int
	Y         int
	Width     int
	Height    int
	Maximized bool // window maximized state (capture only, restore not yet implemented)
}

// GetDefaultWindowConfig returns sensible default window configuration
func GetDefaultWindowConfig() WindowConfig {
	return WindowConfig{
		X:      100,
		Y:      100,
		Width:  800,
		Height: 600,
	}
}

// ConfigPath returns a standard configuration file path in the user's home directory
// Example: ConfigPath("myapp", "config.json") -> "~/.myapp/config.json"
func ConfigPath(appName, filename string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "error getting user home directory")
	}
	return filepath.Join(home, "."+appName, filename), nil
}

// SaveJSON saves any struct to a JSON file with proper formatting and error handling
// Creates parent directories if they don't exist
func SaveJSON(path string, config interface{}) error {
	// create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrapf(err, "error creating directory '%v'", dir)
	}
	return dd.UnbindJSONFile(config, path)
}

// LoadJSON loads a JSON file into a struct
// If the file doesn't exist, the config parameter is left unchanged (use defaults)
// Returns error only if the file exists but can't be read or parsed
func LoadJSON(path string, config interface{}) error {
	// check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // file doesn't exist, use defaults
	}
	if err := dd.MergeJSONFile(config, path); err != nil {
		return err
	}
	return nil
}

// dashSlots returns a map of slot name to dash pointer for iteration.
func dashSlots(dm *DashManager) map[string]*Dash {
	return map[string]*Dash{
		"top": dm.Top, "left": dm.Left, "right": dm.Right, "bottom": dm.Bottom,
	}
}

// CaptureDashState extracts configuration from a DashManager.
// returns a map with keys: "top", "left", "right", "bottom"
func CaptureDashState(dm *DashManager) map[string]DashConfig {
	config := make(map[string]DashConfig)
	for name, dash := range dashSlots(dm) {
		if dash != nil {
			config[name] = DashConfig{Visible: dash.Visible, Size: dash.TargetSize}
		}
	}
	return config
}

// RestoreDashState applies configuration to a DashManager.
// accepts a map with keys: "top", "left", "right", "bottom"
func RestoreDashState(dm *DashManager, config map[string]DashConfig) {
	for name, dash := range dashSlots(dm) {
		if dash != nil {
			if cfg, ok := config[name]; ok {
				dash.Visible = cfg.Visible
				dash.TargetSize = cfg.Size
				dash.CurrentSize = cfg.Size
			}
		}
	}
}

// CaptureWindowState gets current window state from App
func CaptureWindowState(app *App) WindowConfig {
	x, y := app.GetWindowPos()
	width, height := app.GetWindowSize()

	// TODO: Capture maximized state when backend supports GetWindowMaximized()
	// For now, always set to false
	maximized := false

	return WindowConfig{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Maximized: maximized,
	}
}
