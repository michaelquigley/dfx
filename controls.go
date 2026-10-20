package dfx

import (
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
)

// control constants
const (
	toggleInactiveAlpha = 0.1  // alpha for inactive toggle buttons
	wheelMultiplierFast = 10.0 // ctrl modifier for wheel slider
	wheelMultiplierSlow = 10.0 // alt modifier for wheel slider (divisor)
)

// Controls provides simplified wrappers for imgui widgets that add genuine value.
// trivial pass-through wrappers have been removed; use imgui.* directly for
// Button, Text, Separator, SameLine, Spacing, TreeNode, TreePop, BeginChild,
// EndChild, BeginMenu, EndMenu, BeginMenuBar, EndMenuBar, MenuItem.

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

// Toggle creates a button that acts as a boolean toggle.
// when inactive (false), the button is dimmed. when active (true), it uses the checkmark color.
// returns (newValue, changed) following dfx conventions.
func Toggle(label string, value bool) (bool, bool) {
	return ToggleV(label, value, ToggleOptions{})
}

// ToggleOptions adjusts a single ToggleV; zero values keep the style's defaults.
type ToggleOptions struct {
	OnText *imgui.Vec4 // label color while on, ahead of the style's default
}

// ToggleV is Toggle with options. with the selected style's
// StyleControls.AccentInk, an on toggle keeps its fill when hovered or
// pressed and draws its label in whichever of the theme's background and text
// colors contrasts more with the fill.
func ToggleV(label string, value bool, opts ToggleOptions) (bool, bool) {
	pushed := 0
	if !value {
		// inactive: dim the button
		buttonColor := styleColor(imgui.ColButton)
		buttonColor.W = toggleInactiveAlpha
		imgui.PushStyleColorVec4(imgui.ColButton, buttonColor)
		pushed++
	} else {
		// active: use checkmark color
		fill := styleColor(imgui.ColCheckMark)
		imgui.PushStyleColorVec4(imgui.ColButton, fill)
		pushed++
		text := opts.OnText
		if currentStyleControls().AccentInk {
			if text == nil {
				ink := contrastingInk(fill, styleColor(imgui.ColWindowBg), styleColor(imgui.ColText))
				text = &ink
			}
			// hover and press move the fill away from the label, so neither state
			// can read worse than the toggle at rest
			away := imgui.Vec4{W: fill.W}
			if relativeLuminance(*text) < relativeLuminance(fill) {
				away = imgui.Vec4{X: 1, Y: 1, Z: 1, W: fill.W}
			}
			imgui.PushStyleColorVec4(imgui.ColButtonHovered, mix(fill, away, 0.12))
			imgui.PushStyleColorVec4(imgui.ColButtonActive, mix(fill, away, 0.24))
			pushed += 2
		}
		if text != nil {
			imgui.PushStyleColorVec4(imgui.ColText, *text)
			pushed++
		}
	}
	defer imgui.PopStyleColorV(int32(pushed))

	// render button and toggle on click
	if imgui.Button(label) {
		return !value, true // newValue, changed
	}
	return value, false // no change
}

// contrastingInk returns whichever of dark and light has the higher WCAG contrast
// ratio against fill, opaque. it chooses between two theme colors; it does not
// invent a color.
func contrastingInk(fill, dark, light imgui.Vec4) imgui.Vec4 {
	f := relativeLuminance(fill)
	ratio := func(c imgui.Vec4) float64 {
		l := relativeLuminance(c)
		return (max(f, l) + 0.05) / (min(f, l) + 0.05)
	}
	ink := light
	if ratio(dark) > ratio(light) {
		ink = dark
	}
	ink.W = 1
	return ink
}

func relativeLuminance(c imgui.Vec4) float64 {
	channel := func(v float32) float64 {
		x := float64(v)
		if x <= 0.04045 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(c.X) + 0.7152*channel(c.Y) + 0.0722*channel(c.Z)
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
