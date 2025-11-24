package dfx

import (
	"github.com/AllenDang/cimgui-go/imgui"
)

// Fader creates a vertical slider (fader) with mouse wheel support and right-click reset.
// designed for audio mixing applications. displays current value above the fader.
// range is fixed at 0.0 to 1.0. right-click to reset to defaultValue.
// wheelSteps controls sensitivity (larger value = smaller adjustments per wheel tick).
// modifiers: Ctrl = 10x faster, Alt = 10x slower.
// returns (newValue, changed) following dfx conventions.
func Fader(label string, value, resetValue, width, height float32) (float32, bool) {
	// draw vertical slider
	newValue := value
	size := imgui.Vec2{X: width, Y: height}
	changed := imgui.VSliderFloatV(label, size, &newValue, 0.0, 1.0, "", imgui.SliderFlagsNone)

	// handle right-click reset
	if imgui.IsItemHovered() && imgui.IsMouseClickedBool(imgui.MouseButtonRight) {
		newValue = resetValue
		if newValue != value {
			changed = true
		}
	}

	// handle mouse wheel when hovering
	if imgui.IsItemHovered() {
		wheel := imgui.CurrentIO().MouseWheel()
		if wheel != 0 {
			// clear active state if slider is being dragged
			if imgui.IsItemActive() {
				imgui.InternalClearActiveID()
			}

			// calculate adjustment amount (using 100 steps by default for fine control)
			const wheelSteps = 100.0
			fraction := float32(1.0 / wheelSteps)

			// apply modifiers
			if imgui.CurrentIO().KeyCtrl() {
				fraction *= wheelMultiplierFast
			} else if imgui.CurrentIO().KeyAlt() {
				fraction /= wheelMultiplierSlow
			}

			// adjust value
			newValue += wheel * fraction

			// clamp to 0.0-1.0 range
			if newValue < 0.0 {
				newValue = 0.0
			}
			if newValue > 1.0 {
				newValue = 1.0
			}

			// mark as changed if value actually changed
			if newValue != value {
				changed = true
			}
		}
	}

	return newValue, changed
}
