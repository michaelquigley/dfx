package dfx

import (
	"fmt"
	"runtime"
	"slices"
	"testing"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx/fonts"
)

// headlessBackend stands in for the GLFW backend: it owns an imgui context from
// CreateWindow until Run returns, as the real backend does, and runs one frame,
// or frames, in an 800x600 display unless size is set. unimplemented methods
// nil-panic through the embedded interface.
type headlessBackend struct {
	backend.Backend[glfwbackend.GLFWWindowFlags]
	ctx    *imgui.Context
	panic  bool
	size   imgui.Vec2
	frames int
}

func (b *headlessBackend) CreateWindow(string, int, int) {
	b.ctx = imgui.CreateContext()
	io := imgui.CurrentIO()
	io.SetIniFilename("")
	if b.size == (imgui.Vec2{}) {
		b.size = imgui.Vec2{X: 800, Y: 600}
	}
	io.SetDisplaySize(b.size)
	io.SetDeltaTime(1.0 / 60)
	io.SetBackendFlags(imgui.BackendFlagsRendererHasTextures)
}

func (b *headlessBackend) SetAfterRenderHook(func()) {}

func (b *headlessBackend) Run(loop func()) {
	defer imgui.DestroyContextV(b.ctx)
	for range max(1, b.frames) {
		imgui.NewFrame()
		loop()
		imgui.Render()
	}
	if b.panic {
		panic("backend failed")
	}
}

// runHeadless runs app on a headless backend, restoring the backend seam and any
// globals a failed run could leave behind.
func runHeadless(t *testing.T, app *App, be *headlessBackend) error {
	t.Helper()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	original := createBackend
	createBackend = func() (backend.Backend[glfwbackend.GLFWWindowFlags], error) { return be, nil }
	defer func() { createBackend = original }()
	return app.Run()
}

// sameFont compares the C fonts behind two wrappers; cimgui-go returns a new
// wrapper from each call.
func sameFont(a, b *imgui.Font) bool {
	pa, fa := a.Handle()
	defer fa()
	pb, fb := b.Handle()
	defer fb()
	return pa == pb
}

// spyTheme records its applications and the setup state each one observed.
type spyTheme struct {
	name     string
	applied  int
	padding  imgui.Vec2 // frame padding when applied
	fonts    int        // loaded font slots when applied
	accent   imgui.Vec4
	observed []string
}

func (s *spyTheme) Name() string { return s.name }

func (s *spyTheme) Apply() {
	s.applied++
	s.padding = imgui.CurrentStyle().FramePadding()
	s.fonts = len(Fonts)
	colors := imgui.CurrentStyle().Colors()
	colors[imgui.ColCheckMark] = s.accent
	imgui.CurrentStyle().SetColors(&colors)
}

// customStyle differs from Classic in every startup-visible respect.
func customStyle(theme Theme) Style {
	style := ClassicStyle()
	style.Name = "Custom"
	style.Theme = theme
	style.Metrics.FramePadding = imgui.Vec2{X: 9, Y: 7}
	style.Spacing.SectionGap = 21
	style.Typography.Small.Size = 14
	style.Typography.Heading = FontFace{Data: fonts.GidoleRegular, Size: 24, Icons: true, IconOffset: 6}
	return style
}

func TestRun_StyleSetupOrderPrecedenceAndFlags(t *testing.T) {
	for _, disableFonts := range []bool{false, true} {
		for _, disableTheming := range []bool{false, true} {
			t.Run(fmt.Sprintf("fonts=%v/theming=%v", !disableFonts, !disableTheming), func(t *testing.T) {
				styleTheme := &spyTheme{name: "style", accent: imgui.Vec4{X: 0.1, W: 1}}
				override := &spyTheme{name: "override", accent: imgui.Vec4{Y: 0.2, W: 1}}
				style := customStyle(styleTheme)
				// setup reuses Fonts' backing array, so compare against a separate copy
				sentinel := []*imgui.Font{nil}
				previousFonts, previousSizes := Fonts, fontSizes
				Fonts = slices.Clone(sentinel)
				defer func() { Fonts, fontSizes = previousFonts, previousSizes }()

				setupRan, drew := false, false
				var root Component = NewFunc(func(*State) {
					drew = true
					before := imgui.CurrentFont()
					PushHeadingFont()
					pushed := imgui.CurrentFont()
					PopFont()
					if disableFonts {
						if !sameFont(pushed, before) {
							t.Fatal("heading helper must push the current font when dfx fonts are disabled")
						}
					} else if !sameFont(pushed, Fonts[headingFont]) || fontSizes[headingFont] != 24 {
						t.Fatal("heading helper must push the style's heading face")
					}
				})
				app := New(root, Config{
					Style:          &style,
					Theme:          override,
					DisableFonts:   disableFonts,
					DisableTheming: disableTheming,
					OnSetup: func(*App) {
						setupRan = true
						current := CurrentStyle()
						if current.Name != "Custom" || current.Theme != override || current.Spacing.SectionGap != 21 {
							t.Fatalf("OnSetup observed the wrong style: '%s' theme '%v'", current.Name, current.Theme)
						}
						// metrics apply whether or not colors do
						if p := imgui.CurrentStyle().FramePadding(); p != style.Metrics.FramePadding {
							t.Fatalf("metrics not applied before OnSetup: '%v'", p)
						}
						if disableFonts {
							if !slices.Equal(Fonts, sentinel) || !slices.Equal(fontSizes, previousSizes) {
								t.Fatal("disabled fonts must leave the font globals alone")
							}
						} else if len(Fonts) != 4 || fontSizes[SmallFont] != 14 {
							t.Fatalf("expected the style's four font slots, got '%d' sizes '%v'", len(Fonts), fontSizes)
						}
						checkMark := imgui.CurrentStyle().Colors()[imgui.ColCheckMark]
						if disableTheming {
							if override.applied != 0 || checkMark == override.accent {
								t.Fatal("disabled theming must not apply a color theme")
							}
						} else if override.applied != 1 || checkMark != override.accent {
							t.Fatalf("effective theme not applied once before OnSetup: '%d'", override.applied)
						}
					},
				})
				if err := runHeadless(t, app, &headlessBackend{}); err != nil {
					t.Fatal(err)
				}
				if !setupRan || !drew {
					t.Fatal("setup or draw did not run")
				}
				if styleTheme.applied != 0 {
					t.Fatal("Config.Theme must override the style's default theme")
				}
				if !disableTheming {
					// the theme ran after fonts and metrics
					if override.padding != style.Metrics.FramePadding {
						t.Fatalf("theme applied before metrics: '%v'", override.padding)
					}
					if expected := map[bool]int{false: 4, true: 1}[disableFonts]; override.fonts != expected {
						t.Fatalf("theme applied before fonts: '%d' slots", override.fonts)
					}
				}
				if activeStyle != nil || !slices.Equal(Fonts, sentinel) || !slices.Equal(fontSizes, previousSizes) {
					t.Fatal("Run must restore the style metadata and font globals")
				}
			})
		}
	}
}

func TestRun_StyleThemeAndModernDarkFallback(t *testing.T) {
	styleTheme := &spyTheme{name: "style", accent: imgui.Vec4{Z: 0.3, W: 1}}
	style := customStyle(styleTheme)
	app := New(nil, Config{Style: &style, OnSetup: func(*App) {
		if styleTheme.applied != 1 || CurrentStyle().Theme != styleTheme {
			t.Fatal("the style's default theme must apply without a Config.Theme")
		}
	}})
	if err := runHeadless(t, app, &headlessBackend{}); err != nil {
		t.Fatal(err)
	}

	style.Theme = nil
	app = New(nil, Config{Style: &style, OnSetup: func(*App) {
		if name := CurrentStyle().Theme.Name(); name != ModernDark.Name() {
			t.Fatalf("a nil style theme must fall back to Modern Dark, got '%s'", name)
		}
		expected := imgui.Vec4{X: 1.0, Y: 0.733, Z: 0.124, W: 0.828}
		if got := imgui.CurrentStyle().Colors()[imgui.ColButtonHovered]; got != expected {
			t.Fatalf("Modern Dark not applied: '%v'", got)
		}
	}})
	if err := runHeadless(t, app, &headlessBackend{}); err != nil {
		t.Fatal(err)
	}
}

func TestRun_StyleIsCopiedAtStartup(t *testing.T) {
	style := customStyle(nil)
	app := New(nil, Config{Style: &style, DisableFonts: true, OnSetup: func(*App) {
		style.Name = "Mutated"
		style.Spacing.SectionGap = 99
		if current := CurrentStyle(); current.Name != "Custom" || current.Spacing.SectionGap != 21 {
			t.Fatal("the active style must be a copy of Config.Style")
		}
		copied := CurrentStyle()
		copied.Spacing.SectionGap = 77
		if CurrentStyle().Spacing.SectionGap != 21 {
			t.Fatal("CurrentStyle must return a copy")
		}
	}})
	if err := runHeadless(t, app, &headlessBackend{}); err != nil {
		t.Fatal(err)
	}
}

func TestRun_NextAppDoesNotInheritStyle(t *testing.T) {
	previousFonts, previousSizes := Fonts, fontSizes
	Fonts = nil
	defer func() { Fonts, fontSizes = previousFonts, previousSizes }()

	style := customStyle(nil)
	if err := runHeadless(t, New(nil, Config{Style: &style}), &headlessBackend{}); err != nil {
		t.Fatal(err)
	}
	if CurrentStyle().Name != "Classic" || Fonts != nil {
		t.Fatal("a finished run left its style or fonts behind")
	}

	// the second app neither selects a style nor loads fonts; the first app's
	// selection and heading slot must not reach it.
	drew := false
	root := NewFunc(func(*State) {
		drew = true
		before := imgui.CurrentFont()
		PushHeadingFont()
		if !sameFont(imgui.CurrentFont(), before) {
			t.Fatal("heading helper reached an unloaded slot")
		}
		PopFont()
	})
	app := New(root, Config{DisableFonts: true, OnSetup: func(*App) {
		current := CurrentStyle()
		if current.Name != "Classic" || current.Spacing.SectionGap != ClassicStyle().Spacing.SectionGap {
			t.Fatalf("second app inherited style '%s'", current.Name)
		}
		if p := imgui.CurrentStyle().FramePadding(); p != ClassicStyle().Metrics.FramePadding {
			t.Fatalf("second app inherited metrics '%v'", p)
		}
	}})
	if err := runHeadless(t, app, &headlessBackend{}); err != nil {
		t.Fatal(err)
	}
	if !drew {
		t.Fatal("second app did not draw")
	}
}

func TestRun_RestoresStyleAfterPanic(t *testing.T) {
	previousFonts, previousSizes := Fonts, fontSizes
	Fonts = nil
	defer func() { Fonts, fontSizes = previousFonts, previousSizes }()

	style := customStyle(nil)
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected the backend panic to propagate")
			}
		}()
		_ = runHeadless(t, New(nil, Config{Style: &style}), &headlessBackend{panic: true})
	}()
	if activeStyle != nil || Fonts != nil || !slices.Equal(fontSizes, previousSizes) {
		t.Fatal("a panicking run must restore the style metadata and font globals")
	}
}

func TestRun_RejectsStyleWithoutFonts(t *testing.T) {
	style := ClassicStyle()
	style.Typography.Monospace.Data = nil
	original := createBackend
	createBackend = func() (backend.Backend[glfwbackend.GLFWWindowFlags], error) {
		t.Fatal("an invalid style must be rejected before the window opens")
		return nil, nil
	}
	defer func() { createBackend = original }()
	if err := New(nil, Config{Style: &style}).Run(); err == nil {
		t.Fatal("expected an error for a style missing its monospace font")
	}

	style = ClassicStyle()
	style.Typography.Heading = FontFace{Data: fonts.GidoleRegular}
	if err := New(nil, Config{Style: &style}).Run(); err == nil {
		t.Fatal("expected an error for a heading face without a size")
	}
}
