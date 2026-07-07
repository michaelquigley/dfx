# NodeCanvas

A zoomable, pannable node-graph editing surface for dfx, replacing imnodes. The application owns the graph; the canvas owns the view. First client: baab's visual system editor.

## Why

imnodes cannot zoom — its coordinate model has no canvas transform, and until Dear ImGui 1.92 the font atlas couldn't rasterize at arbitrary sizes anyway. dfx now pins cimgui-go v1.5.0 (imgui 1.92.8), whose dynamic font system (`PushFont` at any size) removes the substrate limitation. The remaining obstacle is imnodes itself, and it earns replacement on more than the zoom axis: it retains node positions, selection, and panning inside its own context, forcing clients into synchronization code; it demands int32 IDs, forcing clients to intern their model objects through lookup tables; and its editor context has lifecycle constraints that force clients to defer view restoration until a draw frame is active. baab carries compensation code for all three.

Node-graph surfaces recur across our product designs. This component is infrastructure, built once in dfx and proven through a dfx example before baab migrates.

## Guiding Principles

**The app owns all graph truth.** Nodes, links, positions, and selection are declared by the application every frame, exactly in the dfx immediate-mode idiom. The canvas never becomes a second database. There is no equivalent of `SetNodeGridSpacePos`, no read-back of positions, no selection sync loop.

**The canvas owns view and gesture state, and nothing else.** Pan offset, zoom detent, and in-flight gestures (a drag in progress, a link being pulled from a pin, a box-select preview) are the canvas's only retained state. View state is plain data the app can read, set, and persist directly — no context-lifecycle hacks.

**The canvas reports intents; the app applies them.** A completed gesture produces an intent — nodes moved, link created, selection changed. The app decides what to do with it (typically: run an undo command). The canvas never mutates the graph, and never assumes an intent was accepted; next frame's declarations are the only truth it renders.

**Detail policy belongs to the app.** The canvas provides zoom mechanics and reports the current detent. What a node displays at each detent is the application's decision, made inside the node's content closure. baab already has a per-node display-state concept (maximized/focused/minimized); detent composes with it app-side. The canvas has no LOD system.

## Concepts

**Canvas space** is the infinite 2D plane nodes live on, in unscaled units. **Screen space** is pixels within the canvas widget's on-screen rectangle. The view transform between them is `screen = (canvas + pan) * zoom + origin`. The canvas exposes both directions as helpers; client code never does this arithmetic by hand.

**Detents** are the only zoom levels. Default set: 0.25, 0.5, 0.75, 1.0, configurable. 1.0 is the maximum. Discrete detents bound the set of font sizes ever rasterized (bounded atlas growth under imgui 1.92's per-size baking) and make zoom levels feel like named views rather than a continuous scrubber.

**Nodes** are rounded-rect cards with a title bar region, arbitrary content, and pin rows. Node size derives from content, as with imnodes. **Pins** are connection points rendered as markers on the node's left (inputs) or right (outputs) edge, each anchored to a content row. **Links** are cubic beziers between an output pin and an input pin.

**IDs** are application-supplied. The canvas is generic over a single comparable ID type shared by nodes, pins, and links: `NodeCanvas[ID comparable]`. baab can pass its model objects (or their `model.Identifier`s) directly; the interning tables die.

## API Contract

This is contract-level, not implementation. Signatures are illustrative of shape and ownership, not final.

```go
type NodeCanvas[ID comparable] struct { /* view + gesture state */ }

func NewNodeCanvas[ID comparable](cfg NodeCanvasConfig) *NodeCanvas[ID]

type NodeCanvasConfig struct {
    Detents      []float32 // default {0.25, 0.5, 0.75, 1.0}; last entry is max
    GridSpacing  float32   // default 50; exposed for app-side snap logic
    Locked       bool      // suppress node dragging (baab's xltLocked)
    Style        NodeCanvasStyle
}
```

Per frame, inside the owning component's `Draw`:

```go
nc.Begin(state)                       // establishes the canvas child region, draws grid, handles view input

nc.Node(id, pos, NodeFlags{Selected: sel}, func(n *NodeContext[ID]) {
    n.TitleBar(func() { /* imgui/dfx widgets: icons, text */ })
    /* arbitrary content: imgui/dfx widgets, legal at detent 1.0 */
    n.Input(pinID, "score in")        // pin row: marker + label
    n.Output(pinID, "score out")
})

nc.Link(id, fromPinID, toPinID, LinkFlags{Selected: sel})

intents := nc.End()                   // finishes drawing, resolves gestures
```

`End` returns the frame's intents:

```go
type Intents[ID comparable] struct {
    NodesMoved       []NodeMove[ID]     // emitted once, on drag release: {ID, From, To}
    LinkCreated      *LinkCreate[ID]    // {FromPin, ToPin}; app validates and applies (or ignores)
    SelectionChanged *SelectionChange[ID] // full replacement sets for nodes and links
}
```

View state and navigation:

```go
func (nc *NodeCanvas[ID]) View() View                  // {Pan Vec2, Zoom float32} — plain, persistable data; zoom snaps to a configured detent
func (nc *NodeCanvas[ID]) SetView(v View)              // callable any time — no draw-frame lifecycle hack; applies at the next frame
func (nc *NodeCanvas[ID]) Detent() float32             // current zoom factor
func (nc *NodeCanvas[ID]) CanvasFromScreen(p Vec2) Vec2
func (nc *NodeCanvas[ID]) ScreenFromCanvas(p Vec2) Vec2
func (nc *NodeCanvas[ID]) ZoomToFit(ids ...ID)         // all nodes when empty
func (nc *NodeCanvas[ID]) CenterOn(ids ...ID)          // pan only, detent unchanged
```

The content closure runs with the canvas's scaled font pushed (`base * detent`). At detent 1.0, imgui/dfx widgets behave normally — this is the editing detent — with one v1 exclusion: content that opens child windows of its own (`BeginChild`-style scroll regions) is deferred until a client needs it, since child windows carry their own drawlists and sit outside the canvas's z-order machinery. Below 1.0, text and drawlist content scale correctly, but interactive imgui widgets are not supported: widget chrome metrics (frame padding, grab sizes) come from global style and do not scale. The app is expected to declare simplified content at reduced detents (labels, values, pins), and that simplified content is non-interactive: below 1.0, content closures must not emit anything hoverable or activatable, or hover capture would carve holes in the canvas's own hit-testing. The `NodeContext` exposes `Detent()` so content closures can branch without reaching back to the canvas, and an interaction-free label primitive so honoring the contract is the path of least resistance.

## Interaction Grammar

**Left button** is selection and manipulation: click a node to select it (ctrl-click toggles, shift-click adds); drag a selected node to move the whole selection; drag on empty canvas for box select (preview drawn by canvas, committed as a `SelectionChanged` intent on release); click a link to select it; click empty canvas to clear selection. Drag from a pin marker pulls a link, with snap when the cursor is within a snap radius of a compatible-side pin (output→input or input→output; the canvas knows sides, not types); release on a pin emits `LinkCreated`, release elsewhere ends the gesture with no intent.

**Middle-drag** pans (space is deliberately not claimed — clients like baab bind it to transport). **Wheel** steps the detent, zooming toward the cursor — the canvas point under the mouse stays fixed through the transition. **Right button** is reserved for future context menus and does nothing in v1.

**Locked mode** suppresses node dragging only; selection, panning, zooming, and link creation remain live.

Node dragging renders from internal gesture state — declared positions plus the in-flight offset — so the graph moves smoothly while the app's model stays untouched until the single `NodesMoved` intent on release. One gesture, one intent, one undo command.

## Rendering

Everything is drawn through `ImDrawList` with the view transform applied canvas-side: dotted or line grid (spacing from config, scaled), node body as a filled rounded rect with border, title bar as a tinted band, pin markers as circles on the node edge, links as horizontal-tangent cubic beziers with hover/selected emphasis. Selection is indicated by border color/thickness. All colors and metrics live in `NodeCanvasStyle`, with a constructor deriving defaults from the active dfx theme — replacing baab's imgui→imnodes color bridge with an imgui→style-struct derivation in dfx itself.

Hit-testing operates in canvas space through the inverse transform and works identically at every detent — nodes can be selected, dragged, and linked while zoomed out; only in-node widget interaction is confined to 1.0.

## Acceptance: the dfx Example

`examples/dfx_example_nodecanvas` is the proving ground and the migration gate. A standalone app declaring a small synthetic graph (a handful of node kinds with title icons, a mix of widget-bearing and label-only content, several links) demonstrating: node dragging with multi-select, box select, click/ctrl/shift selection, link creation with snap, detent zoom toward cursor through all four detents with app-side content simplification below 1.0, pan, zoom-to-fit, center-on-selection, locked mode, view get/set round-trip (simulated persistence), and intent handling wired to a trivial in-memory model with logged intents. When the example feels right under the hand, baab migrates.

## baab Migration (informative)

The migration deletes more than it adds. Gone: `idFor`/`objFor` interning (model objects become IDs directly), the per-frame position diff and `xltOverride` force-write dance (positions are declared; `NodesMoved` becomes an intent handler running the existing `TranslateNodesCommand`), the `pendingPanning` deferral (`SetView` is callable any time, no draw-frame lifecycle constraint), the selection sync-out loop (`SelectionChanged` handler writes `env.sel`), the imnodes style bridge, the minimap, and the imnodes context lifecycle. Kept intact: panning bookmarks (now view bookmarks carrying detent, same persistence extension shape), the node registry popup (`CanvasFromScreen` replaces its manual origin/panning arithmetic), connection bookmarks, snap-to-grid (reads `GridSpacing` from config), fetch-node (via `CanvasFromScreen` of the viewport center), display states (now composed with detent inside the content closure), and all undo command machinery.

## Deferred (and Why)

**Minimap.** The design bet is that detents + bookmarks + zoom-to-fit + center-on-selection cover navigation better than a space-eating minimap. Revisit only if large graphs prove disorienting in practice.

**Link detach-by-drag.** imnodes offers it; baab doesn't use it. Selection + delete covers removal. Add later if patching workflow wants it.

**Subgraph cut/copy/paste and duplicate.** App-domain semantics (what does duplicating a DrumScore mean?) dominate the canvas's share of the work. The intent model accommodates it later without API upheaval.

**Widget interaction below detent 1.0.** Scaling imgui widget chrome per-canvas means shadowing global style metrics — possible, fiddly, and unneeded under the "edit at 100%" model. Revisit only if a real workflow demands turning knobs while zoomed out.

**Continuous zoom.** Detents are a feature (bounded atlas, named views), not a limitation to fix. Revisit only with SDF text rendering, i.e., in a future native substrate.

**Free-form pin placement and non-bezier link routing.** Row-anchored pins and horizontal beziers match every current design. Orthogonal routing or arbitrary pin geometry would complicate hit-testing and layout for no present client.

**Context menus.** Right-button is reserved; baab's INSERT-driven registry popup already covers node creation. Canvas-provided menu affordances wait for a second client's needs.

## Seam Census

**Canvas ↔ application (declaration/intent seam).** The load-bearing boundary of the design. Crossed by: per-frame declarations inbound, intents outbound, view state in both directions. The contract is that the canvas holds no graph truth and the app holds no gesture state. Every future feature request should be tested against this seam first — anything that tempts the canvas to retain graph data (auto-layout state, link validity, clipboard) belongs on the app side or in a separate layer.

**Canvas ↔ imgui substrate.** Deliberately thin: `ImDrawList` primitives, `PushFont` at scaled sizes, input queries, one child region. This is a designed-for seam — the back-burnered native Go immediate-mode library would slot in beneath it by satisfying the same narrow draw/text/input needs. No abstraction layer is built now (one substrate, no second party); the seam is kept honest by concentrating all imgui calls in the canvas's render/input internals and keeping the public API free of imgui types except `Vec2`-equivalents. Revisit condition: the native library becomes a committed project.

**Detent mechanics ↔ detail policy.** Canvas provides zoom and reports detent; app decides content. Crossed by a single float. Keeping LOD policy out of the canvas is what lets baab compose detent with its existing display states without the canvas knowing either concept exists.

**Style ↔ theme.** `NodeCanvasStyle` is plain data; the dfx theme derivation is a constructor, not a live coupling. Apps may override any field. Crossed once at construction (or whenever the app rebuilds style on theme change).

## Open Questions for Planning

Final shapes of the intent structs and `NodeContext` surface; pin marker hit radius and link snap radius defaults; whether `Begin`/`End` or a single closure-taking `Canvas(state, func(...))` entry better fits dfx idiom (both satisfy the contract; pick during implementation design against the example); exact grid rendering style at each detent.

*(All of these are now resolved in the companion work order: concrete API and intent shapes, `Begin`/`End`, starting hit/snap radii tuned against the example, grid style deferred to the example's tuning pass.)*
