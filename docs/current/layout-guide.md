# Dear ImGui Layout & Sizing Guide

A comprehensive guide to understanding widget positioning, child window sizing, and layout mechanics in Dear ImGui (via cimgui-go and dfx).

## Table of Contents

1. [The Core Mental Model](#part-1-the-core-mental-model)
2. [The Cursor System](#part-2-the-cursor-system)
3. [Spacing and Padding](#part-3-spacing-and-padding)
4. [Widget Width Control](#part-4-widget-width-control)
5. [Horizontal Layout](#part-5-horizontal-layout)
6. [Groups](#part-6-groups)
7. [Child Windows](#part-7-child-windows)
8. [Columns (Legacy)](#part-8-columns-legacy)
9. [Tables](#part-9-tables)
10. [Practical Patterns](#part-10-practical-patterns)
11. [Debugging Layout Issues](#part-11-debugging-layout-issues)

---

## Part 1: The Core Mental Model

### Immediate-Mode vs Retained-Mode Layout

Dear ImGui uses an **immediate-mode** approach where the UI is rebuilt every frame. Unlike retained-mode frameworks (React, Qt, SwiftUI) where you declare a tree of widgets that persists, ImGui requires you to emit widgets on every frame.

```go
// This runs every frame
func (c *MyComponent) Draw(state *dfx.State) {
    imgui.Text("Hello")        // Emitted at cursor position
    imgui.Button("Click Me")   // Emitted below the text
}
```

### The Cursor as the "Write Head"

Think of ImGui's layout system like a typewriter or text editor:
- A **cursor** tracks where the next widget will be placed
- Each widget **advances the cursor** after drawing
- By default, the cursor moves **down** after each widget (vertical stacking)
- `SameLine()` moves the cursor **right** instead of down

```
Initial State:        After Text():         After Button():
┌──────────────┐     ┌──────────────┐      ┌──────────────┐
│▌             │     │ Hello        │      │ Hello        │
│              │     │▌             │      │ [Click Me]   │
│              │     │              │      │▌             │
└──────────────┘     └──────────────┘      └──────────────┘
  ▌ = cursor           cursor moved          cursor moved
```

### Why Understanding State Flow Matters

Because ImGui reconstructs the UI each frame, there's no automatic layout memory. You must:
1. Query available space with `ContentRegionAvail()`
2. Set sizes explicitly or let ImGui calculate them
3. Understand how spacing accumulates

---

## Part 2: The Cursor System

### Absolute vs Window-Local Coordinates

Internally, ImGui tracks the cursor in **absolute screen coordinates** for efficiency. However, the public API exposes **window-local coordinates**:

| Function | Coordinates | Use Case |
|----------|-------------|----------|
| `CursorPos()` | Window-local | Calculating widget positions within a window |
| `CursorScreenPos()` | Absolute | Custom drawing, overlays |
| `SetCursorPos(pos)` | Window-local | Jump cursor to specific position |
| `SetCursorScreenPos(pos)` | Absolute | Advanced positioning |

```go
// Get position within the current window
pos := imgui.CursorPos()  // e.g., {8, 8} after WindowPadding

// Get absolute screen position
screenPos := imgui.CursorScreenPos()  // e.g., {108, 208} if window is at {100, 200}

// Jump to a specific position in the window
imgui.SetCursorPos(imgui.NewVec2(100, 50))
imgui.Text("I'm at (100, 50)")
```

### CursorMaxPos and Content Extent Tracking

ImGui tracks the maximum extent of content via `CursorMaxPos`. This is used for:
- Determining scrollbar need
- Calculating auto-resize dimensions
- Reporting content size to parent containers

Every widget updates this value to include its bounding box.

---

## Part 3: Spacing and Padding

### The Three Key Variables

| Variable | Purpose | Default | Applied |
|----------|---------|---------|---------|
| `WindowPadding` | Space from window edge to content | 8×8 | Once, at window edges |
| `FramePadding` | Internal padding in framed widgets | 4×3 | Per framed widget (buttons, inputs) |
| `ItemSpacing` | Gap between widgets | 8×4 | After each widget |

These defaults are ImGui's own. A dfx App replaces them with the selected style's metrics at startup: 12×12 window padding, 8×5 frame padding, and 8×8 item spacing in Faceplate, the default; 4×4 for all three in Classic.

### The Spacing Hierarchy

```
Window Outer Edge
│
├── WindowBorderSize (if border enabled)
│
├── WindowPadding.x / .y
│   │
│   └── Content Region Starts (CursorStartPos)
│       │
│       ├── [Button with FramePadding.y×2 height]
│       │   └── Button text centered within FramePadding
│       │
│       ├── ItemSpacing.y
│       │
│       ├── [Next Widget]
│       │
│       ├── ItemSpacing.y
│       │
│       └── ...
│
└── WindowPadding (implicit at bottom for scrolling)
```

### Practical Effects

```go
style := imgui.CurrentStyle()

// Widget height formula
frameHeight := imgui.FrameHeight()  // = FontSize + FramePadding.y * 2

// Line height with spacing
lineHeight := imgui.FrameHeightWithSpacing()  // = frameHeight + ItemSpacing.y

// Content region is inset by WindowPadding
contentStart := style.WindowPadding()  // Where first widget appears
```

### cimgui-go Note

Access style values via:
```go
style := imgui.CurrentStyle()
windowPadding := style.WindowPadding()   // imgui.Vec2
framePadding := style.FramePadding()     // imgui.Vec2
itemSpacing := style.ItemSpacing()       // imgui.Vec2
```

---

## Part 4: Widget Width Control

### Default Width Behavior

By default, widgets that accept a width use `ItemWidthDefault`, which is approximately 65% of the window's content width. This provides a reasonable default for most forms.

### PushItemWidth / PopItemWidth Stack

`PushItemWidth()` sets the default width for subsequent widgets until `PopItemWidth()`:

```go
imgui.PushItemWidth(200)  // All following inputs will be 200px wide
imgui.InputText("Name", &name)
imgui.InputText("Email", &email)
imgui.PopItemWidth()
```

### SetNextItemWidth for One-Off

`SetNextItemWidth()` affects only the next widget:

```go
imgui.SetNextItemWidth(300)
imgui.InputText("Wide Input", &value)
imgui.InputText("Normal Width", &other)  // Back to default
```

### The Negative Value Trick

**This is crucial**: negative width values mean "pixels from the right edge":

| Width Value | Result |
|-------------|--------|
| `100` | Fixed 100 pixels |
| `0` | Use default (65% of window) |
| `-1` | Fill to 1 pixel from right edge |
| `-100` | Fill to 100 pixels from right edge |
| `-math.SmallestNonzeroFloat32` | Fill to exact right edge |

```go
// Fill entire width
imgui.PushItemWidth(-math.SmallestNonzeroFloat32)
imgui.InputText("##full", &value)  // Fills to edge
imgui.PopItemWidth()

// Leave room for a button
imgui.PushItemWidth(-80)  // Leave 80px on the right
imgui.InputText("##input", &value)
imgui.PopItemWidth()
imgui.SameLine()
imgui.Button("Browse")  // Fits in the 80px
```

### CalcItemWidth Resolution Order

When determining width, ImGui checks (in order):
1. `SetNextItemWidth()` value (if set)
2. `PushItemWidth()` stack top
3. Default width calculation

---

## Part 5: Horizontal Layout

### Default Vertical Stacking

Without any intervention, widgets stack vertically:

```go
imgui.Text("Line 1")     // Cursor advances down
imgui.Text("Line 2")     // Below Line 1
imgui.Button("Submit")   // Below Line 2
```

### SameLine() Mechanics

`SameLine()` prevents the line break and continues horizontally:

```go
imgui.Text("Label:")
imgui.SameLine()         // Stay on same line
imgui.InputText("##input", &value)

// With custom spacing
imgui.Button("A")
imgui.SameLineV(0, 20)   // 20 pixels gap instead of ItemSpacing.x
imgui.Button("B")

// With absolute positioning
imgui.Button("Left")
imgui.SameLineV(200, 0)  // Position at x=200 from window left
imgui.Button("At 200px")
```

### How SameLine() Works Internally

```
Before SameLine():           After SameLine():
┌────────────────────┐      ┌────────────────────┐
│ [Button A]         │      │ [Button A]         │
│ ▌                  │      │           ▌        │
│                    │      │                    │
└────────────────────┘      └────────────────────┘
  Cursor below A              Cursor beside A
```

`SameLine()` does:
1. Sets `IsSameLine = true` (prevents new line height calculation)
2. Restores Y to the previous line's Y
3. Positions X at previous item's right edge + spacing

---

## Part 6: Groups

### BeginGroup / EndGroup

Groups create a bounding box around multiple widgets, treating them as a single item for layout purposes.

```go
imgui.BeginGroup()
imgui.Text("Name:")
imgui.Text("Email:")
imgui.EndGroup()

// Now the group is a single "item"
imgui.SameLine()

imgui.BeginGroup()
imgui.InputText("##name", &name)
imgui.InputText("##email", &email)
imgui.EndGroup()
```

Result:
```
┌────────────────────────────────┐
│ Name:   [___________________] │
│ Email:  [___________________] │
└────────────────────────────────┘
```

### ItemRectSize After EndGroup

After `EndGroup()`, you can query the group's dimensions:

```go
imgui.BeginGroup()
// ... widgets ...
imgui.EndGroup()

groupSize := imgui.ItemRectSize()  // Vec2 with group dimensions
groupMin := imgui.ItemRectMin()    // Top-left corner
groupMax := imgui.ItemRectMax()    // Bottom-right corner
```

### Groups for Complex Inline Layouts

```go
// Create a "panel" with a frame around grouped content
imgui.BeginGroup()
imgui.Text("Panel Title")
imgui.Separator()
imgui.Text("Content here...")
imgui.Button("Action")
imgui.EndGroup()

// Draw a border around the group
min := imgui.ItemRectMin()
max := imgui.ItemRectMax()
drawList := imgui.WindowDrawList()
drawList.AddRect(min, max, imgui.ColorU32Col(imgui.ColBorder))
```

---

## Part 7: Child Windows

Child windows are the primary tool for creating bounded, optionally scrollable regions.

### Size Parameter Semantics

The `BeginChildStrV()` size parameter follows the universal pattern:

| Size | X Behavior | Y Behavior |
|------|------------|------------|
| `(0, 0)` | Fill remaining width | Fill remaining height |
| `(300, 200)` | Fixed 300px | Fixed 200px |
| `(-100, 0)` | 100px from right edge | Fill remaining height |
| `(0, -50)` | Fill remaining width | 50px from bottom |
| `(-1, -1)` | Fill to 1px from edges | Fill to 1px from edges |

```go
// Fill all remaining space
imgui.BeginChildStrV("full", imgui.NewVec2(0, 0), imgui.ChildFlagsNone, imgui.WindowFlagsNone)

// Fixed size scrollable region
imgui.BeginChildStrV("scrollable", imgui.NewVec2(300, 200), imgui.ChildFlagsNone, imgui.WindowFlagsNone)

// Leave 100px on the right for a sidebar
imgui.BeginChildStrV("main", imgui.NewVec2(-100, 0), imgui.ChildFlagsNone, imgui.WindowFlagsNone)
// ... content ...
imgui.EndChild()
imgui.SameLine()
imgui.BeginChildStrV("sidebar", imgui.NewVec2(0, 0), imgui.ChildFlagsNone, imgui.WindowFlagsNone)  // Takes remaining 100px
// ... sidebar content ...
imgui.EndChild()
```

### ImGuiChildFlags

```go
const (
    ChildFlagsNone                   = 0
    ChildFlagsBorders                = 1 << 0  // Draw border, apply WindowPadding
    ChildFlagsAlwaysUseWindowPadding = 1 << 1  // Force WindowPadding even without border
    ChildFlagsResizeX                = 1 << 2  // Allow horizontal resize
    ChildFlagsResizeY                = 1 << 3  // Allow vertical resize
    ChildFlagsAutoResizeX            = 1 << 4  // Size to content width
    ChildFlagsAutoResizeY            = 1 << 5  // Size to content height
    ChildFlagsAlwaysAutoResize       = 1 << 6  // Always measure size
    ChildFlagsFrameStyle             = 1 << 7  // Style like a framed widget
    ChildFlagsNavFlattened           = 1 << 8  // Share navigation scope
)
```

### AutoResize vs Explicit Sizing

```go
// Auto-resize to fit content (no scrolling)
imgui.BeginChildStrV("auto", imgui.NewVec2(0, 0),
    imgui.ChildFlagsAutoResizeY, imgui.WindowFlagsNone)
for i := 0; i < 5; i++ {
    imgui.Text(fmt.Sprintf("Item %d", i))
}
imgui.EndChild()
// Child height = content height (no scrollbar)

// Fixed height with scrolling
imgui.BeginChildStrV("scroll", imgui.NewVec2(0, 100),
    imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
for i := 0; i < 50; i++ {
    imgui.Text(fmt.Sprintf("Item %d", i))
}
imgui.EndChild()
// Child height = 100px, scrollbar appears
```

### Nested Child Windows

Child windows can be nested, and each maintains its own:
- Cursor position
- Clipping region
- Scroll state
- Content extent

```go
imgui.BeginChildStrV("outer", imgui.NewVec2(400, 300), imgui.ChildFlagsNone, imgui.WindowFlagsNone)
    imgui.Text("Outer content")

    imgui.BeginChildStrV("inner", imgui.NewVec2(200, 150), imgui.ChildFlagsNone, imgui.WindowFlagsNone)
        imgui.Text("Inner content")
    imgui.EndChild()

    imgui.Text("More outer content")
imgui.EndChild()
```

---

## Part 8: Columns (Legacy)

### Historical Context

The Columns API was ImGui's original solution for multi-column layouts. It has been **deprecated** in favor of Tables (introduced in v1.80).

### Basic Usage

```go
// Create 3 columns
imgui.ColumnsV(3, "mycolumns", true)

imgui.Text("Column 1")
imgui.NextColumn()
imgui.Text("Column 2")
imgui.NextColumn()
imgui.Text("Column 3")
imgui.NextColumn()

// Second row
imgui.Text("Row 2, Col 1")
imgui.NextColumn()
imgui.Text("Row 2, Col 2")
imgui.NextColumn()
imgui.Text("Row 2, Col 3")
imgui.NextColumn()

imgui.Columns()  // End columns
```

### Width Control

```go
imgui.ColumnsV(2, "cols", true)
imgui.SetColumnWidth(0, 150)  // First column = 150px
// Second column gets remaining space
```

### Why Deprecated

Limitations of the Columns API:
1. No automatic sizing policies
2. No built-in sorting
3. No column hiding/reordering
4. Manual width management
5. Complex nesting behavior

### Migration to Tables

```go
// Old Columns code:
imgui.ColumnsV(2, "old", true)
imgui.Text("Name")
imgui.NextColumn()
imgui.Text("Value")
imgui.NextColumn()
imgui.Columns()

// New Tables code:
if imgui.BeginTableV("new", 2, imgui.TableFlagsBorders, imgui.Vec2{}, 0) {
    imgui.TableNextColumn()
    imgui.Text("Name")
    imgui.TableNextColumn()
    imgui.Text("Value")
    imgui.EndTable()
}
```

---

## Part 9: Tables

Tables are ImGui's modern, feature-rich layout system for columnar data.

### The Four Sizing Policies

| Policy | Behavior | Use Case |
|--------|----------|----------|
| `SizingFixedFit` | Columns fit their content | Variable content widths |
| `SizingFixedSame` | All columns same width (max content) | Uniform columns |
| `SizingStretchProp` | Stretch proportional to content | Flexible, content-aware |
| `SizingStretchSame` | Stretch equally (default) | Even distribution |

### Default Policy Selection

- **With ScrollX**: `SizingFixedFit` (columns fit content, table scrolls)
- **Without ScrollX**: `SizingStretchSame` (columns stretch to fill)
- **AlwaysAutoResize window**: `SizingFixedFit`

### TableSetupColumn Specifications

```go
if imgui.BeginTableV("example", 3, imgui.TableFlagsBorders, imgui.Vec2{}, 0) {
    // Fixed width column
    imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthFixed, 100, 0)

    // Stretch column with weight 2.0
    imgui.TableSetupColumnV("Description", imgui.TableColumnFlagsWidthStretch, 2.0, 0)

    // Stretch column with default weight 1.0
    imgui.TableSetupColumnV("Value", imgui.TableColumnFlagsWidthStretch, 1.0, 0)

    imgui.TableHeadersRow()

    // Data rows...
    imgui.TableNextRow()
    imgui.TableNextColumn()
    imgui.Text("Item 1")
    imgui.TableNextColumn()
    imgui.Text("A long description that will wrap or clip")
    imgui.TableNextColumn()
    imgui.Text("$10.00")

    imgui.EndTable()
}
```

### The Weight System

For stretch columns, `init_width_or_weight` specifies relative sizing:

```go
// Column A gets 1/4, B gets 3/4 of remaining space
imgui.TableSetupColumnV("A", imgui.TableColumnFlagsWidthStretch, 1.0, 0)
imgui.TableSetupColumnV("B", imgui.TableColumnFlagsWidthStretch, 3.0, 0)
```

### Mixing Fixed and Stretch

Best practice: Fixed columns first, stretch columns last:

```go
if imgui.BeginTableV("mixed", 3, imgui.TableFlagsNone, imgui.Vec2{}, 0) {
    imgui.TableSetupColumnV("Icon", imgui.TableColumnFlagsWidthFixed, 30, 0)
    imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthStretch, 0, 0)
    imgui.TableSetupColumnV("Actions", imgui.TableColumnFlagsWidthFixed, 80, 0)
    // ...
    imgui.EndTable()
}
```

---

## Part 10: Practical Patterns

### Pattern 1: Fill Remaining Space

Use `-math.SmallestNonzeroFloat32` (or `-FLT_MIN` in C++) to fill to the edge:

```go
// Button fills entire width
imgui.SetNextItemWidth(-math.SmallestNonzeroFloat32)
if imgui.Button("Full Width Button") {
    // ...
}

// Input fills remaining width after label
imgui.Text("Name:")
imgui.SameLine()
imgui.SetNextItemWidth(-math.SmallestNonzeroFloat32)
imgui.InputText("##name", &name)
```

### Pattern 2: Proportional Layouts

```go
avail := imgui.ContentRegionAvail()

// Two equal columns
imgui.BeginChildStrV("left", imgui.NewVec2(avail.X*0.5, 0), imgui.ChildFlagsNone, imgui.WindowFlagsNone)
// ... left content ...
imgui.EndChild()

imgui.SameLine()

imgui.BeginChildStrV("right", imgui.NewVec2(0, 0), imgui.ChildFlagsNone, imgui.WindowFlagsNone)  // Takes remaining half
// ... right content ...
imgui.EndChild()
```

### Pattern 3: Right-Aligned Elements

```go
// Right-align a button
avail := imgui.ContentRegionAvail()
buttonWidth := imgui.CalcTextSize("Submit").X + style.FramePadding().X*2
imgui.SetCursorPosX(imgui.CursorPosX() + avail.X - buttonWidth)
imgui.Button("Submit")
```

### Pattern 4: Fixed + Flexible Side-by-Side

```go
// 200px sidebar + flexible main content
imgui.BeginChildStrV("sidebar", imgui.NewVec2(200, 0), imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
// ... sidebar content ...
imgui.EndChild()

imgui.SameLine()

imgui.BeginChildStrV("main", imgui.NewVec2(0, 0), imgui.ChildFlagsNone, imgui.WindowFlagsNone)  // Fills remaining
// ... main content ...
imgui.EndChild()
```

### Pattern 5: Scrollable Region with Fixed Header/Footer

```go
// Fixed header
imgui.Text("Header - Always Visible")
imgui.Separator()

// Calculate available height minus footer
footerHeight := imgui.FrameHeightWithSpacing() + 10

// Scrollable middle section
imgui.BeginChildStrV("content", imgui.NewVec2(0, -footerHeight), imgui.ChildFlagsNone, imgui.WindowFlagsNone)
for i := 0; i < 100; i++ {
    imgui.Text(fmt.Sprintf("Scrollable item %d", i))
}
imgui.EndChild()

// Fixed footer
imgui.Separator()
imgui.Text("Footer - Always Visible")
```

---

## Part 11: Debugging Layout Issues

### Using the ImGui Demo Window

The best reference is `imgui.ShowDemoWindow()`. It demonstrates virtually every layout technique:

```go
showDemo := true
imgui.ShowDemoWindowV(&showDemo)
```

Pay special attention to:
- "Layout & Scrolling" section
- "Tables & Columns" section
- Child windows examples

### Inspect Values at Runtime

```go
imgui.Text(fmt.Sprintf("Cursor: (%.0f, %.0f)",
    imgui.CursorPos().X, imgui.CursorPos().Y))
imgui.Text(fmt.Sprintf("Avail: (%.0f, %.0f)",
    imgui.ContentRegionAvail().X, imgui.ContentRegionAvail().Y))
imgui.Text(fmt.Sprintf("state.Size: (%.0f, %.0f)",
    state.Size.X, state.Size.Y))  // dfx-specific
```

### Common Pitfalls

**1. Off-by-one ItemSpacing**

When calculating heights manually, remember to account for ItemSpacing:

```go
// Wrong: doesn't account for spacing between items
height := numItems * itemHeight

// Correct: account for N-1 spacings
height := numItems*itemHeight + (numItems-1)*style.ItemSpacing().Y
```

**2. Forgetting EndChild/EndTable**

Every `BeginChild` must have a matching `EndChild`, even if `BeginChild` returns false:

```go
// WRONG
if imgui.BeginChildStrV("child", imgui.NewVec2(0, 0), imgui.ChildFlagsNone, imgui.WindowFlagsNone) {
    // content
    imgui.EndChild()
}

// CORRECT
if imgui.BeginChildStrV("child", imgui.NewVec2(0, 0), imgui.ChildFlagsNone, imgui.WindowFlagsNone) {
    // content
}
imgui.EndChild()  // Always called
```

**3. Zero-size child with AutoResize**

Don't combine `(0, 0)` size with `AutoResizeX` and `AutoResizeY`:

```go
// Defeats purpose - auto-resize on both axes means no scrolling
imgui.BeginChildStrV("bad", imgui.NewVec2(0, 0),
    imgui.ChildFlagsAutoResizeX|imgui.ChildFlagsAutoResizeY,
    imgui.WindowFlagsNone)
```

**4. Negative size meaning**

Remember: negative size is distance FROM the edge, not a negative dimension:

```go
// This creates a 100px margin from the right, NOT a -100px width
imgui.BeginChildStrV("child", imgui.NewVec2(-100, 0), imgui.ChildFlagsNone, imgui.WindowFlagsNone)
```

---

## dfx Integration Notes

### Using state.Size

In dfx components, `state.Size` provides the available drawing area:

```go
func (c *MyComponent) Draw(state *dfx.State) {
    // state.Size = available space for this component
    imgui.Text(fmt.Sprintf("Available: %.0f x %.0f",
        state.Size.X, state.Size.Y))

    // Use for proportional layouts
    halfWidth := state.Size.X * 0.5
}
```

### Relationship to ContentRegionAvail

`state.Size` and `ContentRegionAvail()` are related but not identical:
- `state.Size` is the space allocated to the component by dfx
- `ContentRegionAvail()` is the remaining space at the current cursor position

dfx's own containers allocate the usable area at the component's drawing origin, so at the start of a component they're equal. As you add widgets, `ContentRegionAvail()` decreases.

### How dfx allocates

Each boundary has one owner for its padding, and a child receives what is left after it:

- **Root.** The root window fills the native client area below the main menu, with square corners. Its padding is the selected style's `Spacing.OuterMargin` (4 px in Classic, 12 px in Faceplate), an inset of content rather than a second window. The root component's `state.Size` is the content region inside that margin; `Position` stays `(0, 0)`, relative to the window.
- **Workspace.** The current item's height is what remains after the selector's actual cursor advance, including the spacing below it. Under a style with `Chrome.ContentPadding` the selector is inset by that padding, like a padded cell, and is followed by item spacing only: padded content below supplies the rest of the separation, so the gaps above and below the selector match. With no content padding (Classic) the selector sits at the origin and adds a line of spacing below, as before.
- **Dash and DashManager.** Dash bounds are relative to the cursor when the dash is drawn: the manager's content origin, not the enclosing window's outer edge. A dash's surface is inset by the panel padding on every side and by its handle row once; its component gets the surface's content region. The manager's inner component gets the inner panel's content region. `DashManager.Margin` remains explicit. `TopMargin` reserves space above the whole manager layout under either precedence: the origin moves down by it and the available height loses it once, so every dash and the inner panel lie in the remainder below it. A dash's panel is clamped to its bounds, so a top or bottom dash taller than the remainder is shortened rather than reaching into the inset.
- **HCollapse.** The expanded header is measured from the toggle's frame height and the panel padding (42 px in Faceplate, 36 px in Classic). Content starts below it, padded by the style's `Chrome.ContentPadding`, and gets the content region inside that padding. A collapsed toggle is centered in its strip with its frame padding narrowed to fit.
- **MultiGrid.** FlexLayout subtracts exactly the 10 px splitter strip it draws; GridLayout places cells from its content origin, takes `Chrome.GridGap` inside interior cell edges, and clips cells to its allocation. Both pad cells by `Chrome.ContentPadding` and pass the padded content region, except a cell holding a nested MultiGrid: its own cells pad their content, so padding the outer cell too would inset nested content twice and out of line with its siblings.
- **Toolbar.** The bar registers its full width and height as one item, so the layout continues one item spacing below it.

Allocations are never negative. An exhausted allocation is skipped rather than passed to `BeginChild`, where a zero size would mean fill-remaining and a negative one fill-relative: a dash its neighbors or a `TopMargin` leave no room is not drawn and cannot hold focus, though its show/hide animation keeps advancing. A positive panel or cell whose padding leaves no usable area is drawn empty; its component is not called with a zero size.

---

## Quick Reference

### Size Parameter Cheat Sheet

| Value | X Behavior | Y Behavior |
|-------|------------|------------|
| `(0, 0)` | Fill width | Fill height |
| `(300, 200)` | 300px | 200px |
| `(-50, 0)` | 50px from right | Fill height |
| `(0, -100)` | Fill width | 100px from bottom |
| `(-FLT_MIN, 0)` | Fill exactly | Fill height |

### Common Functions

| Function | Purpose |
|----------|---------|
| `ContentRegionAvail()` | Remaining space (use this!) |
| `CursorPos()` | Current position (local) |
| `SetCursorPos(pos)` | Jump cursor |
| `SameLine()` | Continue horizontally |
| `PushItemWidth(w)` | Set default width |
| `SetNextItemWidth(w)` | One-time width |
| `BeginChildStrV(id, size, childFlags, windowFlags)` | Create child region |
| `BeginGroup()` | Start grouping |

### cimgui-go Imports

```go
import (
    "github.com/AllenDang/cimgui-go/imgui"
    "math"  // For math.SmallestNonzeroFloat32
)
```
