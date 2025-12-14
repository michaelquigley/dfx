package main

import (
	"fmt"
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx"
)

// This example demonstrates Dear ImGui's layout and sizing system.
// Each collapsible section shows a different layout concept with
// real-time values and interactive examples.
//
// See docs/LAYOUT_GUIDE.md for comprehensive documentation.

func main() {
	root := dfx.NewFunc(func(state *dfx.State) {
		// Show dfx state.Size at the top
		imgui.Text("dfx Layout & Sizing Demo")
		imgui.Separator()
		imgui.Text(fmt.Sprintf("state.Size: (%.0f, %.0f)", state.Size.X, state.Size.Y))
		imgui.Spacing()

		// Each demo in its own collapsing header
		cursorDemo()
		spacingDemo()
		widthControlDemo()
		sameLineDemo()
		groupDemo()
		childWindowDemo()
		columnsDemo()
		tablesDemo()
		practicalPatternsDemo()
	})

	app := dfx.New(root, dfx.Config{
		Title:  "Layout & Sizing Demo",
		Width:  900,
		Height: 700,
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}

// =============================================================================
// Demo 1: Cursor Position System
// =============================================================================

func cursorDemo() {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("1. Cursor Position System", imgui.TreeNodeFlagsNone) {
		return
	}

	imgui.Indent()
	defer imgui.Unindent()

	// Show the difference between local and screen coordinates
	localPos := imgui.CursorPos()
	screenPos := imgui.CursorScreenPos()
	avail := imgui.ContentRegionAvail()

	imgui.Text("Position Comparison:")
	imgui.BulletText(fmt.Sprintf("GetCursorPos() [local]:  (%.0f, %.0f)", localPos.X, localPos.Y))
	imgui.BulletText(fmt.Sprintf("GetCursorScreenPos() [screen]: (%.0f, %.0f)", screenPos.X, screenPos.Y))
	imgui.BulletText(fmt.Sprintf("GetContentRegionAvail(): (%.0f, %.0f)", avail.X, avail.Y))

	imgui.Spacing()

	// Demonstrate SetCursorPos
	imgui.Text("Manual Positioning with SetCursorPos:")

	// Save current position
	startY := imgui.CursorPosY()

	// Draw at offset positions
	imgui.SetCursorPos(imgui.NewVec2(50, startY+20))
	imgui.Text("Positioned at (50, +20)")

	imgui.SetCursorPos(imgui.NewVec2(200, startY+20))
	imgui.Text("Positioned at (200, +20)")

	imgui.SetCursorPos(imgui.NewVec2(350, startY+40))
	imgui.Text("Positioned at (350, +40)")

	// Reset cursor below our manually positioned content
	imgui.SetCursorPosY(startY + 80)

	imgui.Spacing()
	imgui.TextWrapped("Note: GetCursorPos() returns window-local coordinates. " +
		"GetCursorScreenPos() returns absolute screen coordinates. " +
		"Use local coords for layout, screen coords for custom drawing.")
}

// =============================================================================
// Demo 2: Spacing and Padding
// =============================================================================

func spacingDemo() {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("2. Spacing and Padding", imgui.TreeNodeFlagsNone) {
		return
	}

	imgui.Indent()
	defer imgui.Unindent()

	style := imgui.CurrentStyle()

	imgui.Text("Current Style Values:")
	imgui.BulletText(fmt.Sprintf("WindowPadding: (%.0f, %.0f)",
		style.WindowPadding().X, style.WindowPadding().Y))
	imgui.BulletText(fmt.Sprintf("FramePadding:  (%.0f, %.0f)",
		style.FramePadding().X, style.FramePadding().Y))
	imgui.BulletText(fmt.Sprintf("ItemSpacing:   (%.0f, %.0f)",
		style.ItemSpacing().X, style.ItemSpacing().Y))

	imgui.Spacing()
	imgui.Text("Derived Measurements:")
	imgui.BulletText(fmt.Sprintf("GetFrameHeight(): %.0f (FontSize + FramePadding.y*2)",
		imgui.FrameHeight()))
	imgui.BulletText(fmt.Sprintf("GetFrameHeightWithSpacing(): %.0f (+ ItemSpacing.y)",
		imgui.FrameHeightWithSpacing()))
	imgui.BulletText(fmt.Sprintf("GetTextLineHeight(): %.0f (FontSize only)",
		imgui.TextLineHeight()))

	imgui.Spacing()
	imgui.Separator()
	imgui.Text("Visual Spacing Demo:")
	imgui.Spacing()

	// Show widgets with default spacing
	imgui.Button("Button 1")
	imgui.Button("Button 2")
	imgui.Button("Button 3")

	imgui.TextWrapped("The buttons above are separated by ItemSpacing.y pixels. " +
		"Each button's height is FontSize + FramePadding.y*2.")
}

// =============================================================================
// Demo 3: Widget Width Control
// =============================================================================

var (
	demoInput1 = "Normal"
	demoInput2 = "200px fixed"
	demoInput3 = "Fill to edge"
	demoInput4 = "-100 (margin)"
)

func widthControlDemo() {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("3. Widget Width Control", imgui.TreeNodeFlagsNone) {
		return
	}

	imgui.Indent()
	defer imgui.Unindent()

	imgui.Text("Width Semantics:")
	imgui.BulletText("Positive: explicit pixel width")
	imgui.BulletText("Zero: use default (65% of window)")
	imgui.BulletText("Negative: distance from right edge")
	imgui.Spacing()

	// Default width
	imgui.Text("Default Width (no specification):")
	imgui.InputTextWithHint("##default", "", &demoInput1, imgui.InputTextFlagsNone, nil)

	// Fixed width
	imgui.Text("SetNextItemWidth(200):")
	imgui.SetNextItemWidth(200)
	imgui.InputTextWithHint("##fixed", "", &demoInput2, imgui.InputTextFlagsNone, nil)

	// Fill to edge using smallest nonzero float
	imgui.Text("SetNextItemWidth(-SmallestNonzeroFloat32) - fills to edge:")
	imgui.SetNextItemWidth(-float32(math.SmallestNonzeroFloat32))
	imgui.InputTextWithHint("##fill", "", &demoInput3, imgui.InputTextFlagsNone, nil)

	// Negative value with margin
	imgui.Text("SetNextItemWidth(-100) - leaves 100px margin:")
	imgui.SetNextItemWidth(-100)
	imgui.InputTextWithHint("##margin", "", &demoInput4, imgui.InputTextFlagsNone, nil)

	imgui.Spacing()

	// Practical pattern: input with button
	imgui.Text("Practical: Input + Button pattern:")
	imgui.SetNextItemWidth(-80) // Leave room for button
	imgui.InputTextWithHint("##withbutton", "", &demoInput1, imgui.InputTextFlagsNone, nil)
	imgui.SameLine()
	imgui.Button("Browse")

	imgui.Spacing()
	imgui.TextWrapped("The -80 leaves exactly enough room for the button. " +
		"This pattern adapts to any window width!")
}

// =============================================================================
// Demo 4: Horizontal Layout with SameLine
// =============================================================================

func sameLineDemo() {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("4. Horizontal Layout (SameLine)", imgui.TreeNodeFlagsNone) {
		return
	}

	imgui.Indent()
	defer imgui.Unindent()

	imgui.Text("Default vertical stacking:")
	imgui.Button("A")
	imgui.Button("B")
	imgui.Button("C")

	imgui.Spacing()
	imgui.Text("With SameLine():")
	imgui.Button("A")
	imgui.SameLine()
	imgui.Button("B")
	imgui.SameLine()
	imgui.Button("C")

	imgui.Spacing()
	imgui.Text("SameLine with custom spacing (20px):")
	imgui.Button("A")
	imgui.SameLineV(0, 20)
	imgui.Button("B")
	imgui.SameLineV(0, 20)
	imgui.Button("C")

	imgui.Spacing()
	imgui.Text("SameLine with absolute positioning:")
	imgui.Button("At 0")
	imgui.SameLineV(150, 0)
	imgui.Button("At 150")
	imgui.SameLineV(300, 0)
	imgui.Button("At 300")

	imgui.Spacing()
	imgui.TextWrapped("SameLine() cancels the automatic line break, placing the " +
		"next widget to the right of the previous one. The first parameter " +
		"is absolute X position (0 = relative), second is spacing override.")
}

// =============================================================================
// Demo 5: Groups
// =============================================================================

func groupDemo() {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("5. Groups (BeginGroup/EndGroup)", imgui.TreeNodeFlagsNone) {
		return
	}

	imgui.Indent()
	defer imgui.Unindent()

	imgui.Text("Groups treat multiple widgets as a single item:")
	imgui.Spacing()

	// Example: Labels on left, inputs on right
	imgui.BeginGroup()
	imgui.Text("Name:")
	imgui.Text("Email:")
	imgui.Text("Phone:")
	imgui.EndGroup()

	groupSize := imgui.ItemRectSize()

	imgui.SameLine()

	imgui.BeginGroup()
	name := "John Doe"
	email := "john@example.com"
	phone := "555-1234"
	imgui.SetNextItemWidth(200)
	imgui.InputTextWithHint("##name", "", &name, imgui.InputTextFlagsNone, nil)
	imgui.SetNextItemWidth(200)
	imgui.InputTextWithHint("##email", "", &email, imgui.InputTextFlagsNone, nil)
	imgui.SetNextItemWidth(200)
	imgui.InputTextWithHint("##phone", "", &phone, imgui.InputTextFlagsNone, nil)
	imgui.EndGroup()

	imgui.Spacing()
	imgui.Text(fmt.Sprintf("Left group size: (%.0f, %.0f)", groupSize.X, groupSize.Y))

	imgui.Spacing()
	imgui.Separator()
	imgui.Text("Getting group bounds after EndGroup:")

	imgui.BeginGroup()
	imgui.Text("Bounded")
	imgui.Text("Content")
	imgui.Button("Inside Group")
	imgui.EndGroup()

	min := imgui.ItemRectMin()
	max := imgui.ItemRectMax()
	size := imgui.ItemRectSize()

	// Draw a rect around the group to visualize bounds
	drawList := imgui.WindowDrawList()
	drawList.AddRect(min, max, imgui.ColorConvertFloat4ToU32(imgui.NewVec4(1, 1, 0, 1)))

	imgui.Text(fmt.Sprintf("Min: (%.0f, %.0f) Max: (%.0f, %.0f) Size: (%.0f, %.0f)",
		min.X, min.Y, max.X, max.Y, size.X, size.Y))
}

// =============================================================================
// Demo 6: Child Windows
// =============================================================================

var childDemoSelectedSize = 0

func childWindowDemo() {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("6. Child Windows", imgui.TreeNodeFlagsNone) {
		return
	}

	imgui.Indent()
	defer imgui.Unindent()

	imgui.Text("Size Parameter Semantics:")
	imgui.BulletText("(0, 0) = Fill remaining space")
	imgui.BulletText("(300, 200) = Fixed 300x200 pixels")
	imgui.BulletText("(-100, 0) = 100px from right edge, full height")
	imgui.BulletText("(0, -50) = Full width, 50px from bottom")
	imgui.Spacing()

	options := []string{"Fixed 300x150", "Fill remaining", "With margins"}
	if newIdx, changed := dfx.Combo("Size Mode", childDemoSelectedSize, options); changed {
		childDemoSelectedSize = newIdx
	}

	imgui.Spacing()

	var childSize imgui.Vec2
	switch childDemoSelectedSize {
	case 0:
		childSize = imgui.NewVec2(300, 150)
	case 1:
		childSize = imgui.NewVec2(0, 150)
	case 2:
		childSize = imgui.NewVec2(-50, 150)
	}

	imgui.Text(fmt.Sprintf("BeginChild size: (%.0f, %.0f)", childSize.X, childSize.Y))

	if imgui.BeginChildStrV("demo_child", childSize, imgui.ChildFlagsBorders, imgui.WindowFlagsNone) {
		for i := 0; i < 20; i++ {
			imgui.Text(fmt.Sprintf("Scrollable item %d", i))
		}
	}
	imgui.EndChild()

	imgui.Spacing()
	imgui.Separator()
	imgui.Text("Side-by-side children (common pattern):")
	imgui.Spacing()

	// Two children side by side
	if imgui.BeginChildStrV("left_pane", imgui.NewVec2(200, 100), imgui.ChildFlagsBorders, imgui.WindowFlagsNone) {
		imgui.Text("Left Pane")
		imgui.Text("Fixed 200px width")
	}
	imgui.EndChild()

	imgui.SameLine()

	if imgui.BeginChildStrV("right_pane", imgui.NewVec2(0, 100), imgui.ChildFlagsBorders, imgui.WindowFlagsNone) {
		imgui.Text("Right Pane")
		imgui.Text("Fills remaining width")
		avail := imgui.ContentRegionAvail()
		imgui.Text(fmt.Sprintf("Available: %.0f x %.0f", avail.X, avail.Y))
	}
	imgui.EndChild()
}

// =============================================================================
// Demo 7: Columns (Legacy)
// =============================================================================

func columnsDemo() {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("7. Columns (Legacy API)", imgui.TreeNodeFlagsNone) {
		return
	}

	imgui.Indent()
	defer imgui.Unindent()

	imgui.TextColored(imgui.NewVec4(1, 0.8, 0, 1), "Note: Columns API is deprecated. Use Tables instead!")
	imgui.Spacing()

	imgui.Text("Basic 3-column layout:")
	imgui.Separator()

	imgui.ColumnsV(3, "legacy_cols", true)

	imgui.Text("Column 1")
	imgui.NextColumn()
	imgui.Text("Column 2")
	imgui.NextColumn()
	imgui.Text("Column 3")
	imgui.NextColumn()

	imgui.Text("Row 2, A")
	imgui.NextColumn()
	imgui.Text("Row 2, B")
	imgui.NextColumn()
	imgui.Text("Row 2, C")

	imgui.Columns() // End columns

	imgui.Spacing()
	imgui.TextWrapped("Columns lack: sizing policies, sorting, hiding, reordering. " +
		"See Tables demo for the modern approach.")
}

// =============================================================================
// Demo 8: Tables
// =============================================================================

var tableSizingPolicy = 0

func tablesDemo() {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("8. Tables (Modern API)", imgui.TreeNodeFlagsNone) {
		return
	}

	imgui.Indent()
	defer imgui.Unindent()

	imgui.Text("Table Sizing Policies:")
	policies := []string{
		"StretchSame (default)",
		"StretchProp",
		"FixedFit",
		"FixedSame",
	}
	if newIdx, changed := dfx.Combo("Policy", tableSizingPolicy, policies); changed {
		tableSizingPolicy = newIdx
	}

	imgui.Spacing()

	var flags imgui.TableFlags = imgui.TableFlagsBorders | imgui.TableFlagsRowBg
	switch tableSizingPolicy {
	case 0:
		flags |= imgui.TableFlagsSizingStretchSame
	case 1:
		flags |= imgui.TableFlagsSizingStretchProp
	case 2:
		flags |= imgui.TableFlagsSizingFixedFit
	case 3:
		flags |= imgui.TableFlagsSizingFixedSame
	}

	if imgui.BeginTableV("demo_table", 3, flags, imgui.NewVec2(0, 150), 0) {
		imgui.TableSetupColumn("Name")
		imgui.TableSetupColumn("Description")
		imgui.TableSetupColumn("Price")
		imgui.TableHeadersRow()

		data := [][]string{
			{"Widget A", "Small widget", "$10"},
			{"Widget B", "Medium widget with longer description", "$25"},
			{"Widget C", "Large", "$50"},
			{"Widget D", "Extra large widget", "$100"},
		}

		for _, row := range data {
			imgui.TableNextRow()
			for col, cell := range row {
				imgui.TableSetColumnIndex(int32(col))
				imgui.Text(cell)
			}
		}

		imgui.EndTable()
	}

	imgui.Spacing()
	imgui.Text("Column Weight Demo (2:1 ratio):")

	if imgui.BeginTableV("weight_table", 2, imgui.TableFlagsBorders|imgui.TableFlagsSizingStretchProp, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumnV("Wide Column", imgui.TableColumnFlagsWidthStretch, 2.0, 0)
		imgui.TableSetupColumnV("Narrow Column", imgui.TableColumnFlagsWidthStretch, 1.0, 0)
		imgui.TableHeadersRow()

		imgui.TableNextRow()
		imgui.TableNextColumn()
		imgui.Text("2x weight")
		imgui.TableNextColumn()
		imgui.Text("1x weight")

		imgui.EndTable()
	}

	imgui.Spacing()
	imgui.Text("Fixed + Stretch Mixed:")

	if imgui.BeginTableV("mixed_table", 3, imgui.TableFlagsBorders, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumnV("Icon", imgui.TableColumnFlagsWidthFixed, 50, 0)
		imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthStretch, 0, 0)
		imgui.TableSetupColumnV("Actions", imgui.TableColumnFlagsWidthFixed, 80, 0)
		imgui.TableHeadersRow()

		imgui.TableNextRow()
		imgui.TableNextColumn()
		imgui.Text("[*]")
		imgui.TableNextColumn()
		imgui.Text("Stretches to fill remaining space")
		imgui.TableNextColumn()
		imgui.Button("Edit")

		imgui.EndTable()
	}
}

// =============================================================================
// Demo 9: Practical Patterns
// =============================================================================

func practicalPatternsDemo() {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("9. Practical Patterns", imgui.TreeNodeFlagsNone) {
		return
	}

	imgui.Indent()
	defer imgui.Unindent()

	// Pattern 1: Right-aligned button
	imgui.Text("Pattern 1: Right-aligned button")
	avail := imgui.ContentRegionAvail()
	style := imgui.CurrentStyle()
	buttonWidth := imgui.CalcTextSizeV("Right Aligned", false, 0).X + style.FramePadding().X*2
	imgui.SetCursorPosX(imgui.CursorPosX() + avail.X - buttonWidth)
	imgui.Button("Right Aligned")

	imgui.Spacing()
	imgui.Separator()

	// Pattern 2: Proportional split
	imgui.Text("Pattern 2: Proportional 30/70 split")
	avail = imgui.ContentRegionAvail()

	if imgui.BeginChildStrV("prop_left", imgui.NewVec2(avail.X*0.3, 60), imgui.ChildFlagsBorders, imgui.WindowFlagsNone) {
		imgui.Text("30% width")
	}
	imgui.EndChild()

	imgui.SameLine()

	if imgui.BeginChildStrV("prop_right", imgui.NewVec2(0, 60), imgui.ChildFlagsBorders, imgui.WindowFlagsNone) {
		imgui.Text("70% width (remaining)")
	}
	imgui.EndChild()

	imgui.Spacing()
	imgui.Separator()

	// Pattern 3: Header + Scrollable + Footer
	imgui.Text("Pattern 3: Fixed header/footer with scrollable middle")

	headerHeight := imgui.FrameHeightWithSpacing()
	footerHeight := imgui.FrameHeightWithSpacing()

	// Header
	imgui.Text("Fixed Header")
	imgui.Separator()

	// Scrollable middle - subtract header, footer, and separators
	if imgui.BeginChildStrV("scroll_middle", imgui.NewVec2(0, 80), imgui.ChildFlagsNone, imgui.WindowFlagsNone) {
		for i := 0; i < 20; i++ {
			imgui.Text(fmt.Sprintf("Scrollable content line %d", i))
		}
	}
	imgui.EndChild()

	// Footer
	imgui.Separator()
	imgui.Text("Fixed Footer")

	_ = headerHeight
	_ = footerHeight

	imgui.Spacing()
	imgui.Separator()

	// Pattern 4: Toolbar
	imgui.Text("Pattern 4: Toolbar with fill space")

	imgui.Button("File")
	imgui.SameLine()
	imgui.Button("Edit")
	imgui.SameLine()
	imgui.Button("View")
	imgui.SameLine()

	// Spacer pushes remaining items to the right
	avail = imgui.ContentRegionAvail()
	imgui.Dummy(imgui.NewVec2(avail.X-100, 0))
	imgui.SameLine()

	imgui.Button("Help")
	imgui.SameLine()
	imgui.Button("About")
}
