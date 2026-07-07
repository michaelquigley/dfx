# CHANGELOG

## Unreleased

FEATURE: `NodeCanvas[ID]` is a zoomable, pannable node-graph editing surface replacing imnodes-style retained editors with the dfx immediate-mode idiom: the application declares nodes, links, positions, and selection every frame and the canvas owns only view and gesture state, reporting completed gestures as intents (`NodesMoved`, `LinkCreated`, `SelectionChanged`) the app applies — typically as undo commands. Zoom moves through configured detents (default 0.25–1.0) with fonts scaled per detent under imgui 1.92's dynamic font system; app-supplied `comparable` IDs span nodes, pins, and links; interaction covers multi-select, whole-selection dragging, box select, link creation with pin snap, middle-drag pan, wheel zoom toward the cursor, locked mode, `ZoomToFit`/`CenterOn` navigation, and persistable plain-data view state. Documented in `docs/current/node-canvas.md`; `examples/dfx_example_nodecanvas` demonstrates the full surface.

FEATURE: `Config.OnAction func(ActionEvent)` reports every action as it fires, so an application can record which actions and shortcuts actually get used and refine its keybindings from real usage rather than guesswork. Keyboard invocations route through a single dispatch chokepoint that notifies the hook before running the handler; each `ActionEvent` carries the invoked `*Action`, its `Source`, and the `Time`. Menu-click capture is not yet wired — `ActionSourceMenu` is reserved for it.

CHANGE: Reference documentation moved under `docs/current/` (`layout-guide.md` and `child-actions.md`, formerly `docs/LAYOUT_GUIDE.md` and `docs/CHILD_ACTIONS.md`), and a repo-local `AGENTS.md` now carries contributor and agent orientation. README links updated to match.

## v0.1.3

Initial release.