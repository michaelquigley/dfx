# HCollapse

`HCollapse` is a horizontally collapsible panel: a header with a toggle and title, its content below, and an optional resize handle. Collapsed, it narrows to `MinWidth` and shows only the toggle. Expanding and collapsing animate over `TransitionMs`. `examples/dfx_example_hcollapse` shows left- and right-anchored panels side by side.

## Sizing contract

`Draw(state)` takes the container's full available size in `state.Size`, not the panel's own width. The panel draws itself `CurrentWidth` wide at the cursor; the caller lays out the rest of the row around `CurrentWidth`. `state.Size.X` is what bounds a resize, so a caller that passes the panel's own width tells the panel it has no room.

A resize clamps the width in this order:

1. no wider than the container less 50px, which stays for the rest of the layout;
2. no wider than `MaxWidth` (when set);
3. no narrower than `MinWidth`.

`MinWidth` is applied last, so no container and no sequence of drags can take a panel below it. The arithmetic (`applyResize`, `clampWidth`) needs no imgui context and is tested headless.

## Anchor

`HCollapseConfig.Anchor` says which edge of its container the panel sits against.

- **`AnchorLeft` (the default):** the toggle is at the header's left and the resize handle on the panel's right edge. Dragging the handle right widens the panel, and the toggle's chevron points left to collapse.
- **`AnchorRight`:** for a panel at the right of its row, such as an inspector, the header is mirrored. The resize handle is on the panel's left edge and the toggle at the header's right. A drag toward the panel's interior (right) narrows it, a drag away (left) widens it, and the chevron points right to collapse. The caller draws the panel after the row's content, with `SameLine`, and sizes the content to the space the panel leaves.
