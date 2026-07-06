# NodeCanvas — Work Order

Implementation plan for [node-canvas.md](node-canvas.md). The spec carries the vision, the guiding principles, and the deferred list; this document carries the code-shaped translation: concrete API, internal architecture, staging, integration points, and risks. Where the two disagree, the spec's principles win and this document is wrong.

## Decisions Recorded During Planning

These were resolved with Michael before drafting and are settled; mercurius should not re-raise them.

1. **Input arbitration cascades widget-first.** The canvas claims a left-button gesture only when no imgui item inside the canvas child is hovered or active; a hovered/active widget always wins. Wheel steps the detent only when no imgui item is hovered. Node chrome (title band, body, pin markers) and all sub-1.0 content are drawlist output, not imgui items, so gestures and zoom work everywhere except directly over a live widget at detent 1.0.
2. **The canvas retains derived geometry across the frame boundary.** Node and pin rectangles are rebuilt from declarations every frame and are never authoritative, but they persist as internal state: `End()` resolves input against the current frame's geometry, and the previous frame's geometry backs queries that arrive before this frame's declarations complete (`ZoomToFit`, `CenterOn`, hit-testing on the first frame of a gesture). This is an accepted refinement of the spec's "view and gesture state, and nothing else" — derived view-side geometry, not graph truth.
3. **`View` stores the zoom factor, not a detent index.** `View{Pan imgui.Vec2, Zoom float32}`. `SetView` snaps `Zoom` to the nearest configured detent. Persisted views survive changes to the `Detents` config without silently meaning a different zoom.
4. **The public API uses `imgui.Vec2`.** Consistent with the rest of dfx (`State.Size`, `FaderParams`); it is the "Vec2-equivalent" the spec's substrate seam allows.
5. **`Begin`/`End` over a closure-taking entry.** `Node` already takes closures, links are declared between the two calls, and the two-call shape matches both the imgui idiom and dfx draw paths.

## Deliverables

New files, all in `package dfx` at the repo root per convention:

| File | Contents |
|---|---|
| `nodeCanvasGeometry.go` | Pure geometry core: transform, detent math, bezier distance, rect ops. No imgui calls (imgui.Vec2 as a plain struct only). |
| `nodeCanvasInput.go` | Gesture state machine as pure functions over a sampled input snapshot. No imgui calls. |
| `nodeCanvas.go` | `NodeCanvas[ID]`, config, `Begin`/`Node`/`Link`/`End`, `NodeContext`, intents, view/navigation API. All imgui rendering and input sampling lives here. |
| `nodeCanvasStyle.go` | `NodeCanvasStyle` plain-data struct and `DefaultNodeCanvasStyle()` theme derivation. |
| `nodeCanvasGeometry_test.go`, `nodeCanvasInput_test.go` | Headless unit tests (display-free, per the existing `*_test.go` pattern). |
| `examples/dfx_example_nodecanvas/main.go` | The acceptance example from the spec's Acceptance section. |
| `docs/current/node-canvas.md` | Built-behavior documentation, written as implementation lands. |

Modified files: `CHANGELOG.md` (`## Unreleased` slot, `FEATURE` entry), `CLAUDE.md` (one line adding NodeCanvas to the architecture list). No changes to any existing `.go` file are expected.

## Public API

Signatures are the contract for implementation; internal names are illustrative.

```go
type NodeCanvas[ID comparable] struct{ /* view + gesture + derived geometry */ }

func NewNodeCanvas[ID comparable](cfg NodeCanvasConfig) *NodeCanvas[ID]

type NodeCanvasConfig struct {
    Detents     []float32       // default {0.25, 0.5, 0.75, 1.0}; sorted ascending; last entry is max and must be 1.0
    GridSpacing float32         // default 50 canvas units; exposed for app-side snap logic
    Locked      bool            // suppress node dragging only
    Style       NodeCanvasStyle // zero value means DefaultNodeCanvasStyle() at construction
}

func (nc *NodeCanvas[ID]) Begin(state *State)
func (nc *NodeCanvas[ID]) Node(id ID, pos imgui.Vec2, flags NodeFlags, content func(n *NodeContext[ID]))
func (nc *NodeCanvas[ID]) Link(id, fromPin, toPin ID, flags LinkFlags)
func (nc *NodeCanvas[ID]) End() Intents[ID]

type NodeFlags struct{ Selected bool }
type LinkFlags struct{ Selected bool }

type NodeContext[ID comparable] struct{ /* per-node draw state */ }
func (n *NodeContext[ID]) TitleBar(content func())
func (n *NodeContext[ID]) Input(pinID ID, label string)
func (n *NodeContext[ID]) Output(pinID ID, label string)
func (n *NodeContext[ID]) Detent() float32

type Intents[ID comparable] struct {
    NodesMoved       []NodeMove[ID]
    LinkCreated      *LinkCreate[ID]
    SelectionChanged *SelectionChange[ID]
}
type NodeMove[ID comparable] struct{ ID ID; From, To imgui.Vec2 }
type LinkCreate[ID comparable] struct{ FromPin, ToPin ID }
type SelectionChange[ID comparable] struct{ Nodes, Links []ID }

type View struct{ Pan imgui.Vec2; Zoom float32 }
func (nc *NodeCanvas[ID]) View() View
func (nc *NodeCanvas[ID]) SetView(v View) // Zoom snapped to nearest configured detent; effective immediately
func (nc *NodeCanvas[ID]) Detent() float32
func (nc *NodeCanvas[ID]) CanvasFromScreen(p imgui.Vec2) imgui.Vec2
func (nc *NodeCanvas[ID]) ScreenFromCanvas(p imgui.Vec2) imgui.Vec2
func (nc *NodeCanvas[ID]) ZoomToFit(ids ...ID) // all nodes when empty; no-op before the first drawn frame
func (nc *NodeCanvas[ID]) CenterOn(ids ...ID)  // pan only; no-op before the first drawn frame
```

Notes pinned here so they don't get re-decided downstream:

- **One ID space.** Node, pin, and link IDs share the type and the app guarantees uniqueness across all three populations. Document this on `NodeCanvas`; the canvas does not police it.
- **NodeCanvas is a widget, not a `Component`.** Like `FaderN`, it is driven from inside an owning component's `Draw`. It has no `Actions()` registry; keyboard behavior (delete-selection, zoom-to-fit bindings) is app-side through the existing action system, reading the app's own selection. No delete intent exists — the app owns selection and deletion outright.
- **`SelectionChanged` semantics.** Full replacement sets, emitted only on frames where a completed interaction produced sets differing from the declared `Selected` flags. The canvas compares against what the app declared this frame, not against its own memory of past selections.
- **`NodesMoved` semantics.** Emitted once, on drag release, one entry per selected node that moved: `From` is the declared position at gesture start, `To` is `From` plus the gesture's canvas-space offset. During the drag the canvas renders declared positions plus the in-flight offset; the app's model is untouched until the intent lands.
- **imgui ID scoping.** Each node's content runs inside `imgui.PushIDStr(fmt.Sprint(id))` so widget state (active `InputText`, etc.) stays stable across frames regardless of declaration-order changes. `fmt.Sprint` per node per frame is acceptable at tool-scale graph sizes.

## Internal Architecture

### Frame lifecycle

`Begin(state)`:

1. Open the canvas child region: `imgui.BeginChildStrV` sized to `state.Size`, flags `NoScrollbar|NoScrollWithMouse|NoMove`, following the pattern at `dash.go:55` and `multiGrid.go:318`. Record the child's screen origin — it is the `origin` in the view transform `screen = (canvas + pan) * zoom + origin`.
2. Draw the grid directly into the window drawlist (pre-splitter, so it sits under everything).
3. Split the drawlist: channel 0 for links, then two channels per node (chrome, content). Channel count is not knowable up front, so `Split` uses last frame's node count plus headroom (+8); nodes declared beyond capacity share the final channel pair for that frame — a one-frame z-order artifact only when many nodes appear at once, logged at debug level via `dl`.
4. Push the scaled font: capture `imgui.CurrentFont()` and `imgui.FontSize()` at entry, then `imgui.PushFont(font, size * zoom)` — the two-arg dynamic-size form already used at `fonts.go:92`. At detent 1.0 this is a visual no-op and all widgets behave normally.
5. Reset per-frame declaration buffers (nodes, links, pin anchors).

`Node(id, pos, flags, content)`:

1. Select the node's content channel, `imgui.SetCursorScreenPos(ScreenFromCanvas(pos + inFlightOffset))` — the offset applies only to nodes in the active drag set.
2. Run the closure inside `PushIDStr` + `BeginGroup`/`EndGroup`. `TitleBar` runs its closure and records the title region height; `Input`/`Output` render marker-less label rows in flow and record each pin's row Y-anchor and side. Content is measured by the group rect.
3. Select the node's chrome channel and draw body (rounded filled rect + border, padded around the group rect), title band (tinted band across the final node width), and pin markers (circles at the node's left/right edge X, each at its recorded row Y). Chrome draws after content measures it, but the channel split puts it visually behind.
4. Record the node's canvas-space rect and pin positions into this frame's geometry map.

`Link(id, fromPin, toPin, flags)` only records the declaration — pins may belong to nodes declared later, so links resolve and draw in `End`.

`End()`:

1. Resolve hover (topmost-first: pins, then nodes by reverse declaration order, then links by bezier distance) against this frame's geometry via the inverse transform.
2. Draw links into channel 0: horizontal-tangent cubic beziers (`AddBezierCubic`, control points offset horizontally by a style-scaled distance), with hover/selected emphasis. Draw the in-flight link-drag preview and box-select preview into the foreground (post-merge, directly into the drawlist).
3. Merge the splitter, pop the font.
4. Sample input into a snapshot struct and run the gesture state machine (below). View mutations (pan, detent step) apply immediately to `nc.view`; completed gestures produce the returned `Intents`.
5. Swap this frame's geometry map into the retained slot, `imgui.EndChild()`, return intents.

Input resolves at `End` against same-frame geometry, so there is no gesture-vs-geometry lag; the one-frame visual lag between an intent and the app's re-declaration is the normal immediate-mode cadence.

### Gesture state machine (`nodeCanvasInput.go`)

Pure functions: `(gestureState, inputSnapshot, geometry) → (gestureState', viewOps, intents)`. The snapshot carries mouse position, button transitions, drag deltas past imgui's drag threshold, wheel, modifiers, hovered-item flags (`IsAnyItemHovered`/`IsAnyItemActive` sampled inside the child), and window-hovered state. Because the machine never touches imgui, press-drag-release sequences are unit-testable headless.

States: `idle`, `dragNodes`, `boxSelect`, `linkDrag`, `pan`. One gesture at a time. Transitions implement the spec's interaction grammar exactly, with these pinned details:

- **Press routing (left button, no imgui item hovered):** pin hit → arm `linkDrag`; node hit → selection updates on press (plain: replace with `{node}` if unselected; ctrl: toggle; shift: add), then arm `dragNodes`; link hit → selection update on press; empty → arm `boxSelect`.
- **Click vs drag** is imgui's drag threshold: release before threshold is a click (selection already applied on press); crossing it starts the armed gesture. `Locked` blocks only the `dragNodes` arming — press-selection still applies.
- **Box select** previews via canvas-drawn rect, commits a replacement `SelectionChanged` (nodes and links whose geometry intersects the canvas-space rect) on release. Modifier-combining box select is deferred with the spec's deferred list.
- **Link drag** requires side-compatibility only (output→input or input→output). Snap: when the cursor is within `Style.LinkSnapRadius` (screen px) of a compatible pin, the preview locks to it; release while snapped emits `LinkCreated` (normalized to `{FromPin: output, ToPin: input}`), release elsewhere cancels.
- **Middle-drag** pans; **wheel** steps one detent per event sign, zooming toward the cursor: with the cursor's canvas point `c` fixed, `pan' = (mouseScreen − origin)/zoom' − c`.
- Gestures cancel cleanly if the mouse is released outside the child or the button state is lost; no intent is emitted on cancel.

### Geometry core (`nodeCanvasGeometry.go`)

Plain functions and small structs, no imgui context: the view transform pair; detent stepping and nearest-detent snapping; zoom-toward-point pan adjustment; `ZoomToFit` (bounds of the retained rects, largest detent that fits with margin, centered); `CenterOn` (bounds center to view center at current detent); point-to-cubic-bezier distance (fixed subdivision, ~24 segments, against `Style.LinkHitDistance`); rect intersection for box select; pin hit circles (`Style.PinHitRadius`, screen-space constant so pins stay grabbable at 0.25).

### Style (`nodeCanvasStyle.go`)

`NodeCanvasStyle` is plain data: colors as `imgui.Vec4` (grid, node body, node border, node border selected/hovered, title band, title band selected, pin, pin hovered, link, link selected/hovered, box-select fill/border) and metrics as `float32` (node rounding, node padding, border thickness normal/selected, pin radius, pin hit radius, link thickness, link hit distance, link snap radius, link tangent distance). `DefaultNodeCanvasStyle()` derives colors from `imgui.CurrentStyle().Colors()` the way `HueColorScheme.Apply` reads them (`theme.go:60`) — e.g. node body from `ColChildBg`/`ColPopupBg`, title from `ColTitleBgActive`, borders from `ColBorder`, selection emphasis from `ColHeaderActive` — replacing baab's imgui→imnodes bridge as the spec directs. Metric starting points: pin radius 5, pin hit radius 10, link snap radius 24, link hit distance 6, link thickness 2 — all expected to be tuned against the example, which is the spec's stated resolution for these open questions. Exact grid style per detent (line weight/alpha fade at low detents) is likewise tuned in the example.

## Staging

Five stages, each leaving `make test` green and each terminus-gated before Michael reviews, per the pipeline. Stages 2–5 also keep the example runnable at whatever scope exists so far — the example grows with the component rather than arriving at the end.

**Stage 1 — pure core.** `nodeCanvasGeometry.go`, `nodeCanvasInput.go` (state machine over synthetic geometry), and their tests. No rendering. This is where transform math, detent snapping, zoom-toward-cursor, bezier distance, and the full gesture grammar (press/click/ctrl/shift/drag/box/link-snap/cancel, locked mode) get locked in headlessly.

**Stage 2 — canvas shell.** `nodeCanvas.go` skeleton: config defaulting, `Begin`/`End` with child region, grid, font push, middle-drag pan, wheel detent zoom, `View`/`SetView`/`Detent`/transform helpers. Example opens an empty pannable, zoomable grid.

**Stage 3 — declaration and rendering.** `Node`/`NodeContext`/`Link` drawing: splitter channels, title bar, content flow, pin rows and markers, node chrome, bezier links, retained geometry. `nodeCanvasStyle.go` and theme derivation land here. Example declares the synthetic graph (several node kinds, icon-bearing titles, widget-bearing and label-only content, several links) — static, no interaction yet. Icon glyph-offset behavior at each detent is checked here (risk 1 below).

**Stage 4 — interaction.** Wire the stage-1 state machine to real input sampling: hover, selection intents, node dragging with in-flight offset rendering and single `NodesMoved` on release, box select, link creation with snap, locked mode, the widget-first arbitration rule. Example handles intents against its in-memory model, logs them via `dl`, and runs `NodesMoved` through a dfx `UndoSystem` command to demonstrate one-gesture-one-intent-one-command.

**Stage 5 — navigation and closure.** `ZoomToFit`, `CenterOn`, per-detent content simplification in the example (branching on `n.Detent()`), view get/set round-trip with simulated persistence, style/grid tuning pass, `docs/current/node-canvas.md`, CHANGELOG entry, CLAUDE.md line. The spec's acceptance paragraph is the checklist for this stage's example walkthrough.

## Risks and Watch Items

1. **Icon glyph offset below detent 1.0** (planning decision 5). The Material Icons merge uses fixed pixel `GlyphOffset` values tuned per size (`fonts.go:48,75`); under dynamic scaling the offset may not track the pushed size, drifting icon baselines at 0.25–0.75. Surface it in stage 3 with icon-bearing titles at every detent; the fallback is a per-detent Y correction applied in title-bar rendering only. Does not block any other stage.
2. **Splitter capacity growth.** `ChannelsSplit`-style splitters cannot re-split mid-frame; the prev-count+headroom scheme degrades to shared channels only on frames where many nodes appear at once, and only as a z-order artifact. Accepted; noted here so it isn't "fixed" into something more elaborate.
3. **Wheel-over-widget blocks zoom at detent 1.0** — a direct consequence of the sanctioned widget-first cascade. Predictable and small (only while the cursor is directly over a live widget); accepted, not a bug.
4. **Text fidelity at 0.25.** imgui 1.92 bakes per-size glyphs, so quality should be fine and atlas growth is bounded by the detent set (the spec's design bet). If 0.25 text looks poor with the default fonts, that's an example-tuning concern (style/grid pass, stage 5), not an architecture change.
5. **Overlapping-node widget interaction.** imgui item overlap between two nodes' live widgets follows imgui's own rules; per-node channel pairs keep the visuals correct, and interaction correctness for heavily overlapped nodes is accepted as-is for v1.

## Out of Scope

Everything in the spec's Deferred list (minimap, link detach, subgraph clipboard, sub-1.0 widget interaction, continuous zoom, free-form pins, context menus), plus: no `Component` wrapper, no canvas-side actions, no delete intent, no modifier-combining box select. The baab migration is a separate follow-on work order in baab's repo; this work completes at the spec's acceptance gate — the example feeling right under the hand.
