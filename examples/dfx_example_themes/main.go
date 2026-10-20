package main

// the style comparison: one representative working surface, launched once per
// complete style so typography, spacing, and colors are judged together.
//
//	go run ./examples/dfx_example_themes                          # Classic
//	go run ./examples/dfx_example_themes -style faceplate         # Faceplate, default yellow
//	go run ./examples/dfx_example_themes -style faceplate -accent 4da3ff
//
// complete styles (fonts, spacing, and colors) are chosen at startup. the Colors
// menu changes colors only, in the running style; it starts on the colors actually
// applied. views are chosen
// with the workspace selector: Controls, Data, Panels, Layouts, and Canvas. all
// sample data is synthetic and stable: no filesystem, log stream, or audio.

import (
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
	"github.com/michaelquigley/dfx/fonts"
)

func main() {
	styleName := flag.String("style", "classic", "complete style: classic or faceplate")
	accentHex := flag.String("accent", "", "faceplate accent as RRGGBB (faceplate only)")
	flag.Parse()

	// validate everything before a window opens
	style, err := selectStyle(*styleName, *accentHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dfx_example_themes: %v\n", err)
		flag.Usage()
		os.Exit(2)
	}

	g := newGallery()
	app := dfx.New(g.root(), dfx.Config{
		Title:   "dfx style comparison — " + style.Name,
		Width:   1280,
		Height:  860,
		Style:   &style,
		MenuBar: g.menu(),
		OnSetup: g.setup,
	})
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func selectStyle(name, accent string) (dfx.Style, error) {
	var style dfx.Style
	switch strings.ToLower(name) {
	case "classic":
		style = dfx.ClassicStyle()
	case "faceplate":
		style = dfx.FaceplateStyle()
	default:
		return style, fmt.Errorf("unknown style '%s'", name)
	}
	if accent != "" {
		if style.Name != "Faceplate" {
			return style, fmt.Errorf("-accent applies to the faceplate style only")
		}
		color, err := parseHex(accent)
		if err != nil {
			return style, err
		}
		style.Theme = &dfx.FaceplateTheme{Accent: color}
	}
	return style, nil
}

func parseHex(s string) (imgui.Vec4, error) {
	s = strings.TrimPrefix(s, "#")
	v, err := strconv.ParseUint(s, 16, 32)
	if len(s) != 6 || err != nil {
		return imgui.Vec4{}, fmt.Errorf("accent '%s' is not RRGGBB", s)
	}
	return imgui.Vec4{X: float32(v>>16&0xff) / 255, Y: float32(v>>8&0xff) / 255, Z: float32(v&0xff) / 255, W: 1}, nil
}

type gallery struct {
	themes []dfx.Theme
	theme  int

	// controls
	toggleA, toggleB, check bool
	radio                   int
	slider, wheel           float32
	sliderInt               int
	combo                   int
	text, notes             string
	color                   [3]float32

	// data
	tree      *dfx.FileTree
	logs      *dfx.LogViewer
	meter     *dfx.VUMeter
	waterfall *dfx.VUWaterfall
	faders    []float32
	frame     int

	// panels and layouts
	workspace *dfx.Workspace
	dashes    *dfx.DashManager
	layouts   *dfx.MultiGrid

	// canvas: an implicit style re-derived on theme change, and an explicit one
	// the theme selector never touches
	implicit, explicit *dfx.NodeCanvas[string]
}

func newGallery() *gallery {
	g := &gallery{
		toggleB: true, check: true, slider: 0.42, wheel: 0.6, sliderInt: 7,
		text: "editable text", notes: "multiline\nnotes", color: [3]float32{0.3, 0.6, 0.9},
		faders: []float32{-6, 0, -18},
	}
	g.tree = dfx.NewFileTree(sampleTree())
	buffer := dfx.NewLogBuffer(64)
	start := time.Now()
	for i, entry := range []struct {
		level slog.Level
		msg   string
	}{
		{slog.LevelDebug, "scanning patch library"},
		{slog.LevelInfo, "loaded 24 presets"},
		{slog.LevelInfo, "audio engine at 48 kHz"},
		{slog.LevelWarn, "buffer underrun on output 2"},
		{slog.LevelError, "midi device disconnected"},
		{slog.LevelInfo, "session saved"},
	} {
		buffer.Add(dfx.LogMessage{Time: start.Add(time.Duration(i) * 250 * time.Millisecond), Level: entry.level, Func: "gallery.sample", Fields: fmt.Sprintf("n=%d", i), Message: entry.msg})
	}
	g.logs = dfx.NewLogViewer(buffer)
	g.logs.LevelFilter = slog.LevelDebug
	g.logs.AutoScroll = false
	g.meter = dfx.NewVUMeter(2)
	g.meter.SetLabels([]string{"L", "R"})
	g.meter.Height = 180
	g.waterfall = dfx.NewVUWaterfall(2)
	g.waterfall.Height = 180

	g.implicit = dfx.NewNodeCanvas[string](dfx.NodeCanvasConfig{
		MiniMap: dfx.NodeCanvasMiniMapConfig{Visible: true, Corner: dfx.NodeCanvasLowerLeft, Size: imgui.Vec2{X: 160, Y: 100}},
	})
	return g
}

// setup runs after the style's fonts, metrics, and colors: the place to adopt the
// style's spacing through older public fields, and to derive context-dependent
// styles.
func (g *gallery) setup(*dfx.App) {
	current := dfx.CurrentStyle()
	g.themes = []dfx.Theme{current.Theme, dfx.ModernDark, dfx.BlueTheme, dfx.GreenTheme, dfx.RedTheme, dfx.PurpleTheme,
		&dfx.FaceplateTheme{Accent: imgui.Vec4{X: 1, Y: 0xd8 / 255.0, W: 1}}, &dfx.FaceplateTheme{Accent: imgui.Vec4{X: 0x4d / 255.0, Y: 0xa3 / 255.0, Z: 1, W: 1}}}
	g.dashes.Margin = current.Metrics.ItemSpacing.X // the historical public margin, set deliberately

	// an explicit canvas style: application colors survive theme changes
	explicit := dfx.DefaultNodeCanvasStyle()
	explicit.NodeBodyColor = imgui.Vec4{X: 0.12, Y: 0.16, Z: 0.14, W: 1}
	explicit.TitleBandColor = imgui.Vec4{X: 0.2, Y: 0.42, Z: 0.3, W: 1}
	explicit.LinkColor = imgui.Vec4{X: 0.45, Y: 0.85, Z: 0.6, W: 0.8}
	explicit.GridColor = imgui.Vec4{X: 0.3, Y: 0.5, Z: 0.4, W: 0.25}
	g.explicit = dfx.NewNodeCanvas[string](dfx.NodeCanvasConfig{Style: explicit})
}

func (g *gallery) selectTheme(i int) {
	g.theme = i
	dfx.SetTheme(g.themes[i])
	g.implicit.SetStyle(dfx.DefaultNodeCanvasStyle()) // only the implicit canvas follows
}

func (g *gallery) menu() dfx.Component {
	return dfx.NewFunc(func(*dfx.State) {
		if imgui.BeginMenu("File") {
			imgui.MenuItemBoolV("New", "Ctrl+N", false, true)
			imgui.MenuItemBoolV("Open…", "Ctrl+O", false, true)
			imgui.MenuItemBoolV("Recent", "", false, false)
			imgui.Separator()
			imgui.MenuItemBoolV("Quit", "Ctrl+Q", false, true)
			imgui.EndMenu()
		}
		if imgui.BeginMenu("Colors") {
			for i, theme := range g.themes {
				label := theme.Name()
				if f, ok := theme.(*dfx.FaceplateTheme); ok {
					label = fmt.Sprintf("%s #%02x%02x%02x", label, int(f.Accent.X*255+0.5), int(f.Accent.Y*255+0.5), int(f.Accent.Z*255+0.5))
				}
				if i == 0 {
					label += " (startup)"
				}
				if imgui.MenuItemBoolV(fmt.Sprintf("%s##theme%d", label, i), "", g.theme == i, true) {
					g.selectTheme(i)
				}
			}
			imgui.Separator()
			imgui.TextDisabled("fonts and spacing: relaunch with -style")
			imgui.EndMenu()
		}
		imgui.TextDisabled(fmt.Sprintf("style %s · colors %s", dfx.CurrentStyle().Name, g.themes[g.theme].Name()))
	})
}

func (g *gallery) root() dfx.Component {
	g.workspace = dfx.NewWorkspace()
	g.workspace.SelectorLabel = "View"
	g.workspace.Add("controls", fonts.ICON_TUNE+" Controls", dfx.NewFunc(g.drawControls))
	g.workspace.Add("data", fonts.ICON_LIST+" Data", g.dataView())
	g.workspace.Add("panels", fonts.ICON_VIEW_QUILT+" Panels", g.panelsView())
	g.workspace.Add("layouts", fonts.ICON_GRID_ON+" Layouts", g.layoutsView())
	g.workspace.Add("canvas", fonts.ICON_SHARE+" Canvas", dfx.NewFunc(g.drawCanvas))
	return dfx.NewFunc(func(state *dfx.State) {
		g.frame++
		g.animate()
		g.workspace.Draw(state)
	})
}

// section starts a group with a heading, separated deliberately from the last.
func section(title string, first bool) {
	if !first {
		imgui.Dummy(imgui.Vec2{Y: dfx.CurrentStyle().Spacing.SectionGap})
	}
	dfx.PushHeadingFont()
	imgui.Text(title)
	dfx.PopFont()
}

func (g *gallery) drawControls(state *dfx.State) {
	column := (state.Size.X - imgui.CurrentStyle().ItemSpacing().X) / 2
	if column <= 0 || state.Size.Y <= 0 {
		return
	}
	imgui.BeginChildStrV("##controlsLeft", imgui.Vec2{X: column, Y: state.Size.Y}, imgui.ChildFlagsAlwaysUseWindowPadding, 0)
	section(fonts.ICON_TEXT_FIELDS+" Typography", true)
	imgui.Text(fonts.ICON_PLAY_ARROW + " Regular text beside an icon, " + fonts.ICON_SYMBOL_VIEW_REAL_SIZE + " and a supplemental symbol")
	dfx.PushHeadingFont()
	imgui.Text(fonts.ICON_STAR + " Bold heading text beside an icon")
	dfx.PopFont()
	imgui.TextDisabled("Secondary text for annotations and readouts")
	dfx.PushFont(dfx.SmallFont)
	imgui.Text(fonts.ICON_INFO + " Small text, 16 px")
	dfx.PopFont()
	dfx.PushFont(dfx.MonospaceFont)
	imgui.Text("monospace: 0123456789 {} => 0x1f")
	dfx.PopFont()

	section(fonts.ICON_TOUCH_APP+" Buttons and toggles", false)
	imgui.Button("Button")
	imgui.SameLine()
	imgui.SmallButton("Small")
	imgui.SameLine()
	imgui.BeginDisabled()
	imgui.Button("Disabled")
	imgui.EndDisabled()
	g.toggleA, _ = dfx.Toggle("Toggle off##a", g.toggleA)
	imgui.SameLine()
	g.toggleB, _ = dfx.Toggle(fonts.ICON_POWER_SETTINGS_NEW+" Toggle on##b", g.toggleB)
	imgui.SameLine()
	ink := imgui.Vec4{X: 0.55, Y: 0.05, Z: 0.1, W: 1}
	g.toggleB, _ = dfx.ToggleV("Explicit ink##c", g.toggleB, dfx.ToggleOptions{OnText: &ink})
	imgui.SameLine()
	imgui.BeginDisabled()
	dfx.Toggle("Disabled on##d", true)
	imgui.EndDisabled()
	g.check, _ = dfx.Checkbox("Checkbox", g.check)
	imgui.SameLine()
	for i, label := range []string{"One", "Two", "Three"} {
		if imgui.RadioButtonBool(label, g.radio == i) {
			g.radio = i
		}
		imgui.SameLine()
	}
	imgui.NewLine()
	if imgui.Button(fonts.ICON_MORE_VERT + " Popup") {
		imgui.OpenPopupStr("gallery popup")
	}
	imgui.SetItemTooltip("Opens a popup menu")
	if imgui.BeginPopup("gallery popup") {
		imgui.SeparatorText("Actions")
		imgui.MenuItemBoolV("Duplicate", "Ctrl+D", false, true)
		imgui.MenuItemBoolV("Selected item", "", true, true)
		imgui.MenuItemBoolV("Unavailable", "", false, false)
		imgui.EndPopup()
	}

	section(fonts.ICON_EDIT+" Inputs and sliders", false)
	imgui.PushItemWidth(column * 0.6)
	g.text, _ = dfx.Input("Input", g.text)
	g.notes, _ = dfx.InputMultiline("Notes", g.notes, column*0.6, 3*imgui.TextLineHeightWithSpacing())
	g.slider, _ = dfx.Slider("Slider", g.slider, 0, 1)
	g.sliderInt, _ = dfx.SliderInt("Int slider", g.sliderInt, 0, 16)
	g.wheel, _ = dfx.WheelSlider("Wheel slider", g.wheel, 0, 1, 100, "%.2f", imgui.SliderFlagsNone)
	g.combo, _ = dfx.Combo("Combo", g.combo, []string{"Sine", "Square", "Saw", "Noise"})
	g.color[0], g.color[1], g.color[2], _ = dfx.ColorEdit3("Color", g.color[0], g.color[1], g.color[2])
	imgui.ProgressBarV(g.slider, imgui.Vec2{X: column * 0.6}, "")
	imgui.PopItemWidth()
	imgui.EndChild()

	imgui.SameLine()
	imgui.BeginChildStrV("##controlsRight", imgui.Vec2{X: column, Y: state.Size.Y}, imgui.ChildFlagsAlwaysUseWindowPadding, 0)
	section(fonts.ICON_TAB+" Tabs", true)
	if imgui.BeginTabBar("gallery tabs") {
		for _, tab := range []string{"Mixer", "Routing", "Settings"} {
			if imgui.BeginTabItem(tab) {
				imgui.Text("The " + strings.ToLower(tab) + " tab's content.")
				imgui.EndTabItem()
			}
		}
		imgui.EndTabBar()
	}

	section(fonts.ICON_TABLE_CHART+" Table", false)
	flags := imgui.TableFlagsBorders | imgui.TableFlagsRowBg
	if imgui.BeginTableV("ordinary", 3, flags, imgui.Vec2{}, 0) {
		imgui.TableSetupColumn("Channel")
		imgui.TableSetupColumn("Level")
		imgui.TableSetupColumn("State")
		imgui.TableHeadersRow()
		for i, row := range [][3]string{{"Kick", "-6.0 dB", "armed"}, {"Snare", "-9.5 dB", "muted"}, {"Bass", "-3.2 dB", "solo"}, {"Pad", "-14.0 dB", ""}} {
			imgui.TableNextRow()
			for _, cell := range row {
				imgui.TableNextColumn()
				if i == 1 && cell == row[0] {
					imgui.SelectableBoolV(cell, true, imgui.SelectableFlagsSpanAllColumns, imgui.Vec2{})
					continue
				}
				imgui.Text(cell)
			}
		}
		imgui.EndTable()
	}

	// a deliberate local override: dense spacing and a quieter row color for one
	// table, while its panel keeps the style's rhythm
	section(fonts.ICON_DENSITY_SMALL+" Dense table (local override)", false)
	imgui.PushStyleVarVec2(imgui.StyleVarCellPadding, imgui.Vec2{X: 4, Y: 1})
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 4, Y: 1})
	imgui.PushStyleColorVec4(imgui.ColTableRowBgAlt, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.05})
	dfx.PushFont(dfx.SmallFont)
	if imgui.BeginTableV("dense", 5, flags, imgui.Vec2{}, 0) {
		for _, h := range []string{"#", "Note", "Vel", "Len", "Ch"} {
			imgui.TableSetupColumn(h)
		}
		imgui.TableHeadersRow()
		for i := 0; i < 12; i++ {
			imgui.TableNextRow()
			for _, cell := range []string{strconv.Itoa(i + 1), []string{"C3", "E3", "G3", "B3"}[i%4], strconv.Itoa(64 + i*5), "1/8", strconv.Itoa(i%2 + 1)} {
				imgui.TableNextColumn()
				imgui.Text(cell)
			}
		}
		imgui.EndTable()
	}
	dfx.PopFont()
	imgui.PopStyleColor()
	imgui.PopStyleVarV(2)
	imgui.EndChild()
}

func sampleTree() *dfx.FileNode {
	root := &dfx.FileNode{Name: "project", Dir: true}
	add := func(parent *dfx.FileNode, name string, dir bool) *dfx.FileNode {
		n := &dfx.FileNode{Name: name, Dir: dir, Parent: parent}
		parent.Children = append(parent.Children, n)
		return n
	}
	patches := add(root, "patches", true)
	for _, name := range []string{"warm pad.yaml", "bass.yaml", "pluck.yaml"} {
		add(patches, name, false)
	}
	samples := add(root, "samples", true)
	drums := add(samples, "drums", true)
	add(drums, "kick.wav", false)
	add(drums, "snare.wav", false)
	add(samples, "vox.wav", false)
	add(root, "README.md", false)
	return root
}

// animate drives the meters from a deterministic synthetic signal.
func (g *gallery) animate() {
	t := float64(g.frame) / 60
	left := float32(0.55 + 0.35*math.Sin(t*2.1))
	right := float32(0.5 + 0.45*math.Sin(t*1.3+1))
	g.meter.SetLevels([]float32{left, right})
	g.waterfall.SetLevels([]float32{left, right})
}

func (g *gallery) dataView() dfx.Component {
	mg := dfx.NewMultiGrid()
	mg.AddComponent("tree", g.tree)
	mg.AddComponent("logs", g.logs)
	mg.AddComponent("signal", dfx.NewFunc(g.drawSignal))
	layout := dfx.NewFlexLayout([][]string{{"tree", "logs"}, {"signal"}})
	mg.SetLayout(layout)
	return mg
}

func (g *gallery) drawSignal(state *dfx.State) {
	g.meter.Draw(state)
	imgui.SameLine()
	g.waterfall.Draw(state)
	params := dfx.DefaultFaderParams()
	params.Height = min(220, max(60, state.Size.Y-imgui.FrameHeightWithSpacing()))
	for i := range g.faders {
		imgui.SameLine()
		imgui.BeginGroup()
		if i == 2 {
			accent := imgui.Vec4{X: 0.35, Y: 0.75, Z: 0.95, W: 1} // explicit colors survive theme changes
			params.AccentColor = &accent
		}
		g.faders[i], _ = dfx.FaderWithScaleF(fmt.Sprintf("##fader%d", i), g.faders[i], -40, 10, params, dfx.DefaultScaleConfig())
		imgui.Text(fmt.Sprintf("%.1f dB", g.faders[i]))
		imgui.EndGroup()
	}
}

func (g *gallery) panelsView() dfx.Component {
	left := dfx.NewHCollapse(dfx.NewFunc(func(*dfx.State) {
		imgui.Text("Left-anchored panel")
		imgui.TextDisabled("drag the handle to resize")
		for i := 0; i < 4; i++ {
			imgui.SelectableBool(fmt.Sprintf("Item %d", i+1))
		}
	}), dfx.HCollapseConfig{Title: "Browser", ExpandedWidth: 240, Resizable: true, Expanded: true})
	right := dfx.NewHCollapse(dfx.NewFunc(func(*dfx.State) {
		imgui.Text("Right-anchored inspector")
		imgui.TextDisabled("collapse with the chevron")
	}), dfx.HCollapseConfig{Title: "Inspector", ExpandedWidth: 220, Resizable: true, Expanded: true, Anchor: dfx.AnchorRight})
	collapsed := dfx.NewHCollapse(dfx.NewFunc(func(*dfx.State) { imgui.Text("hidden") }), dfx.HCollapseConfig{Title: "Collapsed panel", ExpandedWidth: 200})

	center := dfx.NewFunc(func(state *dfx.State) {
		dfx.ToolbarExLayout(fonts.ICON_BUILD+" Toolbar", func(l *dfx.ToolbarLayout) {
			l.CenterFrame()
			imgui.SmallButton(fonts.ICON_PLAY_ARROW)
			imgui.SameLine()
			l.CenterFrame()
			imgui.SetNextItemWidth(120)
			g.combo, _ = dfx.Combo("##toolbarCombo", g.combo, []string{"Sine", "Square", "Saw", "Noise"})
			imgui.SameLine()
			l.CenterText()
			imgui.TextDisabled("centered text")
		})
		avail := imgui.ContentRegionAvail()
		row := &dfx.State{Size: imgui.Vec2{X: avail.X, Y: max(0, avail.Y)}, IO: state.IO, App: state.App}
		collapsed.Draw(row)
		imgui.SameLine()
		left.Draw(row)
		imgui.SameLine()
		middle := avail.X - collapsed.CurrentWidth - left.CurrentWidth - right.CurrentWidth - 3*imgui.CurrentStyle().ItemSpacing().X
		if middle > 0 && avail.Y > 0 {
			imgui.BeginChildStrV("##panelsMiddle", imgui.Vec2{X: middle, Y: avail.Y}, imgui.ChildFlagsBorders, 0)
			imgui.TextWrapped("Content between two collapsible panels. The left panel is anchored left, the inspector right; both resize from their free edge.")
			imgui.EndChild()
			imgui.SameLine()
		}
		right.Draw(row)
	})

	g.dashes = dfx.NewDashManager()
	g.dashes.Left = dfx.NewDash("Sessions", dfx.NewFunc(func(*dfx.State) {
		imgui.Text(fonts.ICON_FOLDER + " Sessions")
		for _, s := range []string{"tuesday jam", "mix v3", "sketches"} {
			imgui.SelectableBool(s)
		}
	}))
	g.dashes.Left.TargetSize, g.dashes.Left.CurrentSize = 200, 200
	g.dashes.Bottom = dfx.NewDash("Console", g.logs)
	g.dashes.Bottom.TargetSize, g.dashes.Bottom.CurrentSize = 150, 150
	g.dashes.Inner = center
	return g.dashes
}

func (g *gallery) layoutsView() dfx.Component {
	g.layouts = dfx.NewMultiGrid()
	layout := dfx.NewGridLayout(3, 2)
	cells := []struct {
		id, label            string
		row, col, rows, cols int
	}{
		{"wide", "spans two columns", 0, 0, 1, 2},
		{"tall", "spans two rows", 0, 2, 2, 1},
		{"one", "single cell", 1, 0, 1, 1},
		{"two", "single cell", 1, 1, 1, 1},
	}
	for _, c := range cells {
		label := c.label
		g.layouts.AddComponent(c.id, dfx.NewFunc(func(state *dfx.State) {
			dfx.PushHeadingFont()
			imgui.Text(label)
			dfx.PopFont()
			imgui.TextDisabled(fmt.Sprintf("usable %.0f × %.0f", state.Size.X, state.Size.Y))
		}))
		layout.SetCell(c.id, c.row, c.col, c.rows, c.cols)
	}
	g.layouts.SetLayout(layout)
	return g.layouts
}

type canvasNode struct {
	pos   imgui.Vec2
	title string
	ins   []string
	outs  []string
}

var canvasNodes = map[string]canvasNode{
	"osc":    {imgui.Vec2{X: 40, Y: 60}, fonts.ICON_MUSIC_NOTE + " oscillator", nil, []string{"out"}},
	"filter": {imgui.Vec2{X: 300, Y: 30}, fonts.ICON_TUNE + " filter", []string{"in"}, []string{"out"}},
	"amp":    {imgui.Vec2{X: 540, Y: 110}, fonts.ICON_GRAPHIC_EQ + " amp", []string{"in", "cv"}, []string{"out"}},
	"lfo":    {imgui.Vec2{X: 300, Y: 230}, fonts.ICON_WAVES + " lfo", nil, []string{"cv"}},
}

var canvasOrder = []string{"osc", "filter", "amp", "lfo"}

var canvasLinks = [][2]string{{"osc.out", "filter.in"}, {"filter.out", "amp.in"}, {"lfo.cv", "amp.cv"}}

func (g *gallery) drawCanvas(state *dfx.State) {
	spacing := imgui.CurrentStyle().ItemSpacing().X
	half := imgui.Vec2{X: (state.Size.X - spacing) / 2, Y: state.Size.Y - imgui.TextLineHeightWithSpacing()}
	if half.X <= 0 || half.Y <= 0 {
		return
	}
	for i, nc := range []*dfx.NodeCanvas[string]{g.implicit, g.explicit} {
		if i > 0 {
			imgui.SameLine()
		}
		imgui.BeginGroup()
		imgui.TextDisabled([]string{"implicit style, follows the theme", "explicit style, unchanged by the theme"}[i])
		nc.Begin(&dfx.State{Size: half, IO: state.IO, App: state.App})
		for _, id := range canvasOrder {
			n := canvasNodes[id]
			nc.Node(id, n.pos, dfx.NodeFlags{Selected: id == "filter"}, func(ctx *dfx.NodeContext[string]) {
				ctx.TitleBar(func() { ctx.Label(n.title) })
				for _, pin := range n.ins {
					ctx.Input(id+"."+pin, pin)
				}
				for _, pin := range n.outs {
					ctx.Output(id+"."+pin, pin)
				}
			})
		}
		for j, l := range canvasLinks {
			nc.Link(fmt.Sprintf("l%d", j), l[0], l[1], dfx.LinkFlags{Selected: j == 1})
		}
		nc.End()
		imgui.EndGroup()
	}
}
