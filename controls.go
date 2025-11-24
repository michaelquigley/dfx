package dfx

import (
	"github.com/AllenDang/cimgui-go/imgui"
)

// control constants
const (
	toggleInactiveAlpha = 0.1  // alpha for inactive toggle buttons
	wheelMultiplierFast = 10.0 // ctrl modifier for wheel slider
	wheelMultiplierSlow = 10.0 // alt modifier for wheel slider (divisor)
)

// Controls provides simplified wrappers for common imgui widgets.
// these return values directly instead of requiring pointers.

// Button creates a simple button
func Button(label string) bool {
	return imgui.Button(label)
}

// ButtonSize creates a button with specific size
func ButtonSize(label string, width, height float32) bool {
	return imgui.ButtonV(label, imgui.Vec2{X: width, Y: height})
}

// Text displays static text
func Text(text string) {
	imgui.Text(text)
}

// TextColored displays colored text
func TextColored(text string, r, g, b, a float32) {
	imgui.TextColored(imgui.Vec4{X: r, Y: g, Z: b, W: a}, text)
}

// Input is a simplified text input that returns the new value and whether it changed
func Input(label string, value string) (string, bool) {
	// imgui expects a mutable string buffer
	buf := value
	changed := imgui.InputTextWithHint(label, "", &buf, imgui.InputTextFlagsNone, nil)
	return buf, changed
}

// InputMultiline is a multiline text input
func InputMultiline(label string, value string, width, height float32) (string, bool) {
	buf := value
	size := imgui.Vec2{X: width, Y: height}
	changed := imgui.InputTextMultiline(label, &buf, size, imgui.InputTextFlagsNone, nil)
	return buf, changed
}

// Checkbox returns new state and whether it changed
func Checkbox(label string, checked bool) (bool, bool) {
	old := checked
	imgui.Checkbox(label, &checked)
	return checked, checked != old
}

// Slider returns new value and whether it changed
func Slider(label string, value float32, min, max float32) (float32, bool) {
	old := value
	imgui.SliderFloat(label, &value, min, max)
	return value, value != old
}

// SliderInt returns new value and whether it changed
func SliderInt(label string, value int, min, max int) (int, bool) {
	old := value
	v := int32(value)
	imgui.SliderInt(label, &v, int32(min), int32(max))
	value = int(v)
	return value, value != old
}

// Combo creates a dropdown. Returns selected index and whether it changed.
func Combo(label string, current int, items []string) (int, bool) {
	if len(items) == 0 {
		return current, false
	}

	// ensure current is valid
	if current < 0 || current >= len(items) {
		current = 0
	}

	preview := items[current]
	if !imgui.BeginCombo(label, preview) {
		return current, false
	}
	defer imgui.EndCombo()

	newIndex := current
	for i, item := range items {
		selected := i == current
		if imgui.SelectableBoolV(item, selected, 0, imgui.Vec2{}) {
			newIndex = i
		}
		if selected {
			imgui.SetItemDefaultFocus()
		}
	}

	return newIndex, newIndex != current
}

// ColorEdit3 edits RGB color. Returns new color and whether it changed.
func ColorEdit3(label string, r, g, b float32) (float32, float32, float32, bool) {
	col := [3]float32{r, g, b}
	changed := imgui.ColorEdit3(label, &col)
	return col[0], col[1], col[2], changed
}

// ColorEdit4 edits RGBA color. Returns new color and whether it changed.
func ColorEdit4(label string, r, g, b, a float32) (float32, float32, float32, float32, bool) {
	col := [4]float32{r, g, b, a}
	changed := imgui.ColorEdit4(label, &col)
	return col[0], col[1], col[2], col[3], changed
}

// Separator draws a horizontal line
func Separator() {
	imgui.Separator()
}

// SameLine places next widget on same line
func SameLine() {
	imgui.SameLine()
}

// Spacing adds vertical spacing
func Spacing() {
	imgui.Spacing()
}

// BeginChild starts a scrollable region
func BeginChild(id string, width, height float32, border bool) bool {
	flags := imgui.ChildFlagsNone
	if border {
		flags |= imgui.ChildFlagsBorders
	}
	return imgui.BeginChildStrV(id, imgui.Vec2{X: width, Y: height}, flags, imgui.WindowFlagsNone)
}

// EndChild ends a scrollable region
func EndChild() {
	imgui.EndChild()
}

// TreeNode creates an expandable tree node. Returns whether it's open.
func TreeNode(label string) bool {
	return imgui.TreeNodeStr(label)
}

// TreePop closes a tree node
func TreePop() {
	imgui.TreePop()
}

// Tooltip shows a tooltip when hovering
func Tooltip(text string) {
	if imgui.IsItemHovered() {
		imgui.SetTooltip(text)
	}
}

// MenuItem creates a menu item. Returns whether it was clicked.
func MenuItem(label string, shortcut string) bool {
	return imgui.MenuItemBoolV(label, shortcut, false, true)
}

// BeginMenu starts a submenu. Returns whether it's open.
func BeginMenu(label string) bool {
	return imgui.BeginMenu(label)
}

// EndMenu ends a submenu
func EndMenu() {
	imgui.EndMenu()
}

// BeginMenuBar starts a menu bar
func BeginMenuBar() bool {
	return imgui.BeginMenuBar()
}

// EndMenuBar ends a menu bar
func EndMenuBar() {
	imgui.EndMenuBar()
}

// Toggle creates a button that acts as a boolean toggle.
// when inactive (false), the button is dimmed. when active (true), it uses the checkmark color.
// returns (newValue, changed) following dfx conventions.
func Toggle(label string, value bool) (bool, bool) {
	// set button color based on state
	if !value {
		// inactive: dim the button
		buttonColor := imgui.CurrentStyle().Colors()[imgui.ColButton]
		buttonColor.W = toggleInactiveAlpha
		imgui.PushStyleColorVec4(imgui.ColButton, buttonColor)
	} else {
		// active: use checkmark color
		imgui.PushStyleColorVec4(imgui.ColButton, imgui.CurrentStyle().Colors()[imgui.ColCheckMark])
	}
	defer imgui.PopStyleColor()

	// render button and toggle on click
	if imgui.Button(label) {
		return !value, true // newValue, changed
	}
	return value, false // no change
}

// WheelSlider creates a slider that responds to mouse wheel when hovered.
// wheelSteps controls sensitivity (larger value = smaller adjustments per wheel tick).
// modifiers: Ctrl = 10x faster, Alt = 10x slower.
// returns (newValue, changed) following dfx conventions.
func WheelSlider(label string, value, min, max, wheelSteps float32, format string, flags imgui.SliderFlags) (float32, bool) {
	// draw normal slider
	newValue := value
	changed := imgui.SliderFloatV(label, &newValue, min, max, format, flags)

	// handle mouse wheel when hovering
	if imgui.IsItemHovered() {
		wheel := imgui.CurrentIO().MouseWheel()
		if wheel != 0 {
			// clear active state if slider is being dragged
			if imgui.IsItemActive() {
				imgui.InternalClearActiveID()
			}

			// calculate adjustment amount
			fraction := (max - min) / wheelSteps

			// apply modifiers
			if imgui.CurrentIO().KeyCtrl() {
				fraction *= wheelMultiplierFast
			} else if imgui.CurrentIO().KeyAlt() {
				fraction /= wheelMultiplierSlow
			}

			// adjust value
			newValue += wheel * fraction

			// clamp to range
			if newValue < min {
				newValue = min
			}
			if newValue > max {
				newValue = max
			}

			// mark as changed if value actually changed
			if newValue != value {
				changed = true
			}
		}
	}

	return newValue, changed
}
