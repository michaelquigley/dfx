package dfx

import (
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
)

// VUMeterMode defines the visual rendering style of the meter.
type VUMeterMode int

const (
	// VUMeterSolid renders a continuous fill with color zones (default).
	VUMeterSolid VUMeterMode = iota
	// VUMeterHighres renders 1px segments with 1px gaps for high resolution display.
	VUMeterHighres
	// VUMeterSegmented renders discrete segments with configurable count and gap.
	VUMeterSegmented
)

// VUMeter is a vertical level meter component.
// supports any number of channels displayed side by side.
type VUMeter struct {
	Container

	// display mode
	Mode VUMeterMode // rendering style (default: VUMeterSolid)

	// fixed size configuration
	Height       float32 // total height in pixels (default: 200)
	ChannelWidth float32 // width of each channel meter (default: 12)

	// segment configuration
	SegmentCount int     // number of vertical segments (default: 20)
	SegmentGap   float32 // gap between segments in pixels (default: 2)
	ChannelGap   float32 // gap between channel meters (default: 4)

	// peak hold configuration
	PeakHoldMs    int     // peak hold duration in ms, 0 = disabled (default: 1000)
	PeakDecayRate float32 // peak decay rate per second (default: 0.5)

	// clip indicator configuration
	ClipHoldMs int // how long clip indicator stays lit in ms (default: 2000)

	// labels (optional, per-channel)
	Labels      []string // custom labels like "L", "R", "Kick", etc.
	LabelHeight float32  // height reserved for labels (default: 16)

	// colors (configurable, with sensible defaults)
	ColorLow  imgui.Vec4 // green zone (0-60%)
	ColorMid  imgui.Vec4 // yellow zone (60-80%)
	ColorHigh imgui.Vec4 // red zone (80-100%)
	ColorOff  imgui.Vec4 // inactive segment color
	ColorPeak imgui.Vec4 // peak indicator color
	ColorClip imgui.Vec4 // clip indicator color (bright red)

	// internal state
	levels    []float32   // current level per channel (0.0-1.0)
	peaks     []float32   // peak level per channel
	peakTimes []time.Time // when each peak was set
	clipped   []bool      // whether channel has clipped
	clipTimes []time.Time // when each clip occurred
	lastFrame time.Time   // for delta time calculation
}

// NewVUMeter creates a new VU meter with the specified number of channels.
func NewVUMeter(channelCount int) *VUMeter {
	v := &VUMeter{
		// size defaults
		Height:       200,
		ChannelWidth: 12,

		// segment defaults
		SegmentCount: 20,
		SegmentGap:   2,
		ChannelGap:   4,

		// peak defaults
		PeakHoldMs:    1000,
		PeakDecayRate: 0.5,

		// clip defaults
		ClipHoldMs: 2000,

		// label defaults
		LabelHeight: 14,

		// default colors
		ColorLow:  imgui.Vec4{X: 0.2, Y: 0.8, Z: 0.2, W: 1.0},    // green
		ColorMid:  imgui.Vec4{X: 0.9, Y: 0.8, Z: 0.1, W: 1.0},    // yellow
		ColorHigh: imgui.Vec4{X: 0.9, Y: 0.2, Z: 0.2, W: 1.0},    // red
		ColorOff:  imgui.Vec4{X: 0.15, Y: 0.15, Z: 0.15, W: 1.0}, // dark gray
		ColorPeak: imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 0.9},    // white
		ColorClip: imgui.Vec4{X: 1.0, Y: 0.0, Z: 0.0, W: 1.0},    // bright red

		lastFrame: time.Now(),
	}

	v.Visible = true
	v.initChannels(channelCount)

	return v
}

// initChannels initializes or resizes the channel state slices.
func (v *VUMeter) initChannels(count int) {
	now := time.Now()

	v.levels = make([]float32, count)
	v.peaks = make([]float32, count)
	v.peakTimes = make([]time.Time, count)
	v.clipped = make([]bool, count)
	v.clipTimes = make([]time.Time, count)

	for i := 0; i < count; i++ {
		v.peakTimes[i] = now
		v.clipTimes[i] = now
	}
}

// ChannelCount returns the number of channels.
func (v *VUMeter) ChannelCount() int {
	return len(v.levels)
}

// SetChannelCount resizes the meter to the specified number of channels.
func (v *VUMeter) SetChannelCount(count int) {
	if count == len(v.levels) {
		return
	}
	v.initChannels(count)
}

// SetLevel sets the level for a single channel (0.0 to 1.0).
func (v *VUMeter) SetLevel(channel int, level float32) {
	if channel < 0 || channel >= len(v.levels) {
		return
	}
	v.levels[channel] = clamp(level, 0, 1)
}

// SetLevels sets the levels for all channels at once.
func (v *VUMeter) SetLevels(levels []float32) {
	for i := 0; i < len(levels) && i < len(v.levels); i++ {
		v.levels[i] = clamp(levels[i], 0, 1)
	}
}

// SetLabel sets the label for a single channel.
func (v *VUMeter) SetLabel(channel int, label string) {
	// grow labels slice if needed
	for len(v.Labels) <= channel {
		v.Labels = append(v.Labels, "")
	}
	v.Labels[channel] = label
}

// SetLabels sets labels for all channels at once.
func (v *VUMeter) SetLabels(labels []string) {
	v.Labels = labels
}

// Width returns the calculated total width of the meter.
func (v *VUMeter) Width() float32 {
	count := float32(len(v.levels))
	if count == 0 {
		return 0
	}
	return (count * v.ChannelWidth) + ((count - 1) * v.ChannelGap)
}

// Draw renders the VU meter.
func (v *VUMeter) Draw(state *State) {
	if !v.Visible || len(v.levels) == 0 {
		return
	}

	// calculate delta time for peak decay
	now := time.Now()
	deltaTime := float32(now.Sub(v.lastFrame).Seconds())
	v.lastFrame = now

	// update peaks and clip indicators
	v.updatePeaks(deltaTime)
	v.updateClip()

	// get draw position and draw list
	cursor := imgui.CursorScreenPos()
	dl := imgui.WindowDrawList()

	// calculate clip indicator height (fixed size for all modes)
	clipHeight := float32(8)
	clipGap := float32(2)
	meterHeight := v.Height - v.LabelHeight - clipHeight - clipGap

	// draw each channel
	for ch := 0; ch < len(v.levels); ch++ {
		xOffset := float32(ch) * (v.ChannelWidth + v.ChannelGap)

		// draw clip indicator at top
		clipTop := cursor.Y
		clipBottom := clipTop + clipHeight
		clipLeft := cursor.X + xOffset
		clipRight := clipLeft + v.ChannelWidth

		var clipColor imgui.Vec4
		if v.clipped[ch] {
			clipColor = v.ColorClip
		} else {
			clipColor = v.ColorOff
		}
		dl.AddRectFilled(
			imgui.Vec2{X: clipLeft, Y: clipTop},
			imgui.Vec2{X: clipRight, Y: clipBottom},
			imgui.ColorConvertFloat4ToU32(clipColor),
		)

		// draw meter based on mode
		meterTop := clipBottom + clipGap
		level := v.levels[ch]
		peakLevel := v.peaks[ch]

		switch v.Mode {
		case VUMeterHighres:
			v.drawHighresChannel(dl, cursor, ch, xOffset, level, peakLevel, meterTop, meterHeight)
		case VUMeterSegmented:
			v.drawSegmentedChannel(dl, cursor, ch, xOffset, level, peakLevel, meterTop, meterHeight)
		default: // VUMeterSolid
			v.drawSolidChannel(dl, cursor, ch, xOffset, level, peakLevel, meterTop, meterHeight)
		}
	}

	// draw labels at bottom using consistent font metrics
	v.drawLabels(cursor, dl)

	// reserve space for the meter so imgui layout works correctly
	imgui.Dummy(imgui.Vec2{X: v.Width(), Y: v.Height})

	drawContainerExtensions(&v.Container, state)
}

// drawLabels renders channel labels with consistent font metrics.
// uses draw list AddText directly to avoid cursor positioning issues in nested contexts.
func (v *VUMeter) drawLabels(cursor imgui.Vec2, dl *imgui.DrawList) {
	// check if any labels exist
	hasLabels := false
	for ch := 0; ch < len(v.levels); ch++ {
		if ch < len(v.Labels) && v.Labels[ch] != "" {
			hasLabels = true
			break
		}
	}
	if !hasLabels {
		return
	}

	// push font once for consistent metrics
	PushFont(SmallFont)
	fontSize := imgui.TextLineHeight()
	labelColor := imgui.ColorConvertFloat4ToU32(imgui.CurrentStyle().Colors()[imgui.ColText])
	for ch := 0; ch < len(v.levels); ch++ {
		if ch < len(v.Labels) && v.Labels[ch] != "" {
			label := v.Labels[ch]
			labelWidth := imgui.CalcTextSize(label).X
			xOffset := float32(ch) * (v.ChannelWidth + v.ChannelGap)
			labelX := cursor.X + xOffset + (v.ChannelWidth-labelWidth)/2
			labelY := cursor.Y + v.Height - v.LabelHeight + (v.LabelHeight-fontSize)/2
			dl.AddTextFontPtr(imgui.CurrentFont(), imgui.FontSize(), imgui.Vec2{X: labelX, Y: labelY}, labelColor, label)
		}
	}
	PopFont()
}

// segmentColor returns the color for a segment based on its position.
func (v *VUMeter) segmentColor(segment int) imgui.Vec4 {
	// calculate normalized position (0.0 to 1.0)
	pos := float32(segment) / float32(v.SegmentCount)

	if pos < 0.6 {
		return v.ColorLow // green zone
	} else if pos < 0.8 {
		return v.ColorMid // yellow zone
	}
	return v.ColorHigh // red zone
}

// updatePeaks updates peak hold and decay for all channels.
func (v *VUMeter) updatePeaks(deltaTime float32) {
	if v.PeakHoldMs <= 0 {
		return
	}

	now := time.Now()
	for i, level := range v.levels {
		if level > v.peaks[i] {
			v.peaks[i] = level
			v.peakTimes[i] = now
		} else if now.Sub(v.peakTimes[i]).Milliseconds() > int64(v.PeakHoldMs) {
			// decay peak after hold time
			v.peaks[i] -= v.PeakDecayRate * deltaTime
			if v.peaks[i] < level {
				v.peaks[i] = level
			}
			if v.peaks[i] < 0 {
				v.peaks[i] = 0
			}
		}
	}
}

// updateClip updates clip indicators for all channels.
func (v *VUMeter) updateClip() {
	now := time.Now()
	for i, level := range v.levels {
		if level >= 1.0 {
			v.clipped[i] = true
			v.clipTimes[i] = now
		} else if v.clipped[i] && now.Sub(v.clipTimes[i]).Milliseconds() > int64(v.ClipHoldMs) {
			// auto-reset clip indicator after hold time
			v.clipped[i] = false
		}
	}
}

// drawSolidChannel renders a channel as a continuous fill with color zones.
func (v *VUMeter) drawSolidChannel(dl *imgui.DrawList, cursor imgui.Vec2, ch int, xOffset float32, level, peakLevel float32, meterTop, meterHeight float32) {
	segLeft := cursor.X + xOffset
	segRight := segLeft + v.ChannelWidth
	meterBottom := meterTop + meterHeight

	// draw off background first
	dl.AddRectFilled(
		imgui.Vec2{X: segLeft, Y: meterTop},
		imgui.Vec2{X: segRight, Y: meterBottom},
		imgui.ColorConvertFloat4ToU32(v.ColorOff),
	)

	if level > 0 {
		// zone boundaries (as normalized positions)
		greenMax := float32(0.6)
		yellowMax := float32(0.8)

		// fill from bottom up based on level
		fillHeight := level * meterHeight
		fillTop := meterBottom - fillHeight

		// draw color zones
		if level <= greenMax {
			// only green zone lit
			dl.AddRectFilled(
				imgui.Vec2{X: segLeft, Y: fillTop},
				imgui.Vec2{X: segRight, Y: meterBottom},
				imgui.ColorConvertFloat4ToU32(v.ColorLow),
			)
		} else if level <= yellowMax {
			// green fully lit, yellow partially lit
			greenTop := meterBottom - (greenMax * meterHeight)
			dl.AddRectFilled(
				imgui.Vec2{X: segLeft, Y: greenTop},
				imgui.Vec2{X: segRight, Y: meterBottom},
				imgui.ColorConvertFloat4ToU32(v.ColorLow),
			)
			dl.AddRectFilled(
				imgui.Vec2{X: segLeft, Y: fillTop},
				imgui.Vec2{X: segRight, Y: greenTop},
				imgui.ColorConvertFloat4ToU32(v.ColorMid),
			)
		} else {
			// all zones lit
			greenTop := meterBottom - (greenMax * meterHeight)
			yellowTop := meterBottom - (yellowMax * meterHeight)
			dl.AddRectFilled(
				imgui.Vec2{X: segLeft, Y: greenTop},
				imgui.Vec2{X: segRight, Y: meterBottom},
				imgui.ColorConvertFloat4ToU32(v.ColorLow),
			)
			dl.AddRectFilled(
				imgui.Vec2{X: segLeft, Y: yellowTop},
				imgui.Vec2{X: segRight, Y: greenTop},
				imgui.ColorConvertFloat4ToU32(v.ColorMid),
			)
			dl.AddRectFilled(
				imgui.Vec2{X: segLeft, Y: fillTop},
				imgui.Vec2{X: segRight, Y: yellowTop},
				imgui.ColorConvertFloat4ToU32(v.ColorHigh),
			)
		}
	}

	// draw peak indicator as thin line
	if v.PeakHoldMs > 0 && peakLevel > 0 {
		peakY := meterBottom - (peakLevel * meterHeight)
		peakHeight := float32(2)
		dl.AddRectFilled(
			imgui.Vec2{X: segLeft, Y: peakY - peakHeight/2},
			imgui.Vec2{X: segRight, Y: peakY + peakHeight/2},
			imgui.ColorConvertFloat4ToU32(v.ColorPeak),
		)
	}
}

// drawHighresChannel renders a channel with 1px segments and 1px gaps.
func (v *VUMeter) drawHighresChannel(dl *imgui.DrawList, cursor imgui.Vec2, ch int, xOffset float32, level, peakLevel float32, meterTop, meterHeight float32) {
	// fixed 1px segment, 1px gap
	segmentHeight := float32(1)
	segmentGap := float32(1)
	segmentCount := int((meterHeight + segmentGap) / (segmentHeight + segmentGap))

	litSegments := int(level * float32(segmentCount))
	peakSegment := int(peakLevel * float32(segmentCount))

	for seg := 0; seg < segmentCount; seg++ {
		segTop := meterTop + meterHeight - float32(seg+1)*(segmentHeight+segmentGap) + segmentGap
		segBottom := segTop + segmentHeight
		segLeft := cursor.X + xOffset
		segRight := segLeft + v.ChannelWidth

		var segColor imgui.Vec4
		if seg < litSegments {
			// color based on position
			pos := float32(seg) / float32(segmentCount)
			if pos < 0.6 {
				segColor = v.ColorLow
			} else if pos < 0.8 {
				segColor = v.ColorMid
			} else {
				segColor = v.ColorHigh
			}
		} else if seg == peakSegment && v.PeakHoldMs > 0 {
			segColor = v.ColorPeak
		} else {
			segColor = v.ColorOff
		}

		dl.AddRectFilled(
			imgui.Vec2{X: segLeft, Y: segTop},
			imgui.Vec2{X: segRight, Y: segBottom},
			imgui.ColorConvertFloat4ToU32(segColor),
		)
	}
}

// drawSegmentedChannel renders a channel using discrete segments with configurable count and gap.
func (v *VUMeter) drawSegmentedChannel(dl *imgui.DrawList, cursor imgui.Vec2, ch int, xOffset float32, level, peakLevel float32, meterTop, meterHeight float32) {
	segmentHeight := (meterHeight - (float32(v.SegmentCount-1) * v.SegmentGap)) / float32(v.SegmentCount)

	litSegments := int(level * float32(v.SegmentCount))
	peakSegment := int(peakLevel * float32(v.SegmentCount))

	for seg := 0; seg < v.SegmentCount; seg++ {
		segTop := meterTop + meterHeight - float32(seg+1)*(segmentHeight+v.SegmentGap) + v.SegmentGap
		segBottom := segTop + segmentHeight
		segLeft := cursor.X + xOffset
		segRight := segLeft + v.ChannelWidth

		var segColor imgui.Vec4
		if seg < litSegments {
			segColor = v.segmentColor(seg)
		} else if seg == peakSegment && v.PeakHoldMs > 0 {
			segColor = v.ColorPeak
		} else {
			segColor = v.ColorOff
		}

		dl.AddRectFilled(
			imgui.Vec2{X: segLeft, Y: segTop},
			imgui.Vec2{X: segRight, Y: segBottom},
			imgui.ColorConvertFloat4ToU32(segColor),
		)
	}
}
