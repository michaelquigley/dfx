# CHANGELOG

## Unreleased

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
