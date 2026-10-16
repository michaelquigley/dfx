# NodeCanvas mini-map — work order

Implements the [reviewed design](nodegraph-mini-map.md). Stages 1 and 2 are built; stage 2 passed Terminus and awaits Michael's review. Michael found the approximately 5,000-node target acceptable after the stage 1 performance fix, while the detailed hands-on matrix remains unrecorded. The stress example now has map visibility and corner controls; stage 3's performance comparison, further example integration, and final acceptance remain future work. The completed stage instructions below remain as review criteria until the whole work order is realized.

## Code grounding

`NodeCanvas.Begin` in `nodeCanvas.go` applies pending view and mini-map state, computes gesture previews and map layout, prepares widget input, and starts draw channels. `Node` synchronously runs content, measures its rectangle, and records geometry even outside the viewport. `End` resolves links, samples input, merges channels, draws foreground previews, mini-map chrome, and the zoom pill, commits gestures, retains geometry, and optionally measures fit candidates.

`nodeCanvasGeometry.go` contains the pure geometry types and transforms. `nodeGeometry` carries the declared accent into the overview. `nodeCanvasInput.go` is the headless gesture machine. `nodeCanvasWidgets.go` routes new widget mouse input using the previous completed frame and preserves active widget capture. The overlay footprint is reserved at `Begin` to protect synchronous widget callbacks.

`nodeCanvasFit.go` constructs a separate measurement canvas whose `Begin`/`End` reuse the visible drawing path. Mini-map chrome, input, and preparation are suppressed for `measuring` canvases.

Real ImGui frame tests already exist in `nodeCanvasWidgets_test.go`, `nodeCanvasPan_test.go`, and `nodeCanvasFit_test.go`; they require neither a display nor a renderer and must stay on one OS thread without parallel execution. The interactive example is `examples/dfx_example_nodecanvas/main.go`. `make test` runs all ordinary Go tests and vet.

## stage 1 — stress harness and baseline

The reusable stress example, deterministic render-free fixture generator, headless geometry and CPU-frame benchmarks, and timing/reproduction guide are built. [The current stress guide](../current/node-canvas-stress.md) records the commands, environment, measurements, and remaining hands-on matrix. Michael found 3,000 nodes acceptable and 5,000 choppy before a profiled text-color fix reduced the 5,000-node headless CPU frame from about 96 ms to 36 ms. After the fix, he found 5,000 fine, 10,000 mostly fine, and 20,000 a little sluggish. The detailed hands-on matrix and timings remain unrecorded.

Complete the interactive baseline at both 1.0 and the minimum detent with 5,000 nodes, in both layouts, at all three link densities. Exercise idle, pan, zoom, node dragging, and fit; distinguish transient fit measurement cost from steady-state cost. Record viewport, platform, backend, VSync, scene settings, timings, and responsiveness in the current stress guide. Michael's hands-on assessment determines whether the baseline is acceptable; no arbitrary cross-machine FPS assertion goes into CI. If it remains inadequate, resolve the [separate visibility optimization](roadmap/optimize-offscreen-node-canvas-declarations.md) before stage 2. Stage 3 adds the map visibility comparison and map benchmarks. Stage 1 finishes with Terminus and Michael's review.

## stage 2 — geometry, configuration, rendering, and navigation

Built in the working tree. Review and hands-on placement are pending; the instructions below state the intended behavior for that review.

Implement the feature as one integrated stage so interactive chrome never ships with input leaking into covered content. Use `nodeCanvasMiniMap.go` for layout, map projection, drawing, and small supporting types; keep pure gesture transitions in `nodeCanvasInput.go` and new widget arbitration in `nodeCanvasWidgets.go`. Pure projection helpers may live beside the mini-map drawing helpers if their tests remain display-free. Avoid a general overlay framework for this one feature.

### Public configuration

Add `NodeCanvasConfig.MiniMap NodeCanvasMiniMapConfig` with these fields:

```go
type NodeCanvasMiniMapConfig struct {
    Visible bool
    Corner  NodeCanvasCorner
    Size    imgui.Vec2
}
```

`NodeCanvasCorner` enumerates lower-left (zero), lower-right, upper-right, and upper-left. Invalid corners normalize to lower-left. Zero or nonpositive size components independently default to 240 by 160 screen pixels; reject nonfinite size components by normalizing them to the same defaults. Size describes the map panel.

Expose `MiniMapConfig()` and `SetMiniMapConfig(config)` for runtime configuration and persistence, plus `MiniMapVisible()` and `SetMiniMapVisible(bool)` for easy application-control integration. Getters are pending-first. Setters normalize and queue changes for the next `Begin`. The embedding application supplies any visibility affordance; the canvas has no toggle, automatic persistence, or global shortcut. Changing visibility, corner, or size terminates a mini-map capture at the next `Begin`, retaining its last valid pan; it does not cancel unrelated gestures. A hidden map has no canvas footprint.

Add theme-derived mini-map background, border, node fallback, and viewport fill/outline colors to `NodeCanvasStyle`, following its existing complete-value convention. New screen-space metrics cover padding, margin, gap, and minimum viewport grab size. Start with 6 px inner padding, 8 px outer margin, 6 px gap, and 8 px viewport grab size. These are placement defaults for the hands-on pass, not graph-space metrics. Continue deriving the zero config style lazily at first `Begin`.

### Layout and mapping

At `Begin`, before node input routing, compute the frame's panel footprint using the current viewport and applied configuration. Use the unscaled UI font metrics for zoom-pill reservation. Anchor the panel at the selected corner. Reserve the actual zoom pill height plus a gap in lower-right when the pill is enabled. Share a pure pill-layout helper with `drawZoomOverlay` so the reserved and drawn rectangles agree.

Reduce map dimensions to fit the available viewport. If the remaining panel would be smaller than 48 by 48 px, omit the panel for that frame while keeping requested visibility true. Preserve the zoom pill when a lower-right map cannot fit alongside it. A zero-area canvas draws and captures nothing. Layout omission must never leave an invisible input-blocking panel.

At `End`, derive padded graph bounds from current node geometry, using displayed drag-offset positions for nodes currently moving. Carry each node's declared accent in its derived geometry. Map graph positions with one uniform scale, centering unused aspect-ratio space inside the panel. Use a positive minimum graph extent on each axis to avoid division by zero. Empty graphs draw a quiet empty panel; map navigation is disabled. Clip node and viewport drawing to the panel's inner rectangle. Give subpixel node silhouettes a minimum 1 px visual footprint; this does not change graph bounds or navigation coordinates.

Use current-detent measured node sizes; no extra measurement callbacks. Zoom can therefore change the fitted node extents, consistent with content changing size across detents. Main-canvas pan alone never changes the graph mapping. During a mini-map gesture freeze the world-to-map transform taken from the last displayed map; newly declared geometry can draw against that fixed transform until the gesture ends, when fitting resumes.

Project the main viewport by inverting `frameView` at the canvas edges. Draw its actual clipped outline and light fill. If an axis is smaller than the minimum grab size, add a distinct center locator and expand only its grab region, clipped to the map interior; do not pretend the actual viewport is larger. When the viewport is wholly outside the mapped area, draw a direction marker at the edge toward its center; that marker is not a draggable viewport substitute. Clicking the map remains available. Draw map chrome after channel merge and font pop, above nodes and gesture previews.

### Input and view lifecycle

Use the previous completed map transform and viewport grab region for initiating navigation, since that is what the user saw. If layout, visibility, or canvas placement changes, reserve the new visible footprint immediately but swallow navigation presses for the handover frame until it has been drawn. Keep retained map geometry derived and bounded; it is not a node registry.

Before synchronous node callbacks, mask new node hover/wheel input whenever the pointer is over a visible map panel. Reuse the existing scoped mouse-routing technique without clearing active IDs, disabling widgets, or intercepting popup-window input. An already-active widget retains its capture. Suppressed clicks must not become node selection, `NodeRaised`, link creation, main-canvas pan, or wheel zoom later in `End`. Reset wheel accumulation over chrome. Hidden maps and measurement canvases reserve no map footprint.

Sample raw button state separately from `nodeInputReady`: a map click must work even if a covered node is awaiting geometry handover. Gate initiation on canvas hover, valid pointer/focus, no popup, no active canvas widget, and idle canvas gesture state. Existing canvas or widget captures continue through map regions. Hovering the map consumes wheel and new middle/right-button input without assigning those inputs a new action.

Add explicit mini-map capture states to the canvas gesture machine. A left press within the viewport grab region arms a viewport drag. Preserve the press offset; after the normal drag threshold, pan from the absolute pointer displacement divided by the frozen map scale. Moving the map viewport right makes canvas pan more negative. A release without drag recenters on the pressed map position. A left press elsewhere in the map immediately recenters and holds a no-drag capture through release, so the press cannot fall through into the canvas. Padding clicks clamp to the inner rectangle.

For recentering on canvas point `q`, use `Pan = viewport / (2 * Zoom) - q`. Preserve zoom exactly. Map drag previews are computed at `Begin` and commits at `End`, sharing one helper as existing middle-pan does. Valid release outside the map/canvas commits the final drag position. Focus loss, unavailable pointer, or lost button state ends capture at the last valid pan; restored input needs a new press. A canvas viewport resize or origin change during capture cancels it at its last valid pan to avoid mixing coordinate frames. An empty graph arising mid-capture also ends it. No map gesture emits graph-edit intents.

Accepting map navigation cancels pending fit immediately, even before a viewport drag moves. Merely changing visibility does not. Explicit `SetView`, effective `CenterOn`, or accepted `ZoomToFit` requests end a mini-map capture so the newer navigation wins; preserve the existing behavior of those methods for unrelated gestures. If a method is called mid-frame, keep that frame's drawing stable and apply the queued transition at the next `Begin`. Release any new retained state in `Destroy`, and bypass all map behavior in `measuring` canvases.

### Stage 2 validation

Pure tests cover mapping and inverse mapping, aspect ratio, negative/large coordinates, empty and single-node graphs, all corners, tiny canvases, pill reservation, fixed mapping through pan and drag, edge indicators, and true viewport bounds versus enlarged grab regions. Test gesture sequences for click, offset-preserving drag, click-without-drag, release outside, input loss, popup gating, navigation supersession, and config/layout changes during capture.

Real ImGui tests must put active buttons and wheel sliders directly beneath the map panel: press, wheel, and release must leave their values/counters unchanged and produce no canvas edit intents. Also start a slider or canvas gesture outside the map and finish over it to prove existing capture survives. Extend the existing popup-input test with a node popup overlapping the visible map; verify popup controls receive click and wheel input while mini-map navigation and underlying node controls remain inactive. This must exercise real ImGui hover and wheel routing, not only synthesized popup-gating flags. Cover first appearance, runtime corner changes, hidden maps leaving no input footprint, below/above editing detents, two independent canvases, and hidden fit measurement. Count callbacks so opening the map cannot silently add a measurement pass. Reuse the existing test lifecycle and run `make test`. Finish with Terminus and Michael's stage review.

## stage 3 — performance comparison, examples, and acceptance

The stress example's left panel now has a show/hide button and a corner-cycle button through the public API, reading state from the getters. Add a size control. Compare shown/hidden runs using identical graph, link density, viewport, and view presets; timing samples already reset and warm when switching visibility or corner. Add map preparation benchmarks at all preset counts, plus shown/hidden variants of the real-ImGui CPU-frame benchmark. Keep map work linear in declared nodes and avoid allocating a second graph representation per frame; verify hidden mode skips bounds/projection work entirely.

Repeat the stage 1 manual scenarios and record the incremental cost. Include a fixed overhead measurement away from all geometry and a normal navigation run, so fewer underlying hit tests over the overlay cannot masquerade as cheaper rendering. Check that 5,000-node dense regions and distant clusters remain recognizable enough to navigate. Report visual limitations or unacceptable map overhead for a focused decision; do not silently introduce aggregation. The target is useful interaction at approximately 5,000 nodes on the tested machine, not an unmeasured universal frame-rate guarantee.

Add a shortcut and configuration sample to the existing small example. Update `docs/current/node-canvas.md` for the API, overlay arbitration, fit precedence, and current-detent geometry behavior; update the stress guide with the final benchmark commands and manual reproduction details. Add an Unreleased FEATURE paragraph in the repository's changelog format. Run `make test`, benchmarks, and manual acceptance in all four corners, then Terminus and Michael's review.

## Completion and handoff

Each stage must leave the tree buildable and ordinary tests passing. Planning does not implement these stages. Stage 1's interactive baseline is a real dependency before stage 2; a failed baseline produces a separate prerequisite rather than an unplanned optimization inside the mini-map patch.

After implementation, synthesize built behavior and accepted limitations into `docs/current/`, retire the realized spec/work order per the repository workflow, and keep still-live follow-ups separately. Resolve links before removing intent documents. Do not delete the roadmap card or edit roadmap ordering; its disposition is the operator's. Leave all work uncommitted for Michael.
