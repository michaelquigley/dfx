# NodeCanvas mini-map

Design developed from [the roadmap prompt](roadmap/nodegraph-mini-map.md), paired with the [work order](nodegraph-mini-map-work-order.md). Mercurius returned `ready_to_build` in round 1 on 2026-10-15 with no blocking findings. Michael accepted its popup-overlap test advisory, now included in the work order; the design review is closed. Stages 1 and 2 are built in the working tree, with stage 2 review pending. The remaining stage 3 work covers performance comparison, further example integration, and hands-on acceptance.

## Purpose

Large graphs need an overview that stays useful when the graph extends beyond the main canvas's zoom-out limit. Large means both spatially spread out and numerous: approximately 5,000 nodes is the upper-end design target. The mini-map should make distant regions easy to find and reach while preserving the current editing zoom.

The earlier NodeCanvas design deferred a mini-map on the expectation that detents, bookmarks, zoom-to-fit, and center-on-selection would cover navigation. This work revisits that decision for large graphs. It adds an overview to the existing navigation model.

## Presentation and configuration

The mini-map starts hidden by default. Applications choose initial visibility and provide a menu, button, shortcut, or other affordance to control it at runtime. The canvas does not draw its own visibility button.

Placement defaults to the lower-left corner. Applications can choose any of the four corners. Explicit lower-right placement leaves room for the existing zoom percentage rather than covering it.

The map has a fixed on-screen size, configurable by the application. Its footprint shrinks when needed to fit a small canvas. The first version has no interactive resize handles. Colors follow the canvas's theme and style conventions.

The overview draws simple node silhouettes, preserving declared node accent colors, with a viewport outline above them. Labels, pin markers, and links are omitted in the first version. This keeps the overview readable and limits its drawing cost at the target graph size.

## Framing

Fit the whole declared graph into the map with padding and preserved aspect ratio. The map's scale is independent of the main canvas's zoom detents. Main-canvas panning moves the viewport outline without shifting or shrinking the graph in the map.

Graph edits can change the fitted bounds. Freeze the map's coordinate mapping for the duration of a mini-map drag so a changing graph cannot move the navigation target beneath the pointer. Ordinary panning does not expand map bounds to include the viewport.

When the main viewport lies completely outside the mapped graph area, show an edge indicator pointing toward it. Clicking the map still returns the user to the graph. A viewport that becomes very small at large graph extents must remain visually locatable; the implementation plan must distinguish any enlarged visual or grab affordance from the actual viewport extent.

An empty graph retains a quiet visible panel without an invalid transform. A very small canvas omits the panel when it cannot fit; a hidden map draws and captures nothing. Exact fallback metrics belong in the work order.

## Navigation

Click a location in the map to center the main viewport there, preserving the current zoom. Drag the viewport rectangle to pan, preserving where it was grabbed so the initial press does not make it jump. Navigation changes canvas-owned view state; it does not select, move, or raise nodes, create links, or generate graph-edit intents.

The visible map panel owns new mouse input within its footprint. Nodes, embedded widgets, links, and the main canvas underneath must not also react. Existing gestures retain capture through completion, consistent with the canvas's current interaction model. Popup arbitration and lost-input handling must remain consistent with existing canvas navigation.

Mini-map navigation participates in the existing rule that the most recent explicit navigation cancels a pending zoom-to-fit. Showing or hiding the map alone does not navigate.

## Geometry and cost

The mini-map consumes geometry derived from the application's ordinary node declarations. It does not introduce a second graph model or a second node-content callback pass. The application continues to own graph positions, selection, and stacking order; the canvas owns navigation and temporary gesture state.

Start with direct rectangle rendering. Validate both dense graphs and widely separated clusters at approximately 5,000 nodes, comparing the same scene with the map shown and hidden. Measure map preparation and drawing overhead separately from the main canvas's existing cost. Check navigation responsiveness and visual legibility as well as timings; a fast drawing of an unreadable cluster is not sufficient evidence of a useful overview.

The current canvas executes each declared node's content callback every frame, including nodes outside the viewport. That existing cost is a material planning constraint, not something the mini-map itself removes. The [stage 1 stress guide](../current/node-canvas-stress.md) records the headless CPU baseline and profiled text-color improvement. Michael's post-fix hands-on pass found 5,000 nodes fine, 10,000 mostly fine, and 20,000 a little sluggish; the complete scenario matrix and timings remain unrecorded.

## Stress testing

A reusable NodeCanvas stress example and automated benchmarks were built in stage 1, before mini-map implementation. The example supports adjustable node counts through 20,000, deterministic dense and widely separated layouts, representative node content, configurable link density, and frame-time reporting. Its left panel now controls map visibility and cycles corners, allowing the same scene to be compared with the map shown and hidden without rebuilding or moving the graph.

Headless benchmarks characterize geometry and, once implemented, mini-map preparation costs. The interactive example supplies the rendering, navigation-responsiveness, and readability checks that headless timings cannot establish. This harness remains useful for future NodeCanvas regression work.

If the main canvas is already too slow at the target size with the mini-map hidden, main-canvas optimization becomes a separately scoped prerequisite. Measure and identify the bottleneck first; this feature does not silently absorb culling or a rendering rewrite.

## Scenarios

A graph spreads into distant clusters beyond the lowest zoom detent. The user opens the map, identifies a cluster by its position and colors, and clicks it. The main canvas centers on that region at the same zoom.

The user drags the viewport outline to inspect adjacent regions. The graph stays fixed in the map while the outline tracks the main view. Changes to graph bounds during the drag do not change the pointer-to-canvas mapping until the gesture finishes.

The user pans into empty space far beyond the graph. The edge indicator communicates where the viewport went, and a map click returns to a populated region.

An embedded slider sits beneath the map. Clicking the map navigates without activating that slider or emitting a node-selection or raise intent. A slider gesture that began before entering the overlay retains its existing capture.

An application chooses the upper-left corner and opens with the map visible. Resizing its canvas reduces the map's footprint when necessary; hiding it removes the map's canvas footprint entirely.

## Seam census

**Graph / overview: reuse derived canvas geometry.** The overview describes the same per-frame declarations as the main canvas. A separate application-maintained mini-map graph would duplicate truth and make synchronization the application's responsibility.

**View / graph edits: keep navigation canvas-owned.** Click and drag change the view without graph-edit intents, matching existing panning. The application's undo model remains concerned with graph edits.

**Overlay / content input: give the visible overlay first claim on new input in its footprint.** An overlay cannot be useful if navigation also changes controls beneath it. Existing gesture capture remains intact.

**Whole-canvas performance / mini-map scope: measure first, separate necessary optimization.** The stress harness establishes the baseline before the mini-map is implemented. If baseline performance is inadequate, improving it is a separate prerequisite. Offscreen culling touches content-derived sizes, pin availability, fit measurement, and widget routing; it is not assumed to be a small addition to this feature.

## Deferred (and Why)

**Labels, pins, and links.** They add clutter and drawing cost at overview scale; the first version uses spatial arrangement and node colors for orientation.

**Aggregation or density rendering.** Begin with direct rectangles. Introduce a more complex representation only if the representative 5,000-node cases show that readability or map overhead requires it.

**User resizing and free placement.** Application-configurable size and four corner placements cover the agreed presentation without another editing gesture.

**Mini-map zoom and graph editing.** The map supplies overview and pan navigation while the existing canvas supplies zoom and editing.

**Main-canvas optimization.** A separately scoped prerequisite if stage 1 shows it is needed. The harness and baseline come first so the next change follows a measured bottleneck.
