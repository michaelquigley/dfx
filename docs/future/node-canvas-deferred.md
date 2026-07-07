# NodeCanvas — Deferred Concerns

NodeCanvas v1 is built and documented in [`docs/current/node-canvas.md`](../current/node-canvas.md); the originating spec and work order are removed (git history preserves them, and the mercurius session `s_fOO4CQSfeIJC` synopsis records why each contract detail is the way it is). This note carries the still-live deferred list forward, with the revisit condition for each.

## Deferred features (from the spec)

**Minimap.** The design bet is that detents + view bookmarks + zoom-to-fit + center-on-selection cover navigation better than a space-eating minimap. Revisit only if large graphs prove disorienting in practice.

**Link detach-by-drag.** Selection + delete covers removal (deletion is app-side; no delete intent exists by design). Add later if patching workflow wants it.

**Subgraph cut/copy/paste and duplicate.** App-domain semantics dominate the canvas's share of the work. The intent model accommodates it later without API upheaval.

**Widget interaction below detent 1.0.** Scaling imgui widget chrome per-canvas means shadowing global style metrics — possible, fiddly, and unneeded under the "edit at 100%" model. Revisit only if a real workflow demands turning knobs while zoomed out.

**Continuous zoom.** Detents are a feature (bounded font atlas, named views), not a limitation. Revisit only with SDF text rendering, i.e. a future native substrate.

**Free-form pin placement and non-bezier link routing.** Row-anchored pins and horizontal beziers match every current design.

**Context menus.** Right-button is reserved and does nothing in v1. Canvas-provided menu affordances wait for a second client's needs.

**Child-window-producing node content.** `BeginChild`-style scroll regions inside content closures escape the per-node drawlist splitter and the visual/hit z-order invariant. The descendant-window arbitration checks are already in place and become load-bearing if this is ever supported; deferred until a client actually declares one.

**Modifier-combining box select.** Box select replaces the whole selection regardless of modifiers; ctrl/shift-combining variants deferred with the rest.

## Small implementation-level revisit hooks (from the build)

- Wheel detent stepping is honored only in the idle gesture state (conservative reading of the arbitration predicates). Revisit if zoom-during-pan wants support.
- The gesture machine's `viewChanged` reports only actual view mutations; a wheel event already clamped at the top detent does not cancel a pending zoom-to-fit. Revisit if "navigation attempted" semantics ever feel more correct under the hand.
- Output pin labels sit in left-aligned content flow; right-aligning them against the node's final width would need a second layout pass. Cosmetic; revisit with a real client's node designs.

## Follow-on work elsewhere

The baab migration (imnodes → NodeCanvas: interning tables, position sync, `pendingPanning`, selection sync-out, and the style bridge all deleted; bookmarks, registry popup, snap-to-grid, fetch-node, display states kept) is a separate work order in baab's repo, per the spec's migration sketch preserved in git history.
