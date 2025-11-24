package main

import (
	"fmt"

	"github.com/michaelquigley/dfx"
	"github.com/AllenDang/cimgui-go/imgui"
)

// appConfig holds all persistent configuration for the application
type appConfig struct {
	// window configuration
	Window dfx.WindowConfig

	// dashboard configuration
	Dashes map[string]dfx.DashConfig

	// application-specific settings
	Counter      int
	LastMessage  string
	ShowWelcome  bool
	EnableDebug  bool
	SelectedMode int
}

// defaultConfig returns sensible defaults for the application
func defaultConfig() *appConfig {
	return &appConfig{
		Window: dfx.GetDefaultWindowConfig(),
		Dashes: map[string]dfx.DashConfig{
			"top":    {Visible: true, Size: 100},
			"left":   {Visible: true, Size: 250},
			"right":  {Visible: true, Size: 250},
			"bottom": {Visible: true, Size: 50},
		},
		Counter:      0,
		LastMessage:  "",
		ShowWelcome:  true,
		EnableDebug:  false,
		SelectedMode: 0,
	}
}

// appState holds runtime state (not persisted)
type appState struct {
	cfg                *appConfig
	cfgPath            string
	message            string
	dashMgr            *dfx.DashManager
	saveCount          int
	configLoadedFromFS bool
}

func main() {
	// determine config file path
	cfgPath, err := dfx.ConfigPath("dfx-example-config", "config.json")
	if err != nil {
		panic(err)
	}

	// load configuration with defaults
	cfg := defaultConfig()
	if err := dfx.LoadJSON(cfgPath, cfg); err != nil {
		fmt.Printf("error loading config: %v\n", err)
	} else {
		fmt.Printf("loaded config from `%s`\n", cfgPath)
	}

	// create application state
	state := &appState{
		cfg:                cfg,
		cfgPath:            cfgPath,
		message:            "configuration loaded successfully",
		configLoadedFromFS: true,
	}

	// create dashboard panels
	state.dashMgr = dfx.NewDashManager()
	state.dashMgr.Precedence = dfx.HorizontalPrecedence

	// top dashboard - debug info
	state.dashMgr.Top = dfx.NewDash("Debug", dfx.NewFunc(func(s *dfx.State) {
		dfx.Text(fmt.Sprintf("Config Path: %s", state.cfgPath))
		dfx.Text(fmt.Sprintf("Save Count: %d", state.saveCount))
		dfx.Text(fmt.Sprintf("Window: %dx%d at (%d,%d)",
			state.cfg.Window.Width, state.cfg.Window.Height,
			state.cfg.Window.X, state.cfg.Window.Y))
	}))

	// left dashboard - settings
	state.dashMgr.Left = dfx.NewDash("Settings", dfx.NewFunc(func(s *dfx.State) {
		dfx.Text("Application Settings")
		dfx.Separator()

		if newValue, changed := dfx.Checkbox("Show welcome message", state.cfg.ShowWelcome); changed {
			state.cfg.ShowWelcome = newValue
		}

		if newValue, changed := dfx.Checkbox("Enable debug mode", state.cfg.EnableDebug); changed {
			state.cfg.EnableDebug = newValue
		}

		dfx.Spacing()
		modes := []string{"Standard", "Advanced", "Expert"}
		if newIdx, changed := dfx.Combo("Mode", state.cfg.SelectedMode, modes); changed {
			state.cfg.SelectedMode = newIdx
			state.message = fmt.Sprintf("switched to '%s' mode", modes[newIdx])
		}

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		if dfx.Button("Save Config Now") {
			if err := state.saveConfig(); err != nil {
				state.message = fmt.Sprintf("error saving: %v", err)
			} else {
				state.message = "configuration saved manually"
			}
		}

		if state.cfg.EnableDebug {
			dfx.Spacing()
			dfx.Separator()
			dfx.Spacing()
			dfx.Text("Debug Controls")

			if dfx.Button("Reset to Defaults") {
				state.cfg = defaultConfig()
				state.message = "configuration reset to defaults"
			}
		}
	}))

	// right dashboard - info
	state.dashMgr.Right = dfx.NewDash("Info", dfx.NewFunc(func(s *dfx.State) {
		if state.cfg.ShowWelcome {
			dfx.Text("Welcome to Configuration Example!")
			dfx.Spacing()
		}

		dfx.Text("This example demonstrates:")
		imgui.BulletText("Loading/saving JSON configuration")
		imgui.BulletText("Window state persistence")
		imgui.BulletText("Dashboard state persistence")
		imgui.BulletText("Application settings persistence")
		dfx.Spacing()

		dfx.Separator()
		dfx.Spacing()

		dfx.Text("Try the following:")
		imgui.BulletText("Resize the window")
		imgui.BulletText("Move the window")
		imgui.BulletText("Toggle dashboards (Alt+T/L/R/B)")
		imgui.BulletText("Resize dashboards")
		imgui.BulletText("Change settings")
		imgui.BulletText("Close and reopen the app")
		dfx.Spacing()

		dfx.Text("All changes are automatically saved!")
	}))

	// bottom dashboard - status bar
	state.dashMgr.Bottom = dfx.NewDash("Status", dfx.NewFunc(func(s *dfx.State) {
		dfx.Text(fmt.Sprintf("Counter: %d  |  Status: %s", state.cfg.Counter, state.message))
	}))

	// inner component - main content
	state.dashMgr.Inner = dfx.NewFunc(func(s *dfx.State) {
		imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.7, Y: 0.9, Z: 1.0, W: 1.0})
		dfx.Text("Configuration Persistence Demo")
		imgui.PopStyleColor()

		dfx.Separator()
		dfx.Spacing()

		dfx.Text(fmt.Sprintf("Current counter value: %d", state.cfg.Counter))

		if dfx.Button("Increment Counter") {
			state.cfg.Counter++
			state.message = fmt.Sprintf("counter incremented to %d", state.cfg.Counter)
		}

		dfx.SameLine()
		if dfx.Button("Decrement Counter") {
			state.cfg.Counter--
			state.message = fmt.Sprintf("counter decremented to %d", state.cfg.Counter)
		}

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		dfx.Text("Enter a message:")
		if newValue, changed := dfx.Input("##message", state.cfg.LastMessage); changed {
			state.cfg.LastMessage = newValue
			state.message = "message updated"
		}

		if state.cfg.LastMessage != "" {
			dfx.Spacing()
			dfx.Text(fmt.Sprintf("Last saved message: '%s'", state.cfg.LastMessage))
		}

		dfx.Spacing()
		dfx.Separator()
		dfx.Spacing()

		dfx.Text("Dashboard visibility can be toggled with:")
		imgui.BulletText("Alt+T - Toggle top (debug)")
		imgui.BulletText("Alt+L - Toggle left (settings)")
		imgui.BulletText("Alt+R - Toggle right (info)")
		imgui.BulletText("Alt+B - Toggle bottom (status)")
	})

	// restore dashboard state from configuration
	dfx.RestoreDashState(state.dashMgr, state.cfg.Dashes)

	// create root component that wraps the DashManager
	root := dfx.NewFunc(func(s *dfx.State) {
		state.dashMgr.Draw(s)
	})

	// create application with lifecycle callbacks
	app := dfx.New(root, dfx.Config{
		Title:  "Configuration Example",
		Width:  state.cfg.Window.Width,
		Height: state.cfg.Window.Height,
		X:      state.cfg.Window.X,
		Y:      state.cfg.Window.Y,

		OnSetup: func(app *dfx.App) {
			// register keyboard shortcuts for toggling dashboards
			app.Actions().Register("toggle-top", "Alt+T", func() {
				if state.dashMgr.Top != nil {
					state.dashMgr.Top.Visible = !state.dashMgr.Top.Visible
					state.message = fmt.Sprintf("debug dashboard: %v", state.dashMgr.Top.Visible)
				}
			})

			app.Actions().Register("toggle-left", "Alt+L", func() {
				if state.dashMgr.Left != nil {
					state.dashMgr.Left.Visible = !state.dashMgr.Left.Visible
					state.message = fmt.Sprintf("settings dashboard: %v", state.dashMgr.Left.Visible)
				}
			})

			app.Actions().Register("toggle-right", "Alt+R", func() {
				if state.dashMgr.Right != nil {
					state.dashMgr.Right.Visible = !state.dashMgr.Right.Visible
					state.message = fmt.Sprintf("info dashboard: %v", state.dashMgr.Right.Visible)
				}
			})

			app.Actions().Register("toggle-bottom", "Alt+B", func() {
				if state.dashMgr.Bottom != nil {
					state.dashMgr.Bottom.Visible = !state.dashMgr.Bottom.Visible
					state.message = fmt.Sprintf("status dashboard: %v", state.dashMgr.Bottom.Visible)
				}
			})
		},

		OnClose: func(app *dfx.App) {
			// save configuration on close
			state.cfg.Window = dfx.CaptureWindowState(app)
			state.cfg.Dashes = dfx.CaptureDashState(state.dashMgr)

			if err := state.saveConfig(); err != nil {
				fmt.Printf("error saving config on close: %v\n", err)
			} else {
				fmt.Println("configuration saved on close")
			}
		},

		OnSizeChange: func(width, height int) {
			// update window size in configuration
			state.cfg.Window.Width = width
			state.cfg.Window.Height = height
			state.message = fmt.Sprintf("window resized to %dx%d", width, height)
		},
	})

	app.Run()
}

func (s *appState) saveConfig() error {
	s.saveCount++
	return dfx.SaveJSON(s.cfgPath, s.cfg)
}
