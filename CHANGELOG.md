# CHANGELOG

## Unreleased

CHANGE: **dfx's default presentation is now Faceplate: applications that set no `Config.Style` start with Chakra Petch text, roomier spacing and margins, and charcoal colors with a yellow accent instead of Gidole, compact spacing, and Modern Dark.** Select `ClassicStyle()` through `Config.Style` to keep the previous fonts, metrics, colors, and widget treatments. `Config.Theme` and `SetTheme` still change colors only, so an application that set `Theme: dfx.ModernDark` keeps those colors but gets Faceplate's fonts and spacing unless it also selects Classic. The allocation corrections below apply under both styles, so Classic does not restore historical sizing defects. `SetupFonts`, `DefaultStyle`, `ModernDark`, and `CurrentStyle()` outside a running App still mean Classic. Several examples were resized or now measure their remaining space so that their default launch fits the new spacing, and `examples/dfx_example_themes` starts in Faceplate, with `-style classic` for comparison. See "Migrating from Classic" in `docs/current/styles.md`.

CHANGE: **Containers now hand components the usable area at their drawing origin, which can shift existing layouts by a few pixels.** The root component's `State.Size` is the content region inside the root's padding (the style's outer margin, 12 px in Faceplate and 4 px in Classic) rather than the whole window. `DashManager` and `Dash` place panels from the manager's content origin instead of the enclosing window's outer edge, inset dash surfaces by the panel padding on every side, and pass components their actual content region. `Workspace` reserves the selector's measured height instead of an approximate 30 px. FlexLayout subtracts the 10 px splitter strip it draws instead of 11 px, and GridLayout places cells from its content origin and clips them to its allocation. No container passes a zero or negative size to `BeginChild` any longer: exhausted panels and cells, including a dash its neighbors leave no room, are skipped, and a component whose container's padding leaves no usable area is not drawn. Classic's fonts, metrics, colors, and widget treatments are otherwise unchanged. See the allocation notes in `docs/current/layout-guide.md`.

CHANGE: **`DashManager.TopMargin` now reserves space above the entire manager layout under both precedence modes, which moves panels in layouts that set it.** The manager's origin moves down by the margin and its available height loses it once, so every dash and the inner panel are placed in the remainder. Previously, under `VerticalPrecedence`, side dashes moved down without getting shorter, so they overflowed the manager, and the inner panel ignored the margin and overlapped the top dash. Under `HorizontalPrecedence`, the margin applied only when a top dash was present. A dash taller than the space left to it is shortened to fit, and a margin that takes the whole height skips every panel.

FEATURE: `Config.Style` selects a complete style — typography, ImGui metrics, default color theme, spacing values, and the color roles dfx's chrome reads — at startup. `ClassicStyle()` returns the existing fonts, metrics, and Modern Dark colors as a customizable configuration copy with shared immutable font assets. `Config.Theme` still overrides the colors, `DisableTheming` still applies metrics, and `CurrentStyle()` reports the running selection. `PushHeadingFont` pushes a style's optional heading face, falling back to the main font. Fader fill and border colors resolve through the style's roles, which Classic binds to the previous slots. `Run` returns an error for a style missing required font data. See `docs/current/styles.md`.

FEATURE: `FaceplateStyle()` gives dfx the faceplate family's face: bundled Chakra Petch Regular and real Bold, charcoal surfaces, one identity accent, and roomier spacing, with icons merged into every UI face. Plain child windows stay unpainted, as under Modern Dark, while Dash and HCollapse panels carry the panel surface. `FaceplateTheme{Accent}` changes the accent and can also be applied alone with `SetTheme`. The running App owns its copy of a `FaceplateTheme`, whether supplied through the style or `Config.Theme`, and each `CurrentStyle()` snapshot carries its own copy, so editing one changes neither the running selection nor another snapshot. Dash, HCollapse, MultiGrid, and Toolbar take their rounding, padding, gaps, and insets from the style's `Chrome`, and toolbar and drag-handle colors from its roles, with Classic keeping the previous values. `NodeCanvas`'s derived selection emphasis follows the style's accent role. `examples/dfx_example_themes` is now a full comparison surface, launched with `-style faceplate|classic` and an optional `-accent RRGGBB`.

FEATURE: A style's `Controls.AccentInk` draws labels on accent fills in whichever of the theme's background and text colors contrasts more: on toggles, which also keep their accent fill when hovered and pressed, and on `NodeCanvas` title labels against the node's band, so a selected Faceplate node's yellow title band carries charcoal text. Faceplate enables it; Classic's toggles and canvas titles are unchanged. `ToggleV` takes `ToggleOptions`, whose `OnText` sets an explicit on-state label color.

CHANGE: `ModernDark` starts from ImGui's dark colors before applying its own, so applying it after another theme no longer leaves that theme's colors in the slots Modern Dark does not set. The result on a fresh context is unchanged.

## v0.1.15

FEATURE: `NodeCanvas` offers an optional mini-map with configurable visibility, size, and corner, defaulting to the lower-left. It draws node silhouettes and a viewport outline for large-graph navigation; clicking recenters and dragging the viewport pans without changing zoom or emitting graph edits. Applications own the visibility control through `SetMiniMapVisible`; the node-canvas example binds M, and the stress example exposes visibility, corner, and size controls.

FEATURE: The NodeCanvas stress example adds deterministic grid and cluster scenes with 100 to 20,000-node presets, adjustable counts, link densities, and separate CPU submission and displayed-frame cadence timings. Headless geometry and ImGui CPU-frame benchmarks cover the same presets.

FIX: `NodeCanvas` text labels and pin rows read ImGui's text color directly instead of copying the full style-color array for every item. This reduces CPU submission cost on large graphs and honors the current style alpha.

## v0.1.14

CHANGE: Faders use a narrow slot, accent fill, and a rounded handle with a position marker, all using the active theme colors by default. `FaderParams` adds accent color, track width, and handle dimensions while retaining the existing slider interactions and value APIs. Scale marks now align with the handle's inset travel; `DefaultScaleConfig` enables horizontal guide lines. The mixer example opens on a larger five-channel reference bank.

FEATURE: `Config.AppID` sets the window's Wayland app id and X11 class and instance names, so the desktop pairs a running window with the launcher entry of the same name and shows its icon. Unset, the window reports no app id on Wayland and its title as the X11 class, as before.

## v0.1.13

CHANGE: **Node content callbacks can run again during `End` while fitting.** This hidden measurement pass disables widget input and reports `NodeContext.Measuring()`. Callbacks must keep per-node data valid through `End`, derive layout from `NodeContext.Detent()`, and suppress non-widget side effects during measurement.

FIX: `NodeCanvas.ZoomToFit` keeps the current view visible while measuring candidate zoom levels, then moves directly to the fitted view. The intermediate zoom-in and step-down are no longer displayed.

## v0.1.12

FIX: `HCollapse` resizing applies its `MinWidth` clamp last, so no container width or sequence of drags can take a panel below `MinWidth`. Before, a caller that passed the panel's own width as the container could drag the panel out of existence. `docs/current/hcollapse.md` states the sizing contract: `Draw` takes the container's full available size.

FEATURE: `HCollapseConfig.Anchor` (`AnchorLeft`, the default, or `AnchorRight`) lets a panel sit against the right edge of its row. Its resize handle moves to its left edge, dragging it left widens the panel, and its toggle moves to the header's right. The hcollapse example adds a right-anchored panel.

## v0.1.11

FEATURE: `NodeCanvas` declarations can carry their own colors. `NodeFlags.Accent` colors a node's title band and pins, and `LinkFlags.Color` sets a link's normal color; the zero value keeps the style. A selected colored node or link shows its selection as a thicker line in a brighter highlight of its own color rather than the style's selection color, so selection never clashes with the color. Hover keeps the style's color. The node-canvas example accents its filter node and one link.

## v0.1.10

FEATURE: `SetupFonts` merges supplemental Material Symbols into the main and small UI fonts, including the 1:1 `fonts.ICON_SYMBOL_VIEW_REAL_SIZE` glyph. The fixed-style supplement excludes all codepoints already supplied by Material Icons, preserving existing icon artwork and font slots. New supported symbols use the additive `ICON_SYMBOL_*` constant namespace; source, licensing, generation, and supported codepoint range are documented in `docs/current/icon-fonts.md`.

## v0.1.9

FEATURE: `NodeCanvas` shows the current zoom as a whole percentage in a small overlay anchored to the canvas's lower-right corner, drawn in the unscaled UI font so it reads the same at every detent. It is pure drawlist chrome rather than an imgui item, so it never captures input or carves a hole in hit-testing. `NodeCanvasConfig.HideZoomOverlay` turns it off; `NodeCanvasStyle` gains `ZoomOverlayBgColor`, `ZoomOverlayTextColor`, and screen-pixel `ZoomOverlayPadding`, `ZoomOverlayMargin`, and `ZoomOverlayRounding`, derived from the theme by `DefaultNodeCanvasStyle`. Custom styles built from a zero value rather than the default will draw the overlay with zero-alpha colors and no padding until those fields are set.

CHANGE: **`NodeCanvas` zoom now requires Ctrl+wheel; a bare wheel over the canvas does nothing.** The wheel tends to get nudged while the middle button is held for a pan, and an unmodified wheel that zoomed turned pans into pan-plus-zoom. Wheel travel banks only while Ctrl is held, so notches scrolled without the modifier never combine with a later Ctrl notch into a step. `WheelSlider`'s Ctrl fast mode is unaffected: over a hovered slider the widget still owns the wheel, modified or not.

## v0.1.8

CHANGE: Upgraded cimgui-go to v1.6.0 and adapted `NodeCanvas` rectangle and grid drawing calls to the updated draw-list API.

## v0.1.7

FIX: The full-window root uses square corners so its background covers the native window's corners. The rounding override applies only to the root; application content retains the configured style.

FIX: `NodeCanvas.Destroy()` releases the owned native draw-list splitter and its channel buffers when an editor is closed or replaced. Cleanup is explicit, safe before the first draw and on repeated calls, and supported after ImGui context shutdown; the node-canvas example wires it into `OnShutdown`. A destroyed canvas cannot be drawn again, and destruction during `Begin`/`End` is rejected.

FIX: `NodeCanvas` panning preserves the last valid view when application focus, mouse position, or held-button state is lost. Unavailable coordinates are rejected before both preview rendering and view commit, preventing the graph from disappearing or grid drawing from stalling. Normal pan releases outside the canvas still commit their final position.

FIX: Starting a middle-button pan immediately cancels a pending `NodeCanvas` zoom-to-fit, even before the pointer moves, so automatic fitting no longer competes with manual navigation.

FIX: Overlapping `NodeCanvas` widgets route new mouse input to the frontmost node from the last completed frame, so covered sliders no longer share wheel adjustments with visible controls. Callbacks remain synchronous, active widget drags retain capture, and popup input stays independent. New or changed input geometry waits for a measured frame; changes declared later in the frame settle on the following frame. The new `NodeRaised` intent lets applications bring a clicked node forward by changing declaration order; the node-canvas example demonstrates this alongside wheel sliders.

FEATURE: `NodeCanvas[ID]` is a zoomable, pannable node-graph editing surface replacing imnodes-style retained editors with the dfx immediate-mode idiom: the application declares nodes, links, positions, and selection every frame and the canvas owns only view and gesture state, reporting completed gestures as intents (`NodesMoved`, `LinkCreated`, `SelectionChanged`) the app applies — typically as undo commands. Zoom moves through configured detents (defaults to a 0.1-step set from 0.25 to 1.5, 1.0 being the editing detent; see the CHANGE entries below) with fonts scaled per detent under imgui 1.92's dynamic font system; app-supplied `comparable` IDs span nodes, pins, and links; interaction covers multi-select, whole-selection dragging, box select, link creation with pin snap, middle-drag pan, wheel zoom toward the cursor, locked mode, `ZoomToFit`/`CenterOn` navigation, and persistable plain-data view state. Documented in `docs/current/node-canvas.md`; `examples/dfx_example_nodecanvas` demonstrates the full surface.

CHANGE: `NodeCanvas` wheel zoom accumulates wheel travel and steps a detent per `Config.WheelStepsPerZoomLevel` notches (default 1.0 — one notch, one detent, the classic behavior). At most one detent steps per frame; sub-threshold remainder carries across frames, a direction change resets the banked travel, and the bank also resets while the canvas cannot step (not hovered, popup open, an in-canvas item hovered, or a gesture in flight), so scrolling elsewhere in the app or mid-gesture never produces a phantom step. Fine-scroll devices that report fractional ticks step far less often than before, and apps whose wheels feel too sensitive can raise the value — `examples/dfx_example_nodecanvas` sets 2.

CHANGE: `NodeCanvas` default detents are now `{0.25, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1, 1.2, 1.3, 1.4, 1.5}` — finer 0.1-step granularity across the working range (one coarser step from 0.4 to the 0.25 zoom-out floor), and detents past 100%, so the wheel traverses the zoom range in small steps. The detent contract relaxes from "last entry must be 1.0" to "1.0 — the editing detent — must be a member of the set", with entries above it allowed (full interactive content continues to apply at detent 1.0 and above). A canvas now opens at the editing detent rather than the largest configured one, and `ZoomToFit` can resolve above 1.0 for small graphs.

## v0.1.6

FIX: **`App.SetWindowSize` and `App.SetWindowPos` now take effect at the next frame boundary rather than at the moment they are called.** Applying window geometry from inside a frame crashes on Wayland: the GLFW call dispatches a surface configure synchronously, which re-enters the render loop through GLFW's window-refresh callback and opens a second ImGui frame inside the current one, tripping ImGui's "Forgot to call Render() or EndFrame()" assertion. X11 issues the same call asynchronously to the window manager, which is why the hazard stayed invisible there. Both setters now record the request and flush it from the backend's `afterRender` hook, outside any frame; a later request before that boundary supersedes an earlier one, and a boundary with nothing pending does not reassert geometry. Applications that called either setter and immediately read the geometry back will now observe the old value until the frame ends.

## v0.1.5

FEATURE: `App.SetWindowPos(x, y int)` and `App.SetWindowSize(w, h int)` expose runtime window geometry control, delegating to the backend behind the same nil guard as `SetWindowTitle`. Previously geometry was fixed once at `Run()` from `Config`; an application can now move and resize the window live (for example, transitioning from a compact launcher surface to a larger main surface). Both must be called on the UI goroutine, since the underlying GLFW operations are main-thread only.

## v0.1.4

FEATURE: `Config.OnAction func(ActionEvent)` reports every action as it fires, so an application can record which actions and shortcuts actually get used and refine its keybindings from real usage rather than guesswork. Keyboard invocations route through a single dispatch chokepoint that notifies the hook before running the handler; each `ActionEvent` carries the invoked `*Action`, its `Source`, and the `Time`. Menu-click capture is not yet wired — `ActionSourceMenu` is reserved for it.

CHANGE: Reference documentation moved under `docs/current/` (`layout-guide.md` and `child-actions.md`, formerly `docs/LAYOUT_GUIDE.md` and `docs/CHILD_ACTIONS.md`), and a repo-local `AGENTS.md` now carries contributor and agent orientation. README links updated to match.

## v0.1.3

Initial release.
