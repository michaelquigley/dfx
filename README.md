# dfx

dfx is a simplified second-generation immediate-mode GUI framework built on top of Dear ImGui. It provides a clean, Go-idiomatic API for building desktop applications with a focus on simplicity and ease of use.

## Overview

dfx is a complete rewrite of the original imapp framework (a personal project, never released), designed to provide the same core functionality with a much simpler and more intuitive API. Key improvements include:

- **50% less code** - Eliminated redundant abstractions
- **Simpler mental model** - Everything is a Component
- **Better composition** - Components can have children
- **Type safety** - Structured events instead of raw IO polling
- **Conflict detection** - Actions prevent key binding conflicts
- **Built-in theming** - Comprehensive font and theme system

The [dfx Roadmap](https://github.com/users/michaelquigley/projects/2/views/1) is always up-to-date with the current and planned work on the project.

## Core Concepts

### Component Interface

The fundamental abstraction in dfx is the `Component`:

```go
type Component interface {
    Draw(state *State)
    Actions() *ActionRegistry
}
```

Components receive a `State` containing all drawing context and can define keyboard actions.

### State

The `State` struct consolidates all drawing parameters:

```go
type State struct {
    Size     imgui.Vec2  // Available drawing area
    Position imgui.Vec2  // Position within parent
    IO       *imgui.IO  // ImGui input/output
    App      *App       // Application reference
    Parent   Component  // Parent component (nil for root)
}
```

### Component Types

#### Func - Simple Function Components
The simplest way to create a component:

```go
root := dfx.NewFunc(func(state *dfx.State) {
    imgui.Text("Hello World!")
    if imgui.Button("Click Me") {
        fmt.Println("Button clicked!")
    }
})
```

#### Container - Composable Components
For more complex components with state and children:

```go
type MyComponent struct {
    dfx.Container
    counter int
}

func NewMyComponent() *MyComponent {
    c := &MyComponent{}
    c.Visible = true
    c.OnDraw = func(state *dfx.State) {
        imgui.Text(fmt.Sprintf("Counter: %d", c.counter))
        if imgui.Button("Increment") {
            c.counter++
        }
    }
    return c
}
```

## Quick Start

### Basic Application

```go
package main

import "github.com/michaelquigley/dfx"

func main() {
    root := dfx.NewFunc(func(state *dfx.State) {
        imgui.Text("Hello from dfx!")
        if imgui.Button("Click Me") {
            // handle button click
        }
    })

    app := dfx.New(root, dfx.Config{
        Title:  "My App",
        Width:  800,
        Height: 600,
    })

    app.Run()
}
```

### With Menu Bar

```go
menuBar := dfx.NewFunc(func(state *dfx.State) {
    if imgui.BeginMenu("File") {
        if imgui.MenuItemBoolV("New", "Ctrl+N", false, true) {
            // handle new
        }
        if imgui.MenuItemBoolV("Open", "Ctrl+O", false, true) {
            // handle open
        }
        imgui.EndMenu()
    }
})

app := dfx.New(root, dfx.Config{
    Title:   "My App",
    MenuBar: menuBar,
})
```

## Theming System

dfx includes a comprehensive theming system with both predefined and customizable themes.

### Predefined Themes

```go
app := dfx.New(root, dfx.Config{
    Title: "Themed App",
    Theme: dfx.BlueTheme,    // or GreenTheme, RedTheme, PurpleTheme, ModernDark
})
```

### Custom HSV Themes

```go
customTheme := dfx.NewHueColorScheme("Custom", 180, 60, 200)
app := dfx.New(root, dfx.Config{
    Title: "Custom Themed App",
    Theme: customTheme,
})
```

### Runtime Theme Switching

```go
// Change theme during runtime
dfx.SetTheme(dfx.ModernDark)
```

## Font System

dfx provides three font constants with Material Icons merged where applicable:
- **MainFont** (20px) - Gidole Regular with Material Icons
- **MonospaceFont** (16px) - JetBrains Mono for code
- **SmallFont** (16px) - Gidole Regular small with Material Icons, for labels/indicators

### Using Different Fonts

```go
// Default font (with icons)
imgui.Text("Regular text " + string(fonts.ICON_FAVORITE))

// Monospace font
dfx.PushFont(dfx.MonospaceFont)
imgui.Text("Monospace code text")
dfx.PopFont()

// Small font for labels and indicators
dfx.PushFont(dfx.SmallFont)
imgui.Text("CH1")
dfx.PopFont()
```

### Disabling Font/Theme System

```go
app := dfx.New(root, dfx.Config{
    Title:          "Minimal App",
    DisableFonts:   true,  // Use default ImGui fonts
    DisableTheming: true,  // Use default ImGui theme
})
```

## Controls

For trivial ImGui operations (Button, Text, Separator, SameLine, Spacing, TreeNode, TreePop, BeginChild, EndChild, BeginMenu, EndMenu, BeginMenuBar, EndMenuBar, MenuItem), call `imgui.*` directly. dfx provides value-add wrappers for controls that benefit from a cleaner Go-idiomatic API, returning `(newValue, changed)` tuples instead of requiring pointers:

```go
// Text input
text, changed := dfx.Input("Label", currentText)
if changed {
    // handle text change
}

// Slider
value, changed := dfx.Slider("Volume", currentValue, 0.0, 1.0)

// Checkbox
checked, changed := dfx.Checkbox("Enable feature", isEnabled)

// Button (use imgui directly)
if imgui.Button("Submit") {
    // handle button click
}

// Combo/Dropdown
items := []string{"Option 1", "Option 2", "Option 3"}
selected, changed := dfx.Combo("Choose", currentIndex, items)
```

**Value-add wrappers** (in `controls.go`): Input, InputMultiline, Checkbox, Slider, SliderInt, Combo, ColorEdit3, ColorEdit4, Toggle, WheelSlider.

### Enhanced Controls

dfx provides several enhanced controls with additional features beyond standard ImGui widgets:

**Toolbar** - Full-width header bar for section labels:
```go
// simple toolbar with label
dfx.Toolbar("Settings")

// toolbar with extra controls on the right
dfx.ToolbarEx("Actions", func() {
    if imgui.Button("Add") {
        // handle add
    }
    imgui.SameLine()
    if imgui.Button("Remove") {
        // handle remove
    }
})
```

Features:
- Draws full-width background rectangle using `ColHeader` color
- Automatically handles padding and cursor positioning
- `ToolbarEx` allows additional controls to the right of the label via callback

**Toggle** - Boolean toggle button with visual feedback:
```go
// inactive (false): dimmed appearance
// active (true): checkmark color
enabled, changed := dfx.Toggle("Play", playEnabled)
```

**WheelSlider** - Horizontal slider with mouse wheel support:
```go
// hover and scroll to adjust, Ctrl = 10x faster, Alt = 10x slower
value, changed := dfx.WheelSlider("Volume", volume, 0.0, 1.0, 100, "%.2f", imgui.SliderFlagsNone)
```

**Fader** - Advanced vertical fader designed for audio mixing applications with support for logarithmic tapers, range limits, and multiple value representations:

**FaderN** - Normalized fader (0.0 to 1.0):
```go
params := dfx.DefaultFaderParams()
params.Taper = dfx.AudioTaper()
params.Format = func(norm float32) string {
    return fmt.Sprintf("%.2f", norm)
}
value, changed := dfx.FaderN("##fader", normalizedValue, params)
```

**FaderF** - Float fader (arbitrary min/max range):
```go
// Example: -60.0 dB to +12.0 dB with audio taper
params := dfx.DefaultFaderParams()
params.Taper = dfx.AudioTaper()
params.Format = func(norm float32) string {
    db := norm*72.0 - 60.0
    if db <= -59.9 {
        return "-∞ dB"
    }
    return fmt.Sprintf("%.1f dB", db)
}
dbValue, changed := dfx.FaderF("##db", gainDB, -60.0, 12.0, params)
```

**FaderI** - Integer fader (arbitrary min/max range):
```go
// Example: 0 to 32767 for hardware control
params := dfx.DefaultFaderParams()
params.MinStop = 0.1  // limit to 10%-90% of range
params.MaxStop = 0.9
hwValue, changed := dfx.FaderI("##hw", hardwareValue, 0, 32767, params)
```

**FaderParams** provides extensive configuration:
- `Taper` - Response curve (Linear, Log, Audio, or Custom)
- `MinStop` / `MaxStop` - Range limits in normalized 0-1 space
- `ResetValue` - Right-click reset target (normalized 0-1 space)
- `Width` / `Height` - Fader dimensions
- `Format` - Custom tooltip formatting function
- `ShowTooltip` - Enable/disable value tooltip (default: true)
- `WheelSteps` - Mouse wheel sensitivity (default: 100.0)

**Built-in Tapers:**
- `LinearTaper()` - No taper, 1:1 mapping (default)
- `LogTaper(steepness)` - Logarithmic curve (steepness: 1.0 = gentle, 3.0 = moderate, 10.0 = steep)
- `AudioTaper()` - Standard audio fader curve (gentle bottom, steep top, optimized for dB scales)
- `DecibelTaper(dbRange)` - UI position linear with dB; for hardware values proportional to amplitude
- `CustomTaper(apply, invert)` - User-defined taper functions

**Multi-Representation Pattern:**
Advanced faders support maintaining multiple value representations (normalized, hardware, display) synchronized via conversion functions:

```go
type FaderState struct {
    normalized float32  // 0.0 - 1.0 (master value)
    hardware   int      // 0 - 32767
    decibels   float32  // -60.0 to +12.0
}

func updateFromNormalized(state *FaderState, norm float32) {
    state.normalized = norm
    state.hardware = int(norm * 32767)
    state.decibels = norm*72.0 - 60.0
}

// User chooses which API to use based on their needs
// FaderN for normalized, FaderI for hardware, FaderF for display values
```

**Faders with Scales:**
The `FaderWithScaleN/F/I` functions add tick marks and labels next to faders, perfect for audio applications that need visual reference marks:

```go
// Example: dB fader with scale
params := dfx.DefaultFaderParams()
params.Taper = dfx.AudioTaper()

scale := dfx.DefaultScaleConfig()
scale.Marks = []float32{0.0, 0.417, 0.667, 0.833, 1.0}
scale.Labels = map[float32]string{
    0.0:   "-60",
    0.417: "-30",
    0.667: "-12",
    0.833: "0",
    1.0:   "+12",
}

dbValue, changed := dfx.FaderWithScaleF("##gain", gainDB, -60.0, 12.0, params, scale)
```

**ScaleConfig** provides:
- `Marks` - Array of normalized positions (0-1) for tick marks
- `Labels` - Map of position → label text for specific marks
- `TickLength` - Tick mark length in pixels (default: 5.0)
- `LabelOffset` - Distance from ticks to labels (default: 3.0)
- `Position` - "left" or "right" side placement (default: "left")

**Key features:**
- **Taper-aware**: Tick marks automatically respect the fader's taper curve for visual accuracy
- **Theme integration**: Uses colors from the current theme
- **Flexible**: Add scales to any normalized, float, or integer range fader

See `examples/dfx_example_mixer` for a complete demonstration with horizontally scrollable mixer interface showcasing all fader types and scales.

**VUMeter** - Vertical level meter with multi-channel support and three display modes:

```go
// create a stereo meter
meter := dfx.NewVUMeter(2)
meter.SetLabels([]string{"L", "R"})

// set display mode (optional - VUMeterSolid is default)
meter.Mode = dfx.VUMeterSolid     // continuous fill
meter.Mode = dfx.VUMeterHighres   // 1px segments with 1px gaps
meter.Mode = dfx.VUMeterSegmented // configurable segments

// update levels each frame (0.0 to 1.0)
meter.SetLevels([]float32{leftLevel, rightLevel})

// draw the meter
meter.Draw(state)
```

**Configuration:**
- `Mode` - Display mode: `VUMeterSolid` (default), `VUMeterHighres`, `VUMeterSegmented`
- `Height` - Total height in pixels (default: 200)
- `ChannelWidth` - Width of each channel meter (default: 12)
- `SegmentCount` - Number of vertical segments, applies to VUMeterSegmented mode (default: 20)
- `SegmentGap` / `ChannelGap` - Spacing between segments and channels
- `PeakHoldMs` - Peak hold duration in ms, 0 = disabled (default: 1000)
- `PeakDecayRate` - Peak decay rate per second (default: 0.5)
- `ClipHoldMs` - Clip indicator hold time in ms (default: 2000)
- `Labels` - Custom labels per channel (e.g., "L", "R", "Kick")
- `ColorLow/Mid/High/Off/Peak/Clip` - Customizable segment colors

**Display Modes:**
- **VUMeterSolid**: Continuous fill with stacked color zones - clean, modern look
- **VUMeterHighres**: Fixed 1px segments with 1px gaps - high resolution digital look
- **VUMeterSegmented**: Configurable segments via `SegmentCount` and `SegmentGap`

**Features:**
- **Multi-channel**: Supports any number of channels displayed side-by-side
- **Color zones**: Green (0-60%), yellow (60-80%), red (80-100%)
- **Peak hold**: Displays peak level with configurable hold and decay
- **Clip indicator**: Top indicator lights red when signal clips, auto-resets
- **Custom labels**: Per-channel labels displayed below meters

See `examples/dfx_example_vumeter` for a complete demonstration.

**VUWaterfall** - Scrolling history display of VU levels over time:

```go
// create a stereo waterfall
waterfall := dfx.NewVUWaterfall(2)
waterfall.Height = 150
waterfall.HistorySize = 100 // samples to retain

// each frame, add current levels to history
waterfall.SetLevels([]float32{leftLevel, rightLevel})

// draw the waterfall
waterfall.Draw(state)
```

**Configuration:**
- `Height` - Total height in pixels (default: 200)
- `ChannelWidth` - Width per channel (default: 40)
- `ChannelGap` - Gap between channels (default: 4)
- `RowHeight` - Height of each history row (default: 2)
- `RowGap` - Gap between rows (default: 0)
- `HistorySize` - Number of samples to retain (default: 100)
- `ColorLow/Mid/High/Off` - Zone colors (same defaults as VUMeter)

**Features:**
- **Vertical scrolling**: New data appears at bottom, scrolls upward
- **Multi-channel**: Channels displayed side-by-side
- **Color zones**: Green (0-60%), yellow (60-80%), red (80-100%)
- **Centered bars**: Level represented by bar width, centered in channel

See `examples/dfx_example_vumeter` for a complete demonstration.

**LogViewer** - Buffered log display with configurable empty-state behavior:

```go
buffer := dfx.NewLogBuffer(1000)
viewer := dfx.NewLogViewer(buffer)

viewer.Visible = true
viewer.ShowDisabledMessage = true
viewer.DisabledMessage = "logging capture disabled"
```

**Visibility behavior:**
- `Visible == false` renders nothing
- `Visible == true` and `Buffer != nil` renders the log stream
- `Visible == true` and `Buffer == nil` renders `DisabledMessage` only when `ShowDisabledMessage == true`

Use `NewSlogHandler(...)` with a shared `LogBuffer` to route `slog` output into the viewer.

### FileNode Search/Filter

`FileNode` provides a `Find` method for searching trees, along with predicate constructors for common patterns:

```go
// find all .go files
goFiles := root.Find(dfx.MatchExt(".go"))

// find by name regex
pred, err := dfx.MatchName(`^main\.`)
if err != nil {
    // handle invalid regex
}
mains := root.Find(pred)

// find by full path regex
pred, err = dfx.MatchPath(`src/.*\.go$`)
if err != nil {
    // handle invalid regex
}
srcGoFiles := root.Find(pred)

// find all directories with an inline predicate
dirs := root.Find(func(n *dfx.FileNode) bool { return n.Dir })
```

**Methods:**
- **`Find(predicate func(*FileNode) bool) []*FileNode`** - Returns all matching nodes in depth-first pre-order. Returns nil on a nil receiver.

**Predicate Constructors:**
- **`MatchExt(ext string) func(*FileNode) bool`** - Matches non-directory nodes by file extension (case-insensitive). The ext parameter should include the dot (e.g. ".go").
- **`MatchName(pattern string) (func(*FileNode) bool, error)`** - Matches nodes whose Name matches the given regex.
- **`MatchPath(pattern string) (func(*FileNode) bool, error)`** - Matches nodes whose full Path() matches the given regex.

Predicates are composable with `FileTree.Filter`:
```go
fileTree.Filter = dfx.MatchExt(".go")
```

## Actions and Keyboard Shortcuts

dfx provides a hierarchical action system with conflict detection:

### Global Actions

Register application-wide keyboard shortcuts:

```go
app := dfx.New(root, dfx.Config{
    Title: "App with Shortcuts",
    OnSetup: func(app *dfx.App) {
        // Register global shortcuts
        app.Actions().Register("save", "Ctrl+S", func() {
            // handle save
        })

        app.Actions().Register("quit", "Ctrl+Q", func() {
            app.Stop()
        })
    },
})
```

### Component-Local Actions

Components can define their own keyboard shortcuts that automatically override global actions:

```go
myComponent := &dfx.Container{
    Visible: true,
    OnDraw: func(state *dfx.State) {
        imgui.Text("Component with local actions")
    },
}

// Add component-specific actions
myComponent.Actions().Register("increment", "Up", func() {
    // handle up arrow - only when this component has focus
})

myComponent.Actions().Register("decrement", "Down", func() {
    // handle down arrow
})
```

The action system provides:
- **Automatic conflict detection** within components
- **Hierarchical override behavior** - component actions override global actions
- **Simple key binding syntax** - "Ctrl+S", "Alt+F4", "Up", etc.
- **No boilerplate** - just define actions and they work

### Action Traversal

For custom composite components, implement:

- `ChildActions() []Component` to expose children for traversal
- `LocalActions() *ActionRegistry` to expose local actions

Action precedence is:
- child component actions first
- parent-local actions next
- app-global actions last

### Menu-Compatible Actions

For applications with menu bars, dfx provides menu-compatible actions that work both as keyboard shortcuts and menu items:

```go
// create menu actions
fileNew := dfx.NewMenuAction("New", "Ctrl+N", func() {
    // handle new file
})

fileSave := dfx.NewMenuAction("Save", "Ctrl+S", func() {
    // handle save
})

fileQuit := dfx.NewMenuAction("Quit", "Ctrl+Q", func() {
    app.Stop()
})

// create menu bar component
// NOTE: dfx.Config.MenuBar already wraps this in BeginMainMenuBar/EndMainMenuBar
menuBar := dfx.NewFunc(func(state *dfx.State) {
    if imgui.BeginMenu("File") {
        fileNew.DrawMenuItem()    // renders as menu item with shortcut label
        imgui.Separator()
        fileSave.DrawMenuItem()
        imgui.Separator()
        fileQuit.DrawMenuItem()
        imgui.EndMenu()
    }
})

// register for keyboard shortcuts
app.Actions().MustRegisterAction(fileNew)
app.Actions().MustRegisterAction(fileSave)
app.Actions().MustRegisterAction(fileQuit)

// use menu bar in config
app := dfx.New(root, dfx.Config{
    MenuBar: menuBar,
})
```

Menu actions provide:
- **Dual functionality** - work as both menu items and keyboard shortcuts
- **Automatic shortcut labels** - keyboard shortcuts display in menus
- **Single definition** - define once, use in both menu and keyboard
- **Consistent behavior** - clicking menu or pressing keys calls the same handler

See `examples/dfx_example_menu` for a complete demonstration.

## Layout and Composition

For a comprehensive guide to Dear ImGui's layout system including child windows, sizing semantics, and practical patterns, see [`docs/LAYOUT_GUIDE.md`](docs/LAYOUT_GUIDE.md). The interactive demo in `examples/dfx_example_layout` demonstrates all concepts with real-time values.

Components can contain children for complex layouts:

```go
container := &dfx.Container{
    Visible: true,
    Children: []dfx.Component{
        header,
        content,
        footer,
    },
    OnDraw: func(state *dfx.State) {
        // Custom layout logic for this container.
        // Children are drawn automatically by Container.Draw().
    },
}
```

### HCollapse - Horizontal Collapsible Panel

The `HCollapse` component provides a horizontal collapsible panel that contains content to its right. When collapsed, only the toggle button is visible. When expanded, it shows a header bar with title and the content below.

```go
// create a collapsible sidebar
sidebar := dfx.NewHCollapse(
    sidebarContent,
    dfx.HCollapseConfig{
        Title:         "Sidebar",
        ExpandedWidth: 250,
        TransitionMs:  100,
        Resizable:     true,
        Expanded:      true,
    },
)

// optional: add keyboard shortcut for toggle
sidebar.Actions().Register("toggle-sidebar", "[", func() {
    sidebar.Toggle()
})

// in Draw, use SameLine() to place content to the right
func (m *MyApp) Draw(state *dfx.State) {
    sidebar.Draw(state)
    imgui.SameLine()

    // main content fills remaining width
    remaining := state.Size.X - sidebar.CurrentWidth
    imgui.BeginChildStrV("main", imgui.Vec2{X: remaining, Y: state.Size.Y}, 0, 0)
    mainContent.Draw(state)
    imgui.EndChild()
}
```

**Configuration:**
- `Title` - displayed in header when expanded (also used for unique imgui ID)
- `ExpandedWidth` - width when fully expanded
- `MinWidth` - collapsed width (defaults to 36px, toggle button only)
- `MaxWidth` - maximum width when resizing (0 = no limit)
- `TransitionMs` - animation duration (default: 80ms)
- `Resizable` - allow drag-to-resize when expanded
- `Expanded` - initial state

**Features:**
- **Animated transitions** - smooth expand/collapse animation
- **Header bar** - toggle button on left, title when expanded, resize handle on right
- **Drag-to-resize** - adjust width by dragging the right edge
- **Collapsed tooltip** - hovering over collapsed toggle shows title
- **Toggle callback** - `OnToggle func(expanded bool)` for state change notifications
- **CurrentWidth** - read current width for layout calculations

**Note:** When using custom-drawn components (like VUMeter, Fader) inside tables within an HCollapse, use `imgui.TableFlagsNoClip` and `imgui.TableColumnFlagsNoClip` to prevent cell clipping.

### Workspace - View Switching

The `Workspace` component provides high-level management of multiple named views with easy switching. It separates stable identifiers from display names, allowing display names to include icons and formatting without affecting code that switches workspaces.

```go
// create workspaces
editor := dfx.NewFunc(func(state *dfx.State) {
    imgui.Text("Editor View")
    // editor UI...
})

viewer := dfx.NewFunc(func(state *dfx.State) {
    imgui.Text("Viewer")
    // viewer UI...
})

// create workspace manager with IDs and display names
ws := dfx.NewWorkspace()
ws.Add("editor", "📝 Editor", editor)  // ID, display name, component
ws.Add("viewer", "👁️ Viewer", viewer)
ws.ShowSelector = true      // shows combo selector
ws.SelectorLabel = "View"

// callback receives stable IDs
ws.OnSwitch = func(oldID, newID string) {
    fmt.Printf("switched from '%s' to '%s'\n", oldID, newID)
}

// add keyboard shortcuts using stable IDs
ws.Actions().MustRegister("Switch to Editor", "Ctrl+1", func() {
    ws.Switch("editor")  // won't break if display name changes
})
ws.Actions().MustRegister("Switch to Viewer", "Ctrl+2", func() {
    ws.Switch("viewer")
})

// change display name without affecting code
ws.SetName("editor", "✏️ Code Editor")

app := dfx.New(ws, dfx.Config{...})
```

**API Methods:**
- `NewWorkspace()` - create workspace manager
- `Add(id, name, component)` - add/replace workspace with ID and display name
- `Remove(id)` - remove workspace by ID
- `Switch(id)` - switch to workspace by ID
- `SwitchByIndex(index)` - switch by index
- `Current()` - get current workspace ID
- `CurrentName()` - get current display name
- `CurrentComponent()` - get current component
- `SetName(id, name)` - change display name (ID unchanged)
- `GetName(id)` - get display name for ID
- `WorkspaceIds()` - get list of workspace IDs
- `WorkspaceNames()` - get list of display names

**Configuration:**
- `ShowSelector` - show/hide combo selector (default: true)
- `SelectorLabel` - label for combo (default: "Workspace")
- `SelectorWidth` - width of selector (default: 200, -1 for auto)
- `OnSwitch` - callback when workspace changes (receives IDs)

**Benefits of ID/Name Separation:**
- Stable IDs for code, config files, keyboard shortcuts
- Display names can include icons, emoji, formatting
- Change display names without breaking code
- Easy localization (same ID, different display names)

See `examples/dfx_example_workspace` for a complete demonstration.

## Configuration Persistence

dfx provides optional utilities for configuration management in `config.go`. These helpers simplify common patterns like saving/loading JSON configuration, persisting window state, and managing dashboard layouts.

### Basic Configuration Pattern

```go
type Config struct {
    Window dfx.WindowConfig              `json:"window"`
    Dashes map[string]dfx.DashConfig     `json:"dashes"`
    // ... your app-specific settings
}

func main() {
    // determine config file path
    cfgPath, _ := dfx.ConfigPath("myapp", "config.json")

    // load with defaults
    cfg := defaultConfig()
    dfx.LoadJSON(cfgPath, cfg)

    // create app with saved window size and position
    app := dfx.New(root, dfx.Config{
        Title:  "My App",
        Width:  cfg.Window.Width,
        Height: cfg.Window.Height,
        X:      cfg.Window.X,
        Y:      cfg.Window.Y,

        OnClose: func(app *dfx.App) {
            cfg.Window = dfx.CaptureWindowState(app)
            dfx.SaveJSON(cfgPath, cfg)
        },

        OnSizeChange: func(width, height int) {
            cfg.Window.Width = width
            cfg.Window.Height = height
        },
    })

    app.Run()
}
```

### Dashboard State Persistence

```go
// capture dashboard state
cfg.Dashes = dfx.CaptureDashState(dashMgr)

// save to file
dfx.SaveJSON(cfgPath, cfg)

// later, restore dashboard state
dfx.RestoreDashState(dashMgr, cfg.Dashes)
```

### Configuration Helper Functions

- **`ConfigPath(appName, filename string) (string, error)`** - Returns standard config file path in user home directory (e.g., `~/.myapp/config.json`)
- **`SaveJSON(path string, config interface{}) error`** - Saves struct to JSON file with formatting
- **`LoadJSON(path string, config interface{}) error`** - Loads JSON file into struct (silent if file doesn't exist)
- **`CaptureDashState(dm *DashManager) map[string]DashConfig`** - Extracts dashboard visibility and sizes
- **`RestoreDashState(dm *DashManager, config map[string]DashConfig)`** - Applies configuration to dashboards
- **`CaptureWindowState(app *App) WindowConfig`** - Gets current window position, size, and state

**Note:** `WindowConfig` includes a `Maximized` field for future compatibility, but maximized state capture/restore is not yet implemented (requires backend enhancements).

### Example

See `examples/dfx_example_config` for a complete demonstration of configuration persistence including window state, dashboard layouts, and application settings.

## Container-Based Architecture with df/da

For larger applications, dfx integrates with the `df/da` dependency injection container and `df/dd` data-driven serialization packages. This pattern provides a more scalable architecture with factory-based component creation and automatic lifecycle management.

### Pattern Overview

The container-based pattern uses three packages:
- **da.Application[C]** - Application lifecycle with typed configuration
- **dfx** - GUI framework
- **dd** - Struct-to-YAML/JSON bidirectional binding

The lifecycle flow is: Configure → Build (factories) → Link → Start → (user interaction) → Stop → Save

### Basic Structure

```go
// config.go - typed configuration
type config struct {
    WindowX      int
    WindowY      int
    WindowWidth  int
    WindowHeight int
    Counter      int
}

func defaultConfig() config {
    return config{
        WindowX:      100,
        WindowY:      100,
        WindowWidth:  800,
        WindowHeight: 600,
    }
}
```

```go
// shellFactory.go - factory creates the main window
type shellFactory struct{}

func (f *shellFactory) Build(a *da.Application[config]) error {
    shl := &shell{
        cfg:       &a.Cfg,  // reference to mutable config
        workspace: dfx.NewWorkspace(),
    }

    shl.app = dfx.New(shl.root, dfx.Config{
        Title:  "my app",
        Width:  shl.cfg.WindowWidth,
        Height: shl.cfg.WindowHeight,
        X:      shl.cfg.WindowX,
        Y:      shl.cfg.WindowY,
        OnSizeChange: func(w, h int) {
            shl.cfg.WindowWidth = w
            shl.cfg.WindowHeight = h
        },
        OnClose: func(app *dfx.App) {
            shl.cfg.WindowX, shl.cfg.WindowY = app.GetWindowPos()
        },
    })

    da.Set(a.C, shl)  // register in container
    return nil
}

// Link wires up tagged workspace components
func (s *shell) Link(c *da.Container) error {
    for i, ws := range da.TaggedAsType[dfx.Component](c, "workspaces") {
        s.workspace.Add(fmt.Sprintf("ws-%d", i), fmt.Sprintf("workspace %d", i+1), ws)
    }
    return nil
}

func (s *shell) Start() error {
    go s.app.Run()
    return nil
}
```

```go
// panelFactory.go - factory registers tagged workspace component
type panelFactory struct{}

func (f *panelFactory) Build(a *da.Application[config]) error {
    panel := &myPanel{cfg: &a.Cfg}
    da.AddTagged(a.C, "workspaces", panel)  // tagged registration
    return nil
}
```

```go
// main.go - application lifecycle
func main() {
    cfgPath, _ := configPath()

    app := da.NewApplication[config](defaultConfig())
    app.Factories = append(app.Factories, &panelFactory{})
    app.Factories = append(app.Factories, &shellFactory{})

    app.InitializeWithPaths(da.OptionalPath(cfgPath))
    app.Start()

    // wait for GUI to close
    if shl, ok := da.Get[*shell](app.C); ok {
        shl.app.Wait()
    }

    // save config
    dd.UnbindYAMLFile(app.Cfg, cfgPath)
    app.Stop()
}
```

### Key Benefits

- **Factory pattern** - Components created via `Build()` methods with access to typed config
- **Tagged components** - `da.AddTagged()` for modular registration without naming conflicts
- **Link phase** - Wire up dependencies after all factories complete
- **Config mutation** - Components hold `*config` reference for real-time state updates
- **Automatic persistence** - `dd.UnbindYAMLFile()` saves config struct to YAML on shutdown

### Example

See `examples/dfx_example_container` for a complete demonstration of container-based architecture with tagged workspace components.

## Examples

See the `examples/` directory for complete working examples:

- `dfx_example_simple` - Basic usage
- `dfx_example_actions` - Keyboard shortcuts
- `dfx_example_custom_component` - Custom component creation
- `dfx_example_composition` - Complex UI with menu bars
- `dfx_example_themes` - Theming and font demonstration
- `dfx_example_filetree` - Filesystem tree viewer
- `dfx_example_logviewer` - Log viewer with df/dl integration
- `dfx_example_controls` - Control wrappers (Combo, Toggle, WheelSlider)
- `dfx_example_mixer` - Advanced fader demonstration with tapers, range limits, and horizontal scrolling mixer
- `dfx_example_vumeter` - VU meter and waterfall with display modes and scrolling history
- `dfx_example_hcollapse` - Horizontal collapsible panels with faders and meters
- `dfx_example_simple_hcollapse` - Minimal HCollapse example
- `dfx_example_workspace` - Workspace switching with multiple views
- `dfx_example_lifecycle` - Window lifecycle callbacks
- `dfx_example_config` - Configuration persistence with window and dashboard state
- `dfx_example_container` - Container-based lifecycle with df/da dependency injection
- `dfx_example_layout` - Comprehensive ImGui layout and sizing tutorial (see [`docs/LAYOUT_GUIDE.md`](docs/LAYOUT_GUIDE.md))
- `dfx_example_multigrid` - MultiGrid layout system
- `dfx_example_dash` - DashManager panel system
- `dfx_example_undo` - Undo/redo system demo
- `dfx_example_menu` - Menu-compatible actions

## Building Examples

```bash
# Build all examples
go build ./examples/dfx_example_simple
go build ./examples/dfx_example_actions
go build ./examples/dfx_example_themes

# Run an example
./dfx_example_themes
```

## Migration from imapp v1

dfx is designed as a replacement for imapp v1:

1. Replace `imapp.Surface` usage with `dfx.Component`
2. Convert `Surface.DrawF` functions to `dfx.Func` components
3. Replace action registration with new conflict-detecting system
4. Update control usage to new return-value API

The migration should be straightforward due to conceptual similarity, but the new API is much cleaner and more Go-idiomatic.

## Architecture Notes

### Full-Window Rendering
Components render within an invisible, borderless ImGui window that fills the entire backend window. This matches imapp v1's behavior exactly and provides a transparent "canvas" for drawing.

### No Layout System
dfx deliberately does not include a layout system, allowing components to handle their own positioning. This keeps the framework simple while enabling maximum flexibility.

### Single Backend
Currently supports only the GLFW backend, matching imapp v1's approach.

## License

Part of the baab project.
