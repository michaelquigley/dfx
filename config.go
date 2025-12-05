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

// CaptureDashState extracts configuration from a DashManager
// Returns a map with keys: "top", "left", "right", "bottom"
func CaptureDashState(dm *DashManager) map[string]DashConfig {
	config := make(map[string]DashConfig)

	if dm.Top != nil {
		config["top"] = DashConfig{
			Visible: dm.Top.Visible,
			Size:    dm.Top.TargetSize,
		}
	}

	if dm.Left != nil {
		config["left"] = DashConfig{
			Visible: dm.Left.Visible,
			Size:    dm.Left.TargetSize,
		}
	}

	if dm.Right != nil {
		config["right"] = DashConfig{
			Visible: dm.Right.Visible,
			Size:    dm.Right.TargetSize,
		}
	}

	if dm.Bottom != nil {
		config["bottom"] = DashConfig{
			Visible: dm.Bottom.Visible,
			Size:    dm.Bottom.TargetSize,
		}
	}

	return config
}

// RestoreDashState applies configuration to a DashManager
// Accepts a map with keys: "top", "left", "right", "bottom"
func RestoreDashState(dm *DashManager, config map[string]DashConfig) {
	if dm.Top != nil {
		if cfg, ok := config["top"]; ok {
			dm.Top.Visible = cfg.Visible
			dm.Top.TargetSize = cfg.Size
			dm.Top.CurrentSize = cfg.Size
		}
	}

	if dm.Left != nil {
		if cfg, ok := config["left"]; ok {
			dm.Left.Visible = cfg.Visible
			dm.Left.TargetSize = cfg.Size
			dm.Left.CurrentSize = cfg.Size
		}
	}

	if dm.Right != nil {
		if cfg, ok := config["right"]; ok {
			dm.Right.Visible = cfg.Visible
			dm.Right.TargetSize = cfg.Size
			dm.Right.CurrentSize = cfg.Size
		}
	}

	if dm.Bottom != nil {
		if cfg, ok := config["bottom"]; ok {
			dm.Bottom.Visible = cfg.Visible
			dm.Bottom.TargetSize = cfg.Size
			dm.Bottom.CurrentSize = cfg.Size
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
