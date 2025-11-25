package dfx

import (
	"fmt"
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
)

// ============================================================================
// Helper Functions
// ============================================================================

// clamp restricts a value to the range [min, max]
func clamp(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ============================================================================
// Taper Interface and Implementations
// ============================================================================

// Taper defines non-linear response curves for faders.
// Tapers affect the UI feel without changing the underlying value range.
type Taper interface {
	// Apply taper: normalized (0-1) -> tapered (0-1) for UI positioning
	Apply(normalized float32) float32

	// Invert taper: tapered (0-1) -> normalized (0-1) from UI position
	Invert(tapered float32) float32
}

// linearTaper implements a 1:1 linear mapping (no taper).
type linearTaper struct{}

func (t linearTaper) Apply(normalized float32) float32 {
	return normalized
}

func (t linearTaper) Invert(tapered float32) float32 {
	return tapered
}

// LinearTaper returns a taper with no curve (1:1 mapping).
func LinearTaper() Taper {
	return linearTaper{}
}

// logTaper implements a logarithmic taper curve.
// Common for audio level controls and frequency controls.
type logTaper struct {
	steepness float32
}

func (t logTaper) Apply(normalized float32) float32 {
	if normalized <= 0 {
		return 0
	}
	if normalized >= 1 {
		return 1
	}
	// Logarithmic mapping: y = (e^(s*x) - 1) / (e^s - 1)
	// where s is steepness
	return float32((math.Exp(float64(t.steepness*normalized)) - 1.0) / (math.Exp(float64(t.steepness)) - 1.0))
}

func (t logTaper) Invert(tapered float32) float32 {
	if tapered <= 0 {
		return 0
	}
	if tapered >= 1 {
		return 1
	}
	// Inverse: x = ln(y * (e^s - 1) + 1) / s
	expS := math.Exp(float64(t.steepness))
	return float32(math.Log(float64(tapered)*(expS-1.0)+1.0) / float64(t.steepness))
}

// LogTaper returns a logarithmic taper with configurable steepness.
// steepness controls the curve intensity:
//   - 1.0 = gentle curve
//   - 3.0 = moderate curve (good default)
//   - 10.0 = steep curve
func LogTaper(steepness float32) Taper {
	if steepness <= 0 {
		steepness = 3.0 // sensible default
	}
	return logTaper{steepness: steepness}
}

// audioTaper implements a standard audio fader taper.
// Approximates analog audio fader behavior with gentle bottom and steep top.
type audioTaper struct{}

func (t audioTaper) Apply(normalized float32) float32 {
	if normalized <= 0 {
		return 0
	}
	if normalized >= 1 {
		return 1
	}
	// Audio taper: roughly -60dB to 0dB mapping
	// Using power curve: y = x^3 gives gentle bottom, steep top
	return normalized * normalized * normalized
}

func (t audioTaper) Invert(tapered float32) float32 {
	if tapered <= 0 {
		return 0
	}
	if tapered >= 1 {
		return 1
	}
	// Inverse: x = y^(1/3)
	return float32(math.Pow(float64(tapered), 1.0/3.0))
}

// AudioTaper returns a standard audio fader taper.
// Optimized for dB scales, approximates analog audio faders.
func AudioTaper() Taper {
	return audioTaper{}
}

// customTaper allows users to provide their own taper functions.
type customTaper struct {
	apply  func(float32) float32
	invert func(float32) float32
}

func (t customTaper) Apply(normalized float32) float32 {
	return t.apply(normalized)
}

func (t customTaper) Invert(tapered float32) float32 {
	return t.invert(tapered)
}

// CustomTaper creates a taper from user-provided functions.
// apply: normalized (0-1) -> tapered (0-1)
// invert: tapered (0-1) -> normalized (0-1)
func CustomTaper(apply, invert func(float32) float32) Taper {
	return customTaper{apply: apply, invert: invert}
}

// ============================================================================
// Fader Parameters
// ============================================================================

// FaderParams configures extended fader behavior.
type FaderParams struct {
	// Taper curve (affects UI feel, not values)
	// nil = linear taper
	Taper Taper

	// Range stops (in normalized 0-1 space, applied after taper)
	MinStop float32 // minimum value (default 0.0)
	MaxStop float32 // maximum value (default 1.0)

	// Reset value (in normalized 0-1 space)
	ResetValue float32 // default 0.0

	// Dimensions
	Width  float32 // default 30.0
	Height float32 // default 300.0

	// Display options
	Format      func(normalized float32) string // optional: custom tooltip format
	ShowTooltip bool                            // show value on hover (default true)

	// Mouse wheel sensitivity
	WheelSteps float32 // default 100.0 (finer = more steps)
}

// DefaultFaderParams returns sensible default parameters.
func DefaultFaderParams() FaderParams {
	return FaderParams{
		Taper:       LinearTaper(),
		MinStop:     0.0,
		MaxStop:     1.0,
		ResetValue:  0.0,
		Width:       30.0,
		Height:      300.0,
		ShowTooltip: true,
		WheelSteps:  100.0,
	}
}

// ============================================================================
// Fader API Functions
// ============================================================================

// FaderN draws a vertical fader working in normalized 0.0-1.0 space.
// This is the foundation for FaderF and FaderI.
func FaderN(label string, value float32, params FaderParams) (float32, bool) {
	// Apply defaults
	if params.Taper == nil {
		params.Taper = LinearTaper()
	}
	if params.Width == 0 {
		params.Width = 30.0
	}
	if params.Height == 0 {
		params.Height = 300.0
	}
	if params.WheelSteps == 0 {
		params.WheelSteps = 100.0
	}
	if params.MaxStop == 0 {
		params.MaxStop = 1.0
	}

	// Clamp value to range stops
	value = clamp(value, params.MinStop, params.MaxStop)

	// Apply taper to get UI position
	uiPosition := params.Taper.Apply(value)

	// Draw vertical slider
	newUIPosition := uiPosition
	size := imgui.Vec2{X: params.Width, Y: params.Height}
	changed := imgui.VSliderFloatV(label, size, &newUIPosition, 0.0, 1.0, "", imgui.SliderFlagsNone)

	// Invert taper to get normalized value
	newValue := params.Taper.Invert(newUIPosition)

	// Handle right-click reset
	if imgui.IsItemHovered() && imgui.IsMouseClickedBool(imgui.MouseButtonRight) {
		newValue = params.ResetValue
		if newValue != value {
			changed = true
		}
	}

	// Handle mouse wheel
	if imgui.IsItemHovered() {
		wheel := imgui.CurrentIO().MouseWheel()
		if wheel != 0 {
			// Clear drag state if needed
			if imgui.IsItemActive() {
				imgui.InternalClearActiveID()
			}

			// Calculate adjustment
			fraction := float32(1.0 / params.WheelSteps)

			// Apply modifiers (Ctrl = 10x faster, Alt = 10x slower)
			if imgui.CurrentIO().KeyCtrl() {
				fraction *= wheelMultiplierFast // 10.0
			} else if imgui.CurrentIO().KeyAlt() {
				fraction /= wheelMultiplierSlow // 10.0
			}

			// Adjust and clamp
			newValue += wheel * fraction
			newValue = clamp(newValue, params.MinStop, params.MaxStop)

			if newValue != value {
				changed = true
			}
		}
	}

	// Show tooltip
	if params.ShowTooltip && imgui.IsItemHovered() {
		var tooltipText string
		if params.Format != nil {
			tooltipText = params.Format(newValue)
		} else {
			tooltipText = fmt.Sprintf("%.3f", newValue)
		}
		imgui.SetTooltip(tooltipText)
	}

	// Clamp final value to range stops
	newValue = clamp(newValue, params.MinStop, params.MaxStop)

	return newValue, changed
}

// FaderF draws a vertical fader working in an arbitrary float range.
// Internally converts to/from normalized 0-1 space.
// Example: -60.0 to +12.0 dB, 20.0 to 20000.0 Hz
func FaderF(label string, value, min, max float32, params FaderParams) (float32, bool) {
	// Normalize value to 0-1
	normalized := (value - min) / (max - min)
	normalized = clamp(normalized, 0.0, 1.0)

	// Call FaderN
	newNormalized, changed := FaderN(label, normalized, params)

	// Denormalize to original range
	newValue := newNormalized*(max-min) + min

	return newValue, changed
}

// FaderI draws a vertical fader working in an integer range.
// Internally converts to/from normalized 0-1 space.
// Example: 0 to 32767 for hardware, 0 to 127 for MIDI
func FaderI(label string, value int, min, max int, params FaderParams) (int, bool) {
	// Normalize value to 0-1
	rangeF := float32(max - min)
	normalized := float32(value-min) / rangeF
	normalized = clamp(normalized, 0.0, 1.0)

	// Call FaderN
	newNormalized, changed := FaderN(label, normalized, params)

	// Denormalize to original range and round
	newValue := int(newNormalized*rangeF + 0.5) + min

	// Clamp to range
	if newValue < min {
		newValue = min
	}
	if newValue > max {
		newValue = max
	}

	return newValue, changed
}
