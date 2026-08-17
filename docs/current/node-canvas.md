# NodeCanvas

`NodeCanvas[ID]` is a zoomable, pannable node-graph editing surface: rounded-rect node cards with title bars, pin rows, and cubic-bezier links, over a grid. It replaces imnodes-style retained editors with dfx's immediate-mode idiom — the application owns the graph; the canvas owns the view.

`examples/dfx_example_nodecanvas` demonstrates everything on this page.

## Ownership model

**The app owns all graph truth.** Nodes, links, positions, and selection are declared every frame between `Begin` and `End`. The canvas never mutates the graph and holds no graph state — there is no position read-back, no selection sync, no node registry.

**The canvas owns view and gesture state only.** Pan, zoom detent, and in-flight gestures (a drag in progress, a link being pulled, a box-select preview) are its only retained state, plus derived per-frame geometry that is rebuilt from declarations every frame and is never authoritative.

**The canvas reports intents; the app applies them.** A completed gesture produces an intent — nodes moved, link created, selection changed — returned from `End`. The app decides what to do with it (typically: run an undo command). The canvas never assumes an intent was accepted; the next frame's declarations are the only truth it renders.

## Declaration cycle

`NodeCanvas` is a widget, not a `Component`: drive it from inside an owning component's `Draw`.

```go
nc := dfx.NewNodeCanvas[string](dfx.NodeCanvasConfig{})

// each frame:
nc.Begin(state)

nc.Node("filter", pos, dfx.NodeFlags{Selected: sel}, func(n *dfx.NodeContext[string]) {
    n.TitleBar(func() { n.Label(fonts.ICON_TUNE + " filter") })
    if n.Detent() < 1.0 {
        n.Label(fmt.Sprintf("cutoff %.0f", cutoff)) // simplified, non-interactive
    } else {
        imgui.PushItemWidth(140)
        imgui.SliderFloat("cutoff", &cutoff, 20, 20000)
        imgui.PopItemWidth()
    }
    n.Input("filter.in", "in")
    n.Output("filter.out", "out")
})

nc.Link("l1", "source.out", "filter.in", dfx.LinkFlags{Selected: sel})

intents := nc.End()
```

- `Node`'s `pos` is the outer node rect's top-left in **canvas space** — the app-owned anchor. Padding and title metrics offset content inward from it; content growth extends the rect right/down while the anchor stays fixed. Node size derives from content.
- `TitleBar`, when used, must be the first call in the content closure; its measured extent is the title band height (violations are debug-logged).
- `Input`/`Output` declare pin rows: a label in flow plus a circular marker on the node's left (inputs) or right (outputs) edge, anchored to the row.
- `Label` draws text through the drawlist and advances layout with an ID-less spacer — it measures like an item but stays invisible to hover/active arbitration.
- `Link` records a declaration; links resolve and draw at `End`. **Pin presence is a per-frame invariant**: a link whose endpoint pin was not declared this frame is skipped (debug-logged), never drawn from stale geometry.

### IDs

Node, pin, and link IDs share one `comparable` type and one ID space; the app guarantees uniqueness across all three populations. IDs must be unique and stable frame-to-frame in their `fmt.Sprint` form, which scopes per-node imgui widget state; pointer-kind IDs format by address (`%p`) instead, so model pointers work directly as IDs.

### Content and detents

The content closure runs with the canvas's font pushed at `base * detent`. At detent 1.0 and above — the editing range — imgui/dfx widgets behave normally (wrap them in `PushItemWidth`/`PopItemWidth`; the canvas window's default item width is meaningless inside a node). **Below 1.0, content must not emit anything interactive** — nothing ID-bearing, hoverable, or activatable — or hover capture would carve holes in the canvas's hit-testing. Declare simplified content instead: labels, values, pins. `NodeContext.Detent()` lets closures branch; `Label` is the primitive that honors the contract. Pins must be declared at every detent so links keep their anchors.

Content that opens child windows of its own (`BeginChild`-style scroll regions) is not supported.

## Intents

```go
type Intents[ID comparable] struct {
    NodesMoved       []NodeMove[ID]       // once, on drag release: {ID, From, To} per moved node
    LinkCreated      *LinkCreate[ID]      // {FromPin, ToPin}, normalized output→input
    SelectionChanged *SelectionChange[ID] // full replacement sets for nodes and links
}
```

- `NodesMoved` emits once per completed drag — one gesture, one intent, one undo command. `From` is the declared position at gesture start; `To` is `From` plus the gesture's canvas-space offset. During the drag the canvas renders declared positions plus the in-flight offset; the model stays untouched until release.
- `LinkCreated` is side-compatible by construction (the canvas knows sides, not types); the app validates semantics and applies or ignores it.
- `SelectionChanged` carries full replacement sets, emitted only when an interaction produced sets differing from the `Selected` flags declared that frame. Selection is one combined set spanning nodes and links.
- All ID slices are in declaration order — deterministic for tests and logs.

## Interaction grammar

**Left button** — click a node to select it (ctrl-click toggles, shift-click adds; ctrl wins when both are held); drag a selected node to move the whole selection; drag on empty canvas box-selects (preview drawn by the canvas, committed on release); click a link to select it; click empty canvas to clear. A plain press on an already-selected node preserves the set so a multi-drag can start from any member; the collapse to that node emits on release-without-drag instead. Drag from a pin pulls a link with snap when within the snap radius of a compatible-side pin; release on a pin emits `LinkCreated`, release elsewhere ends the gesture with no intent.

**Middle-drag** pans. **Wheel** steps the zoom detent toward the cursor — the canvas point under the mouse stays fixed. Wheel travel accumulates, and a detent steps per `Config.WheelStepsPerZoomLevel` notches (imgui wheel units; one classic notch is 1.0, default 1.0): fine-scroll devices report fractional ticks, so the same value governs every wheel type, and at most one detent steps per frame — sub-threshold remainder carries across frames, a direction change resets it, and the canvas's own banking resets whenever it could not step (not hovered, popup open, an in-canvas item hovered, or a gesture in flight), so scrolling elsewhere in the app or mid-gesture never produces a phantom step. Raise the value when the wheel feels too sensitive. **Right button** is reserved and does nothing.

A release is a release wherever the mouse is — a node dragged past the canvas edge still emits its `NodesMoved`. Genuinely lost button state (focus loss) cancels with no intent.

**Widget-first arbitration**: the canvas claims a left gesture only when no imgui item inside it is hovered or active; the wheel steps the detent only when no item is hovered; middle-drag pan needs neither. While any popup is open, the canvas initiates nothing. Item state elsewhere in the app never suppresses canvas input.

**Locked mode** (`Config.Locked` or `SetLocked`) suppresses node dragging only; selection, panning, zooming, and link creation stay live.

Hit-testing operates in canvas space through the inverse transform and works identically at every detent — hit and snap tolerances are screen-pixel constants, so targets stay equally grabbable zoomed out. Hit order matches visual z-order: topmost node first, each node's pins before its body, links only after every node misses.

## View and navigation

```go
type View struct { Pan imgui.Vec2; Zoom float32 } // plain, persistable data
```

- `View()` / `SetView(v)` — get/set view state, callable any time (no draw-frame lifecycle constraint). `SetView` snaps `Zoom` to the nearest configured detent and applies at the next `Begin`; reads are pending-first, so a write-then-read round-trips.
- `Detent()` — the current zoom factor. `GridSpacing()` — effective grid spacing for app-side snap logic.
- `CanvasFromScreen` / `ScreenFromCanvas` — the transform pair (`screen = (canvas + pan) * zoom + origin`), computed against the last begun canvas rect; keyboard action handlers (which run before component drawing) see the previous frame's rect, one frame stale and visually indistinguishable.
- `ZoomToFit(ids...)` — fits nodes (all when empty; non-node IDs ignored). Because node bounds are detent-dependent, the fit resolves over the next few frames, descending from the top detent until the content actually declared at a detent fits — bounded by the detent count. The most recent navigation wins: any other explicit navigation (wheel, `SetView`, `CenterOn`, pan) cancels a pending fit.
- `CenterOn(ids...)` — pan only, detent unchanged.
- `SetStyle(s)` — replace the style, e.g. after a theme change.

**Detents** are the only zoom levels (`Config.Detents`, default `{0.25, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1, 1.2, 1.3, 1.4, 1.5}` — 0.1 steps across the working range, one coarser step from 0.4 to the 0.25 zoom-out floor — sorted ascending). **1.0 — the editing detent — must be a member of the set; entries above it are allowed**, so the wheel can zoom in past 100% (full interactive content applies at every detent ≥ 1.0). Discrete detents bound the set of font sizes ever rasterized and make zoom levels feel like named views. A canvas opens at the editing detent (1.0), and `ZoomToFit` — which descends from the largest configured detent — can resolve above 1.0 for small graphs. Wheel sensitivity is tuned per canvas with `Config.WheelStepsPerZoomLevel` (see Interaction grammar).

## Style

`NodeCanvasStyle` is plain data: colors plus metrics, with unit spaces pinned per group — render metrics (rounding, padding, border and link thickness, pin radius, link tangent) are canvas-space and scale with zoom; hit and snap tolerances (pin hit radius, link hit distance, link snap radius) are screen-pixel constants. A zero-valued `Config.Style` derives `DefaultNodeCanvasStyle()` from the active dfx theme lazily at the first `Begin` (components are constructed before the imgui context exists). The style is a complete value, not a sparse overlay: for partial customization, start from `DefaultNodeCanvasStyle()` and mutate — in `OnSetup` or later via `SetStyle`.

## Rendering internals (for maintainers)

Everything draws through `ImDrawList` with the view transform applied canvas-side. The canvas owns an `imgui.DrawListSplitter` (never the drawlist's `ChannelsSplit`, which cannot nest with the splits imgui widgets perform internally): channel 0 for links, then a chrome/content channel pair per node, so chrome draws after content measures it but sits visually behind. Splitter capacity comes from last frame's node count plus eight spare nodes; overflow shares the final pair for one frame. In-flight previews (box select, link drag) draw after the merge, as true foreground. The scaled font push uses `imgui.CurrentStyle().FontSizeBase()` — never `imgui.FontSize()`, which is post-global-scale and would double-apply DPI scaling.

Input commits only at `End`: the gesture state machine (pure functions over an input snapshot, headless-tested in `nodeCanvasInput_test.go`) resolves against the same frame's geometry. `Begin` computes frame-local draw values for already-active gestures from absolute anchors — never accumulated deltas — so drawn and committed values agree by construction.
