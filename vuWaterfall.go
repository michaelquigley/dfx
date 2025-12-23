package dfx

import (
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
)

// VUWaterfall is a scrolling history display of VU levels over time.
// new data appears at the bottom and scrolls upward.
// each row shows a horizontal bar whose width represents the level at that time slice.
type VUWaterfall struct {
	Container

	// dimensions
	Height       float32 // total height in pixels (default: 200)
	ChannelWidth float32 // width per channel (default: 40)
	ChannelGap   float32 // gap between channels (default: 4)
	RowHeight    float32 // height of each history row (default: 2)
	RowGap       float32 // gap between rows (default: 0)

	// history configuration
	HistorySize    int           // number of samples to keep (default: 100)
	SampleInterval time.Duration // minimum time between samples (default: 16ms)

	// display mode
	Highres bool // when true, alternates row opacity for scanline effect

	// colors (same as VUMeter for consistency)
	ColorLow  imgui.Vec4 // green zone (0-60%)
	ColorMid  imgui.Vec4 // yellow zone (60-80%)
	ColorHigh imgui.Vec4 // red zone (80-100%)
	ColorOff  imgui.Vec4 // background/inactive

	// internal state
	history      [][]float32 // circular buffer: history[row][channel]
	historyHead  int         // index where next entry will be written
	historyLen   int         // current number of valid entries
	channelCount int         // number of channels
	lastSample   time.Time   // when last sample was added
}

// NewVUWaterfall creates a new waterfall display with the specified number of channels.
func NewVUWaterfall(channelCount int) *VUWaterfall {
	w := &VUWaterfall{
		// dimensions
		Height:       200,
		ChannelWidth: 40,
		ChannelGap:   4,
		RowHeight:    2,
		RowGap:       0,

		// history
		HistorySize:    100,
		SampleInterval: 16 * time.Millisecond, // ~60 samples per second

		// colors (match VUMeter defaults)
		ColorLow:  imgui.Vec4{X: 0.2, Y: 0.8, Z: 0.2, W: 1.0},    // green
		ColorMid:  imgui.Vec4{X: 0.9, Y: 0.8, Z: 0.1, W: 1.0},    // yellow
		ColorHigh: imgui.Vec4{X: 0.9, Y: 0.2, Z: 0.2, W: 1.0},    // red
		ColorOff:  imgui.Vec4{X: 0.15, Y: 0.15, Z: 0.15, W: 1.0}, // dark gray

		channelCount: channelCount,
	}

	w.Visible = true
	w.initHistory()

	return w
}

// initHistory initializes or resets the history buffer.
func (w *VUWaterfall) initHistory() {
	w.history = make([][]float32, w.HistorySize)
	for i := range w.history {
		w.history[i] = make([]float32, w.channelCount)
	}
	w.historyHead = 0
	w.historyLen = 0
}

// ChannelCount returns the number of channels.
func (w *VUWaterfall) ChannelCount() int {
	return w.channelCount
}

// SetChannelCount resizes the waterfall to the specified number of channels.
// this clears the history buffer.
func (w *VUWaterfall) SetChannelCount(count int) {
	if count == w.channelCount {
		return
	}
	w.channelCount = count
	w.initHistory()
}

// SetHistorySize sets the number of samples to keep and reinitializes the buffer.
// this clears the history buffer.
func (w *VUWaterfall) SetHistorySize(size int) {
	if size == w.HistorySize {
		return
	}
	w.HistorySize = size
	w.initHistory()
}

// SetLevel sets the level for a single channel and adds a new history entry.
// note: this creates a new row with only this channel set; prefer SetLevels for multi-channel.
// If SampleInterval is set, samples are throttled to maintain consistent scroll speed.
func (w *VUWaterfall) SetLevel(channel int, level float32) {
	if channel < 0 || channel >= w.channelCount {
		return
	}

	// throttle samples based on time interval
	now := time.Now()
	if w.SampleInterval > 0 && time.Since(w.lastSample) < w.SampleInterval {
		return // skip this sample
	}
	w.lastSample = now

	// write to current head position
	for i := range w.history[w.historyHead] {
		w.history[w.historyHead][i] = 0
	}
	w.history[w.historyHead][channel] = clamp(level, 0, 1)

	// advance head
	w.historyHead = (w.historyHead + 1) % w.HistorySize
	if w.historyLen < w.HistorySize {
		w.historyLen++
	}
}

// SetLevels sets levels for all channels at once and adds a new history entry.
// If SampleInterval is set, samples are throttled to maintain consistent scroll speed.
func (w *VUWaterfall) SetLevels(levels []float32) {
	// throttle samples based on time interval
	now := time.Now()
	if w.SampleInterval > 0 && time.Since(w.lastSample) < w.SampleInterval {
		return // skip this sample
	}
	w.lastSample = now

	// copy levels into current head position
	for i := 0; i < w.channelCount; i++ {
		if i < len(levels) {
			w.history[w.historyHead][i] = clamp(levels[i], 0, 1)
		} else {
			w.history[w.historyHead][i] = 0
		}
	}

	// advance head
	w.historyHead = (w.historyHead + 1) % w.HistorySize
	if w.historyLen < w.HistorySize {
		w.historyLen++
	}
}

// Width returns the calculated total width of the waterfall.
func (w *VUWaterfall) Width() float32 {
	if w.channelCount == 0 {
		return 0
	}
	return float32(w.channelCount)*w.ChannelWidth + float32(w.channelCount-1)*w.ChannelGap
}

// Draw renders the VU waterfall.
func (w *VUWaterfall) Draw(state *State) {
	if !w.Visible {
		return
	}

	cursor := imgui.CursorScreenPos()
	dl := imgui.WindowDrawList()

	totalWidth := w.Width()

	// draw background
	dl.AddRectFilled(
		cursor,
		imgui.Vec2{X: cursor.X + totalWidth, Y: cursor.Y + w.Height},
		imgui.ColorConvertFloat4ToU32(w.ColorOff),
	)

	if w.historyLen == 0 {
		// reserve space and return
		imgui.Dummy(imgui.Vec2{X: totalWidth, Y: w.Height})
		if w.OnDraw != nil {
			w.OnDraw(state)
		}
		return
	}

	// calculate row positions
	// we want newest at bottom, oldest at top
	// so we draw from top (oldest) to bottom (newest)
	rowStep := w.RowHeight + w.RowGap

	// calculate how many rows fit in the available height
	maxVisibleRows := int(w.Height / rowStep)
	visibleRows := w.historyLen
	if visibleRows > maxVisibleRows {
		visibleRows = maxVisibleRows
	}

	// calculate starting index (skip older entries that don't fit)
	// we want the newest entries, so skip (historyLen - visibleRows) oldest entries
	skipCount := w.historyLen - visibleRows
	startIdx := (w.historyHead - w.historyLen + skipCount + w.HistorySize) % w.HistorySize

	// calculate vertical offset to align rows at bottom of display
	totalRowsHeight := float32(visibleRows) * rowStep
	yOffset := w.Height - totalRowsHeight

	for row := 0; row < visibleRows; row++ {
		histIdx := (startIdx + row) % w.HistorySize
		rowY := cursor.Y + yOffset + float32(row)*rowStep

		for ch := 0; ch < w.channelCount; ch++ {
			level := w.history[histIdx][ch]
			if level <= 0 {
				continue
			}

			// calculate bar position and size
			chX := cursor.X + float32(ch)*(w.ChannelWidth+w.ChannelGap)
			barWidth := level * w.ChannelWidth

			// center the bar horizontally within the channel
			barLeft := chX + (w.ChannelWidth-barWidth)/2
			barRight := barLeft + barWidth

			// determine color based on level
			color := w.levelColor(level)

			// in highres mode, reduce opacity on every other row for scanline effect
			if w.Highres && row%2 == 1 {
				color.W *= 0.3 // reduce alpha to 30%
			}

			dl.AddRectFilled(
				imgui.Vec2{X: barLeft, Y: rowY},
				imgui.Vec2{X: barRight, Y: rowY + w.RowHeight},
				imgui.ColorConvertFloat4ToU32(color),
			)
		}
	}

	// reserve space for layout
	imgui.Dummy(imgui.Vec2{X: totalWidth, Y: w.Height})

	// call base container draw
	if w.OnDraw != nil {
		w.OnDraw(state)
	}
}

// levelColor returns the color for a given level based on zone thresholds.
func (w *VUWaterfall) levelColor(level float32) imgui.Vec4 {
	if level < 0.6 {
		return w.ColorLow // green zone
	} else if level < 0.8 {
		return w.ColorMid // yellow zone
	}
	return w.ColorHigh // red zone
}

// Clear resets the history buffer.
func (w *VUWaterfall) Clear() {
	w.historyHead = 0
	w.historyLen = 0
	for i := range w.history {
		for j := range w.history[i] {
			w.history[i][j] = 0
		}
	}
}
