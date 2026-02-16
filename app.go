package dfx

import (
	"image"
	"time"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
)

type App struct {
	backend   backend.Backend[glfwbackend.GLFWWindowFlags]
	root      Component
	config    Config
	running   bool
	actions   *ActionRegistry
	startTime time.Time
	done      chan struct{} // signals Run() completion
	runErr    error         // stores error from Run()
}

type Config struct {
	Title          string
	Width          int
	Height         int
	X              int            // window X position (0 = don't set)
	Y              int            // window Y position (0 = don't set)
	OnSetup        func(*App)     // called once after imgui context created
	OnShutdown     func(*App)     // called before shutdown
	OnTick         func(*App)     // called each frame before drawing
	OnClose        func(*App)     // called when window is about to close (can call SetShouldClose to cancel)
	OnSizeChange   func(int, int) // called when window is resized
	MenuBar        Component      // optional menu bar component
	Theme          Theme          // optional theme (defaults to DefaultTheme)
	DisableFonts   bool           // if true, skip font setup (use default ImGui fonts)
	DisableTheming bool           // if true, skip theme setup (use default ImGui theme)
	Icons          []image.Image  // optional window icons
}

func New(root Component, config Config) *App {
	// set defaults
	if config.Title == "" {
		config.Title = "dfx"
	}
	if config.Width == 0 {
		config.Width = 800
	}
	if config.Height == 0 {
		config.Height = 600
	}

	return &App{
		root:    root,
		config:  config,
		actions: NewActionRegistry(),
		done:    make(chan struct{}),
	}
}

func (app *App) Run() error {
	defer close(app.done)

	// record start time for log timestamps
	app.startTime = time.Now()

	app.backend, _ = backend.CreateBackend(glfwbackend.NewGLFWBackend())
	app.backend.CreateWindow(app.config.Title, app.config.Width, app.config.Height)

	// set window position if specified
	if app.config.X != 0 || app.config.Y != 0 {
		app.backend.SetWindowPos(app.config.X, app.config.Y)
	}

	// set window icons if specified
	if len(app.config.Icons) > 0 {
		app.backend.SetIcons(app.config.Icons...)
	}

	// setup fonts and styling
	app.setupFontsAndTheme()

	// user setup
	if app.config.OnSetup != nil {
		app.config.OnSetup(app)
	}
	imgui.CurrentIO().SetConfigFlags(imgui.ConfigFlagsNone)

	// setup window callbacks
	if app.config.OnClose != nil {
		app.backend.SetCloseCallback(func() {
			app.config.OnClose(app)
		})
	}
	if app.config.OnSizeChange != nil {
		app.backend.SetSizeChangeCallback(func(width, height int) {
			app.config.OnSizeChange(width, height)
		})
	}

	// run the main loop
	app.running = true
	app.backend.Run(func() {
		if !app.running {
			app.backend.SetShouldClose(true)
			return
		}

		// user tick
		if app.config.OnTick != nil {
			app.config.OnTick(app)
		}

		// draw menu bar if configured
		if app.config.MenuBar != nil {
			if imgui.BeginMainMenuBar() {
				menuState := &State{
					Size:     imgui.Vec2{X: 0, Y: 0}, // menu bar size is managed by imgui
					Position: imgui.Vec2{},
					IO:       imgui.CurrentIO(),
					App:      app,
					Parent:   nil,
				}
				app.config.MenuBar.Draw(menuState)
				imgui.EndMainMenuBar()
			}
		}

		// create an invisible full-window imgui window
		size := imgui.WindowViewport().Size()
		rootFlags := imgui.WindowFlagsAlwaysAutoResize |
			imgui.WindowFlagsNoSavedSettings |
			imgui.WindowFlagsNoTitleBar |
			imgui.WindowFlagsNoScrollbar |
			imgui.WindowFlagsNoScrollWithMouse

		// adjust for menu bar if present
		var windowPos imgui.Vec2
		var windowSize imgui.Vec2
		if app.config.MenuBar != nil {
			windowPos = imgui.Vec2{X: 0, Y: 25} // menu bar height
			windowSize = size.Sub(imgui.Vec2{X: 0, Y: 25})
		} else {
			windowPos = imgui.Vec2{X: 0, Y: 0}
			windowSize = size
		}

		imgui.SetNextWindowPos(windowPos)
		imgui.SetNextWindowSize(windowSize)

		if imgui.BeginV("##dfx_root", nil, rootFlags) {
			// create state for root component
			io := imgui.CurrentIO()
			state := &State{
				Size:     windowSize,
				Position: imgui.Vec2{}, // position is relative to window
				IO:       io,
				App:      app,
				Parent:   nil,
			}

			// handle events
			app.processEvents(state)

			// draw root component
			if app.root != nil {
				app.root.Draw(state)
			}
		}
		imgui.End()
	})

	// shutdown
	if app.config.OnShutdown != nil {
		app.config.OnShutdown(app)
	}

	app.runErr = nil
	return app.runErr
}

// Stop signals the app to stop running
func (app *App) Stop() {
	app.running = false
}

// Wait blocks until Run() completes and returns any error from Run()
func (app *App) Wait() error {
	<-app.done
	return app.runErr
}

// SetRoot changes the root component
func (app *App) SetRoot(root Component) {
	app.root = root
}

// Actions returns the action registry
func (app *App) Actions() *ActionRegistry {
	return app.actions
}

// SetWindowTitle updates the window title
func (app *App) SetWindowTitle(title string) {
	if app.backend != nil {
		app.backend.SetWindowTitle(title)
	}
}

// SetShouldClose sets whether the window should close
// this can be used in OnClose callback to cancel closing
func (app *App) SetShouldClose(shouldClose bool) {
	if app.backend != nil {
		app.backend.SetShouldClose(shouldClose)
	}
}

// GetWindowSize returns the current window size
func (app *App) GetWindowSize() (int, int) {
	if app.backend != nil {
		w, h := app.backend.DisplaySize()
		return int(w), int(h)
	}
	return 0, 0
}

// GetWindowPos returns the current window position
func (app *App) GetWindowPos() (int, int) {
	if app.backend != nil {
		x, y := app.backend.GetWindowPos()
		return int(x), int(y)
	}
	return 0, 0
}

// setupFontsAndTheme initializes fonts and applies theme
func (app *App) setupFontsAndTheme() {
	// setup fonts unless disabled
	if !app.config.DisableFonts {
		SetupFonts()
	}

	// apply default style
	DefaultStyle()

	// apply theme unless disabled
	if !app.config.DisableTheming {
		theme := app.config.Theme
		if theme == nil {
			theme = &ModernTheme{}
		}
		SetTheme(theme)
	}
}

// processEvents converts imgui events to our event system
func (app *App) processEvents(state *State) {
	// collect all actions to check (component actions first, then global)
	var actionsToCheck []*ActionRegistry

	// gather component actions hierarchically
	if app.root != nil {
		actionsToCheck = app.gatherComponentActions(app.root)
	}

	// add global actions last
	actionsToCheck = append(actionsToCheck, app.actions)

	// get current modifiers once
	currentMods := app.getModifiers()

	// check each action to see if its key combo is pressed
	for _, registry := range actionsToCheck {
		for _, action := range registry.actions {
			if imgui.IsKeyPressedBool(action.key) {
				if action.mods == currentMods {
					if action.Handler != nil {
						action.Handler()
						return // stop processing after first match
					}
				}
			}
		}
	}
}

func (app *App) getModifiers() KeyModifier {
	var mod KeyModifier
	io := imgui.CurrentIO()
	if io.KeyCtrl() {
		mod |= ModCtrl
	}
	if io.KeyShift() {
		mod |= ModShift
	}
	if io.KeyAlt() {
		mod |= ModAlt
	}
	if io.KeySuper() {
		mod |= ModSuper
	}
	return mod
}

// gatherComponentActions collects all component actions hierarchically (children first)
func (app *App) gatherComponentActions(comp Component) []*ActionRegistry {
	var registries []*ActionRegistry

	// gather children's actions first (reverse order for proper z-ordering)
	if box, ok := comp.(*Container); ok {
		for i := len(box.Children) - 1; i >= 0; i-- {
			registries = append(registries, app.gatherComponentActions(box.Children[i])...)
		}
	}

	// then add this component's actions
	if actions := comp.Actions(); actions != nil && len(actions.actions) > 0 {
		registries = append(registries, actions)
	}

	return registries
}
