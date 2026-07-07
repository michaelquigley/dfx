# NodeCanvas — Work Order

Implementation plan for [node-canvas.md](node-canvas.md). The spec carries the vision, the guiding principles, and the deferred list; this document carries the code-shaped translation: concrete API, internal architecture, staging, integration points, and risks. Where the two disagree, the spec's principles win and this document is wrong.

## Decisions Recorded During Planning

These were resolved with Michael before drafting and are settled; mercurius should not re-raise them.

1. **Input arbitration cascades widget-first.** The canvas claims a left-button gesture only when no imgui item inside the canvas child is hovered or active; a hovered/active widget always wins. Wheel steps the detent only when no imgui item is hovered. Node chrome (title band, body, pin markers) and all sub-1.0 content are drawlist output, never interactive imgui items, so gestures and zoom work everywhere except directly over a live widget at detent 1.0.
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

Modified files: `CHANGELOG.md` (`## Unreleased` slot, `FEATURE` entry), `AGENTS.md` (one line adding NodeCanvas to the architecture list; `CLAUDE.md` is a symlink to it, so the change lands in both names). No changes to any existing `.go` file are expected.

## Public API

Signatures are the contract for implementation; internal names are illustrative.

```go
type NodeCanvas[ID comparable] struct{ /* view + gesture + derived geometry */ }

func NewNodeCanvas[ID comparable](cfg NodeCanvasConfig) *NodeCanvas[ID]

type NodeCanvasConfig struct {
    Detents     []float32       // default {0.25, 0.5, 0.75, 1.0}; sorted ascending; last entry is max and must be 1.0
    GridSpacing float32         // default 50 canvas units; exposed for app-side snap logic
    Locked      bool            // suppress node dragging only
    Style       NodeCanvasStyle // zero value: derived lazily via DefaultNodeCanvasStyle() at first Begin
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
func (n *NodeContext[ID]) Label(text string) // drawlist text + ID-less Dummy spacer: measures, invisible to hover/active arbitration; safe at every detent
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
func (nc *NodeCanvas[ID]) SetView(v View) // Zoom snapped to nearest configured detent; callable any time (no context-lifecycle constraint), applied at the next Begin
func (nc *NodeCanvas[ID]) Detent() float32
func (nc *NodeCanvas[ID]) GridSpacing() float32 // effective value, including the applied default
func (nc *NodeCanvas[ID]) CanvasFromScreen(p imgui.Vec2) imgui.Vec2
func (nc *NodeCanvas[ID]) ScreenFromCanvas(p imgui.Vec2) imgui.Vec2
func (nc *NodeCanvas[ID]) ZoomToFit(ids ...ID) // len(ids)==0: all nodes; non-node IDs ignored; len(ids)>0 filtering to zero node IDs: no-op; no-op before the first drawn frame
func (nc *NodeCanvas[ID]) CenterOn(ids ...ID)  // pan only; same argument contract as ZoomToFit
func (nc *NodeCanvas[ID]) SetStyle(s NodeCanvasStyle) // replace style, e.g. on theme change (spec: style seam crossed at construction or rebuild)
func (nc *NodeCanvas[ID]) SetLocked(locked bool)      // live toggle; gates only dragNodes arming, like the config flag
```

Notes pinned here so they don't get re-decided downstream:

- **One ID space.** Node, pin, and link IDs share the type and the app guarantees uniqueness across all three populations. Document this on `NodeCanvas`; the canvas does not police it.
- **View mutators are frame-boundary calls.** `SetView`, `ZoomToFit`, and `CenterOn` are intended outside the `Begin`/`End` window (action handlers, `OnTick`, intent handling); a call made mid-frame — say from a button inside node content — takes effect at the next frame's `Begin`, never mid-declaration, so a frame's transform is stable across everything it draws. The natural implementation follows: mutators write pending view state, `Begin` applies it. Read-after-write is pending-first: `View()` returns the pending view when one exists, else the applied view — so a persistence path that calls `SetView` then `View()` round-trips correctly — while the *draw* transform remains frame-stable regardless.
- **Transform helpers use the last begun canvas rect.** `CanvasFromScreen`/`ScreenFromCanvas` need the child's screen origin, which is recorded at each `Begin` and retained. This matters because dfx dispatches keyboard actions before component drawing (`app.go:171`), so an action handler calling a helper — baab's fetch-node does exactly this — computes against the previous frame's rect: one frame stale, visually indistinguishable, and correct as long as the implementation retains the rect rather than assuming in-draw calls only. Before the first `Begin`, the helpers apply the view transform with a zero origin. The helpers follow `View()`'s pending-first rule: after `SetView` (or any pending navigation), they compute against the pending view — a handler that writes a view and then transforms a point sees its own write, consistent with `View()`.
- **NodeCanvas is a widget, not a `Component`.** Like `FaderN`, it is driven from inside an owning component's `Draw`. It has no `Actions()` registry; keyboard behavior (delete-selection, zoom-to-fit bindings) is app-side through the existing action system, reading the app's own selection. No delete intent exists — the app owns selection and deletion outright.
- **`SelectionChanged` semantics.** Full replacement sets, emitted only on frames where an interaction produced sets differing from the declared `Selected` flags (timing — press vs release — per the gesture section's normative rules). The canvas compares against what the app declared this frame, not against its own memory of past selections. Selection is one combined set that spans both populations — plain gestures replace the whole set, modified gestures edit single elements within it; the transition table in the gesture state machine section is normative. `SelectionChange.Nodes`, `SelectionChange.Links`, and `NodesMoved` entries are emitted in declaration order — not semantically meaningful, but deterministic for tests, logs, and app-side comparisons.
- **`NodesMoved` semantics.** Emitted once, on drag release, one entry per selected node that moved: `From` is the declared position at gesture start, `To` is `From` plus the gesture's canvas-space offset. During the drag the canvas renders declared positions plus the in-flight offset; the app's model is untouched until the intent lands.
- **imgui ID scoping.** Each node's content runs inside `imgui.PushIDStr(scope(id))` so widget state (active `InputText`, etc.) stays stable across frames regardless of declaration-order changes. `scope` special-cases pointers: a pointer-kind ID (one `reflect.ValueOf(id).Kind()` check) formats as `fmt.Sprintf("%p", id)` — the address, stable and unique — because plain `fmt.Sprint` renders a struct pointer as `&{field values}`, which mutates when the model mutates and collides when two models compare equal. Non-pointer IDs use `fmt.Sprint(id)` and must be unique *and stable across frames* in that form — formatting that collapses distinct values silently shares widget state between nodes; formatting that changes frame-to-frame silently resets it. The per-node-per-frame reflection and formatting cost is acceptable at tool-scale graph sizes. Document alongside the uniqueness requirement on `NodeCanvas`.
- **The reduced-detent content contract.** Below detent 1.0, content closures must not emit *interactive* imgui items — nothing ID-bearing, hoverable, or activatable. The widget-first arbitration cascade (planning decision 1) depends on this: a hoverable item in a reduced-detent node would capture hover and suppress node selection, dragging, and wheel zoom over its rect, contradicting the spec's promise that hit-testing works identically at every detent. Layout, however, requires items: imgui group bounds grow only when items are submitted, so pure drawlist output would measure a node as zero. `NodeContext.Label` therefore draws its text via the drawlist and submits an `imgui.Dummy` of the text size to advance layout — `Dummy` submits no ID and does not affect the canvas's `IsAnyItemHovered`/active-item arbitration flags (the precise property the contract needs; note that rect-based queries like `IsItemHovered` *can* report hover over a `Dummy`, so arbitration must stay on the ID-based flags), keeping group measurement correct while staying invisible to gesture routing. `Input`/`Output` pin rows render their labels through the same mechanism. The example's sub-1.0 branches use `Label` for all non-pin content — with `Input`/`Output` rows still declared at every detent per the pin-presence invariant — making the correct pattern the visible one.
- **`ZoomToFit` resolves over frames, not instantly.** Node bounds are detent-dependent (apps simplify content below 1.0), so a fit computed from zoomed-out geometry can overshoot: bounds measured at 0.25 may claim everything fits at 1.0, where full-detail content no longer does. `ZoomToFit` therefore records a pending fit operation in view state and walks down from the top: the calling frame sets the largest configured detent; each subsequent `End()` measures the bounds actually declared at the current detent and either finishes (they fit — center and clear the pending op) or steps down one detent and stays pending. The first detent that fits its own declared content is, by construction, the largest such detent — the promise satisfied by definition, with no candidate arithmetic, no visited set, and no possible cycle. Bounded by `len(Detents)` frames (four at the default set, imperceptible at frame rate); if even the smallest detent cannot fit the bounds, it settles there, centered. Ownership of the pending fit follows one rule — the most recent navigation wins: a new `ZoomToFit` replaces any pending fit, and every other explicit navigation (wheel detent step, `SetView`, `CenterOn`, middle-drag pan) cancels the pending fit before applying its own change, so a half-finished fit never clobbers newer intent. `CenterOn` changes no detent and needs none of this machinery beyond that cancellation.

## Internal Architecture

### Frame lifecycle

`Begin(state)`:

1. Open the canvas child region: `imgui.BeginChildStrV` sized to `state.Size`, flags `NoScrollbar|NoScrollWithMouse|NoMove`, following the pattern at `dash.go:55` and `multiGrid.go:318`. Record the child's screen origin — it is the `origin` in the view transform `screen = (canvas + pan) * zoom + origin`.
2. Draw the grid directly into the window drawlist (pre-splitter, so it sits under everything).
3. Split the drawlist: channel 0 for links, then two channels per node (chrome, content). The split MUST use a persistent `imgui.DrawListSplitter` owned by the canvas (`Split`/`SetCurrentChannel`/`Merge`), never the drawlist's own `ChannelsSplit`/`ChannelsMerge` convenience API — imgui widgets (tables in particular) split the drawlist internally, and the convenience path shares channel state and cannot nest; an owned splitter composes with whatever node content does inside its channel. Channel count is not knowable up front, so capacity comes from last frame's node count with node headroom: `channelCount = 1 + 2*(lastFrameNodeCount + 8)` — the `+8` is spare *nodes*, not spare channels. Nodes declared beyond capacity share the final channel pair for that frame — a one-frame z-order artifact only when more than eight nodes appear at once, logged at debug level via `dl`.
4. Push the scaled font: capture `imgui.CurrentFont()` and `imgui.CurrentStyle().FontSizeBase()` at entry, then `imgui.PushFont(font, baseSize * zoom)`. `PushFont`'s size parameter is the *unscaled* base size (`font_size_base_unscaled`); `imgui.FontSize()` returns the post-global-scale value and must never be passed back in, or DPI/global scaling double-applies. This matches dfx's existing pattern at `fonts.go:92`, which passes configured base sizes. At detent 1.0 this is a visual no-op and all widgets behave normally.
5. Reset per-frame declaration buffers (nodes, links, pin anchors).

`Node(id, pos, flags, content)`:

1. `pos` is the **outer node rect's top-left, in canvas space** — the app-owned anchor. It is style-independent: padding and title metrics offset content *inward* from `pos`, never the card away from it, so theme changes restyle the card without moving persisted layouts, and content growth extends the rect right/down while the anchor stays fixed. `NodeMove.From`/`To` use this same anchor. Select the node's content channel, then `imgui.SetCursorScreenPos(ScreenFromCanvas(pos + inFlightOffset + padding))` — padding is canvas-space (style rules), so it rides inside the transform; padding only, the title band needs no advance height (next step). The offset applies only to nodes in the active drag set.
2. Run the closure inside `PushIDStr` + `BeginGroup`/`EndGroup`. Layout is one pass, in flow: `TitleBar`, when used, must be the *first* call in the content closure (debug-log a violation) — its closure runs and is measured in flow like everything else, and its measured extent *is* the recorded title-band height, so no fixed title metric exists and nothing needs the height in advance. Body content continues below it in the same flow; `Input`/`Output` render marker-less label rows and record each pin's row Y-anchor and side. Content is measured by the group rect.
3. Select the node's chrome channel and draw body (rounded filled rect + border, anchored at `pos` and sized from the group rect plus insets), title band (tinted band drawn behind `pos` through the measured title bottom, across the final node width), and pin markers (circles at the node's left/right edge X, each at its recorded row Y). Chrome draws after content measures it, but the channel split puts it visually behind.
4. Record the node's canvas-space rect — anchored at `pos` — and pin positions into this frame's geometry map.

`Link(id, fromPin, toPin, flags)` only records the declaration — pins may belong to nodes declared later, so links resolve and draw in `End`. **Pin presence is a per-frame invariant:** anything that should render, hit-test, snap, or anchor a link at the current detent must be declared in that frame — pins included, at every detent. A link whose endpoint pin was not declared this frame is skipped for rendering and hit-testing, logged at debug level via `dl`; it is never drawn from retained geometry, which is derived view-side state and must not get promoted to graph truth by a declaration gap.

`End()`:

1. Resolve hover against this frame's geometry via the inverse transform, per node in reverse declaration order (topmost first): test each node's own pins, then that same node's body; first hit wins. Only after every node misses do links get tested by bezier distance. Pins outrank their own node's body (markers on the edge stay grabbable) but never reach through a node covering them — hit order must match visual z-order.
2. Draw links into channel 0: horizontal-tangent cubic beziers (`AddBezierCubic`, control points offset horizontally by a style-scaled distance), with hover/selected emphasis.
3. Merge the splitter, then draw the in-flight link-drag preview and box-select preview directly into the drawlist — true foreground, above all merged channels. Pop the font.
4. Sample input into a snapshot struct and run the gesture state machine (below). View mutations (pan, detent step) apply immediately to `nc.view`; completed gestures produce the returned `Intents`.
5. Swap this frame's geometry map into the retained slot, `imgui.EndChild()`, return intents.

Input resolves at `End` against same-frame geometry, so there is no gesture-vs-geometry lag; the one-frame visual lag between an intent and the app's re-declaration is the normal immediate-mode cadence. One ordering rule so in-flight visuals don't trail the mouse, with single ownership so nothing double-applies: `End` is the *only* phase that commits — state transitions, view mutations, and intents all happen there. `Begin` computes **frame-local draw values only** for already-active gestures (effective pan, in-flight drag offset, preview endpoint), and derives them from the *absolute* mouse position against the gesture's anchor — never by accumulating per-frame deltas. Anchor-relative math is idempotent, so `Begin`'s drawn values and `End`'s committed values agree by construction: latency for what's moving, correctness for what's starting, and no path where the same input applies twice.

### Gesture state machine (`nodeCanvasInput.go`)

Pure functions: `(gestureState, inputSnapshot, geometry) → (gestureState', viewOps, intents)`. The snapshot carries mouse position, button transitions, drag deltas past imgui's drag threshold, wheel, modifiers, window-hovered state, and three canvas-scoped arbitration flags. The flags are *not* the raw global queries — `IsAnyItemHovered`/`IsAnyItemActive` report app-wide state, and a text field left active in a sidebar would otherwise suppress every canvas gesture. Instead:

"In canvas" means the canvas child window *plus its descendant windows*. Child-window-producing node content (`BeginChild` inside a content closure — a scrollable region) is **not supported in v1**: child windows carry their own drawlists, so they escape the per-node splitter and would break the visual/hit z-order invariant, and no present client declares one. The descendant-window arbitration checks below are still specified — they are cheap, they defend against accidental violations, and they become load-bearing if child-window content is ever supported. Popups are handled by a separate blanket rule below, because imgui popups (a combo's dropdown, an app's floating panel) are top-level windows with no child relationship to the canvas. Concretely (all present in cimgui-go v1.5.0):

- `itemHoveredInCanvas` — global `IsAnyItemHovered()` gated on `IsWindowHoveredV(HoveredFlagsChildWindows)` of the canvas child, so a hovered item inside an inner child window still counts. Within that gate the only hoverable items are canvas content items (chrome and reduced-detent content emit none).
- `itemActiveInCanvas` — the OR of `imgui.IsItemActive()` sampled immediately after each node's content group (`EndGroup` propagates item status flags within the same window), OR'd with a window-hierarchy check: the active item counts when its owning window (context `ActiveIdWindow`) is the canvas child or a descendant of it (`InternalIsWindowChildOf`). Widgets active elsewhere in the app never set it.
- `anyPopupOpen` — `imgui.IsPopupOpenStrV("", PopupFlagsAnyPopupId|PopupFlagsAnyPopupLevel)`. While *any* popup is open, the canvas initiates no new gesture and steps no detent; in-flight gestures run to completion. No ownership test: it doesn't matter whose popup it is, because interacting with any popup means the user is not gesturing at the canvas. This deliberately covers both directions — a combo dropdown opened from node content, and an app-level popup floated over the canvas (baab's INSERT-driven node-registry popup, with its searchable type list, is exactly this shape; an ownership test scoped to node content would miss it and the canvas would eat clicks aimed at the popup).

Because the machine never touches imgui, press-drag-release sequences are unit-testable headless.

States: `idle`, `dragNodes`, `boxSelect`, `linkDrag`, `pan`. One gesture at a time. Transitions implement the spec's interaction grammar exactly, with these pinned details:

- **The idle gate.** imgui mouse state is global, so initiation is normatively gated: while `idle`, left press, middle press, and wheel are ignored unless the canvas child (or an allowed descendant) is hovered — the same `IsWindowHoveredV(HoveredFlagsChildWindows)` test the hover flag uses — and `anyPopupOpen` is false. Without this gate a sidebar click would clear the canvas selection and a sidebar wheel would step the detent. In-flight gestures are exempt by design: they ignore hover and complete on release, per the release-is-a-release rule below.
- **The full initiation predicates** — decision 1's arbitration cascade composed per input, so no bullet has to be cross-read against another: left-button initiation requires canvas-hovered ∧ `!anyPopupOpen` ∧ `!itemHoveredInCanvas` ∧ `!itemActiveInCanvas`; a wheel detent step requires canvas-hovered ∧ `!anyPopupOpen` ∧ `!itemHoveredInCanvas`; middle-drag pan requires only canvas-hovered ∧ `!anyPopupOpen`. These are the normative truth tables; the bullets below assume them.

- **Press routing (left button, initiation predicates satisfied):** pin hit → arm `linkDrag`; node hit → selection transition (table below), then arm `dragNodes` only if the pressed node is selected in the *post-transition* set — ctrl-toggling a node off is a selection-only interaction, and crossing the drag threshold afterward does nothing; link hit → selection transition; empty → arm `boxSelect`.
- **Selection transitions.** Selection is one combined set spanning nodes and links: plain gestures replace the whole set, modified gestures edit single elements. The table below is normative; the timing rules follow it. Modifier normalization: when Ctrl and Shift are both held, Ctrl wins (toggle beats add, matching the table's column order); Alt and Super do not participate in canvas mouse semantics and pass through untouched, leaving them free for app-side meanings.

| Gesture | Plain | Ctrl | Shift |
|---|---|---|---|
| Click node N | nodes = {N}, links = {} | toggle N; all else kept | add N; all else kept |
| Click link L | links = {L}, nodes = {} | toggle L; all else kept | add L; all else kept |
| Click empty | both cleared | both cleared | both cleared |
| Box select | both replaced by intersecting elements | — deferred — | — deferred — |

Node and link selection intents emit on press (which is also what arms a drag). Presses on *empty canvas* emit nothing: they arm a pending gesture, and exactly one of three outcomes follows — release before the drag threshold emits the clear (it was a click); drag then release emits the box-select replacement; cancel emits nothing. This keeps a canceled or completed box select from also firing a spurious clear. One further exception keeps the table coherent with "drag a selected node to move the whole selection": a plain press on an *already-selected* node preserves the current set (so the whole selection can be dragged from any member) and the collapse to `{N}` emits on release-without-drag instead.

- **Click vs drag** is imgui's drag threshold: release before threshold is a click (selection already applied per the table); crossing it starts the armed gesture. `Locked` blocks only the `dragNodes` arming — press-selection still applies. Locked mode also bypasses the already-selected-node exception: that delay exists solely to keep a multi-drag startable, so with dragging impossible the plain-click table applies on press, collapse included. No pending state survives a locked press.
- **Box select** previews via canvas-drawn rect, commits a replacement `SelectionChanged` on release. Inclusion tests are normative: a node is in when its rect intersects the selection rect; a link is in when any segment of the same fixed cubic subdivision used for link hit-testing (~24 segments) intersects the selection rect — endpoint-inside falls out of segment intersection for free. Modifier-combining box select is deferred with the spec's deferred list.
- **Link drag** requires side-compatibility only (output→input or input→output). Snap: when the cursor is within `Style.LinkSnapRadius` (screen px) of a compatible pin, the preview locks to it; release while snapped emits `LinkCreated` (normalized to `{FromPin: output, ToPin: input}`), release elsewhere ends the gesture with no intent.
- **Candidate arbitration is deterministic: nearest wins.** When several pins fall within the snap radius (likely at low detents, where tolerances are screen-pixel constants), snap chooses the nearest compatible pin in screen space; when several links fall within `LinkHitDistance`, hover/click chooses the closest by bezier distance. Exact ties break by reverse declaration order — topmost, consistent with node hit-testing. The geometry snapshot carries declaration indexes for this; no resolution may depend on Go map iteration order.
- **Middle-drag** pans; **wheel** steps one detent per event sign, zooming toward the cursor: with the cursor's canvas point `c` fixed, `pan' = (mouseScreen − origin)/zoom' − c`.
- A release is a release wherever the mouse is: gestures complete on button release using the last sampled position, even outside the child — a node dragged past the canvas edge still emits its `NodesMoved`, a box select finished off-canvas still commits. Cancellation is reserved for genuinely lost button state (focus loss, button state vanishing without a release event) and emits no intent. Link drag needs no special case: its grammar already makes release anywhere off a compatible pin a no-intent end.

### Geometry core (`nodeCanvasGeometry.go`)

Plain functions and small structs, no imgui context: the view transform pair; detent stepping and nearest-detent snapping; zoom-toward-point pan adjustment; the `ZoomToFit` fit test (do the retained rects fit the viewport at a given detent, with margin, centered — driven by the descend-from-top pending-fit loop described in the API notes); `CenterOn` (bounds center to view center at current detent); point-to-cubic-bezier distance (fixed subdivision, ~24 segments, against `Style.LinkHitDistance` — a screen-pixel tolerance like `PinHitRadius` and `LinkSnapRadius`, divided by the current zoom before the canvas-space comparison so links stay equally clickable at every detent); rect intersection for box select (node rects rect-vs-rect; links by segment-vs-rect over the shared cubic subdivision); pin hit circles (`Style.PinHitRadius`, screen-space constant so pins stay grabbable at 0.25).

### Style (`nodeCanvasStyle.go`)

`NodeCanvasStyle` is plain data: colors as `imgui.Vec4` (grid, node body, node border, node border selected/hovered, title band, title band selected, pin, pin hovered, link, link selected/hovered, box-select fill/border) and metrics as `float32`, each with its unit space pinned: render metrics (node rounding, node padding, border thickness normal/selected, pin radius, link thickness, link tangent distance) are canvas-space and are multiplied by the current zoom before reaching the drawlist, so chrome scales with the nodes it decorates; hit and snap tolerances (pin hit radius, link hit distance, link snap radius) are screen-pixel constants, so targets stay equally grabbable at every detent. Zero-value style defaulting is lazy: components are constructed before `app.Run()` creates the imgui context and applies the theme (`app.go:75-93`), so `NewNodeCanvas` must not read theme state — a zero-valued `Style` is resolved by calling `DefaultNodeCanvasStyle()` at the first `Begin`, when a context and theme are guaranteed live. `Style` is a complete value, not a sparse overlay: a partially filled struct is used as-is, zero fields included. Callers wanting partial customization start from `DefaultNodeCanvasStyle()` and mutate fields — necessarily after an imgui context exists (e.g. in `OnSetup`, or later via `SetStyle`). Apps that rebuild style on theme change assign `nc` a fresh `DefaultNodeCanvasStyle()` (or their own) at any time. `DefaultNodeCanvasStyle()` derives colors from `imgui.CurrentStyle().Colors()` the way `HueColorScheme.Apply` reads them (`theme.go:60`) — e.g. node body from `ColChildBg`/`ColPopupBg`, title from `ColTitleBgActive`, borders from `ColBorder`, selection emphasis from `ColHeaderActive` — replacing baab's imgui→imnodes bridge as the spec directs. Metric starting points: pin radius 5, pin hit radius 10, link snap radius 24, link hit distance 6, link thickness 2 — all expected to be tuned against the example, which is the spec's stated resolution for these open questions. Exact grid style per detent (line weight/alpha fade at low detents) is likewise tuned in the example.

## Staging

Five stages, each leaving `make test` green and each terminus-gated before Michael reviews, per the pipeline. Stages 2–5 also keep the example runnable at whatever scope exists so far — the example grows with the component rather than arriving at the end.

**Stage 1 — pure core.** `nodeCanvasGeometry.go`, `nodeCanvasInput.go` (state machine over synthetic geometry), and their tests. No rendering. This is where transform math, detent snapping, zoom-toward-cursor, bezier distance, and the full gesture grammar (press/click/ctrl/shift/drag/box/link-snap/cancel, locked mode) get locked in headlessly.

**Stage 2 — canvas shell.** `nodeCanvas.go` skeleton: config defaulting, `Begin`/`End` with child region, grid, font push, middle-drag pan, wheel detent zoom, `View`/`SetView`/`Detent`/transform helpers. `nodeCanvasStyle.go` lands here as the struct plus lazy zero-value defaulting (stage 2 references the type, so it must exist for the stage to compile); theme derivation can be a placeholder. Example opens an empty pannable, zoomable grid.

**Stage 3 — declaration and rendering.** `Node`/`NodeContext`/`Link` drawing: splitter channels, title bar, content flow, pin rows and markers, node chrome, bezier links, retained geometry. `DefaultNodeCanvasStyle()`'s real theme derivation lands here, replacing the stage-2 placeholder. Example declares the synthetic graph (several node kinds, icon-bearing titles, widget-bearing and label-only content, several links) — static, no interaction yet. Icon glyph-offset behavior at each detent is checked here (risk 1 below).

**Stage 4 — interaction.** Wire the stage-1 state machine to real input sampling: hover, selection intents, node dragging with in-flight offset rendering and single `NodesMoved` on release, box select, link creation with snap, locked mode, the widget-first arbitration rule. Example handles intents against its in-memory model, logs them via `dl`, and runs `NodesMoved` through a dfx `UndoSystem` command to demonstrate one-gesture-one-intent-one-command. Widget-bearing node content in the example wraps its widgets in `PushItemWidth`/`PopItemWidth` — node content owns its widget widths; the canvas window's default item width is meaningless inside a node, and the example is where that pattern is made visible.

**Stage 5 — navigation and closure.** `ZoomToFit`, `CenterOn`, per-detent content simplification in the example (branching on `n.Detent()`: `n.Label` for non-pin content below 1.0, with `Input`/`Output` pin rows still declared at every detent so links keep their anchors — the spec's reduced-detent content list is "labels, values, pins"), view get/set round-trip with simulated persistence, a live locked-mode toggle wired to `SetLocked`, style/grid tuning pass, `docs/current/node-canvas.md`, CHANGELOG entry, AGENTS.md line. The spec's acceptance paragraph is the checklist for this stage's example walkthrough.

## Risks and Watch Items

1. **Icon glyph offset below detent 1.0** (flagged during planning). The Material Icons merge uses fixed pixel `GlyphOffset` values tuned per size (`fonts.go:48,75`); under dynamic scaling the offset may not track the pushed size, drifting icon baselines at 0.25–0.75. Surface it in stage 3 with icon-bearing titles at every detent; the fallback is a per-detent Y correction applied in title-bar rendering only. Does not block any other stage.
2. **Splitter capacity growth.** The owned `DrawListSplitter` cannot re-split mid-frame; the `1 + 2*(lastFrameNodeCount + 8)` capacity formula degrades to shared channels only on frames where more than eight new nodes appear at once, and only as a one-frame z-order artifact. Accepted; noted here so it isn't "fixed" into something more elaborate.
3. **Wheel-over-widget blocks zoom at detent 1.0** — a direct consequence of the sanctioned widget-first cascade. Predictable and small (only while the cursor is directly over a live widget); accepted, not a bug.
4. **Text fidelity at 0.25.** imgui 1.92 bakes per-size glyphs, so quality should be fine and atlas growth is bounded by the detent set (the spec's design bet). If 0.25 text looks poor with the default fonts, that's an example-tuning concern (style/grid pass, stage 5), not an architecture change.
5. **Overlapping-node widget interaction.** imgui item overlap between two nodes' live widgets follows imgui's own rules; per-node channel pairs keep the visuals correct, and interaction correctness for heavily overlapped nodes is accepted as-is for v1.

## Out of Scope

Everything in the spec's Deferred list (minimap, link detach, subgraph clipboard, sub-1.0 widget interaction, continuous zoom, free-form pins, context menus), plus: no `Component` wrapper, no canvas-side actions, no delete intent, no modifier-combining box select, and no child-window-producing node content (`BeginChild` in a content closure) — child windows own their drawlists and escape the per-node z-order machinery; deferred until a client actually declares one. The baab migration is a separate follow-on work order in baab's repo; this work completes at the spec's acceptance gate — the example feeling right under the hand.
