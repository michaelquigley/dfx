# CHANGELOG

## Unreleased

FIX: `NodeCanvas` panning preserves the last valid view when application focus, mouse position, or held-button state is lost. Unavailable coordinates are rejected before both preview rendering and view commit, preventing the graph from disappearing or grid drawing from stalling. Normal pan releases outside the canvas still commit their final position.

FIX: Overlapping `NodeCanvas` widgets route new mouse input to the frontmost node from the last completed frame, so covered sliders no longer share wheel adjustments with visible controls. Callbacks remain synchronous, active widget drags retain capture, and popup input stays independent. New or changed input geometry waits for a measured frame; changes declared later in the frame settle on the following frame. The new `NodeRaised` intent lets applications bring a clicked node forward by changing declaration order; the node-canvas example demonstrates this alongside wheel sliders.

FEATURE: `NodeCanvas[ID]` is a zoomable, pannable node-graph editing surface replacing imnodes-style retained editors with the dfx immediate-mode idiom: the application declares nodes, links, positions, and selection every frame and the canvas owns only view and gesture state, reporting completed gestures as intents (`NodesMoved`, `LinkCreated`, `SelectionChanged`) the app applies — typically as undo commands. Zoom moves through configured detents (defaults to a 0.1-step set from 0.25 to 1.5, 1.0 being the editing detent; see the CHANGE entries below) with fonts scaled per detent under imgui 1.92's dynamic font system; app-supplied `comparable` IDs span nodes, pins, and links; interaction covers multi-select, whole-selection dragging, box select, link creation with pin snap, middle-drag pan, wheel zoom toward the cursor, locked mode, `ZoomToFit`/`CenterOn` navigation, and persistable plain-data view state. Documented in `docs/current/node-canvas.md`; `examples/dfx_example_nodecanvas` demonstrates the full surface.

CHANGE: `NodeCanvas` wheel zoom accumulates wheel travel and steps a detent per `Config.WheelStepsPerZoomLevel` notches (default 1.0 — one notch, one detent, the classic behavior). At most one detent steps per frame; sub-threshold remainder carries across frames, a direction change resets the banked travel, and the bank also resets while the canvas cannot step (not hovered, popup open, an in-canvas item hovered, or a gesture in flight), so scrolling elsewhere in the app or mid-gesture never produces a phantom step. Fine-scroll devices that report fractional ticks step far less often than before, and apps whose wheels feel too sensitive can raise the value — `examples/dfx_example_nodecanvas` sets 2.

CHANGE: `NodeCanvas` default detents are now `{0.25, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1, 1.2, 1.3, 1.4, 1.5}` — finer 0.1-step granularity across the working range (one coarser step from 0.4 to the 0.25 zoom-out floor), and detents past 100%, so the wheel traverses the zoom range in small steps. The detent contract relaxes from "last entry must be 1.0" to "1.0 — the editing detent — must be a member of the set", with entries above it allowed (full interactive content continues to apply at detent 1.0 and above). A canvas now opens at the editing detent rather than the largest configured one, and `ZoomToFit` can resolve above 1.0 for small graphs.

FEATURE: `Config.OnAction func(ActionEvent)` reports every action as it fires, so an application can record which actions and shortcuts actually get used and refine its keybindings from real usage rather than guesswork. Keyboard invocations route through a single dispatch chokepoint that notifies the hook before running the handler; each `ActionEvent` carries the invoked `*Action`, its `Source`, and the `Time`. Menu-click capture is not yet wired — `ActionSourceMenu` is reserved for it.

CHANGE: Reference documentation moved under `docs/current/` (`layout-guide.md` and `child-actions.md`, formerly `docs/LAYOUT_GUIDE.md` and `docs/CHILD_ACTIONS.md`), and a repo-local `AGENTS.md` now carries contributor and agent orientation. README links updated to match.

## v0.1.3

Initial release.
